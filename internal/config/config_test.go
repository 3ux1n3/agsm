package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultsWhenFileMissing(t *testing.T) {
	t.Setenv("AGSM_CONFIG_HOME", t.TempDir())

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if !cfg.Agents.OpenCode.Enabled {
		t.Fatal("expected opencode to be enabled by default")
	}
}

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	data := []byte("sort_by = \"name\"\nsort_order = \"asc\"\n[agents.opencode]\nenabled = true\nsession_path = \"/tmp/opencode\"\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.SortBy != "name" || cfg.SortOrder != "asc" {
		t.Fatalf("unexpected sort config: %#v", cfg)
	}
	if cfg.Agents.OpenCode.SessionPath != "/tmp/opencode" {
		t.Fatalf("unexpected session path: %s", cfg.Agents.OpenCode.SessionPath)
	}
}
