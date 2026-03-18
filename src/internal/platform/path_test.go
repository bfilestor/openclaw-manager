package platform

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultPathsNotEmpty(t *testing.T) {
	openclawHome := DefaultOpenClawHome()
	managerHome := DefaultManagerHome()

	if openclawHome == "" {
		t.Fatalf("DefaultOpenClawHome should not be empty")
	}
	if managerHome == "" {
		t.Fatalf("DefaultManagerHome should not be empty")
	}
}

func TestDefaultPathsBasename(t *testing.T) {
	openclawHome := filepath.Base(DefaultOpenClawHome())
	managerHome := filepath.Base(DefaultManagerHome())

	if openclawHome != ".openclaw" {
		t.Fatalf("openclaw home basename mismatch: %s", openclawHome)
	}
	if managerHome != ".openclaw-manager" {
		t.Fatalf("manager home basename mismatch: %s", managerHome)
	}
}

func TestDefaultPathsUseOSFlavor(t *testing.T) {
	openclawHome := DefaultOpenClawHome()
	managerHome := DefaultManagerHome()

	if runtime.GOOS == "windows" {
		if filepath.VolumeName(openclawHome) == "" && filepath.VolumeName(managerHome) == "" {
			t.Fatalf("expected windows style absolute-like paths, got %s and %s", openclawHome, managerHome)
		}
	}
}
