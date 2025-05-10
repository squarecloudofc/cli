package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Xuanwo/go-locale"
)

var SupportedLanguages = []string{"en", "pt", "es", "zh"}

//go:embed data/*.json
var localeData embed.FS

var LocaleContents = make(map[string]map[string]any)

func init() {
	data, err := localeData.ReadDir("data")
	if err != nil {
		panic(fmt.Errorf("failed to read locale directory: %w", err))
	}

	englishData := make(map[string]any)

	for _, d := range data {
		localeName := strings.Split(d.Name(), ".")[0]

		lang, err := localeData.ReadFile("data/" + d.Name())
		if err != nil {
			fmt.Printf("failed to read locale file %s: %v\n", d.Name(), err)
			continue
		}

		var langData map[string]any
		if err := json.Unmarshal(lang, &langData); err != nil {
			fmt.Printf("failed to parse locale file %s: %v\n", d.Name(), err)
			continue
		}

		flatLangData := toFlatMap(langData)
		LocaleContents[localeName] = flatLangData

		if localeName == "en" {
			englishData = flatLangData
		}
	}

	for lang, langData := range LocaleContents {
		if lang == "en" {
			continue
		}

		for key, value := range englishData {
			if _, exists := langData[key]; !exists {
				langData[key] = value
			}
		}

		LocaleContents[lang] = langData
	}
}

func DetectSystemLanguage() string {
	result, err := locale.Detect()
	if err != nil {
		return "en"
	}

	value, _ := result.Base()
	for _, lang := range SupportedLanguages {
		if value.String() == lang {
			return lang
		}
	}

	return "en"
}
