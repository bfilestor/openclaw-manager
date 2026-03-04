package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenInitializesDBAndTables(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "manager.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}
	t.Cleanup(func() { _ = db.SQL.Close() })

	mustTableExists(t, db.SQL, "users")
	mustTableExists(t, db.SQL, "refresh_tokens")
	mustTableExists(t, db.SQL, "token_blacklist")
	mustTableExists(t, db.SQL, "tasks")
	mustTableExists(t, db.SQL, "revisions")
	mustTableExists(t, db.SQL, "backups")
}

func TestOpenIsIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "manager.db")
	db1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("first open failed: %v", err)
	}
	defer db1.SQL.Close()

	if _, err := db1.SQL.Exec(`INSERT INTO users(user_id, username, password_hash, created_at) VALUES('u1','alice','hash','now')`); err != nil {
		t.Fatalf("insert user failed: %v", err)
	}

	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("second open failed: %v", err)
	}
	defer db2.SQL.Close()

	var cnt int
	if err := db2.SQL.QueryRow(`SELECT COUNT(1) FROM users WHERE user_id='u1'`).Scan(&cnt); err != nil {
		t.Fatalf("count user failed: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 user, got %d", cnt)
	}
}

func TestPragmaSettings(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "manager.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}
	defer db.SQL.Close()

	var mode string
	if err := db.SQL.QueryRow(`PRAGMA journal_mode;`).Scan(&mode); err != nil {
		t.Fatalf("read journal_mode failed: %v", err)
	}
	if mode != "wal" {
		t.Fatalf("expected wal, got %s", mode)
	}

	var fk int
	if err := db.SQL.QueryRow(`PRAGMA foreign_keys;`).Scan(&fk); err != nil {
		t.Fatalf("read foreign_keys failed: %v", err)
	}
	if fk != 1 {
		t.Fatalf("expected foreign_keys=1, got %d", fk)
	}
}

func TestForeignKeyConstraint(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "manager.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}
	defer db.SQL.Close()

	_, err = db.SQL.Exec(`INSERT INTO tasks(task_id, task_type, created_by, created_at) VALUES('t1','gateway.start','not-exists','now')`)
	if err == nil {
		t.Fatal("expected foreign key constraint error, got nil")
	}
}

func TestCreateDBWhenDirMissing(t *testing.T) {
	base := filepath.Join(t.TempDir(), "a", "b", "c")
	dbPath := filepath.Join(base, "manager.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}
	defer db.SQL.Close()

	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("db file not created: %v", err)
	}
}

func mustTableExists(t *testing.T, db *sql.DB, table string) {
	t.Helper()
	var cnt int
	if err := db.QueryRow(`SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&cnt); err != nil {
		t.Fatalf("query table %s failed: %v", table, err)
	}
	if cnt != 1 {
		t.Fatalf("table %s not found", table)
	}
}
