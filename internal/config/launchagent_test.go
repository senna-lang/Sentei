package config

import (
	"strings"
	"testing"
)

func TestGeneratePlist_ContainsBinaryPath(t *testing.T) {
	plist := GeneratePlist("/usr/local/bin/sentei")

	if !strings.Contains(plist, "/usr/local/bin/sentei") {
		t.Error("plist should contain the binary path")
	}
	if !strings.Contains(plist, "<string>serve</string>") {
		t.Error("plist should contain 'serve' argument")
	}
}

func TestGeneratePlist_ContainsLabel(t *testing.T) {
	plist := GeneratePlist("/usr/local/bin/sentei")

	if !strings.Contains(plist, "com.sentei.daemon") {
		t.Error("plist should contain label 'com.sentei.daemon'")
	}
}

func TestGeneratePlist_HasRunAtLoad(t *testing.T) {
	plist := GeneratePlist("/usr/local/bin/sentei")

	if !strings.Contains(plist, "<key>RunAtLoad</key>") {
		t.Error("plist should have RunAtLoad key")
	}
	if !strings.Contains(plist, "<true/>") {
		t.Error("plist should have RunAtLoad set to true")
	}
}

func TestGeneratePlist_HasLogPaths(t *testing.T) {
	plist := GeneratePlist("/usr/local/bin/sentei")

	if !strings.Contains(plist, "sentei.log") {
		t.Error("plist should contain stdout log path")
	}
	if !strings.Contains(plist, "sentei.err.log") {
		t.Error("plist should contain stderr log path")
	}
}

func TestPlistPath_IsInLaunchAgents(t *testing.T) {
	path := PlistPath()

	if !strings.Contains(path, "Library/LaunchAgents") {
		t.Errorf("PlistPath = %q, should be in Library/LaunchAgents", path)
	}
	if !strings.HasSuffix(path, "com.sentei.daemon.plist") {
		t.Errorf("PlistPath = %q, should end with com.sentei.daemon.plist", path)
	}
}
