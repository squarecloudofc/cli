package squarecloud

type ApplicationFileInfo struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	Size         int    `json:"size"`
	LastModified int64  `json:"lastModified"`
}

type ApplicationFileData struct {
	Type string `json:"type"`
	Data []int  `json:"data"`
}
