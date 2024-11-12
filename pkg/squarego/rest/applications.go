package rest

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gabriel-vasile/mimetype"
	"github.com/squarecloudofc/cli/pkg/squarego/square"
)

var _ Applications = (*applicationsImpl)(nil)

func NewApplications(client Client) Applications {
	return &applicationsImpl{client: client}
}

type Applications interface {
	GetApplications(options ...RequestOpt) ([]square.UserApplication, error)
	GetApplicationListStatus(options ...RequestOpt) ([]square.ApplicationListStatus, error)

	PostApplications(reader io.Reader, options ...RequestOpt) (*square.ApplicationUploaded, error)

	GetApplication(appId string, options ...RequestOpt) (square.Application, error)
	GetApplicationStatus(appId string, options ...RequestOpt) (square.ApplicationStatus, error)

	GetApplicationLogs(appId string, options ...RequestOpt) (square.ApplicationLogs, error)
	PostApplicationSignal(appId string, signal square.ApplicationSignal, options ...RequestOpt) error
	PostApplicationCommit(appId string, reader io.Reader, options ...RequestOpt) error

	GetApplicationBackups(appId string, options ...RequestOpt) ([]square.ApplicationBackup, error)
	CreateApplicationBackup(appId string, options ...RequestOpt) (square.ApplicationBackupCreated, error)

	// GetApplicationFileContent(appId string, path string, options ...RequestOption) error
	// GetApplicationFiles(appId string, path string, options ...RequestOption) error
	// CreateApplicationFile(appId string, path string, options ...RequestOption) error
	// PatchApplicationFile(appId string, path string, to string, options ...RequestOption) error
	// DeleteApplicationFile(appId string, path string, options ...RequestOption) error
	//
	// GetApplicationDeployments(appId string, options ...RequestOption) error
	// GetApplicationCurrentDeployments(appId string, options ...RequestOption) error
	// PostApplicationDeployWebhook(appId string, options ...RequestOption) error
	//
	// GetApplicationDNSRecords(appId string, options ...RequestOption) error
	// GetApplicationAnalytics(appId string, options ...RequestOption) error
	// PostApplicationCustomDomain(appId string, options ...RequestOption) error
	// PostApplicationPurgeCache(appId string, options ...RequestOption) error
	//
	DeleteApplication(appId string, options ...RequestOpt) error
}

type applicationsImpl struct {
	client Client
}

func (s *applicationsImpl) GetApplications(opts ...RequestOpt) ([]square.UserApplication, error) {
	var r square.APIResponse[square.User]
	err := s.client.Request(http.MethodGet, EndpointUser(), nil, &r)

	return r.Response.Applications, err
}

func (s *applicationsImpl) GetApplicationListStatus(opts ...RequestOpt) ([]square.ApplicationListStatus, error) {
	var r square.APIResponse[[]square.ApplicationListStatus]
	err := s.client.Request(http.MethodGet, EndpointApplicationListStatus(), nil, &r)

	return r.Response, err
}

func (s *applicationsImpl) PostApplications(reader io.Reader, opts ...RequestOpt) (*square.ApplicationUploaded, error) {
	bodyBuffer := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuffer)

	mime, err := mimetype.DetectReader(reader)
	if err != nil {
		return nil, err
	}

	part, err := writer.CreateFormFile("file", fmt.Sprintf("commit%s", mime.Extension()))
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

	var r square.APIResponse[square.ApplicationUploaded]

	err = s.client.Request(http.MethodPost, EndpointApplication(), bodyBuffer.Bytes(), &r, opts...)
	if err != nil {
		return nil, err
	}
	return &r.Response, err
}

func (s *applicationsImpl) GetApplication(appId string, opts ...RequestOpt) (square.Application, error) {
	var r square.APIResponse[square.Application]
	err := s.client.Request(http.MethodGet, EndpointApplicationInformation(appId), nil, &r)

	return r.Response, err
}

func (s *applicationsImpl) GetApplicationStatus(appId string, opts ...RequestOpt) (square.ApplicationStatus, error) {
	var r square.APIResponse[square.ApplicationStatus]
	err := s.client.Request(http.MethodGet, EndpointApplicationStatus(appId), nil, &r)

	return r.Response, err
}

func (s *applicationsImpl) GetApplicationLogs(appId string, opts ...RequestOpt) (square.ApplicationLogs, error) {
	var r square.APIResponse[square.ApplicationLogs]
	err := s.client.Request(http.MethodGet, EndpointApplicationLogs(appId), nil, &r)

	return r.Response, err
}

func (s *applicationsImpl) PostApplicationSignal(appId string, signal square.ApplicationSignal, opts ...RequestOpt) error {
	var r square.APIResponse[any]
	var endpoint string

	switch signal {
	case square.ApplicationSignalStart:
		endpoint = EndpointApplicationStart(appId)
	case square.ApplicationSignalRestart:
		endpoint = EndpointApplicationRestart(appId)
	case square.ApplicationSignalStop:
		endpoint = EndpointApplicationStop(appId)
	}

	return s.client.Request(http.MethodPost, endpoint, nil, &r)
}

func (c *applicationsImpl) PostApplicationCommit(appId string, reader io.Reader, options ...RequestOpt) error {
	bodyBuffer := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuffer)

	mime, err := mimetype.DetectReader(reader)
	if err != nil {
		return err
	}

	part, err := writer.CreateFormFile("file", fmt.Sprintf("commit%s", mime.Extension()))
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

	var r square.APIResponse[any]
	return c.client.Request(http.MethodPost, EndpointApplicationCommit(appId), bodyBuffer.Bytes(), &r, options...)
}

func (s *applicationsImpl) GetApplicationBackups(appId string, opts ...RequestOpt) ([]square.ApplicationBackup, error) {
	var r square.APIResponse[[]square.ApplicationBackup]
	err := s.client.Request(http.MethodGet, EndpointApplicationBackup(appId), nil, &r)

	return r.Response, err
}

func (s *applicationsImpl) CreateApplicationBackup(appId string, opts ...RequestOpt) (square.ApplicationBackupCreated, error) {
	var r square.APIResponse[square.ApplicationBackupCreated]
	err := s.client.Request(http.MethodPost, EndpointApplicationBackup(appId), nil, &r)

	return r.Response, err
}

func (s *applicationsImpl) DeleteApplication(appId string, opts ...RequestOpt) error {
	var r square.APIResponse[any]
	return s.client.Request(http.MethodDelete, EndpointApplicationInformation(appId), nil, &r)
}
