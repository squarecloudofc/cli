package square

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
