package client

import (
	"context"

	"github.com/rancher/harvester/pkg/apis/harvester.cattle.io/v1alpha1"
	"github.com/rancher/harvester/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

type HarvesterClient struct {
	context   context.Context
	namespace string
	clientset *versioned.Clientset
}

func NewHarvesterStore(ctx context.Context, namespace string, config *rest.Config) (*HarvesterClient, error) {
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &HarvesterClient{
		context:   ctx,
		namespace: namespace,
		clientset: clientset,
	}, nil
}

func (h *HarvesterClient) GetSupportBundleState(name string) (string, error) {
	sb, err := h.clientset.HarvesterV1alpha1().SupportBundles(h.namespace).Get(h.context, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return sb.Status.State, nil
}

func (h *HarvesterClient) UpdateSupportBundleState(name string, state string) error {
	sb, err := h.clientset.HarvesterV1alpha1().SupportBundles(h.namespace).Get(h.context, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	toUpdate := sb.DeepCopy()
	toUpdate.Status.State = state
	_, err = h.clientset.HarvesterV1alpha1().SupportBundles(h.namespace).Update(h.context, toUpdate, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (h *HarvesterClient) GetAllKeypairs() (runtime.Object, error) {
	return h.clientset.HarvesterV1alpha1().KeyPairs(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllPreferences() (runtime.Object, error) {
	return h.clientset.HarvesterV1alpha1().Preferences(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllSettings() (runtime.Object, error) {
	return h.clientset.HarvesterV1alpha1().Settings().List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetSettingByName(name string) (*v1alpha1.Setting, error) {
	return h.clientset.HarvesterV1alpha1().Settings().Get(h.context, name, metav1.GetOptions{})
}

func (h *HarvesterClient) GetAllUpgrades() (runtime.Object, error) {
	return h.clientset.HarvesterV1alpha1().Upgrades(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllUsers() (runtime.Object, error) {
	return h.clientset.HarvesterV1alpha1().Users().List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineBackups() (runtime.Object, error) {
	return h.clientset.HarvesterV1alpha1().VirtualMachineBackups(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineBackupContents() (runtime.Object, error) {
	return h.clientset.HarvesterV1alpha1().VirtualMachineBackupContents(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineImages() (runtime.Object, error) {
	return h.clientset.HarvesterV1alpha1().VirtualMachineImages(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineRestores() (runtime.Object, error) {
	return h.clientset.HarvesterV1alpha1().VirtualMachineRestores(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineTemplates() (runtime.Object, error) {
	return h.clientset.HarvesterV1alpha1().VirtualMachineTemplates(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineTemplateVersions() (runtime.Object, error) {
	return h.clientset.HarvesterV1alpha1().VirtualMachineTemplateVersions(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachines() (runtime.Object, error) {
	return h.clientset.KubevirtV1().VirtualMachines(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineInstances() (runtime.Object, error) {
	return h.clientset.KubevirtV1().VirtualMachineInstances(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineInstanceMigrations() (runtime.Object, error) {
	return h.clientset.KubevirtV1().VirtualMachineInstanceMigrations(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllDataVolumes() (runtime.Object, error) {
	return h.clientset.CdiV1beta1().DataVolumes(h.namespace).List(h.context, metav1.ListOptions{})
}
