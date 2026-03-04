package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

var ErrConfigInvalid = errors.New("config invalid")

type Config struct {
	Server ServerConfig `toml:"server"`
	Auth   AuthConfig   `toml:"auth"`
	Paths  PathsConfig  `toml:"paths"`
}

type ServerConfig struct {
	Listen string `toml:"listen"`
}

type AuthConfig struct {
	JWTSecret       string        `toml:"jwt_secret"`
	AccessTokenTTL  time.Duration `toml:"access_token_ttl"`
	RefreshTokenTTL time.Duration `toml:"refresh_token_ttl"`
	PublicRegister  bool          `toml:"public_registration"`
	PasswordMinLen  int           `toml:"password_min_length"`
}

type PathsConfig struct {
	OpenClawHome string `toml:"openclaw_home"`
	ManagerHome  string `toml:"manager_home"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("%w: empty path", ErrConfigInvalid)
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}

	applyDefaults(&cfg)
	applyEnv(&cfg)

	if err := normalizePaths(&cfg); err != nil {
		return nil, err
	}
	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Listen == "" {
		cfg.Server.Listen = "127.0.0.1:18790"
	}
	if cfg.Auth.AccessTokenTTL == 0 {
		cfg.Auth.AccessTokenTTL = 15 * time.Minute
	}
	if cfg.Auth.RefreshTokenTTL == 0 {
		cfg.Auth.RefreshTokenTTL = 7 * 24 * time.Hour
	}
	if cfg.Auth.PasswordMinLen == 0 {
		cfg.Auth.PasswordMinLen = 8
	}
}

func applyEnv(cfg *Config) {
	if v := strings.TrimSpace(os.Getenv("OPENCLAW_JWT_SECRET")); v != "" {
		cfg.Auth.JWTSecret = v
	}
}

func validate(cfg *Config) error {
	if len(cfg.Auth.JWTSecret) < 32 {
		return fmt.Errorf("%w: jwt_secret must be at least 32 bytes", ErrConfigInvalid)
	}

	host, port, err := net.SplitHostPort(cfg.Server.Listen)
	if err != nil || host == "" || port == "" {
		return fmt.Errorf("%w: listen must be host:port", ErrConfigInvalid)
	}

	p, err := strconv.Atoi(port)
	if err != nil || p <= 0 || p > 65535 {
		return fmt.Errorf("%w: listen port invalid", ErrConfigInvalid)
	}
	return nil
}

func normalizePaths(cfg *Config) error {
	var err error
	cfg.Paths.OpenClawHome, err = expandTilde(cfg.Paths.OpenClawHome)
	if err != nil {
		return err
	}
	cfg.Paths.ManagerHome, err = expandTilde(cfg.Paths.ManagerHome)
	if err != nil {
		return err
	}
	return nil
}

func expandTilde(p string) (string, error) {
	if p == "" {
		return p, nil
	}
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if p == "~" {
			return home, nil
		}
		return filepath.Join(home, strings.TrimPrefix(p, "~/")), nil
	}
	return p, nil
}
