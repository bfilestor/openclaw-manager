package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	s := New("127.0.0.1:0", "")
	r := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	s.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestNotFoundJSON(t *testing.T) {
	s := New("127.0.0.1:0", "")
	r := httptest.NewRequest(http.MethodGet, "/api/v1/not-exists", nil)
	w := httptest.NewRecorder()

	s.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"code":"NOT_FOUND"`) {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestCORSHeader(t *testing.T) {
	s := New("127.0.0.1:0", "")
	r := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	r.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()

	s.Handler().ServeHTTP(w, r)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("missing cors header: %v", w.Header())
	}
}

func TestStaticIndex(t *testing.T) {
	dir := t.TempDir()
	idx := filepath.Join(dir, "index.html")
	if err := os.WriteFile(idx, []byte("hello-spa"), 0o644); err != nil {
		t.Fatalf("write index failed: %v", err)
	}

	s := New("127.0.0.1:0", dir)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "hello-spa") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	h := recoverMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "INTERNAL_ERROR") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestGracefulShutdown(t *testing.T) {
	s := New("127.0.0.1:0", "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("get health failed: %v", err)
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
}
