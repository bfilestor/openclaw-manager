package agent

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mexec struct {
	fn func(ctx context.Context, name string, args ...string) ([]byte, error)
}

type mworkspaces struct {
	paths map[string]string
}

func (e mexec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return e.fn(ctx, name, args...)
}

func (w mworkspaces) GetWorkspacePath(ctx context.Context, agentID string) (string, error) {
	if path, ok := w.paths[agentID]; ok {
		return path, nil
	}
	return "", ErrNotFound
}

func TestCreateDeleteAgent(t *testing.T) {
	root := t.TempDir()
	tpl := filepath.Join(root, "workspace-main")
	if err := os.MkdirAll(tpl, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range templateMarkdownFiles {
		if err := os.WriteFile(filepath.Join(tpl, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfgPath := filepath.Join(root, "openclaw.json")
	if err := os.WriteFile(cfgPath, []byte(`{"agents":{"list":[{"id":"main","workspace":"`+tpl+`"}]}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	h := &ManageHandler{
		Exec: mexec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return []byte("ok"), nil
		}},
		Workspaces:       mworkspaces{paths: map[string]string{"main": tpl}},
		OpenClawJSONPath: cfgPath,
	}
	w1 := httptest.NewRecorder()
	h.CreateAgent(w1, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"agent_id":"a1","template_agent_id":"main"}`)))
	if w1.Code != http.StatusOK {
		t.Fatalf("expect 200 got %d", w1.Code)
	}

	created := filepath.Join(root, "workspace-a1")
	if _, err := os.Stat(filepath.Join(created, "AGENTS.md")); err != nil {
		t.Fatalf("expect template copied: %v", err)
	}

	w2 := httptest.NewRecorder()
	h.DeleteAgent(w2, httptest.NewRequest(http.MethodDelete, "/api/v1/agents/a1", nil))
	if w2.Code != http.StatusAccepted {
		t.Fatalf("expect 202 got %d", w2.Code)
	}
}

func TestCreateInvalidAndDeleteErr(t *testing.T) {
	h := &ManageHandler{Exec: mexec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) >= 2 && args[1] == "delete" {
			return nil, errors.New("boom")
		}
		return []byte("ok"), nil
	}}}
	w1 := httptest.NewRecorder()
	h.CreateAgent(w1, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"agent_id":"../x"}`)))
	if w1.Code != http.StatusBadRequest {
		t.Fatalf("expect 400 got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	h.DeleteAgent(w2, httptest.NewRequest(http.MethodDelete, "/api/v1/agents/a1", nil))
	if w2.Code != http.StatusInternalServerError {
		t.Fatalf("expect 500 got %d", w2.Code)
	}
}
