package platform

import (
	"os"
	"path/filepath"
	"runtime"
)

// DefaultOpenClawHome returns the OS-specific default openclaw home path.
func DefaultOpenClawHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(home, ".openclaw")
	}
	return filepath.Join(home, ".openclaw")
}

// DefaultManagerHome returns the OS-specific manager data path.
func DefaultManagerHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(home, ".openclaw-manager")
	}
	return filepath.Join(home, ".openclaw-manager")
}
