package i18n

import (
	"maps"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"
	"text/template"
)

var (
	actionRe = regexp.MustCompile(`\{\{.*?\}\}`)
	fieldRe  = regexp.MustCompile(`\.[A-Za-z_][\w.]*`)
)

// fields returns the .Field references of s's template actions. Only the set
// counts: a translation may shape its if/else branches differently.
func fields(s string) []string {
	set := map[string]bool{}
	for _, action := range actionRe.FindAllString(s, -1) {
		for _, f := range fieldRe.FindAllString(action, -1) {
			set[f] = true
		}
	}
	return slices.Sorted(maps.Keys(set))
}

// TestLocaleParity keeps every resource file in step with en.json, the
// source the translations are made from.
func TestLocaleParity(t *testing.T) {
	en := LocaleContents["en"]
	if len(en) == 0 {
		t.Fatal("en.json is missing or failed to parse")
	}

	for _, lang := range SupportedLanguages {
		if _, ok := LocaleContents[lang]; !ok {
			t.Errorf("%s is in SupportedLanguages but resources/%s.json is missing or failed to parse", lang, lang)
		}
	}

	for lang, data := range LocaleContents {
		if !slices.Contains(SupportedLanguages, lang) {
			t.Errorf("resources/%s.json exists but %s is not in SupportedLanguages", lang, lang)
		}

		for key, value := range data {
			s, _ := value.(string)
			if strings.ContainsAny(s, string([]rune{0x2013, 0x2014})) {
				t.Errorf("%s: %s has an en or em dash: %q", lang, key, s)
			}
			// T() falls back to the raw text on a parse error: {{ .X }} on screen.
			if _, err := template.New("").Parse(s); err != nil {
				t.Errorf("%s: %s is not a valid template: %v", lang, key, err)
			}

			enValue, ok := en[key]
			if !ok {
				t.Errorf("%s: %s is not in en.json", lang, key)
				continue
			}
			enString, _ := enValue.(string)
			if got, want := fields(s), fields(enString); !slices.Equal(got, want) {
				t.Errorf("%s: %s uses %v, en.json uses %v", lang, key, got, want)
			}
		}

		for key := range en {
			if _, ok := data[key]; !ok {
				t.Errorf("%s: %s is missing (it is in en.json)", lang, key)
			}
		}
	}
}

func TestDetectSystemLanguage(t *testing.T) {
	t.Setenv("LANGUAGE", "") // empty counts as unset
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "")
	for env, want := range map[string]string{
		"pt_BR.UTF-8": "pt",
		"pt_PT":       "pt",
		"zh_TW.UTF-8": "zh",
		"zh-CN":       "zh",
		"es_MX.UTF-8": "es",
		"de_AT.UTF-8": "de",
		"en_GB.UTF-8": "en",
		"C":           "en",
		"xx_YY":       "en",
	} {
		t.Setenv("LC_ALL", env)
		if got := DetectSystemLanguage(); got != want {
			t.Errorf("LC_ALL=%s: DetectSystemLanguage() = %q, want %q", env, got, want)
		}
	}
}

// Without locale variables Windows answers with its display language, not
// the regional format.
func TestDetectSystemLanguageWithoutEnv(t *testing.T) {
	for _, env := range []string{"LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG"} {
		t.Setenv(env, "")
	}
	langs := uiLanguages()
	if runtime.GOOS == "windows" && len(langs) == 0 {
		t.Fatal("Windows returned no display language")
	}
	if len(langs) > 0 {
		if got, want := DetectSystemLanguage(), matchLanguage(langs); got != want {
			t.Errorf("DetectSystemLanguage() = %q, display languages %v give %q", got, langs, want)
		}
	}
}

func TestMatchLanguage(t *testing.T) {
	for _, c := range []struct {
		tags []string
		want string
	}{
		{[]string{"pt-BR"}, "pt"},
		{[]string{"zh-Hans-CN"}, "zh"},
		{[]string{"ja-JP"}, "ja"},
		{[]string{"FR_ca"}, "fr"},
		{[]string{"ru-RU", "it-IT"}, "it"}, // no Russian: the next language the user added
		{[]string{"xx-YY"}, "en"},
		{nil, "en"},
	} {
		if got := matchLanguage(c.tags); got != c.want {
			t.Errorf("matchLanguage(%v) = %q, want %q", c.tags, got, c.want)
		}
	}
}
