/**
 * macOS LaunchAgent plist 生成
 * sentei init で LaunchAgent の自動起動設定を作成する
 */
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// GeneratePlist は LaunchAgent の plist XML を生成する
func GeneratePlist(binaryPath string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.sentei.daemon</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>serve</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>%s/sentei.log</string>
  <key>StandardErrorPath</key>
  <string>%s/sentei.err.log</string>
</dict>
</plist>
`, binaryPath, ConfigDir(), ConfigDir())
}

// PlistPath は LaunchAgent plist のインストール先パスを返す
func PlistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", "com.sentei.daemon.plist")
}
