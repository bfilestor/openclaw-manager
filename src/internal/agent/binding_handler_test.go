package agent

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type bExec struct {
	fn func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func (e bExec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return e.fn(ctx, name, args...)
}

func TestBindingListAndApply(t *testing.T) {
	h := NewBindingHandler(bExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		s := strings.Join(append([]string{name}, args...), " ")
		if strings.Contains(s, "list") {
			return []byte(`{"agents":[{"id":"a1","bindings":[{"peer":"p1"}]}]}`), nil
		}
		return []byte("ok"), nil
	}})

	w1 := httptest.NewRecorder()
	h.ListBindings(w1, httptest.NewRequest(http.MethodGet, "/", nil))
	if w1.Code != http.StatusOK || !strings.Contains(w1.Body.String(), "\"agents\"") || !strings.Contains(w1.Body.String(), "bindings") {
		t.Fatalf("list failed code=%d body=%s", w1.Code, w1.Body.String())
	}

	w2 := httptest.NewRecorder()
	body := `{"add":[{"agent_id":"a1","channel":"telegram","account":"default","peer":"p1"}],"remove":[{"agent_id":"a1","channel":"telegram","account":"default","peer":"p2"}]}`
	h.ApplyBindings(w2, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)))
	if w2.Code != http.StatusAccepted || !strings.Contains(w2.Body.String(), "SUCCEEDED") {
		t.Fatalf("apply failed code=%d body=%s", w2.Code, w2.Body.String())
	}
}

func TestBindingApplyPartialFailed(t *testing.T) {
	h := NewBindingHandler(bExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 1 && args[1] == "bind" {
			return nil, errors.New("bind fail")
		}
		return []byte("ok"), nil
	}})
	w := httptest.NewRecorder()
	body := `{"add":[{"agent_id":"a1","channel":"telegram","account":"default","peer":"p1"}]}`
	h.ApplyBindings(w, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)))
	if w.Code != http.StatusAccepted || !strings.Contains(w.Body.String(), "FAILED") {
		t.Fatalf("expect failed status code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestBindingListNewCLIJSONShape(t *testing.T) {
	h := NewBindingHandler(bExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return []byte(`[{"id":"a1","bindings":2},{"id":"a2","bindings":[]}]`), nil
	}})

	w := httptest.NewRecorder()
	h.ListBindings(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected code=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"id":"a1"`) || !strings.Contains(w.Body.String(), `"count":2`) {
		t.Fatalf("unexpected normalized response: %s", w.Body.String())
	}
}
