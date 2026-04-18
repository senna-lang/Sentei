package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig_HasSaneDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Daemon.Addr != "127.0.0.1:7890" {
		t.Errorf("Daemon.Addr = %q, want %q", cfg.Daemon.Addr, "127.0.0.1:7890")
	}
	if cfg.Bonsai.URL != "http://127.0.0.1:8080" {
		t.Errorf("Bonsai.URL = %q, want %q", cfg.Bonsai.URL, "http://127.0.0.1:8080")
	}
	if !cfg.Plugins.Git.Enabled {
		t.Error("Plugins.Git.Enabled should be true by default")
	}
	if cfg.Plugins.Git.PollIntervalSec != 60 {
		t.Errorf("PollIntervalSec = %d, want 60", cfg.Plugins.Git.PollIntervalSec)
	}
	if cfg.Plugins.Git.SurveyIntervalSec != 3600 {
		t.Errorf("SurveyIntervalSec = %d, want 3600", cfg.Plugins.Git.SurveyIntervalSec)
	}
	if len(cfg.Plugins.Git.Repos) != 3 {
		t.Errorf("Repos count = %d, want 3", len(cfg.Plugins.Git.Repos))
	}
}

func TestLoad_ParsesTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[daemon]
addr = "0.0.0.0:9999"

[bonsai]
url = "http://localhost:9090"

[plugins.git]
enabled = false
poll_interval_sec = 30
survey_interval_sec = 1800
repos = ["owner/repo1", "owner/repo2"]
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Daemon.Addr != "0.0.0.0:9999" {
		t.Errorf("Daemon.Addr = %q, want %q", cfg.Daemon.Addr, "0.0.0.0:9999")
	}
	if cfg.Bonsai.URL != "http://localhost:9090" {
		t.Errorf("Bonsai.URL = %q, want %q", cfg.Bonsai.URL, "http://localhost:9090")
	}
	if cfg.Plugins.Git.Enabled {
		t.Error("Plugins.Git.Enabled should be false")
	}
	if cfg.Plugins.Git.PollIntervalSec != 30 {
		t.Errorf("PollIntervalSec = %d, want 30", cfg.Plugins.Git.PollIntervalSec)
	}
	if len(cfg.Plugins.Git.Repos) != 2 {
		t.Errorf("Repos count = %d, want 2", len(cfg.Plugins.Git.Repos))
	}
}

func TestLoad_PartialOverride_KeepsDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	// daemon セクションのみ変更、他はデフォルトを保持
	content := `
[daemon]
addr = "0.0.0.0:8888"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Daemon.Addr != "0.0.0.0:8888" {
		t.Errorf("Daemon.Addr = %q, want %q", cfg.Daemon.Addr, "0.0.0.0:8888")
	}
	// デフォルト値が保持されること
	if cfg.Bonsai.URL != "http://127.0.0.1:8080" {
		t.Errorf("Bonsai.URL should keep default, got %q", cfg.Bonsai.URL)
	}
	if !cfg.Plugins.Git.Enabled {
		t.Error("Plugins.Git.Enabled should keep default true")
	}
}

func TestLoad_MissingFile_ReturnsError(t *testing.T) {
	_, err := Load("/nonexistent/config.toml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestExpandPath_TildeExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	got := ExpandPath("~/foo/bar")
	want := filepath.Join(home, "foo/bar")
	if got != want {
		t.Errorf("ExpandPath(~/foo/bar) = %q, want %q", got, want)
	}
}

func TestExpandPath_AbsolutePathUnchanged(t *testing.T) {
	got := ExpandPath("/absolute/path")
	if got != "/absolute/path" {
		t.Errorf("ExpandPath(/absolute/path) = %q, want /absolute/path", got)
	}
}

func TestDefaultTOML_IsValid(t *testing.T) {
	// DefaultTOML がパース可能であることを確認
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(DefaultTOML()), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("DefaultTOML is not valid TOML: %v", err)
	}
	if cfg.Daemon.Addr != "127.0.0.1:7890" {
		t.Errorf("parsed default has wrong addr: %q", cfg.Daemon.Addr)
	}
}
