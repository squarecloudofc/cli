package squarecloud

import "time"

type ApplicationDeploys [][]struct {
	Date  time.Time `json:"date"`
	ID    string    `json:"id"`
	State string    `json:"state"`
}

type ApplicationGithubWebhook struct {
	Webhook string `json:"webhook"`
}

type ApplicaitonNetworkData struct {
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
