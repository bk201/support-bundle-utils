package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/bk201/support-bundle-utils/pkg/client/utils"
)

type SupportBundleClient struct {
	url        string
	User       string
	Password   string
	NoAuth     bool
	OutputFile string
	Insecure   bool

	IssueURL         string
	IssueDescription string
	r                *RESTClient
}

type SupportBundleInitateInput struct {
	IssueURL    string `json:"issueURL"`
	Description string `json:"description"`
}

type BundleState string

const (
	BundleStateInProgress       = BundleState("InProgress")
	BundleStateReadyForDownload = BundleState("ReadyForDownload")
	BundleStateError            = BundleState("Error")
)

type BundleError string

type SupportBundleResource struct {
	PodID              string      `json:"podID"`
	NodeID             string      `json:"nodeID"`
	State              BundleState `json:"state"`
	Name               string      `json:"name"`
	ErrorMessage       BundleError `json:"errorMessage"`
	ProgressPercentage int         `json:"progressPercentage"`
}

func (sbr *SupportBundleResource) BackendID() string {
	// This is just to be compatible with Longhorn support bundle
	if sbr.PodID != "" {
		return sbr.PodID
	}
	return sbr.NodeID
}

func (c *SupportBundleClient) Run(url string) error {
	c.url = url

	c.r = NewRESTClient(context.TODO(), c.url, c.User, c.Password, c.Insecure)
	if !c.NoAuth {
		err := c.r.Login()
		if err != nil {
			return fmt.Errorf("fail to login: %s", err)
		}
		defer func() {
			err = c.r.Logout()
			if err != nil {
				fmt.Fprintf(os.Stderr, "fail to logout: %s\n", err)
			}
		}()
	}

	sbr, err := c.create()
	if err != nil {
		return err
	}
	fmt.Printf("bundle %s is being generated...", sbr.Name)

	err = c.wait(sbr)
	if err != nil {
		return err
	}

	saved, err := c.download(sbr, c.OutputFile)
	if err != nil {
		return err
	}
	fmt.Printf("bundle is saved to %s\n", saved)
	return nil
}

func (c *SupportBundleClient) create() (*SupportBundleResource, error) {
	sbi := SupportBundleInitateInput{
		IssueURL:    c.IssueURL,
		Description: c.IssueDescription,
	}
	url := c.url + "/v1/supportbundles"

	data, err := json.Marshal(sbi)
	if err != nil {
		return nil, err
	}
	resp, err := c.r.Post(url, data)
	if err != nil {
		return nil, err
	}

	var sbr SupportBundleResource
	err = json.Unmarshal(resp, &sbr)
	if err != nil {
		return nil, err
	}

	return &sbr, nil
}

func (c *SupportBundleClient) wait(sbr *SupportBundleResource) error {
	utils.WriteStdout("\n")
	interval := time.Second * 5
	retries := 20
	previousProgress := 0
	for retries > 0 {
		url := fmt.Sprintf("%s/v1/supportbundles/%s/%s", c.url, sbr.BackendID(), sbr.Name)
		resp, err := c.r.Get(url)
		if err != nil {
			return err
		}
		var newSbr SupportBundleResource
		err = json.Unmarshal(resp, &newSbr)
		if err != nil {
			return err
		}

		if newSbr.State == BundleStateReadyForDownload {
			utils.WriteStdout(".100\n")
			return nil
		}

		utils.WriteStdout(".")
		if previousProgress != newSbr.ProgressPercentage {
			utils.WriteStdout(fmt.Sprintf("%d", newSbr.ProgressPercentage))
			previousProgress = newSbr.ProgressPercentage
		}
		time.Sleep(interval)
		retries--
	}
	return fmt.Errorf("timeout for waiting a bundle")
}

func (c *SupportBundleClient) download(sbr *SupportBundleResource, path string) (string, error) {
	url := fmt.Sprintf("%s/v1/supportbundles/%s/%s/download", c.url, sbr.BackendID(), sbr.Name)
	return c.r.Download(url, path)
}
