//go:build darwin

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

	plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.wapclip.plist")
	
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.wapclip</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>daemon</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>`, exe)

	dir := filepath.Dir(plistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(plistPath, []byte(plistContent), 0644)
}

func deregisterOS() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.wapclip.plist")
	return os.Remove(plistPath)
}
