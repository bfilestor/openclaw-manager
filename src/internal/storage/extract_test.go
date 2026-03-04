package storage

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func makeZip(t *testing.T, entries map[string][]byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "a.zip")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for n, b := range entries {
		w, err := zw.Create(n)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(b); err != nil {
			t.Fatal(err)
		}
	}
	_ = zw.Close()
	_ = f.Close()
	return p
}

func makeTarGz(t *testing.T, entries map[string][]byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "a.tar.gz")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for n, b := range entries {
		h := &tar.Header{Name: n, Mode: 0o644, Size: int64(len(b))}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(b); err != nil {
			t.Fatal(err)
		}
	}
	_ = tw.Close()
	_ = gz.Close()
	_ = f.Close()
	return p
}

func TestSafeExtractZipOK(t *testing.T) {
	z := makeZip(t, map[string][]byte{"x/a.txt": []byte("ok")})
	d := t.TempDir()
	if err := SafeExtract(z, d); err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(d, "x", "a.txt")); err != nil {
		t.Fatalf("file not extracted: %v", err)
	}
}

func TestSafeExtractZipSlip(t *testing.T) {
	z := makeZip(t, map[string][]byte{"../../etc/passwd": []byte("bad")})
	err := SafeExtract(z, t.TempDir())
	if !errors.Is(err, ErrZipSlip) {
		t.Fatalf("expected ErrZipSlip, got %v", err)
	}
}

func TestSafeExtractAbsolutePath(t *testing.T) {
	z := makeZip(t, map[string][]byte{"/etc/passwd": []byte("bad")})
	err := SafeExtract(z, t.TempDir())
	if !errors.Is(err, ErrZipSlip) {
		t.Fatalf("expected ErrZipSlip, got %v", err)
	}
}

func TestSafeExtractTarGzOK(t *testing.T) {
	tg := makeTarGz(t, map[string][]byte{"a/b.txt": []byte("ok")})
	d := t.TempDir()
	if err := SafeExtract(tg, d); err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(d, "a", "b.txt")); err != nil {
		t.Fatalf("file not extracted: %v", err)
	}
}

func TestSafeExtractTarGzSlip(t *testing.T) {
	tg := makeTarGz(t, map[string][]byte{"../evil": []byte("bad")})
	err := SafeExtract(tg, t.TempDir())
	if !errors.Is(err, ErrZipSlip) {
		t.Fatalf("expected ErrZipSlip, got %v", err)
	}
}

func TestSafeExtractFileTooLarge(t *testing.T) {
	big := make([]byte, maxSingleFile+1)
	z := makeZip(t, map[string][]byte{"big.bin": big})
	err := SafeExtract(z, t.TempDir())
	if !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestSafeExtractEmptyArchive(t *testing.T) {
	z := makeZip(t, map[string][]byte{})
	if err := SafeExtract(z, t.TempDir()); err != nil {
		t.Fatalf("empty archive should pass, got %v", err)
	}
}
