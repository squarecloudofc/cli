//go:build windows

package integration

import (
	"errors"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

// System is the current user's registry; nil where there is none.
var System Registry = currentUser{}

type currentUser struct{}

func (currentUser) SetString(key, name, data string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, key, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetStringValue(name, data)
}

func (currentUser) DeleteKey(key string) error {
	if err := registry.DeleteKey(registry.CURRENT_USER, key); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return err
	}
	return nil
}

// Notify shows a native notification, best effort.
func Notify(title, message string) error {
	encoded, env := toastCommand(title, message)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-EncodedCommand", encoded)
	cmd.Env = append(os.Environ(), env...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}
