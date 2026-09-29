package i18n

import "golang.org/x/sys/windows"

// uiLanguages is the Windows display language first, then the other
// languages the user added, as tags like "pt-BR".
func uiLanguages() []string {
	langs, err := windows.GetUserPreferredUILanguages(windows.MUI_LANGUAGE_NAME)
	if err != nil {
		return nil
	}
	return langs
}
