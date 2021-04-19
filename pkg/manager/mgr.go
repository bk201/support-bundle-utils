package manager

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/bk201/support-bundle-utils/pkg/manager/client"
	"github.com/bk201/support-bundle-utils/pkg/manager/external"
	"github.com/bk201/support-bundle-utils/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

var debugf = logrus.Debugf
var debug = logrus.Debug

type SupportBundleManager struct {
	Namespace   string
	BundleName  string
	OutputDir   string
	NodeCount   int
	WaitTimeout time.Duration
	LonghornAPI string

	context context.Context

	restConfig *rest.Config
	k8s        *client.KubernetesClient
	harvester  *client.HarvesterClient

	ch    chan struct{}
	done  bool
	lock  sync.Mutex
	nodes map[string]string
}

func (m *SupportBundleManager) check() error {
	if m.Namespace == "" {
		return errors.New("namespace is not specified")
	}
	if m.BundleName == "" {
		return errors.New("support bundle name is not specified")
	}
	if m.NodeCount == 0 {
		return errors.New("node count is not specified")
	}
	if m.OutputDir == "" {
		m.OutputDir = filepath.Join(os.TempDir(), "harvester-support-bundle")
	}
	if err := os.MkdirAll(m.getWorkingDir(), os.FileMode(0755)); err != nil {
		return err
	}
	return nil
}

func (m *SupportBundleManager) getWorkingDir() string {
	return filepath.Join(m.OutputDir, "bundle")
}

func (m *SupportBundleManager) getBundlefile() string {
	return filepath.Join(m.OutputDir, "bundle.zip")
}

func (m *SupportBundleManager) Run() error {
	if err := m.check(); err != nil {
		return err
	}

	m.context = context.Background()

	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	m.restConfig = config
	hvst, err := client.NewHarvesterStore(m.context, m.Namespace, m.restConfig)
	if err != nil {
		return err
	}
	m.harvester = hvst

	k8s, err := client.NewKubernetesStore(m.context, m.Namespace, m.restConfig)
	if err != nil {
		return err
	}
	m.k8s = k8s

	// generate cluster level support bundle
	if err := m.GenerateClusterBundle(m.getWorkingDir()); err != nil {
		logrus.Errorf("fail to generate cluster bundle: %s", err)
		return err
	}
	m.harvester.UpdateSupportBundleState(m.BundleName, "managerdone")

	m.ch = make(chan struct{})
	// create a http server to receive node level support bundles
	s := HttpServer{context: m.context}
	go s.Run(m)

	logrus.Infof("wating node bundles for %s...", m.WaitTimeout)
	select {
	case <-m.ch:
		debug("wait done!")
		if err := m.compressBundle(); err != nil {
			logrus.Errorf("fail to compress bundle: %s", err)
		} else {
			m.harvester.UpdateSupportBundleState(m.BundleName, "agentdone")
		}
	case <-time.After(m.WaitTimeout):
		m.harvester.UpdateSupportBundleState(m.BundleName, "error")
		debug("wait timeout!")
	}

	select {}
}

func (m *SupportBundleManager) getBundle(w http.ResponseWriter, req *http.Request) {
	debugf("getBundle")

	f, err := os.Open(m.getBundlefile())
	if err != nil {
		e := errors.Wrap(err, "fail to open bundle file")
		logrus.Error(e)
		utils.HttpResponseError(w, http.StatusNotFound, e)
		return
	}
	defer f.Close()

	fstat, err := f.Stat()
	if err != nil {
		e := errors.Wrap(err, "fail to stat bundle file")
		logrus.Error(e)
		utils.HttpResponseError(w, http.StatusNotFound, e)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Length", strconv.FormatInt(fstat.Size(), 10))
	if _, err := io.Copy(w, f); err != nil {
		utils.HttpResponseError(w, http.StatusInternalServerError, err)
		return
	}
}

func (m *SupportBundleManager) createNodeBundle(w http.ResponseWriter, req *http.Request) {
	debug("enter createNodeBundle")
	node := mux.Vars(req)["nodeName"]
	if node == "" {
		return
	}

	nodesDir := filepath.Join(m.getWorkingDir(), "nodes")
	err := os.MkdirAll(nodesDir, os.FileMode(0775))
	if err != nil {
		logrus.Errorf("fail to create directory %s: %s", nodesDir, err)
		return
	}

	nodeBundle := filepath.Join(nodesDir, node+".zip")
	f, err := os.Create(nodeBundle)
	if err != nil {
		logrus.Errorf("fail to create file %s: %s", nodeBundle, err)
		return
	}
	_, err = io.Copy(f, req.Body)
	if err != nil {
		fmt.Println(err)
	}
	f.Close()

	err = m.verifyNodeBundle(nodeBundle)
	if err != nil {
		logrus.Errorf("fail to verify file %s: %s", nodeBundle, err)
		utils.HttpResponseError(w, http.StatusBadRequest, err)
		return
	}
	m.completeNode(node)
	utils.HttpResponseStatus(w, http.StatusCreated)
	debug("exit createNodeBundle")
}

func (m *SupportBundleManager) verifyNodeBundle(file string) error {
	_, err := zip.OpenReader(file)
	return err
}

func (m *SupportBundleManager) completeNode(node string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.nodes == nil {
		m.nodes = make(map[string]string)
	}
	if _, ok := m.nodes[node]; !ok {
		m.nodes[node] = "done"
		debugf("complete node %s", node)
	}

	if len(m.nodes) == m.NodeCount {
		if !m.done {
			m.ch <- struct{}{}
			m.done = true
			debugf("send All nodes completed")
		}
		debugf("All nodes completed")
	}
}

func (m *SupportBundleManager) compressBundle() error {
	cmd := exec.Command("zip", "-r", m.getBundlefile(), filepath.Base(m.getWorkingDir()))
	cmd.Dir = m.OutputDir
	return cmd.Run()
}

func (m *SupportBundleManager) GenerateClusterBundle(bundleDir string) error {
	debug("generating cluster bundle")
	namespace, err := m.k8s.GetNamespace()
	if err != nil {
		return errors.Wrap(err, "cannot get harvester namespace")
	}
	kubeVersion, err := m.k8s.GetKubernetesVersion()
	if err != nil {
		return errors.Wrap(err, "cannot get kubernetes version")
	}

	createdAt := utils.Now()
	bundleMeta := &BundleMeta{
		ProjectName:          "Harvester",
		ProjectVersion:       "todo", // TODO
		KubernetesVersion:    kubeVersion.GitVersion,
		ProjectNamespaceUUID: string(namespace.UID),
		BundleCreatedAt:      createdAt,
		IssueURL:             "todo", // TODO
		IssueDescription:     "todo", // TODO
	}
	debugf("bundleMeta: %+v", bundleMeta)

	errLog, err := os.Create(filepath.Join(bundleDir, "bundleGenerationError.log"))
	if err != nil {
		logrus.Errorf("Failed to create bundle generation log")
		return err
	}
	defer errLog.Close()

	metaFile := filepath.Join(bundleDir, "metadata.yaml")
	encodeToYAMLFile(bundleMeta, metaFile, errLog)

	yamlsDir := filepath.Join(bundleDir, "yamls")
	m.generateSupportBundleYAMLs(yamlsDir, errLog)
	// sb.ProgressPercentage = BundleProgressPercentageYaml

	logsDir := filepath.Join(bundleDir, "logs")
	m.generateSupportBundleLogs(logsDir, errLog)
	// sb.ProgressPercentage = BundleProgressPercentageLogs

	externalDir := filepath.Join(bundleDir, "external")
	m.getExternalSupportBundles(bundleMeta, externalDir, errLog)
	// sb.ProgressPercentage = BundleProgressExternalBundle

	return nil
}

func (m *SupportBundleManager) generateSupportBundleYAMLs(yamlsDir string, errLog io.Writer) {
	kubernetesDir := filepath.Join(yamlsDir, "kubernetes")
	m.generateSupportBundleYAMLsForKubernetes(kubernetesDir, errLog)
	harvesterDir := filepath.Join(yamlsDir, "harvester")
	m.generateSupportBundleYAMLsForHarvester(harvesterDir, errLog)
}

func (m *SupportBundleManager) generateSupportBundleYAMLsForKubernetes(dir string, errLog io.Writer) {
	getListAndEncodeToYAML("events", m.k8s.GetAllEventsList, dir, errLog)
	getListAndEncodeToYAML("pods", m.k8s.GetAllPodsList, dir, errLog)
	getListAndEncodeToYAML("services", m.k8s.GetAllServicesList, dir, errLog)
	getListAndEncodeToYAML("deployments", m.k8s.GetAllDeploymentsList, dir, errLog)
	getListAndEncodeToYAML("daemonsets", m.k8s.GetAllDaemonSetsList, dir, errLog)
	getListAndEncodeToYAML("statefulsets", m.k8s.GetAllStatefulSetsList, dir, errLog)
	getListAndEncodeToYAML("jobs", m.k8s.GetAllJobsList, dir, errLog)
	getListAndEncodeToYAML("cronjobs", m.k8s.GetAllCronJobsList, dir, errLog)
	getListAndEncodeToYAML("nodes", m.k8s.GetAllNodesList, dir, errLog)
	getListAndEncodeToYAML("configmaps", m.k8s.GetAllConfigMaps, dir, errLog)
	getListAndEncodeToYAML("volumeattachments", m.k8s.GetAllVolumeAttachments, dir, errLog)
}

func (m *SupportBundleManager) generateSupportBundleYAMLsForHarvester(dir string, errLog io.Writer) {

	// Harvester
	for _, ns := range []string{m.Namespace, "default"} {
		harvester, err := client.NewHarvesterStore(m.context, ns, m.restConfig)
		if err != nil {
			fmt.Fprint(errLog, err)
			continue
		}
		toDir := filepath.Join(dir, "harvester", ns)
		getListAndEncodeToYAML("keypairs", harvester.GetAllKeypairs, toDir, errLog)
		getListAndEncodeToYAML("preferences", harvester.GetAllPreferences, toDir, errLog)
		getListAndEncodeToYAML("settings", harvester.GetAllSettings, toDir, errLog)
		getListAndEncodeToYAML("upgrades", harvester.GetAllUpgrades, toDir, errLog)
		getListAndEncodeToYAML("users", harvester.GetAllUsers, toDir, errLog)
		getListAndEncodeToYAML("virtualmachinebackups", harvester.GetAllVirtualMachineBackups, toDir, errLog)
		getListAndEncodeToYAML("virtualmachinebackupcontents", harvester.GetAllVirtualMachineBackupContents, toDir, errLog)
		getListAndEncodeToYAML("virtualmachineimages", harvester.GetAllVirtualMachineImages, toDir, errLog)
		getListAndEncodeToYAML("virtualmachinerestores", harvester.GetAllVirtualMachineRestores, toDir, errLog)
		getListAndEncodeToYAML("virtualmachinetemplates", harvester.GetAllVirtualMachineTemplates, toDir, errLog)
		getListAndEncodeToYAML("virtualmachinetemplateversions", harvester.GetAllVirtualMachineTemplateVersions, toDir, errLog)
	}

	// KubeVirt & CDI
	ns := "default"
	harvester, err := client.NewHarvesterStore(m.context, ns, m.restConfig)
	if err != nil {
		fmt.Fprint(errLog, err)
		return
	}
	toDir := filepath.Join(dir, "kubevirt", ns)
	getListAndEncodeToYAML("virtualmachines", harvester.GetAllVirtualMachines, toDir, errLog)
	getListAndEncodeToYAML("virtualmachineinstances", harvester.GetAllVirtualMachineInstances, toDir, errLog)
	getListAndEncodeToYAML("virtualmachineinstancemigrations", harvester.GetAllVirtualMachineInstanceMigrations, toDir, errLog)

	toDir = filepath.Join(dir, "cdi", ns)
	getListAndEncodeToYAML("datavolumes", harvester.GetAllDataVolumes, toDir, errLog)
}

func encodeToYAMLFile(obj interface{}, path string, errLog io.Writer) {
	var err error
	defer func() {
		if err != nil {
			fmt.Fprintf(errLog, "Support Bundle: failed to generate %v: %v\n", path, err)
		}
	}()
	err = os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	if err != nil {
		return
	}
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	encoder := yaml.NewEncoder(f)
	if err = encoder.Encode(obj); err != nil {
		return
	}
	if err = encoder.Close(); err != nil {
		return
	}
}

type GetRuntimeObjectListFunc func() (runtime.Object, error)

func getListAndEncodeToYAML(name string, getListFunc GetRuntimeObjectListFunc, yamlsDir string, errLog io.Writer) {
	obj, err := getListFunc()
	if err != nil {
		fmt.Fprintf(errLog, "Support Bundle: failed to get %v: %v\n", name, err)
	}
	encodeToYAMLFile(obj, filepath.Join(yamlsDir, name+".yaml"), errLog)
}

func (m *SupportBundleManager) generateSupportBundleLogs(logsDir string, errLog io.Writer) {
	namespaces := []string{m.Namespace, "default", "kube-system", "cattle-system"}

	for _, ns := range namespaces {
		k8s, err := client.NewKubernetesStore(m.context, ns, m.restConfig)
		if err != nil {
			fmt.Fprint(errLog, err)
			continue
		}
		list, err := k8s.GetAllPodsList()
		if err != nil {
			fmt.Fprintf(errLog, "Support bundle: cannot get pod list: %v\n", err)
			return
		}
		podList, ok := list.(*corev1.PodList)
		if !ok {
			fmt.Fprintf(errLog, "BUG: Support bundle: didn't get pod list\n")
			return
		}
		for _, pod := range podList.Items {
			podName := pod.Name
			podDir := filepath.Join(logsDir, ns, podName)
			for _, container := range pod.Spec.Containers {
				req := k8s.GetPodContainerLogRequest(podName, container.Name)
				logFileName := filepath.Join(podDir, container.Name+".log")
				stream, err := req.Stream(m.context)
				if err != nil {
					fmt.Fprintf(errLog, "BUG: Support bundle: cannot get log for pod %v container %v: %v\n",
						podName, container.Name, err)
					continue
				}
				streamLogToFile(stream, logFileName, errLog)
				stream.Close()
			}
		}
	}
}

func streamLogToFile(logStream io.ReadCloser, path string, errLog io.Writer) {
	var err error
	defer func() {
		if err != nil {
			fmt.Fprintf(errLog, "Support Bundle: failed to generate %v: %v\n", path, err)
		}
	}()
	err = os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	if err != nil {
		return
	}
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = io.Copy(f, logStream)
	if err != nil {
		return
	}
}

func (m *SupportBundleManager) getExternalSupportBundles(bundleMeta *BundleMeta, toDir string, errLog io.Writer) {
	var err error
	defer func() {
		if err != nil {
			fmt.Fprintf(errLog, "Support Bundle: failed to get external bundle: %v\n", err)
		}
	}()
	err = os.Mkdir(toDir, os.FileMode(0755))
	if err != nil {
		return
	}

	lh := external.NewLonghornSupportBundleManager(m.context, m.LonghornAPI)
	err = lh.GetLonghornSupportBundle(bundleMeta.IssueURL, bundleMeta.IssueDescription, toDir)
	if err != nil {
		return
	}
}
