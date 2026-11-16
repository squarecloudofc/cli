// Package integration adds "Send to Square Cloud" to the Windows Explorer
// menu of folders and .zip files. Everything lives under HKEY_CURRENT_USER,
// so no administrator rights are needed.
package integration

import (
	"encoding/base64"
	"unicode/utf16"
)

// AppID is the AppUserModelID the result notifications are sent as.
const AppID = "SquareCloud.CLI"

var (
	menuKeys = []string{
		`Software\Classes\Directory\shell\SquareCloud`,
		`Software\Classes\SystemFileAssociations\.zip\shell\SquareCloud`,
	}
	appIDKey = `Software\Classes\AppUserModelId\` + AppID
)

// Value is one registry value under HKEY_CURRENT_USER; Name "" is the
// key's default value.
type Value struct{ Key, Name, Data string }

// Registry writes under HKEY_CURRENT_USER.
type Registry interface {
	SetString(key, name, data string) error
	// DeleteKey removes a key without subkeys; nil when it does not exist.
	DeleteKey(key string) error
}

// Values are what Install writes for the binary at exe, label being the
// menu entry's text.
func Values(exe, label string) []Value {
	command := `"` + exe + `" integration send "%1"`
	var values []Value
	for _, key := range menuKeys {
		values = append(values,
			Value{key, "MUIVerb", label},
			Value{key, "Icon", exe},
			Value{key + `\command`, "", command},
		)
	}
	return append(values, Value{appIDKey, "DisplayName", "Square Cloud"})
}

// Keys are every key Install creates, children first: what Uninstall
// deletes. Shared parents (Directory\shell...) are left alone.
func Keys() []string {
	var keys []string
	for _, key := range menuKeys {
		keys = append(keys, key+`\command`, key)
	}
	return append(keys, appIDKey)
}

func Install(r Registry, values []Value) error {
	for _, v := range values {
		if err := r.SetString(v.Key, v.Name, v.Data); err != nil {
			return err
		}
	}
	return nil
}

// Uninstall deletes every key, going on past a failure; it reports the
// first one.
func Uninstall(r Registry) error {
	var first error
	for _, key := range Keys() {
		if err := r.DeleteKey(key); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// toastScript shows a notification. Its texts come from $env:SQ_TOAST_TITLE
// and $env:SQ_TOAST_BODY and go in as XML text nodes: nothing user-derived is
// ever parsed as PowerShell or as XML, whatever quotes it holds (PowerShell
// also closes strings on U+2018-U+201B).
const toastScript = `[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null
$x = New-Object Windows.Data.Xml.Dom.XmlDocument
$x.LoadXml('<toast><visual><binding template="ToastGeneric"><text/><text/></binding></visual></toast>')
$t = $x.GetElementsByTagName('text')
$t.Item(0).AppendChild($x.CreateTextNode($env:SQ_TOAST_TITLE)) | Out-Null
$t.Item(1).AppendChild($x.CreateTextNode($env:SQ_TOAST_BODY)) | Out-Null
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('` + AppID + `').Show([Windows.UI.Notifications.ToastNotification]::new($x))`

// toastCommand is the -EncodedCommand (base64 UTF-16LE) of toastScript and
// the environment entries that carry the texts.
func toastCommand(title, message string) (encoded string, env []string) {
	units := utf16.Encode([]rune(toastScript))
	raw := make([]byte, 2*len(units))
	for i, u := range units {
		raw[2*i], raw[2*i+1] = byte(u), byte(u>>8)
	}
	return base64.StdEncoding.EncodeToString(raw), []string{"SQ_TOAST_TITLE=" + title, "SQ_TOAST_BODY=" + message}
}
