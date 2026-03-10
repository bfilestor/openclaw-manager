package storage

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

var ErrMigration = errors.New("migration failed")

type DB struct {
	SQL *sql.DB
}

func Open(dbPath string) (*DB, error) {
	if strings.TrimSpace(dbPath) == "" {
		return nil, fmt.Errorf("empty db path")
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// MVP 阶段保持单连接，确保 PRAGMA 行为稳定。
	db.SetMaxOpenConns(1)

	if err := applyPragmas(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := runMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &DB{SQL: db}, nil
}

func applyPragmas(db *sql.DB) error {
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		return err
	}
	if _, err := db.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		return err
	}
	return nil
}

func runMigrations(db *sql.DB) error {
	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  applied_at TEXT NOT NULL
);`); err != nil {
		return err
	}

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return err
	}

	versions := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		versions = append(versions, e.Name())
	}
	sort.Strings(versions)

	for _, v := range versions {
		applied, err := isApplied(db, v)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		sqlBytes, err := migrationFS.ReadFile(filepath.Join("migrations", v))
		if err != nil {
			return err
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("%w: %s: %v", ErrMigration, v, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES(?, datetime('now'));`, v); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

func isApplied(db *sql.DB, version string) (bool, error) {
	var cnt int
	if err := db.QueryRow(`SELECT COUNT(1) FROM schema_migrations WHERE version = ?`, version).Scan(&cnt); err != nil {
		return false, err
	}
	return cnt > 0, nil
}
