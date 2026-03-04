package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func setupValidator(t *testing.T) (*PathValidator, string) {
	t.Helper()
	base := filepath.Join(t.TempDir(), "allow")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir base failed: %v", err)
	}
	v, err := NewPathValidator([]string{base})
	if err != nil {
		t.Fatalf("new validator failed: %v", err)
	}
	return v, base
}

func TestValidateLegalPath(t *testing.T) {
	v, base := setupValidator(t)
	p := filepath.Join(base, "openclaw.json")
	if err := os.WriteFile(p, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	got, err := v.Validate(p)
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if got != p {
		t.Fatalf("path mismatch: got=%s want=%s", got, p)
	}
}

func TestValidatePathTraversal(t *testing.T) {
	v, base := setupValidator(t)
	_, err := v.Validate(filepath.Join(base, "..", "..", "etc", "passwd"))
	if err == nil {
		t.Fatal("expected path not allowed error")
	}
}

func TestValidateAbsoluteOutside(t *testing.T) {
	v, _ := setupValidator(t)
	_, err := v.Validate("/etc/shadow")
	if err == nil {
		t.Fatal("expected path not allowed error")
	}
}

func TestValidateSymlinkOutside(t *testing.T) {
	v, base := setupValidator(t)
	link := filepath.Join(base, "link")
	if err := os.Symlink("/etc/passwd", link); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}
	_, err := v.Validate(link)
	if err == nil {
		t.Fatal("expected path not allowed for symlink out")
	}
}

func TestValidateBasePath(t *testing.T) {
	v, base := setupValidator(t)
	if _, err := v.Validate(base); err != nil {
		t.Fatalf("base path should be valid: %v", err)
	}
}

func TestValidateEmptyPath(t *testing.T) {
	v, _ := setupValidator(t)
	_, err := v.Validate("")
	if err == nil {
		t.Fatal("expected empty path error")
	}
}

func TestValidateNullByte(t *testing.T) {
	v, _ := setupValidator(t)
	_, err := v.Validate("/tmp/\x00evil")
	if err == nil {
		t.Fatal("expected null byte error")
	}
}

func TestJoinAndValidateOK(t *testing.T) {
	v, base := setupValidator(t)
	targetDir := filepath.Join(base, "my-skill")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	got, err := v.JoinAndValidate(base, "my-skill")
	if err != nil {
		t.Fatalf("join validate failed: %v", err)
	}
	if got != targetDir {
		t.Fatalf("path mismatch: got=%s want=%s", got, targetDir)
	}
}

func TestJoinAndValidateTraversal(t *testing.T) {
	v, base := setupValidator(t)
	_, err := v.JoinAndValidate(base, "../../../etc/cron.d/evil")
	if err == nil {
		t.Fatal("expected traversal rejected")
	}
}
