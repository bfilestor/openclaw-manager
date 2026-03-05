package skills

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"openclaw-manager/internal/agent"
)

type aexec struct{}

func (aexec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return []byte(`{"agents":[{"id":"a1","workspace":"/tmp/agent-a1","bindings":[]}]}`), nil
}

func TestListSkillsGlobalAndAgent(t *testing.T) {
	g := "/tmp/skills-global"
	_ = os.MkdirAll(filepath.Join(g, "s1"), 0o755)
	_ = os.WriteFile(filepath.Join(g, "s1", "README.md"), []byte("x"), 0o644)
	_ = os.MkdirAll(filepath.Join(g, "s-skillmd"), 0o755)
	_ = os.WriteFile(filepath.Join(g, "s-skillmd", "SKILL.md"), []byte("x"), 0o644)

	_ = os.MkdirAll("/tmp/agent-a1/skills/s2", 0o755)
	_ = os.WriteFile("/tmp/agent-a1/skills/s2/skill.json", []byte("{}"), 0o644)

	r := agent.NewRepository(aexec{}, nil)
	h := &Handler{AgentRepo: r, GlobalDir: g}

	w1 := httptest.NewRecorder()
	h.ListSkills(w1, httptest.NewRequest(http.MethodGet, "/api/v1/skills?scope=global", nil))
	if w1.Code != http.StatusOK || !strings.Contains(w1.Body.String(), "s1") || !strings.Contains(w1.Body.String(), "s-skillmd") {
		t.Fatalf("bad global resp code=%d body=%s", w1.Code, w1.Body.String())
	}

	w2 := httptest.NewRecorder()
	h.ListSkills(w2, httptest.NewRequest(http.MethodGet, "/api/v1/skills?scope=agent&agent_id=a1", nil))
	if w2.Code != http.StatusOK || !strings.Contains(w2.Body.String(), "s2") {
		t.Fatalf("bad agent resp code=%d body=%s", w2.Code, w2.Body.String())
	}

	w3 := httptest.NewRecorder()
	h.ListSkills(w3, httptest.NewRequest(http.MethodGet, "/api/v1/skills?scope=agent", nil))
	if w3.Code != http.StatusBadRequest {
		t.Fatalf("expect 400 got %d", w3.Code)
	}
}

func TestScanSkillsManyDedup(t *testing.T) {
	b1 := "/tmp/skills-b1"
	b2 := "/tmp/skills-b2"
	_ = os.MkdirAll(filepath.Join(b1, "dup"), 0o755)
	_ = os.WriteFile(filepath.Join(b1, "dup", "SKILL.md"), []byte("x"), 0o644)
	_ = os.MkdirAll(filepath.Join(b2, "dup"), 0o755)
	_ = os.WriteFile(filepath.Join(b2, "dup", "README.md"), []byte("x"), 0o644)
	_ = os.MkdirAll(filepath.Join(b2, "other"), 0o755)

	items, err := scanSkillsMany([]string{b1, b2}, "global", "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected deduped 2 skills, got %d: %+v", len(items), items)
	}
}
