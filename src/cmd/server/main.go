package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	appcfg "openclaw-manager/internal/config"
	"openclaw-manager/internal/server"
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

	dist := resolveStaticDir(*staticDir)
	s := server.New(cfg.Server.Listen, dist)
	fmt.Printf("manager server starting, listen=%s, static_dir=%s\n", cfg.Server.Listen, dist)

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
