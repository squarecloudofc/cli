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
	User struct {
		ID     string `json:"id"`
		Tag    string `json:"tag"`
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
	Applications []struct {
		ID        string `json:"id"`
		Tag       string `json:"tag"`
		RAM       int    `json:"ram"`
		Lang      string `json:"lang"`
		Cluster   string `json:"cluster"`
		IsWebsite bool   `json:"isWebsite"`
		Avatar    string `json:"avatar"`
	} `json:"applications"`
}

// Application

type ResponseApplicationInformation struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Avatar         string `json:"avatar"`
	Owner          string `json:"owner"`
	Cluster        string `json:"cluster"`
	RAM            int    `json:"ram"`
	Language       string `json:"language"`
	Domain         string `json:"domain"`
	Custom         string `json:"custom"`
	IsWebsite      bool   `json:"isWebsite"`
	GitIntegration bool   `json:"gitIntegration"`
}

type ResponseApplicationStatus struct {
	CPU     string `json:"cpu"`
	RAM     string `json:"ram"`
	Status  string `json:"status"`
	Running bool   `json:"running"`
	Storage string `json:"storage"`
	Network struct {
		Total string `json:"total"`
		Now   string `json:"now"`
	} `json:"network"`
	Requests int   `json:"requests"`
	Uptime   int64 `json:"uptime"`
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
	ID          string `json:"id"`
	Tag         string `json:"tag"`
	Description string `json:"description"`
	Subdomain   any    `json:"subdomain"`
	Avatar      string `json:"avatar"`
	RAM         int    `json:"ram"`
	CPU         int    `json:"cpu"`
	Language    struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"language"`
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
	ID    string    `json:"id"`
	State string    `json:"state"`
	Date  time.Time `json:"date"`
}

type ResponseApplicationGithubWebhook struct {
	Webhook string `json:"webhook"`
}

// ApplicationNetwork

type ResponseApplicationNetwork struct {
	Hostname string `json:"hostname"`
	Total    struct {
		Visits    int    `json:"visits"`
		Megabytes string `json:"megabytes"`
		Bytes     int    `json:"bytes"`
	} `json:"total"`
	Countries []struct {
		Country   string `json:"country"`
		Visits    int    `json:"visits"`
		Megabytes string `json:"megabytes"`
		Bytes     int    `json:"bytes"`
	} `json:"countries"`
	Methods []struct {
		Method    string `json:"method"`
		Visits    int    `json:"visits"`
		Megabytes string `json:"megabytes"`
		Bytes     int    `json:"bytes"`
	} `json:"methods"`
	Referers []struct {
		Referer   string `json:"referer"`
		Visits    int    `json:"visits"`
		Megabytes string `json:"megabytes"`
		Bytes     int    `json:"bytes"`
	} `json:"referers"`
	Browsers []struct {
		Browser   string `json:"browser"`
		Visits    int    `json:"visits"`
		Megabytes string `json:"megabytes"`
		Bytes     int    `json:"bytes"`
	} `json:"browsers"`
	DeviceTypes []struct {
		Device    string `json:"device"`
		Visits    int    `json:"visits"`
		Megabytes string `json:"megabytes"`
		Bytes     int    `json:"bytes"`
	} `json:"deviceTypes"`
	OperatingSystems []struct {
		Os        string `json:"os"`
		Visits    int    `json:"visits"`
		Megabytes string `json:"megabytes"`
		Bytes     int    `json:"bytes"`
	} `json:"operatingSystems"`
	Agents []struct {
		Agent     string `json:"agent"`
		Visits    int    `json:"visits"`
		Megabytes string `json:"megabytes"`
		Bytes     int    `json:"bytes"`
	} `json:"agents"`
	Hosts []struct {
		Host      string `json:"host"`
		Visits    int    `json:"visits"`
		Megabytes string `json:"megabytes"`
		Bytes     int    `json:"bytes"`
	} `json:"hosts"`
	Paths []struct {
		Path      string `json:"path"`
		Visits    int    `json:"visits"`
		Megabytes string `json:"megabytes"`
		Bytes     int    `json:"bytes"`
	} `json:"paths"`
}
