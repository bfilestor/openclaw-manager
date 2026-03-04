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

type dExec struct {
	fn func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func (e dExec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return e.fn(ctx, name, args...)
}

func TestDoctorRunRepairAndParse(t *testing.T) {
	h := NewDoctorHandler(dExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return []byte("node_path=/home/mixi/.nvm/versions/node/v20/bin/node"), nil
	}})
	w1 := httptest.NewRecorder()
	h.Run(w1, httptest.NewRequest(http.MethodPost, "/", nil))
	if w1.Code != http.StatusAccepted || !strings.Contains(w1.Body.String(), "doctor.run") || !strings.Contains(w1.Body.String(), "true") {
		t.Fatalf("run bad code=%d body=%s", w1.Code, w1.Body.String())
	}
	w2 := httptest.NewRecorder()
	h.Repair(w2, httptest.NewRequest(http.MethodPost, "/", nil))
	if w2.Code != http.StatusAccepted || !strings.Contains(w2.Body.String(), "doctor.repair") {
		t.Fatalf("repair bad code=%d body=%s", w2.Code, w2.Body.String())
	}
}

func TestDoctorTimeout(t *testing.T) {
	h := NewDoctorHandler(dExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}})
	h.Timeout = 10 * time.Millisecond
	w := httptest.NewRecorder()
	h.Run(w, httptest.NewRequest(http.MethodPost, "/", nil))
	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("expect 504 got %d", w.Code)
	}
}

func TestDoctorExecError(t *testing.T) {
	h := NewDoctorHandler(dExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return nil, errors.New("boom")
	}})
	w := httptest.NewRecorder()
	h.Run(w, httptest.NewRequest(http.MethodPost, "/", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expect 500 got %d", w.Code)
	}
}
