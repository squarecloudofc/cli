package rest

import (
	"fmt"
)

var (
	ApiVersion = 2
	ApiURL     = fmt.Sprintf("https://api.squarecloud.app/v%d", ApiVersion)
)

var (
	// Square Cloud Service
	EndpointServiceStatistics = func() string { return "/service/statistics" }

	// User
	EndpointUser = func() string { return "/users/me" }

	// Application
	EndpointApplication            = func() string { return "/apps" }
	EndpointApplicationListStatus  = func() string { return "/apps/status" }
	EndpointApplicationInformation = func(appId string) string { return fmt.Sprintf("/apps/%s", appId) }
	EndpointApplicationStatus      = func(appId string) string { return fmt.Sprintf("/apps/%s/status", appId) }
	EndpointApplicationLogs        = func(appId string) string { return fmt.Sprintf("/apps/%s/logs", appId) }
	EndpointApplicationStart       = func(appId string) string { return fmt.Sprintf("/apps/%s/start", appId) }
	EndpointApplicationRestart     = func(appId string) string { return fmt.Sprintf("/apps/%s/restart", appId) }
	EndpointApplicationStop        = func(appId string) string { return fmt.Sprintf("/apps/%s/stop", appId) }
	EndpointApplicationBackup      = func(appId string) string { return fmt.Sprintf("/apps/%s/backups", appId) }
	EndpointApplicationCommit      = func(appId string) string { return fmt.Sprintf("/apps/%s/commit", appId) }

	// Application File Manager
	EndpointApplicationFiles    = func(appId, path string) string { return fmt.Sprintf("/apps/%s/files?path=%s", appId, path) }
	EndpointApplicationFileRead = func(appId, path string) string { return fmt.Sprintf("/apps/%s/files/content?path=%s", appId, path) }

	// Application Deploy
	EndpointApplicationDeploys           = func(appId string) string { return fmt.Sprintf("/apps/%s/deploy/list", appId) }
	EndpointApplicationGithubIntegration = func(appId string) string { return fmt.Sprintf("/apps/%s/deploy/git-webhook", appId) }

	// Application Network
	EndpointApplicationNetwork      = func(appId string) string { return fmt.Sprintf("/apps/%s/network/analytics", appId) }
	EndpointApplicationCustomDomain = func(appId, domain string) string { return fmt.Sprintf("/apps/%s/network/custom/%s", appId, domain) }
)
