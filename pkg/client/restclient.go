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
	"time"

	"github.com/bk201/support-bundle-utils/pkg/utils"
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

	httpClient *http.Client
	token      string
}

type JWTAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type JWTAuthResponse struct {
	JWEToken string `json:"jweToken"`
}

func NewRESTClient(ctx context.Context, apiURL string, username string, password string, insecure bool) *RESTClient {
	httpClient := http.Client{
		Timeout: 15 * time.Second,
	}
	if insecure {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	return &RESTClient{
		context:    ctx,
		apiURL:     apiURL,
		username:   username,
		password:   password,
		insecure:   insecure,
		httpClient: &httpClient,
	}
}

func (r *RESTClient) Login() error {
	auth := JWTAuthRequest{Username: r.username, Password: r.password}
	data, err := json.Marshal(auth)
	if err != nil {
		return err
	}
	resp, err := r.Post(r.apiURL+HarvesterURLAuthLogin, data)
	if err != nil {
		return err
	}

	var authResp JWTAuthResponse
	err = json.Unmarshal(resp, &authResp)
	if err != nil {
		return err
	}
	r.token = authResp.JWEToken
	return nil
}

func (r *RESTClient) Logout() error {
	if r.token == "" {
		return errors.New("not login")
	}
	_, err := r.Post(r.apiURL+HarvesterURLAuthLogout, nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *RESTClient) Get(url string) ([]byte, error) {
	return r.Request(http.MethodGet, url, nil)
}

func (r *RESTClient) Post(url string, data []byte) ([]byte, error) {
	return r.Request(http.MethodPost, url, data)
}

func (r *RESTClient) Request(method string, url string, data []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(r.context, method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if r.token != "" {
		req.Header.Set("jweToken", r.token)
	}

	resp, err := r.httpClient.Do(req)
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

func (r *RESTClient) Download(url string, path string) (string, error) {
	req, err := http.NewRequestWithContext(r.context, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	if r.token != "" {
		req.Header.Set("jweToken", r.token)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("unexpected status: %s", resp.Status)
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			errMsg = fmt.Sprintf("%s: %s", errMsg, body)
		}
		return "", errors.New(errMsg)
	}

	filename := path
	if filename == "" {
		filename, err = utils.HttpGetDispositionFilename(resp.Header.Get("Content-Disposition"))
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
