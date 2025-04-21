package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jeandeaual/go-locale"
	"golang.org/x/text/language"
)

var SupportedLanguages = []language.Tag{
	language.English,
	language.Portuguese,
	language.Spanish,
	language.Chinese,
}

var matcher = language.NewMatcher(SupportedLanguages)

//go:embed data/*
var localeData embed.FS

var LocaleContents = make(map[language.Tag]map[string]any)

func init() {
	data, err := localeData.ReadDir("data")
	if err != nil {
		panic(fmt.Errorf("failed to read locale directory: %w", err))
	}

	englishLocale := language.English
	englishData := make(map[string]any)

	for _, d := range data {
		localeName := strings.Split(d.Name(), ".")[0]
		tag, err := language.Parse(localeName)
		if err != nil {
			fmt.Printf("invalid language tag %s: %v\n", localeName, err)
			continue
		}

		lang, err := localeData.ReadFile(filepath.Join("data", d.Name()))
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
		LocaleContents[tag] = flatLangData

		if tag == englishLocale {
			englishData = flatLangData
		}
	}

	for tag, langData := range LocaleContents {
		if tag == englishLocale {
			continue
		}

		for key, value := range englishData {
			if _, exists := langData[key]; !exists {
				langData[key] = value
			}
		}

		LocaleContents[tag] = langData
	}
}

func DetectSystemLanguage() language.Tag {
	result := language.English

	osLocale, err := locale.GetLanguage()
	if err == nil {
		if osLocale, err := language.Parse(osLocale); err == nil {
			result, _ = language.MatchStrings(matcher, osLocale.String())
		}
	}

	return result
}
