//go:build !windows

package i18n

// uiLanguages is empty here: Linux has no display language apart from the
// locale variables, and on macOS go-locale's AppleLocale ("en_BR") already
// starts with the display language.
func uiLanguages() []string { return nil }
