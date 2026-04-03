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

func TestConfigDirUsesOverrideAsIs(t *testing.T) {
	override := filepath.Join(t.TempDir(), "agsm")
	t.Setenv("AGSM_CONFIG_HOME", override)

	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir returned error: %v", err)
	}
	if dir != override {
		t.Fatalf("expected override path %q, got %q", override, dir)
	}
}

func TestLoadInvalidSortFallsBackToDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	data := []byte("sort_by = \"nme\"\nsort_order = \"up\"\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.SortBy != Default().SortBy || cfg.SortOrder != Default().SortOrder {
		t.Fatalf("expected default sort config, got %#v", cfg)
	}
}
