package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateConfigPath(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(cfg, []byte("[server]\nlisten='127.0.0.1:18790'\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := validateConfigPath(cfg); err != nil {
		t.Fatalf("expected valid config path, got error: %v", err)
	}
}

func TestValidateConfigPathNotFound(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.toml")
	if err := validateConfigPath(missing); err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}

func TestDefaultConfigPath(t *testing.T) {
	got := defaultConfigPath()
	if filepath.Base(got) != "config.toml" {
		t.Fatalf("expected basename config.toml, got %s", filepath.Base(got))
	}
}
