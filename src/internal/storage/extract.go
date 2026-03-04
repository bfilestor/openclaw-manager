package storage

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrZipSlip        = errors.New("zip slip detected")
	ErrFileTooLarge   = errors.New("file too large")
	ErrExtractTooLarge = errors.New("extract too large")
)

const (
	maxSingleFile = 50 * 1024 * 1024
	maxTotalSize  = 200 * 1024 * 1024
)

func SafeExtract(archivePath, destBase string) error {
	if err := os.MkdirAll(destBase, 0o755); err != nil {
		return err
	}
	lower := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return extractZip(archivePath, destBase)
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		return extractTarGz(archivePath, destBase)
	default:
		return fmt.Errorf("unsupported archive format")
	}
}

func secureJoin(base, name string) (string, error) {
	if filepath.IsAbs(name) || strings.Contains(name, "..") {
		return "", ErrZipSlip
	}
	clean := filepath.Clean(name)
	target := filepath.Join(base, clean)
	rel, err := filepath.Rel(base, target)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", ErrZipSlip
	}
	return target, nil
}

func extractZip(path, dest string) error {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer zr.Close()

	var total int64
	for _, f := range zr.File {
		target, err := secureJoin(dest, f.Name)
		if err != nil {
			return err
		}
		if f.FileInfo().Mode()&os.ModeSymlink != 0 {
			return ErrZipSlip
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if f.UncompressedSize64 > maxSingleFile {
			return ErrFileTooLarge
		}
		total += int64(f.UncompressedSize64)
		if total > maxTotalSize {
			return ErrExtractTooLarge
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		if err := writeLimitedFile(target, rc, int64(f.UncompressedSize64)); err != nil {
			rc.Close()
			return err
		}
		rc.Close()
	}
	return nil
}

func extractTarGz(path, dest string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	var total int64
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		target, err := secureJoin(dest, hdr.Name)
		if err != nil {
			return err
		}
		if hdr.Typeflag == tar.TypeSymlink || hdr.Typeflag == tar.TypeLink {
			return ErrZipSlip
		}
		if hdr.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}

		if hdr.Size > maxSingleFile {
			return ErrFileTooLarge
		}
		total += hdr.Size
		if total > maxTotalSize {
			return ErrExtractTooLarge
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := writeLimitedFile(target, tr, hdr.Size); err != nil {
			return err
		}
	}
	return nil
}

func writeLimitedFile(target string, r io.Reader, size int64) error {
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	written, err := io.CopyN(out, r, size)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if written > maxSingleFile {
		return ErrFileTooLarge
	}
	return nil
}
