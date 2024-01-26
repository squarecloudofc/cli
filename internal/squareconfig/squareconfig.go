package squareconfig

import (
	"os"
	"sync"
)

var (
	SquareConfigNames    = []string{"squarecloud.config", "squarecloud.app"}
	readSquareConfigFile = new(sync.Once)
	filename             = ""
)

func ConfigFile() string {
	readSquareConfigFile.Do(func() {
		for _, name := range SquareConfigNames {
			_, err := os.Lstat(name)
			if err != nil {
				continue
			}

			filename = name
		}

		if filename == "" {
			filename = SquareConfigNames[0]
		}
	})

	return filename
}
