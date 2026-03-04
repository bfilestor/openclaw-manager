package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type hExec struct{ out []byte }

func (e *hExec) Run(ctx context.Context, name string, args ...string) ([]byte, error) { return e.out, nil }

func TestAgentHandlers(t *testing.T) {
	ex := &hExec{out: []byte(`{"agents":[{"id":"a1","workspace":"/tmp/w1","bindings":[]}]}`)}
	r := NewRepository(ex, nil)
	h := &Handler{Repo: r}

	w1 := httptest.NewRecorder()
	h.ListAgents(w1, httptest.NewRequest(http.MethodGet, "/api/v1/agents", nil))
	if w1.Code != http.StatusOK {
		t.Fatalf("expect 200 got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	h.GetAgent(w2, httptest.NewRequest(http.MethodGet, "/api/v1/agents/a1", nil))
	if w2.Code != http.StatusOK {
		t.Fatalf("expect 200 got %d", w2.Code)
	}

	w3 := httptest.NewRecorder()
	h.GetAgent(w3, httptest.NewRequest(http.MethodGet, "/api/v1/agents/x", nil))
	if w3.Code != http.StatusNotFound {
		t.Fatalf("expect 404 got %d", w3.Code)
	}
}
