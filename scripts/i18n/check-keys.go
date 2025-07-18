package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/squarecloudofc/cli/i18n"
)

func GetKeysInUse() []string {
	pattern := regexp.MustCompile(`T\("([^"]+)"(\)|,\s+?map.*)`)
	var found []string

	filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		matches := pattern.FindAllStringSubmatch(string(data), -1)
		for _, m := range matches {
			key := m[1]
			if !slices.Contains(found, key) {
				found = append(found, key)
			}
		}
		return nil
	})

	return found
}

func main() {
	localizer := i18n.NewLocalizer("pt")
	localeData := localizer.LocaleData()
	keysInUse := GetKeysInUse()

	var inUse, notInUse, invalid []string

	for _, key := range keysInUse {
		if _, ok := localeData[key]; ok {
			inUse = append(inUse, key)
		} else {
			invalid = append(invalid, key)
		}
	}

	for key := range localeData {
		if !slices.Contains(keysInUse, key) {
			notInUse = append(notInUse, key)
		}
	}

	result := map[string][]string{
		"in_use":     inUse,
		"not_in_use": notInUse,
		"invalid":    invalid,
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
}
