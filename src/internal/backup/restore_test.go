package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"openclaw-manager/internal/storage"
)

func TestRestoreDryRunAndApply(t *testing.T) {
	db := storage.NewTestDB(t)
	home := t.TempDir()
	mgr := t.TempDir()
	bk := t.TempDir()
	_ = os.WriteFile(filepath.Join(home, "openclaw.json"), []byte(`{"x":1}`), 0o644)
	_, _ = db.SQL.Exec(`INSERT INTO users(user_id,username,password_hash,role,status,created_at) VALUES(?,?,?,?,?,?)`, "u1", "u1", "x", "Admin", "active", time.Now().UTC().Format(time.RFC3339))
	s := &Service{DB: db.SQL, BackupHome: bk, OpenclawHome: home, ManagerHome: mgr}
	id, err := s.Create([]string{"openclaw_json"}, "b1", "u1")
	if err != nil { t.Fatal(err) }

	_ = os.WriteFile(filepath.Join(home, "openclaw.json"), []byte(`{"x":999}`), 0o644)
	r1, err := s.Restore(id, true, false, "u1")
	if err != nil || !r1.DryRun || len(r1.WillOverwrite) == 0 { t.Fatalf("dry run failed r=%+v err=%v", r1, err) }
	b, _ := os.ReadFile(filepath.Join(home, "openclaw.json"))
	if string(b) != `{"x":999}` { t.Fatalf("dry run should not modify file") }
}

func TestAllowedRestoreTarget(t *testing.T) {
	if !isAllowedRestoreTarget("/tmp/a/b", []string{"/tmp/a"}) { t.Fatalf("expected allowed") }
	if isAllowedRestoreTarget("/etc/passwd", []string{"/tmp/a"}) { t.Fatalf("expected denied") }
}
