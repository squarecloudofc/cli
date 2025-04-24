package i18n

import (
	"bytes"
	"html/template"
)

type Localizer interface {
	Locale() string
	LocaleData() map[string]any
	T(key string, params ...map[string]any) string
}

type localizerImpl struct {
	locale string
}

func NewLocalizer(lang string) Localizer {
	return &localizerImpl{
		locale: lang,
	}
}

func (l *localizerImpl) Locale() string {
	return l.locale
}

func (l *localizerImpl) LocaleData() map[string]any {
	return LocaleContents[l.locale]
}

func (l *localizerImpl) T(key string, params ...map[string]any) string {
	value, ok := l.LocaleData()[key]
	if !ok {
		return key
	}

	strValue, ok := value.(string)
	if !ok {
		return key
	}

	if len(params) == 0 {
		return strValue
	}

	tmpl, err := template.New("translation").Parse(strValue)
	if err != nil {
		return strValue
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params[0]); err != nil {
		return strValue
	}

	return buf.String()
}
