package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/Xuanwo/go-locale"
)

var SupportedLanguages = []string{"en", "pt", "es", "zh", "de", "fr", "it", "ja"}

//go:embed resources/*.json
var localeData embed.FS

var LocaleContents = make(map[string]map[string]any)

func init() {
	data, err := localeData.ReadDir("resources")
	if err != nil {
		panic(fmt.Errorf("failed to read locale directory: %w", err))
	}

	for _, d := range data {
		localeName := strings.Split(d.Name(), ".")[0]

		lang, err := localeData.ReadFile("resources/" + d.Name())
		if err != nil {
			fmt.Printf("failed to read locale file %s: %v\n", d.Name(), err)
			continue
		}

		var langData map[string]any
		if err := json.Unmarshal(lang, &langData); err != nil {
			fmt.Printf("failed to parse locale file %s: %v\n", d.Name(), err)
			continue
		}

		LocaleContents[localeName] = toFlatMap(langData)
	}
}

// DetectSystemLanguage picks the language of the CLI's messages: the locale
// variables when set, else the OS display language. On Windows go-locale
// would fall back to the regional format, which only says how dates and
// numbers look: English Windows with Brazilian formats is still English.
func DetectSystemLanguage() string {
	for _, env := range []string{"LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG"} {
		if os.Getenv(env) != "" {
			return detectViaLocale()
		}
	}
	if langs := uiLanguages(); len(langs) > 0 {
		return matchLanguage(langs)
	}
	return detectViaLocale()
}

func detectViaLocale() string {
	tag, err := locale.Detect()
	if err != nil {
		return "en"
	}
	return matchLanguage([]string{tag.String()})
}

// matchLanguage returns the first supported language among tags such as
// "pt-BR" or "zh-Hans-CN", in the user's order of preference.
func matchLanguage(tags []string) string {
	for _, tag := range tags {
		base, _, _ := strings.Cut(strings.ReplaceAll(tag, "_", "-"), "-")
		if slices.Contains(SupportedLanguages, strings.ToLower(base)) {
			return strings.ToLower(base)
		}
	}
	return "en"
}
