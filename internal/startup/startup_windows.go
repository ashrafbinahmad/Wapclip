//go:build windows

package startup

import (
	"fmt"
	"os"
	"path/filepath"

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

	home, err := os.UserHomeDir()
	if err == nil {
		vbsPath := filepath.Join(home, ".wapclip", "startup.vbs")
		vbsContent := fmt.Sprintf(`Set WshShell = CreateObject("WScript.Shell")
WshShell.Run """%s"" daemon", 0, False`, exe)
		os.WriteFile(vbsPath, []byte(vbsContent), 0644)
		
		val := `wscript.exe "` + vbsPath + `"`
		return key.SetStringValue("WapClip", val)
	}

	val := `"` + exe + `" start`
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
