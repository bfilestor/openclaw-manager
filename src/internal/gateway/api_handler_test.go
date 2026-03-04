package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type apiExec struct {
	fn func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func (e apiExec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return e.fn(ctx, name, args...)
}

func TestGatewayStatusAndActions(t *testing.T) {
	ex := apiExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		s := strings.Join(append([]string{name}, args...), " ")
		switch {
		case strings.HasPrefix(s, "systemctl --user show"):
			return []byte("ActiveState=active\nSubState=running\n"), nil
		case s == "openclaw gateway status --deep":
			return []byte("bind=127.0.0.1:18789\nnode_path=/usr/bin/node\n"), nil
		case strings.HasPrefix(s, "systemctl --user start") || strings.HasPrefix(s, "systemctl --user stop") || strings.HasPrefix(s, "systemctl --user restart"):
			return []byte(""), nil
		default:
			return nil, errors.New("unexpected cmd")
		}
	}}
	h := &APIHandler{Service: NewSystemctlService(ex)}

	w1 := httptest.NewRecorder()
	h.Status(w1, httptest.NewRequest(http.MethodGet, "/", nil))
	if w1.Code != http.StatusOK || !strings.Contains(w1.Body.String(), "bind_addr") {
		t.Fatalf("status failed code=%d body=%s", w1.Code, w1.Body.String())
	}

	w2 := httptest.NewRecorder()
	h.Start(w2, httptest.NewRequest(http.MethodPost, "/", nil))
	if w2.Code != http.StatusAccepted {
		t.Fatalf("start failed code=%d body=%s", w2.Code, w2.Body.String())
	}
}

func TestGatewayConflict409(t *testing.T) {
	ex := apiExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "systemctl" {
			time.Sleep(50 * time.Millisecond)
			return []byte(""), nil
		}
		return []byte(""), nil
	}}
	h := &APIHandler{Service: NewSystemctlService(ex)}

	done := make(chan struct{})
	go func() {
		w := httptest.NewRecorder()
		h.Start(w, httptest.NewRequest(http.MethodPost, "/", nil))
		close(done)
	}()
	time.Sleep(10 * time.Millisecond)

	w2 := httptest.NewRecorder()
	h.Start(w2, httptest.NewRequest(http.MethodPost, "/", nil))
	if w2.Code != http.StatusConflict {
		t.Fatalf("expect 409 got %d body=%s", w2.Code, w2.Body.String())
	}
	<-done
}
