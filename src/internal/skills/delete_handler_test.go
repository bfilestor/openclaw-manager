package skills

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"openclaw-manager/internal/agent"
	"openclaw-manager/internal/storage"
)

type delexec struct{}

func (delexec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return []byte(`{"agents":[{"id":"a1","workspace":"/tmp/agent-del","bindings":[]}]}`), nil
}

func TestDeleteSkillGlobalAndAgent(t *testing.T) {
	g := "/tmp/skills-del"
	_ = os.MkdirAll(filepath.Join(g, "s1"), 0o755)
	_ = os.MkdirAll("/tmp/agent-del/skills/s2", 0o755)

	v, _ := storage.NewPathValidator([]string{g, "/tmp/agent-del"})
	h := &DeleteHandler{AgentRepo: agent.NewRepository(delexec{}, nil), GlobalDir: g, Validator: v}

	w1 := httptest.NewRecorder()
	h.DeleteSkill(w1, httptest.NewRequest(http.MethodDelete, "/api/v1/skills/s1?scope=global", nil))
	if w1.Code != http.StatusAccepted {
		t.Fatalf("expect 202 got %d", w1.Code)
	}
	if _, err := os.Stat(filepath.Join(g, "s1")); !os.IsNotExist(err) {
		t.Fatalf("global skill not removed")
	}

	w2 := httptest.NewRecorder()
	h.DeleteSkill(w2, httptest.NewRequest(http.MethodDelete, "/api/v1/skills/s2?scope=agent&agent_id=a1", nil))
	if w2.Code != http.StatusAccepted {
		t.Fatalf("expect 202 got %d", w2.Code)
	}

	w3 := httptest.NewRecorder()
	h.DeleteSkill(w3, httptest.NewRequest(http.MethodDelete, "/api/v1/skills/sx?scope=agent", nil))
	if w3.Code != http.StatusBadRequest {
		t.Fatalf("expect 400 got %d", w3.Code)
	}
}
