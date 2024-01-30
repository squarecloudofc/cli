package squareconfig

import (
	"os"
	"runtime"
	"sync"
)

var (
	SquareConfigNames    = []string{"squarecloud.config", "squarecloud.app"}
	readSquareConfigFile = new(sync.Once)
	filename             = ""
)

func GetDefaultConfig() string {
	if runtime.GOOS == "darwin" {
		return SquareConfigNames[0]
	}

	return SquareConfigNames[1]
}

func GetConfigFile() string {
	readSquareConfigFile.Do(func() {
		for _, name := range SquareConfigNames {
			_, err := os.Lstat(name)
			if err != nil {
				continue
			}

			filename = name
		}

		if filename == "" {
			filename = GetDefaultConfig()
		}
	})

	return filename
}
