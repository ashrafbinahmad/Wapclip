//go:build linux

package startup

import (
	"fmt"
	"os"
	"path/filepath"
)

func registerOS() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	desktopPath := filepath.Join(home, ".config", "autostart", "wapclip.desktop")
	
	desktopContent := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=WapClip
Exec="%s" daemon
X-GNOME-Autostart-enabled=true
NoDisplay=true
Terminal=false
`, exe)

	dir := filepath.Dir(desktopPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(desktopPath, []byte(desktopContent), 0644)
}

func deregisterOS() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	desktopPath := filepath.Join(home, ".config", "autostart", "wapclip.desktop")
	return os.Remove(desktopPath)
}
