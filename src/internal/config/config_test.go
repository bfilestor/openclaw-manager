package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}
	return p
}

func TestLoadSuccess(t *testing.T) {
	os.Unsetenv("OPENCLAW_JWT_SECRET")
	p := writeConfig(t, `
[server]
listen = "127.0.0.1:18790"

[auth]
jwt_secret = "12345678901234567890123456789012"
access_token_ttl = "10m"
refresh_token_ttl = "72h"
password_min_length = 10

[paths]
openclaw_home = "~/.openclaw"
manager_home = "~/.openclaw-manager"
`)

	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if cfg.Server.Listen != "127.0.0.1:18790" {
		t.Fatalf("listen mismatch: %s", cfg.Server.Listen)
	}
	if cfg.Auth.AccessTokenTTL != 10*time.Minute {
		t.Fatalf("access ttl mismatch: %v", cfg.Auth.AccessTokenTTL)
	}
	if !strings.HasPrefix(cfg.Paths.ManagerHome, "/") {
		t.Fatalf("manager_home should be absolute: %s", cfg.Paths.ManagerHome)
	}
}

func TestLoadJWTSecretTooShort(t *testing.T) {
	os.Unsetenv("OPENCLAW_JWT_SECRET")
	p := writeConfig(t, `
[server]
listen = "127.0.0.1:18790"

[auth]
jwt_secret = "short"
`)

	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "jwt_secret") {
		t.Fatalf("expected jwt_secret error, got: %v", err)
	}
}

func TestLoadEnvOverrideJWTSecret(t *testing.T) {
	t.Setenv("OPENCLAW_JWT_SECRET", "abcdefghijklmnopqrstuvwxyz123456")
	p := writeConfig(t, `
[server]
listen = "127.0.0.1:18790"

[auth]
jwt_secret = "short"
`)

	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.Auth.JWTSecret != "abcdefghijklmnopqrstuvwxyz123456" {
		t.Fatalf("env override failed, got: %s", cfg.Auth.JWTSecret)
	}
}

func TestLoadInvalidListen(t *testing.T) {
	p := writeConfig(t, `
[server]
listen = "invalid"

[auth]
jwt_secret = "12345678901234567890123456789012"
`)

	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "listen") {
		t.Fatalf("expected listen error, got: %v", err)
	}
}

func TestLoadDefaultAccessTTL(t *testing.T) {
	p := writeConfig(t, `
[server]
listen = "127.0.0.1:18790"

[auth]
jwt_secret = "12345678901234567890123456789012"
`)

	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.Auth.AccessTokenTTL != 15*time.Minute {
		t.Fatalf("expected default 15m, got %v", cfg.Auth.AccessTokenTTL)
	}
}
