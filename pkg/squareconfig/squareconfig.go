package squareconfig

import (
	"os"
	"path/filepath"
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

// FileIn is the config file of the project in dir ("" for the current
// folder): the last of SquareConfigNames there, else GetDefaultConfig; found
// is whether it exists. What the CLI reads and what init overwrites.
func FileIn(dir string) (name string, found bool) {
	name = GetDefaultConfig()
	for _, n := range SquareConfigNames {
		if _, err := os.Lstat(filepath.Join(dir, n)); err == nil {
			name, found = n, true
		}
	}
	return name, found
}

func GetConfigFile() string {
	readSquareConfigFile.Do(func() { filename, _ = FileIn("") })
	return filename
}
