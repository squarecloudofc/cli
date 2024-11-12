package square

import "time"

type Application struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Owner    string `json:"owner"`
	Cluster  string `json:"cluster"`
	Language string `json:"language"`
	Domain   string `json:"domain"`
	Custom   string `json:"custom"`
	RAM      int    `json:"ram"`
}

type ApplicationSignal string

const (
	ApplicationSignalStart   ApplicationSignal = "START"
	ApplicationSignalStop    ApplicationSignal = "STOP"
	ApplicationSignalRestart ApplicationSignal = "RESTART"
)

type ApplicationStatus struct {
	Network struct {
		Total string `json:"total"`
		Now   string `json:"now"`
	} `json:"network"`
	CPU      string `json:"cpu"`
	RAM      string `json:"ram"`
	Status   string `json:"status"`
	Storage  string `json:"storage"`
	Requests int    `json:"requests"`
	Uptime   int64  `json:"uptime"`
	Running  bool   `json:"running"`
}

type ApplicationListStatus []struct {
	ID      string `json:"id"`
	CPU     string `json:"cpu"`
	RAM     string `json:"ram"`
	Running bool   `json:"running"`
}

type ApplicationLogs struct {
	Logs string `json:"logs"`
}

type ApplicationBackup struct {
	Name     string    `json:"name"`
	Size     int       `json:"size"`
	Modified time.Time `json:"modified"`
	Key      string    `json:"key"`
}

type ApplicationBackupCreated struct {
	URL string `json:"url"`
	key string `json:"key"`
}

type ApplicationUploaded struct {
	Language struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"language"`

	ID          string `json:"id"`
	Tag         string `json:"tag"`
	Description string `json:"description"`
	Subdomain   string `json:"subdomain"`
	Avatar      string `json:"avatar"`
	RAM         int    `json:"ram"`
	CPU         int    `json:"cpu"`
}
