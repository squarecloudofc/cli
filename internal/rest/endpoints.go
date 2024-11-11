package rest

import (
	"fmt"
	"strconv"
)

var (
	ApiVersion  = 2
	EndpointAPI = fmt.Sprintf("https://api.squarecloud.app/v%d", ApiVersion)
)

var (
	// Square Cloud Service
	EndpointServiceStatistics = func() string { return "/service/statistics" }

	// User
	EndpointUser = func() string { return "/users/me" }

	// Application
	EndpointApplication            = func() string { return "/apps" }
	EndpointApplicationInformation = func(appId string) string { return fmt.Sprintf("/apps/%s", appId) }
	EndpointApplicationStatus      = func(appId string) string { return fmt.Sprintf("/apps/%s/status", appId) }
	EndpointApplicationLogs        = func(appId string) string { return fmt.Sprintf("/apps/%s/logs", appId) }
	EndpointApplicationStart       = func(appId string) string { return fmt.Sprintf("/apps/%s/start", appId) }
	EndpointApplicationRestart     = func(appId string) string { return fmt.Sprintf("/apps/%s/restart", appId) }
	EndpointApplicationStop        = func(appId string) string { return fmt.Sprintf("/apps/%s/stop", appId) }
	EndpointApplicationBackup      = func(appId string) string { return fmt.Sprintf("/apps/%s/backups", appId) }
	EndpointApplicationCommit      = func(appId string, restart bool) string {
		return fmt.Sprintf("/apps/%s/commit?restart=%s", appId, strconv.FormatBool(restart))
	}
	EndpointApplicationDelete  = func(appId string) string { return fmt.Sprintf("/apps/%s/delete", appId) } // ROUTE DEPRECATED USE -> DELETE AT /apps
	EndpointApplicationsStatus = func() string { return "/apps/status" }

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

func MakeURL(path string) string {
	return fmt.Sprint(EndpointAPI, path)
}
