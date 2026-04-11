//go:build windows

package startup

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

func registerOS() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	val := `"` + exe + `" daemon`
	return key.SetStringValue("WapClip", val)
}

func deregisterOS() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.DeleteValue("WapClip")
}
