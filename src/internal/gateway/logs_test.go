package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type logExec struct{ out []byte; err error }

func (e logExec) Run(ctx context.Context, name string, args ...string) ([]byte, error) { return e.out, e.err }

func TestGetLogsFileAndValidation(t *testing.T) {
	base := "/tmp/openclaw"
	_ = os.MkdirAll(base, 0o755)
	f := filepath.Join(base, "openclaw-2099-01-01.log")
	_ = os.WriteFile(f, []byte("l1\nl2\nl3\n"), 0o644)

	h := NewLogsHandler(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/gateway/logs?lines=2&source=file", nil)
	h.GetLogs(w, r)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "l2") || !strings.Contains(w.Body.String(), "l3") {
		t.Fatalf("unexpected resp code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestGetLogsJournaldAndLimits(t *testing.T) {
	h := NewLogsHandler(logExec{out: []byte("a\nb\n")})

	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest(http.MethodGet, "/api/v1/gateway/logs?source=journald&lines=1001", nil)
	h.GetLogs(w1, r1)
	if w1.Code != http.StatusOK || !strings.Contains(w1.Body.String(), "a") {
		t.Fatalf("unexpected journald resp code=%d body=%s", w1.Code, w1.Body.String())
	}

	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodGet, "/api/v1/gateway/logs?lines=-1", nil)
	h.GetLogs(w2, r2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expect 400 got %d", w2.Code)
	}
}
