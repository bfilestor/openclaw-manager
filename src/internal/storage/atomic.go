package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var atomicMu sync.Mutex

// AtomicWriteFile 原子写文件：先写同目录临时文件，再 rename 覆盖。
func AtomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); err != nil {
		return err
	}

	atomicMu.Lock()
	defer atomicMu.Unlock()

	tmp, err := os.CreateTemp(dir, ".tmp_*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }
	defer cleanup()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("atomic rename failed: %w", err)
	}
	return nil
}
