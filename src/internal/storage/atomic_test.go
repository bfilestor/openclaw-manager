package storage

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestAtomicWriteFileOK(t *testing.T) {
	p := filepath.Join(t.TempDir(), "a.txt")
	if err := AtomicWriteFile(p, []byte("hello"), 0o640); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(b) != "hello" {
		t.Fatalf("content mismatch: %s", b)
	}
}

func TestAtomicWriteFileDirNotExist(t *testing.T) {
	p := filepath.Join(t.TempDir(), "not", "exist", "a.txt")
	if err := AtomicWriteFile(p, []byte("x"), 0o644); err == nil {
		t.Fatal("expected error for missing dir")
	}
}

func TestAtomicWriteConcurrent(t *testing.T) {
	p := filepath.Join(t.TempDir(), "c.txt")
	var wg sync.WaitGroup
	payloads := []string{"aaaa", "bbbb", "cccc", "dddd"}
	for _, s := range payloads {
		wg.Add(1)
		go func(v string) {
			defer wg.Done()
			_ = AtomicWriteFile(p, []byte(v), 0o644)
		}(s)
	}
	wg.Wait()

	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	got := string(b)
	ok := false
	for _, s := range payloads {
		if got == s {
			ok = true
			break
		}
	}
	if !ok {
		t.Fatalf("unexpected content: %s", got)
	}
}
