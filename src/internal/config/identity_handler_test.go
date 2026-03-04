package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"openclaw-manager/internal/agent"
	"openclaw-manager/internal/storage"
)

type iexec struct{}

func (iexec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return []byte(`{"agents":[{"id":"a1","workspace":"/tmp/agent-id1","bindings":[]}]}`), nil
}

func TestIdentityHandlerReadWriteAndRevisions(t *testing.T) {
	_ = os.MkdirAll("/tmp/agent-id1", 0o755)
	_ = os.WriteFile("/tmp/agent-id1/IDENTITY.md", []byte("old"), 0o644)

	db := storage.NewTestDB(t)
	rev := NewRevisionRepository(db.SQL)
	v, _ := storage.NewPathValidator([]string{"/tmp/agent-id1"})
	h := &IdentityHandler{AgentRepo: agent.NewRepository(iexec{}, nil), Revisions: rev, Validator: v}

	w1 := httptest.NewRecorder()
	h.GetIdentity(w1, httptest.NewRequest(http.MethodGet, "/api/v1/agents/a1/identity", nil))
	if w1.Code != http.StatusOK || !strings.Contains(w1.Body.String(), "old") {
		t.Fatalf("get failed code=%d body=%s", w1.Code, w1.Body.String())
	}

	w2 := httptest.NewRecorder()
	h.PutIdentity(w2, httptest.NewRequest(http.MethodPut, "/api/v1/agents/a1/identity", strings.NewReader(`{"content":"new"}`)))
	if w2.Code != http.StatusOK {
		t.Fatalf("put failed code=%d body=%s", w2.Code, w2.Body.String())
	}
	b, _ := os.ReadFile(filepath.Join("/tmp/agent-id1", "IDENTITY.md"))
	if string(b) != "new" {
		t.Fatalf("write failed: %s", string(b))
	}

	w3 := httptest.NewRecorder()
	h.ListIdentityRevisions(w3, httptest.NewRequest(http.MethodGet, "/api/v1/agents/a1/identity/revisions", nil))
	if w3.Code != http.StatusOK || !strings.Contains(w3.Body.String(), "revision_id") {
		t.Fatalf("revision list failed code=%d body=%s", w3.Code, w3.Body.String())
	}
}
