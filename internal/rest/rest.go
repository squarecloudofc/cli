package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

var UserAgent = "Square Cloud CLI (v1.0.0)"

type ApiResponse[T any] struct {
	Status   string `json:"status"`
	Code     string `json:"code"`
	Response T      `json:"response"`
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

func (c *RestClient) Request(method, url string, b []byte, options ...RequestOption) (response []byte, err error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(b))
	if err != nil {
		return
	}

	cfg := newRequestConfig(req)
	for _, opt := range options {
		opt(cfg)
	}
	req = cfg.Request

	if c != nil {
		req.Header.Set("Authorization", c.token)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := cfg.Client.Do(req)
	if err != nil {
		return
	}

	response, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return
}

func unmarshal(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		return err
	}

	return nil
}

func (c *RestClient) ServiceStatistics(options ...RequestOption) (result *ResponseServiceStatistics, err error) {
	body, err := c.Request(http.MethodGet, MakeURL(EndpointServiceStatistics()), nil, options...)
	if err != nil {
		return
	}

	var r ApiResponse[ResponseServiceStatistics]
	err = unmarshal(body, &r)
	return &r.Response, err
}

func (c *RestClient) SelfUser(options ...RequestOption) (result *ResponseUser, err error) {
	body, err := c.Request(http.MethodGet, MakeURL(EndpointUser()), nil, options...)
	if err != nil {
		return
	}

	var r ApiResponse[ResponseUser]
	err = unmarshal(body, &r)
	return &r.Response, err
}

func (c *RestClient) Application(appId string, options ...RequestOption) (result *ResponseApplicationInformation, err error) {
	body, err := c.Request(http.MethodGet, MakeURL(EndpointApplication(appId)), nil, options...)
	if err != nil {
		return
	}

	var r ApiResponse[ResponseApplicationInformation]
	err = unmarshal(body, &r)
	return &r.Response, err
}

func (c *RestClient) ApplicationStatus(appId string, options ...RequestOption) (result *ResponseApplicationStatus, err error) {
	body, err := c.Request(http.MethodGet, MakeURL(EndpointApplicationStatus(appId)), nil, options...)
	if err != nil {
		return
	}

	var r ApiResponse[ResponseApplicationStatus]
	err = unmarshal(body, &r)
	return &r.Response, err
}

func (c *RestClient) AllApplicationStatus(options ...RequestOption) (result *ResponseApplicationStatus, err error) {
	body, err := c.Request(http.MethodGet, MakeURL(EndpointApplicationsStatus()), nil, options...)
	if err != nil {
		return
	}

	var r ApiResponse[ResponseApplicationStatus]
	err = unmarshal(body, &r)
	return &r.Response, err
}

func (c *RestClient) ApplicationLogs(appId string, options ...RequestOption) (result *ResponseApplicationLogs, err error) {
	body, err := c.Request(http.MethodGet, MakeURL(EndpointApplicationLogs(appId)), nil, options...)
	if err != nil {
		return
	}

	var r ApiResponse[ResponseApplicationLogs]
	err = unmarshal(body, &r)
	return &r.Response, err
}

func (c *RestClient) ApplicationStart(appId string, options ...RequestOption) (_ bool, err error) {
	body, err := c.Request(http.MethodPost, MakeURL(EndpointApplicationStart(appId)), nil, options...)
	if err != nil {
		return
	}

	var r ApiResponse[any]
	err = unmarshal(body, &r)
	return r.Status == "success", err
}

func (c *RestClient) ApplicationStop(appId string, options ...RequestOption) (_ bool, err error) {
	body, err := c.Request(http.MethodPost, MakeURL(EndpointApplicationStop(appId)), nil, options...)
	if err != nil {
		return
	}

	var r ApiResponse[any]
	err = unmarshal(body, &r)
	return r.Status == "success", err
}

func (c *RestClient) ApplicationRestart(appId string, options ...RequestOption) (_ bool, err error) {
	body, err := c.Request(http.MethodPost, MakeURL(EndpointApplicationRestart(appId)), nil, options...)
	if err != nil {
		return
	}

	var r ApiResponse[any]
	err = unmarshal(body, &r)
	return r.Status == "success", err
}

func (c *RestClient) ApplicationBackup(appId string, options ...RequestOption) (result *ResponseApplicationBackup, err error) {
	body, err := c.Request(http.MethodGet, MakeURL(EndpointApplicationBackup(appId)), nil, options...)
	if err != nil {
		return
	}

	var r ApiResponse[ResponseApplicationBackup]
	err = unmarshal(body, &r)
	return &r.Response, err
}

func (c *RestClient) UploadApplication(options ...RequestOption) (result *ResponseUploadApplication, err error) {
	body, err := c.Request(http.MethodGet, MakeURL(EndpointApplicationUpload()), nil, options...)
	if err != nil {
		return
	}

	var r ApiResponse[ResponseUploadApplication]
	err = unmarshal(body, &r)
	return &r.Response, err
}

func (c *RestClient) ApplicationFiles(appId string, path string, options ...RequestOption) (result *ResponseApplicationFiles, err error) {
	body, err := c.Request(http.MethodGet, MakeURL(EndpointApplicationFiles(appId, path)), nil, options...)
	if err != nil {
		return
	}

	var r ApiResponse[ResponseApplicationFiles]
	err = unmarshal(body, &r)
	return &r.Response, err
}

func (c *RestClient) ApplicationFile(appId string, path string, options ...RequestOption) (result *ResponseApplicationFile, err error) {
	body, err := c.Request(http.MethodGet, MakeURL(EndpointApplicationFile(appId, path)), nil, options...)
	if err != nil {
		return
	}

	var r ApiResponse[ResponseApplicationFile]
	err = unmarshal(body, &r)
	return &r.Response, err
}

func (c *RestClient) ApplicationDeploys(appId string, options ...RequestOption) (result *ResponseApplicationFile, err error) {
	body, err := c.Request(http.MethodGet, MakeURL(EndpointApplicationDeploys(appId)), nil, options...)
	if err != nil {
		return
	}

	var r ApiResponse[ResponseApplicationFile]
	err = unmarshal(body, &r)
	return &r.Response, err
}

func (c *RestClient) ApplicationGithubWebhook(appId string, accessToken string, options ...RequestOption) (result *ResponseApplicationGithubWebhook, err error) {
	mapData := map[string]string{
		accessToken: accessToken,
	}

	data, err := json.Marshal(mapData)
	if err != nil {
		return
	}

	body, err := c.Request(http.MethodPost, MakeURL(EndpointApplicationGithubIntegration(appId)), data, options...)
	if err != nil {
		return
	}

	var r ApiResponse[ResponseApplicationGithubWebhook]
	err = unmarshal(body, &r)
	return &r.Response, err
}

func (c *RestClient) ApplicationNetwork(appId string, options ...RequestOption) (result *ResponseApplicationNetwork, err error) {
	body, err := c.Request(http.MethodGet, MakeURL(EndpointApplicationNetwork(appId)), nil, options...)
	if err != nil {
		return
	}

	var r ApiResponse[ResponseApplicationNetwork]
	err = unmarshal(body, &r)
	return &r.Response, err
}

func (c *RestClient) ApplicationDelete(appId string, options ...RequestOption) (bool, error) {
	body, err := c.Request(http.MethodDelete, MakeURL(EndpointApplicationDelete(appId)), nil, options...)
	if err != nil {
		return false, err
	}

	var r ApiResponse[any]
	err = unmarshal(body, &r)
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

	body, err := c.Request(http.MethodPost, MakeURL(EndpointApplicationCommit(appId)), bodyBuffer.Bytes(), options...)
	if err != nil {
		return false, err
	}

	var r ApiResponse[any]
	err = unmarshal(body, &r)

	return r.Status == "success", err
}
