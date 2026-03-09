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
	_ = os.MkdirAll(filepath.Join(home, "agents", "main", "agent"), 0o755)
	_ = os.WriteFile(filepath.Join(home, "openclaw.json"), []byte(`{"a":1}`), 0o644)
	_ = os.WriteFile(filepath.Join(home, "models.json"), []byte(`{"default":"gpt"}`), 0o644)
	_ = os.WriteFile(filepath.Join(home, "agents", "main", "agent", "models.json"), []byte(`{"agent":"main"}`), 0o644)
	_ = os.WriteFile(filepath.Join(home, "skills", "x.txt"), []byte("x"), 0o644)

	_, _ = db.SQL.Exec(`INSERT INTO users(user_id,username,password_hash,role,status,created_at) VALUES(?,?,?,?,?,?)`, "u1", "u1", "x", "Admin", "active", time.Now().UTC().Format(time.RFC3339))

	s := &Service{DB: db.SQL, BackupHome: bk, OpenclawHome: home, ManagerHome: mgr}
	id, err := s.Create([]string{"openclaw_json", "global_skills"}, "test", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(bk, id+".tar.gz")); err != nil {
		t.Fatalf("archive missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(bk, id+".manifest.json")); err != nil {
		t.Fatalf("manifest missing: %v", err)
	}
}

func TestResolveScopeIncludesCoreConfigFiles(t *testing.T) {
	db := storage.NewTestDB(t)
	home := t.TempDir()
	mgr := t.TempDir()
	s := &Service{DB: db.SQL, OpenclawHome: home, ManagerHome: mgr}

	paths := s.resolveScope([]string{"openclaw_json"})
	want := map[string]bool{
		filepath.Join(home, "openclaw.json"):                         false,
		filepath.Join(home, "models.json"):                           false,
		filepath.Join(home, "agents", "main", "agent", "models.json"): false,
	}
	for _, p := range paths {
		if _, ok := want[p]; ok {
			want[p] = true
		}
	}
	for p, ok := range want {
		if !ok {
			t.Fatalf("missing core config path in scope: %s, got=%v", p, paths)
		}
	}
}

func TestResolveScopeIncludesMultiAgentWorkspaces(t *testing.T) {
	db := storage.NewTestDB(t)
	home := t.TempDir()
	mgr := t.TempDir()

	cfg := `{
		"agents": {
			"defaults": {
				"workspace": "` + home + `/workspace"
			},
			"list": [
				{"id":"main"},
				{"id":"xcoder"},
				{"id":"pm","workspace":"` + home + `/workspace-pm"}
			]
		}
	}`
	if err := os.WriteFile(filepath.Join(home, "openclaw.json"), []byte(cfg), 0o644); err != nil {
		t.Fatalf("write openclaw.json: %v", err)
	}

	s := &Service{DB: db.SQL, OpenclawHome: home, ManagerHome: mgr}
	paths := s.resolveScope([]string{"workspaces"})

	want := map[string]bool{
		filepath.Join(home, "workspace"):        false,
		filepath.Join(home, "workspace-xcoder"): false,
		filepath.Join(home, "workspace-pm"):     false,
	}
	for _, p := range paths {
		if _, ok := want[p]; ok {
			want[p] = true
		}
	}
	for p, ok := range want {
		if !ok {
			t.Fatalf("missing workspace path in scope: %s, got=%v", p, paths)
		}
	}
}
