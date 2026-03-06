package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeWorkspaceResolver struct {
	path string
	err  error
}

func (r fakeWorkspaceResolver) GetWorkspacePath(ctx context.Context, agentID string) (string, error) {
	if r.err != nil {
		return "", r.err
	}
	return r.path, nil
}

func TestMigrateWorkspaceSuccess(t *testing.T) {
	base := t.TempDir()
	oldDir := filepath.Join(base, "workspace-old")
	newDir := filepath.Join(base, "workspace-new")
	if err := os.MkdirAll(filepath.Join(oldDir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldDir, "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldDir, "sub", "b.txt"), []byte("world"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(base, "openclaw.json")
	cfg := `{
	  "agents": {
	    "defaults": {"workspace": "` + oldDir + `"},
	    "list": [{"id":"a1","workspace":"` + oldDir + `"}]
	  }
	}`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	restarted := false
	h := &ManageHandler{
		Exec: mexec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			if name == "openclaw" && len(args) == 2 && args[0] == "gateway" && args[1] == "restart" {
				restarted = true
			}
			return []byte("ok"), nil
		}},
		Workspaces:       fakeWorkspaceResolver{path: oldDir},
		OpenClawJSONPath: cfgPath,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/a1/workspace/migrate", strings.NewReader(`{"new_workspace_path":"`+newDir+`"}`))
	w := httptest.NewRecorder()
	h.MigrateWorkspace(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expect 200 got %d body=%s", w.Code, w.Body.String())
	}
	if !restarted {
		t.Fatal("gateway restart not called")
	}
	if _, err := os.Stat(filepath.Join(newDir, "a.txt")); err != nil {
		t.Fatalf("moved file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(newDir, "sub", "b.txt")); err != nil {
		t.Fatalf("moved nested file missing: %v", err)
	}

	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		t.Fatal(err)
	}
	agents := root["agents"].(map[string]any)
	list := agents["list"].([]any)
	item := list[0].(map[string]any)
	if got := strings.TrimSpace(item["workspace"].(string)); got != filepath.Clean(newDir) {
		t.Fatalf("workspace not updated, got=%s want=%s", got, filepath.Clean(newDir))
	}
}

func TestMigrateWorkspaceConflict(t *testing.T) {
	base := t.TempDir()
	oldDir := filepath.Join(base, "workspace-old")
	newDir := filepath.Join(base, "workspace-new")
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldDir, "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "a.txt"), []byte("exists"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(base, "openclaw.json")
	if err := os.WriteFile(cfgPath, []byte(`{"agents":{"list":[{"id":"a1","workspace":"`+oldDir+`"}]}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	h := &ManageHandler{
		Exec:             mexec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) { return []byte("ok"), nil }},
		Workspaces:       fakeWorkspaceResolver{path: oldDir},
		OpenClawJSONPath: cfgPath,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/a1/workspace/migrate", strings.NewReader(`{"new_workspace_path":"`+newDir+`"}`))
	w := httptest.NewRecorder()
	h.MigrateWorkspace(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("expect 409 got %d body=%s", w.Code, w.Body.String())
	}
}
