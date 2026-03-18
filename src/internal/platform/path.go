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

// DefaultGatewayLogDir returns the conventional log directory used by openclaw.
func DefaultGatewayLogDir() string {
	if runtime.GOOS == "windows" {
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			home, _ := os.UserHomeDir()
			base = filepath.Join(home, "AppData", "Local")
		}
		if base == "" {
			return filepath.Join(os.TempDir(), "openclaw")
		}
		return filepath.Join(base, "openclaw")
	}
	return filepath.Join(os.TempDir(), "openclaw")
}
