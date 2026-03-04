package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	configPath := flag.String("config", defaultConfigPath(), "path to config.toml")
	flag.Parse()

	if err := validateConfigPath(*configPath); err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("manager server bootstrap ok, config=%s\n", *configPath)
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
