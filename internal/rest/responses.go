package rest

import "time"

type ResponseServiceStatistics struct {
	Worker     int `json:"worker"`
	Statistics struct {
		Users    int   `json:"users"`
		Apps     int   `json:"apps"`
		Websites int   `json:"websites"`
		Ping     int   `json:"ping"`
		Time     int64 `json:"time"`
	} `json:"statistics"`
}

type ResponseUser struct {
	Applications []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Lang    string `json:"lang"`
		Cluster string `json:"cluster"`
		Avatar  string `json:"avatar"`
		RAM     int    `json:"ram"`
	} `json:"applications"`

	User struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Locale string `json:"locale"`
		Email  string `json:"email"`
		Plan   struct {
			Name   string `json:"name"`
			Memory struct {
				Limit     int `json:"limit"`
				Available int `json:"available"`
				Used      int `json:"used"`
			} `json:"memory"`
			Duration int64 `json:"duration"`
		} `json:"plan"`
	} `json:"user"`
}

// Application

type ResponseApplicationInformation struct {
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

type ResponseApplicationStatus struct {
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

type ResponseApplicationsStatus []struct {
	ID      string `json:"id"`
	CPU     string `json:"cpu"`
	RAM     string `json:"ram"`
	Running bool   `json:"running"`
}

type ResponseApplicationLogs struct {
	Logs string `json:"logs"`
}

type ResponseApplicationBackup struct {
	DownloadURL string `json:"downloadURL"`
}

type ResponseUploadApplication struct {
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

// Application File Manager

type ResponseApplicationFiles []struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	Size         int    `json:"size"`
	LastModified int64  `json:"lastModified"`
}

type ResponseApplicationFile struct {
	Type string `json:"type"`
	Data []int  `json:"data"`
}

// Application Deploy

type ResponseApplicationDeploys [][]struct {
	Date  time.Time `json:"date"`
	ID    string    `json:"id"`
	State string    `json:"state"`
}

type ResponseApplicationGithubWebhook struct {
	Webhook string `json:"webhook"`
}

// ApplicationNetwork

type ResponseApplicationNetwork struct {
	Hostname  string `json:"hostname"`
	Countries []struct {
		Country   string `json:"country"`
		Megabytes string `json:"megabytes"`
		Visits    int    `json:"visits"`
		Bytes     int    `json:"bytes"`
	} `json:"countries"`
	Methods []struct {
		Method    string `json:"method"`
		Megabytes string `json:"megabytes"`
		Visits    int    `json:"visits"`
		Bytes     int    `json:"bytes"`
	} `json:"methods"`
	Referers []struct {
		Referer   string `json:"referer"`
		Megabytes string `json:"megabytes"`
		Visits    int    `json:"visits"`
		Bytes     int    `json:"bytes"`
	} `json:"referers"`
	Browsers []struct {
		Browser   string `json:"browser"`
		Megabytes string `json:"megabytes"`
		Visits    int    `json:"visits"`
		Bytes     int    `json:"bytes"`
	} `json:"browsers"`
	DeviceTypes []struct {
		Device    string `json:"device"`
		Megabytes string `json:"megabytes"`
		Visits    int    `json:"visits"`
		Bytes     int    `json:"bytes"`
	} `json:"deviceTypes"`
	OperatingSystems []struct {
		Os        string `json:"os"`
		Megabytes string `json:"megabytes"`
		Visits    int    `json:"visits"`
		Bytes     int    `json:"bytes"`
	} `json:"operatingSystems"`
	Agents []struct {
		Agent     string `json:"agent"`
		Megabytes string `json:"megabytes"`
		Visits    int    `json:"visits"`
		Bytes     int    `json:"bytes"`
	} `json:"agents"`
	Hosts []struct {
		Host      string `json:"host"`
		Megabytes string `json:"megabytes"`
		Visits    int    `json:"visits"`
		Bytes     int    `json:"bytes"`
	} `json:"hosts"`
	Paths []struct {
		Path      string `json:"path"`
		Megabytes string `json:"megabytes"`
		Visits    int    `json:"visits"`
		Bytes     int    `json:"bytes"`
	} `json:"paths"`
	Total struct {
		Megabytes string `json:"megabytes"`
		Visits    int    `json:"visits"`
		Bytes     int    `json:"bytes"`
	} `json:"total"`
}
