package integration

import (
	"encoding/base64"
	"slices"
	"strings"
	"testing"
	"unicode/utf16"
)

// FakeRegistry is an in-memory HKEY_CURRENT_USER: key -> name -> data.
type FakeRegistry map[string]map[string]string

func (f FakeRegistry) SetString(key, name, data string) error {
	if f[key] == nil {
		f[key] = map[string]string{}
	}
	f[key][name] = data
	return nil
}

func (f FakeRegistry) DeleteKey(key string) error {
	delete(f, key)
	return nil
}

func TestInstallThenUninstallLeavesNothing(t *testing.T) {
	exe := `C:\Users\me\AppData\Local\squarecloud\bin\squarecloud.exe`
	reg := FakeRegistry{}
	if err := Install(reg, Values(exe, "Send to Square Cloud")); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{
		`Software\Classes\Directory\shell\SquareCloud\command`,
		`Software\Classes\SystemFileAssociations\.zip\shell\SquareCloud\command`,
	} {
		if got, want := reg[key][""], `"`+exe+`" integration send "%1"`; got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
	if reg[`Software\Classes\Directory\shell\SquareCloud`]["MUIVerb"] != "Send to Square Cloud" {
		t.Errorf("folder menu label = %v", reg[`Software\Classes\Directory\shell\SquareCloud`])
	}

	// Every key written is one Uninstall removes.
	if err := Uninstall(reg); err != nil {
		t.Fatal(err)
	}
	if len(reg) != 0 {
		t.Fatalf("left behind: %v", reg)
	}
}

// TestToastCommandKeepsTextsOutOfTheScript: the texts reach PowerShell as
// environment variables only. Quotes (ASCII and the U+2018-U+201B PowerShell
// also treats as quotes), $(), backticks and markup never become script.
func TestToastCommandKeepsTextsOutOfTheScript(t *testing.T) {
	var first string
	for _, text := range []string{
		`it's "done"`,
		"a\u2018b", "a\u2019+(Start-Process calc)+\u2019.zip", "a\u201ab", "a\u201bb",
		`$(Start-Process calc)`, "`whoami`",
		`x'); Remove-Item C:\ -Recurse; ('<b>&`,
		"Échec de l\u2019envoi de « l\u2019app »",
	} {
		encoded, env := toastCommand(text, "body "+text)
		raw, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			t.Fatal(err)
		}
		units := make([]uint16, len(raw)/2)
		for i := range units {
			units[i] = uint16(raw[2*i]) | uint16(raw[2*i+1])<<8
		}
		script := string(utf16.Decode(units))

		if strings.Contains(script, text) {
			t.Errorf("script contains %q:\n%s", text, script)
		}
		if first == "" {
			first = script
		} else if script != first {
			t.Errorf("script changed with the text %q", text)
		}
		if want := []string{"SQ_TOAST_TITLE=" + text, "SQ_TOAST_BODY=body " + text}; !slices.Equal(env, want) {
			t.Errorf("env = %q, want %q", env, want)
		}
	}
	if !strings.Contains(first, "$env:SQ_TOAST_TITLE") || !strings.Contains(first, "$env:SQ_TOAST_BODY") {
		t.Errorf("script does not read the texts from the environment:\n%s", first)
	}
}
