package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"openclaw-manager/internal/storage"
)

func TestBackupCreate(t *testing.T) {
	db := storage.NewTestDB(t)
	home := t.TempDir()
	mgr := t.TempDir()
	bk := t.TempDir()
	_ = os.MkdirAll(filepath.Join(home, "skills"), 0o755)
	_ = os.WriteFile(filepath.Join(home, "openclaw.json"), []byte(`{"a":1}`), 0o644)
	_ = os.WriteFile(filepath.Join(home, "skills", "x.txt"), []byte("x"), 0o644)

	_, _ = db.SQL.Exec(`INSERT INTO users(user_id,username,password_hash,role,status,created_at) VALUES(?,?,?,?,?,?)`, "u1", "u1", "x", "Admin", "active", time.Now().UTC().Format(time.RFC3339))

	s := &Service{DB: db.SQL, BackupHome: bk, OpenclawHome: home, ManagerHome: mgr}
	id, err := s.Create([]string{"openclaw_json", "global_skills"}, "test", "u1")
	if err != nil { t.Fatal(err) }
	if _, err := os.Stat(filepath.Join(bk, id+".tar.gz")); err != nil { t.Fatalf("archive missing: %v", err) }
	if _, err := os.Stat(filepath.Join(bk, id+".manifest.json")); err != nil { t.Fatalf("manifest missing: %v", err) }
}
