package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const (
	HarvesterURLAuthLogin  = "/v1-public/auth?action=login"
	HarvesterURLAuthLogout = "/v1-public/auth?action=logout"
)

type RESTClient struct {
	context  context.Context
	apiURL   string
	username string
	password string
	insecure bool

	token string
}

type JWTAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type JWTAuthResponse struct {
	JWEToken string `json:"jweToken"`
}

func NewRESTClient(ctx context.Context, apiURL string, username string, password string, insecure bool) *RESTClient {
	return &RESTClient{
		context:  ctx,
		apiURL:   apiURL,
		username: username,
		password: password,
		insecure: insecure,
	}
}

func (c *RESTClient) getHTTPClient() http.Client {
	client := http.Client{}
	if c.insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	return client
}

func (c *RESTClient) Login() error {
	auth := JWTAuthRequest{Username: c.username, Password: c.password}
	data, err := json.Marshal(auth)
	if err != nil {
		return err
	}
	resp, err := c.Post(c.apiURL+HarvesterURLAuthLogin, data)
	if err != nil {
		return err
	}

	var authResp JWTAuthResponse
	err = json.Unmarshal(resp, &authResp)
	if err != nil {
		return err
	}
	c.token = authResp.JWEToken
	return nil
}

func (c *RESTClient) Logout() error {
	if c.token == "" {
		return errors.New("not login")
	}
	_, err := c.Post(c.apiURL+HarvesterURLAuthLogout, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *RESTClient) Get(url string) ([]byte, error) {
	return c.Request(http.MethodGet, url, nil)
}

func (c *RESTClient) Post(url string, data []byte) ([]byte, error) {
	return c.Request(http.MethodPost, url, data)
}

func (c *RESTClient) Request(method string, url string, data []byte) ([]byte, error) {
	client := c.getHTTPClient()
	req, err := http.NewRequestWithContext(c.context, method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		req.Header.Set("jweToken", c.token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s. Body: %s", resp.Status, respBody)
	}
	return respBody, nil
}

func (c *RESTClient) Download(url string, path string) (string, error) {
	client := c.getHTTPClient()
	req, err := http.NewRequestWithContext(c.context, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	if c.token != "" {
		req.Header.Set("jweToken", c.token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}

	filename := path
	if filename == "" {
		filename, err = c.getFilename(resp.Header.Get("Content-Disposition"))
		if err != nil {
			return "", fmt.Errorf("fail to parse filename from response header: %s", err)
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return filename, err
}

// getFilename parse value of "Content-Disposition" header
// e.g., extract "abc.zip" from "attachment; filename=abc.zip"
func (c *RESTClient) getFilename(disposition string) (string, error) {
	errMsg := fmt.Errorf("unexpected disposition value: %s", disposition)

	if disposition == "" {
		return "", errMsg
	}

	var attachement bool
	var filename string

	for _, param := range strings.Split(disposition, ";") {
		p := strings.TrimSpace(param)
		if p == "attachment" {
			attachement = true
		}

		if strings.HasPrefix(p, "filename=") {
			filename = strings.Trim(strings.SplitN(p, "filename=", 2)[1], "\"")
		}
	}
	if attachement && filename != "" {
		return filename, nil
	}
	return "", errMsg
}
