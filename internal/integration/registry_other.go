//go:build !windows

package integration

// System is nil: the Explorer integration exists on Windows only.
var System Registry

// Notify does nothing outside Windows.
func Notify(title, message string) error { return nil }
