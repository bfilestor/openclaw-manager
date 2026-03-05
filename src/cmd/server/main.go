package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"openclaw-manager/internal/auth"
	appcfg "openclaw-manager/internal/config"
	"openclaw-manager/internal/server"
	"openclaw-manager/internal/storage"
	"openclaw-manager/internal/user"
)

func main() {
	configPath := flag.String("config", defaultConfigPath(), "path to config.toml")
	staticDir := flag.String("static-dir", "", "path to frontend dist directory")
	flag.Parse()

	if err := validateConfigPath(*configPath); err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	cfg, err := appcfg.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config error: %v\n", err)
		os.Exit(1)
	}

	dbPath := resolveDBPath(cfg)
	db, err := storage.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open db error: %v\n", err)
		os.Exit(1)
	}
	defer db.SQL.Close()

	passSvc := auth.NewPasswordService()
	if cfg.Auth.PasswordMinLen > 0 {
		passSvc.MinLength = cfg.Auth.PasswordMinLen
	}
	tokenRepo := auth.NewTokenRepository(db.SQL)
	jwtSvc := &auth.JWTService{
		Secret:           []byte(cfg.Auth.JWTSecret),
		AccessTokenTTL:   cfg.Auth.AccessTokenTTL,
		RefreshTokenTTL:  cfg.Auth.RefreshTokenTTL,
		BlacklistChecker: tokenRepo,
	}
	authHandler := &auth.Handler{
		Repo:      user.NewRepository(db.SQL),
		Pass:      passSvc,
		Config:    cfg,
		JWT:       jwtSvc,
		TokenRepo: tokenRepo,
	}

	dist := resolveStaticDir(*staticDir)
	s := server.New(cfg.Server.Listen, dist, authHandler)
	fmt.Printf("manager server starting, listen=%s, static_dir=%s, db=%s\n", cfg.Server.Listen, dist, dbPath)

	if err := server.RunWithSignals(s); err != nil {
		fmt.Fprintf(os.Stderr, "server run error: %v\n", err)
		os.Exit(1)
	}
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "~/.openclaw-manager/config.toml"
	}
	return filepath.Join(home, ".openclaw-manager", "config.toml")
}

func validateConfigPath(path string) error {
	if path == "" {
		return fmt.Errorf("config path is empty")
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}

func resolveStaticDir(fromFlag string) string {
	if fromFlag != "" {
		return fromFlag
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(cwd, "frontend", "dist")
}

func resolveDBPath(cfg *appcfg.Config) string {
	if cfg != nil && cfg.Paths.ManagerHome != "" {
		return filepath.Join(cfg.Paths.ManagerHome, "manager.db")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "manager.db")
	}
	return filepath.Join(home, ".openclaw-manager", "manager.db")
}
