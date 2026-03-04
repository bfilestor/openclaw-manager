package storage

import (
	"path/filepath"
	"testing"
)

// NewTestDB 创建一个临时 SQLite 数据库，供单元/集成测试复用。
func NewTestDB(t *testing.T) *DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	t.Cleanup(func() { _ = db.SQL.Close() })
	return db
}
