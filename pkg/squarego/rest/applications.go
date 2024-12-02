package rest

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/squarecloudofc/cli/pkg/squarego/squarecloud"
)

var _ Applications = (*applicationsImpl)(nil)

func NewApplications(client Client) Applications {
	return &applicationsImpl{client: client}
}

type Applications interface {
	GetApplications(options ...RequestOpt) ([]squarecloud.UserApplication, error)
	GetApplicationListStatus(options ...RequestOpt) ([]squarecloud.ApplicationListStatus, error)

	PostApplications(reader io.Reader, options ...RequestOpt) (*squarecloud.ApplicationUploaded, error)

	GetApplication(appId string, options ...RequestOpt) (squarecloud.Application, error)
	GetApplicationStatus(appId string, options ...RequestOpt) (squarecloud.ApplicationStatus, error)

	GetApplicationLogs(appId string, options ...RequestOpt) (squarecloud.ApplicationLogs, error)
	PostApplicationSignal(appId string, signal squarecloud.ApplicationSignal, options ...RequestOpt) error
	PostApplicationCommit(appId string, reader io.Reader, options ...RequestOpt) error

	GetApplicationBackups(appId string, options ...RequestOpt) ([]squarecloud.ApplicationBackup, error)
	CreateApplicationBackup(appId string, options ...RequestOpt) (squarecloud.ApplicationBackupCreated, error)

	// GetApplicationFileContent(appId string, path string, options ...RequestOption) error
	// GetApplicationFiles(appId string, path string, options ...RequestOption) error
	// CreateApplicationFile(appId string, path string, options ...RequestOption) error
	// PatchApplicationFile(appId string, path string, to string, options ...RequestOption) error
	// DeleteApplicationFile(appId string, path string, options ...RequestOption) error

	// GetApplicationDeployments(appId string, options ...RequestOption) error
	// GetApplicationCurrentDeployments(appId string, options ...RequestOption) error
	// PostApplicationDeployWebhook(appId string, options ...RequestOption) error

	// GetApplicationDNSRecords(appId string, options ...RequestOption) error
	// GetApplicationAnalytics(appId string, options ...RequestOption) error
	// PostApplicationCustomDomain(appId string, options ...RequestOption) error
	// PostApplicationPurgeCache(appId string, options ...RequestOption) error

	DeleteApplication(appId string, options ...RequestOpt) error
}

type applicationsImpl struct {
	client Client
}

func (s *applicationsImpl) GetApplications(opts ...RequestOpt) ([]squarecloud.UserApplication, error) {
	var r squarecloud.APIResponse[responseUser]
	err := s.client.Request(http.MethodGet, EndpointUser(), nil, &r, opts...)

	return r.Response.Applications, err
}

func (s *applicationsImpl) GetApplicationListStatus(opts ...RequestOpt) ([]squarecloud.ApplicationListStatus, error) {
	var r squarecloud.APIResponse[[]squarecloud.ApplicationListStatus]
	err := s.client.Request(http.MethodGet, EndpointApplicationListStatus(), nil, &r, opts...)

	return r.Response, err
}

func (s *applicationsImpl) PostApplications(reader io.Reader, opts ...RequestOpt) (*squarecloud.ApplicationUploaded, error) {
	bodyBuffer := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuffer)

	part, err := writer.CreateFormFile("file", "upload.zip")
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(part, reader); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	opts = append(opts, WithHeader("Content-Type", writer.FormDataContentType()))

	var r squarecloud.APIResponse[squarecloud.ApplicationUploaded]

	if err = s.client.Request(http.MethodPost, EndpointApplication(), bodyBuffer.Bytes(), &r, opts...); err != nil {
		return nil, err
	}

	return &r.Response, err
}

func (s *applicationsImpl) GetApplication(appId string, opts ...RequestOpt) (squarecloud.Application, error) {
	var r squarecloud.APIResponse[squarecloud.Application]
	err := s.client.Request(http.MethodGet, EndpointApplicationInformation(appId), nil, &r, opts...)

	return r.Response, err
}

func (c *applicationsImpl) PostApplicationCommit(appId string, reader io.Reader, options ...RequestOpt) error {
	bodyBuffer := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuffer)

	part, err := writer.CreateFormFile("file", "commit.zip")
	if err != nil {
		return err
	}

	if _, err := io.Copy(part, reader); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	options = append(options, WithHeader("Content-Type", writer.FormDataContentType()))

	var r squarecloud.APIResponse[any]
	return c.client.Request(http.MethodPost, EndpointApplicationCommit(appId), bodyBuffer.Bytes(), &r, options...)
}

func (s *applicationsImpl) GetApplicationStatus(appId string, opts ...RequestOpt) (squarecloud.ApplicationStatus, error) {
	var r squarecloud.APIResponse[squarecloud.ApplicationStatus]
	err := s.client.Request(http.MethodGet, EndpointApplicationStatus(appId), nil, &r, opts...)

	return r.Response, err
}

func (s *applicationsImpl) GetApplicationLogs(appId string, opts ...RequestOpt) (squarecloud.ApplicationLogs, error) {
	var r squarecloud.APIResponse[squarecloud.ApplicationLogs]
	err := s.client.Request(http.MethodGet, EndpointApplicationLogs(appId), nil, &r, opts...)

	return r.Response, err
}

func (s *applicationsImpl) PostApplicationSignal(appId string, signal squarecloud.ApplicationSignal, opts ...RequestOpt) error {
	var r squarecloud.APIResponse[any]
	var endpoint string

	switch signal {
	case squarecloud.ApplicationSignalStart:
		endpoint = EndpointApplicationStart(appId)
	case squarecloud.ApplicationSignalRestart:
		endpoint = EndpointApplicationRestart(appId)
	case squarecloud.ApplicationSignalStop:
		endpoint = EndpointApplicationStop(appId)
	}

	return s.client.Request(http.MethodPost, endpoint, nil, &r, opts...)
}

func (s *applicationsImpl) GetApplicationBackups(appId string, opts ...RequestOpt) ([]squarecloud.ApplicationBackup, error) {
	var r squarecloud.APIResponse[[]squarecloud.ApplicationBackup]
	err := s.client.Request(http.MethodGet, EndpointApplicationBackup(appId), nil, &r, opts...)

	return r.Response, err
}

func (s *applicationsImpl) CreateApplicationBackup(appId string, opts ...RequestOpt) (squarecloud.ApplicationBackupCreated, error) {
	var r squarecloud.APIResponse[squarecloud.ApplicationBackupCreated]
	err := s.client.Request(http.MethodPost, EndpointApplicationBackup(appId), nil, &r, opts...)

	return r.Response, err
}

func (s *applicationsImpl) DeleteApplication(appId string, opts ...RequestOpt) error {
	var r squarecloud.APIResponse[any]
	return s.client.Request(http.MethodDelete, EndpointApplicationInformation(appId), nil, &r, opts...)
}
