package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/squarecloudofc/cli/internal/build"
)

var UserAgent = fmt.Sprintf("Square Cloud CLI (%s)", build.Version)

type ApiResponse[T any] struct {
	Response T      `json:"response"`
	Status   string `json:"status"`
	Code     string `json:"code"`
}

type RequestConfig struct {
	Request *http.Request
	Client  *http.Client
}

func newRequestConfig(req *http.Request) *RequestConfig {
	return &RequestConfig{
		Client:  &http.Client{},
		Request: req,
	}
}

func (c *RestClient) Request(method, url string, body []byte, respBody interface{}, options ...RequestOption) error {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	cfg := newRequestConfig(req)
	for _, opt := range options {
		opt(cfg)
	}
	req = cfg.Request

	if c.Token() != "" {
		req.Header.Set("Authorization", c.Token())
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := cfg.Client.Do(req)
	if err != nil {
		return err
	}

	var rawResponse []byte
	if rawResponse, err = io.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if err := json.Unmarshal(rawResponse, respBody); err != nil {
		return fmt.Errorf("error unmarshalling response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		return nil
	default:
		var r ApiResponse[any]
		if err := json.Unmarshal(rawResponse, &r); err != nil {
			return fmt.Errorf("error unmarshalling response body: %w", err)
		}
		return ParseError(&r)
	}
}

func (c *RestClient) ServiceStatistics(options ...RequestOption) (*ResponseServiceStatistics, error) {
	var r ApiResponse[ResponseServiceStatistics]
	err := c.Request(http.MethodGet, MakeURL(EndpointServiceStatistics()), nil, &r, options...)
	return &r.Response, err
}

func (c *RestClient) SelfUser(options ...RequestOption) (*ResponseUser, error) {
	var r ApiResponse[ResponseUser]
	err := c.Request(http.MethodGet, MakeURL(EndpointUser()), nil, &r, options...)
	return &r.Response, err
}

func (c *RestClient) Application(appId string, options ...RequestOption) (*ResponseApplicationInformation, error) {
	var r ApiResponse[ResponseApplicationInformation]
	err := c.Request(http.MethodGet, MakeURL(EndpointApplication(appId)), nil, &r, options...)
	return &r.Response, err
}

func (c *RestClient) ApplicationStatus(appId string, options ...RequestOption) (*ResponseApplicationStatus, error) {
	var r ApiResponse[ResponseApplicationStatus]
	err := c.Request(http.MethodGet, MakeURL(EndpointApplicationStatus(appId)), nil, &r, options...)
	return &r.Response, err
}

func (c *RestClient) AllApplicationStatus(options ...RequestOption) (*ResponseApplicationStatus, error) {
	var r ApiResponse[ResponseApplicationStatus]
	err := c.Request(http.MethodGet, MakeURL(EndpointApplicationsStatus()), nil, &r, options...)
	return &r.Response, err
}

func (c *RestClient) ApplicationLogs(appId string, options ...RequestOption) (*ResponseApplicationLogs, error) {
	var r ApiResponse[ResponseApplicationLogs]
	err := c.Request(http.MethodGet, MakeURL(EndpointApplicationLogs(appId)), nil, &r, options...)
	return &r.Response, err
}

func (c *RestClient) ApplicationStart(appId string, options ...RequestOption) (bool, error) {
	var r ApiResponse[any]
	err := c.Request(http.MethodPost, MakeURL(EndpointApplicationStart(appId)), nil, &r, options...)
	return r.Status == "success", err
}

func (c *RestClient) ApplicationStop(appId string, options ...RequestOption) (bool, error) {
	var r ApiResponse[ResponseServiceStatistics]
	err := c.Request(http.MethodPost, MakeURL(EndpointApplicationStop(appId)), nil, &r, options...)
	if err != nil {
		return false, err
	}

	return r.Status == "success", err
}

func (c *RestClient) ApplicationRestart(appId string, options ...RequestOption) (bool, error) {
	var r ApiResponse[ResponseServiceStatistics]
	err := c.Request(http.MethodPost, MakeURL(EndpointApplicationRestart(appId)), nil, &r, options...)
	if err != nil {
		return false, err
	}

	return r.Status == "success", err
}

func (c *RestClient) ApplicationBackup(appId string, options ...RequestOption) (*ResponseApplicationBackup, error) {
	var r ApiResponse[ResponseApplicationBackup]
	err := c.Request(http.MethodGet, MakeURL(EndpointApplicationBackup(appId)), nil, &r, options...)
	if err != nil {
		return nil, err
	}

	return &r.Response, err
}

func (c *RestClient) UploadApplication(options ...RequestOption) (*ResponseUploadApplication, error) {
	var r ApiResponse[ResponseUploadApplication]
	err := c.Request(http.MethodGet, MakeURL(EndpointApplicationUpload()), nil, &r, options...)
	if err != nil {
		return nil, err
	}

	return &r.Response, err
}

func (c *RestClient) ApplicationFiles(appId string, path string, options ...RequestOption) (*ResponseApplicationFiles, error) {
	var r ApiResponse[ResponseApplicationFiles]
	err := c.Request(http.MethodGet, MakeURL(EndpointApplicationFiles(appId, path)), nil, &r, options...)
	if err != nil {
		return nil, err
	}

	return &r.Response, err
}

func (c *RestClient) ApplicationFile(appId string, path string, options ...RequestOption) (*ResponseApplicationFile, error) {
	var r ApiResponse[ResponseApplicationFile]
	err := c.Request(http.MethodGet, MakeURL(EndpointApplicationFile(appId, path)), nil, &r, options...)
	if err != nil {
		return nil, err
	}

	return &r.Response, err
}

func (c *RestClient) ApplicationDeploys(appId string, options ...RequestOption) (*ResponseApplicationFile, error) {
	var r ApiResponse[ResponseApplicationFile]
	err := c.Request(http.MethodGet, MakeURL(EndpointApplicationDeploys(appId)), nil, &r, options...)
	if err != nil {
		return nil, err
	}

	return &r.Response, err
}

func (c *RestClient) ApplicationGithubWebhook(appId string, accessToken string, options ...RequestOption) (*ResponseApplicationGithubWebhook, error) {
	mapData := map[string]string{
		accessToken: accessToken,
	}

	data, err := json.Marshal(mapData)
	if err != nil {
		return nil, err
	}

	var r ApiResponse[ResponseApplicationGithubWebhook]
	err = c.Request(http.MethodPost, MakeURL(EndpointApplicationGithubIntegration(appId)), data, &r, options...)
	if err != nil {
		return nil, err
	}

	return &r.Response, err
}

func (c *RestClient) ApplicationNetwork(appId string, options ...RequestOption) (*ResponseApplicationNetwork, error) {
	var r ApiResponse[ResponseApplicationNetwork]
	err := c.Request(http.MethodGet, MakeURL(EndpointApplicationNetwork(appId)), nil, &r, options...)
	if err != nil {
		return nil, err
	}

	return &r.Response, err
}

func (c *RestClient) ApplicationDelete(appId string, options ...RequestOption) (bool, error) {
	var r ApiResponse[any]
	err := c.Request(http.MethodDelete, MakeURL(EndpointApplicationDelete(appId)), nil, &r, options...)
	if err != nil {
		return false, err
	}

	return r.Status == "success", err
}

func (c *RestClient) ApplicationCommit(appId string, filep string, options ...RequestOption) (bool, error) {
	bodyBuffer := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuffer)

	file, err := os.Open(filep)
	if err != nil {
		return false, err
	}
	defer file.Close()

	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))

	if _, err := io.Copy(part, file); err != nil {
		return false, err
	}

	if err := writer.Close(); err != nil {
		return false, err
	}

	options = append(options, WithHeader("Content-Type", writer.FormDataContentType()))

	var r ApiResponse[any]
	err = c.Request(http.MethodPost, MakeURL(EndpointApplicationCommit(appId)), bodyBuffer.Bytes(), &r, options...)
	if err != nil {
		return false, err
	}

	return r.Status == "success", err
}
