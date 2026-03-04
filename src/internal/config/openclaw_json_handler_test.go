package config

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"openclaw-manager/internal/storage"
)

func TestOpenClawJSONCRUDAndRevisions(t *testing.T) {
	db := storage.NewTestDB(t)
	rev := NewRevisionRepository(db.SQL)
	d := t.TempDir()
	p := filepath.Join(d, "openclaw.json")
	v, _ := storage.NewPathValidator([]string{d})
	h := &OpenClawJSONHandler{FilePath: p, Validator: v, Revisions: rev}

	// get not exists
	w1 := httptest.NewRecorder()
	h.GetOpenClawJSON(w1, httptest.NewRequest(http.MethodGet, "/", nil))
	if w1.Code != http.StatusOK || !strings.Contains(w1.Body.String(), "null") {
		t.Fatalf("unexpected get-empty code=%d body=%s", w1.Code, w1.Body.String())
	}

	// put invalid
	w2 := httptest.NewRecorder()
	h.PutOpenClawJSON(w2, httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"content":"{"}`)))
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expect 400 got %d", w2.Code)
	}

	// put valid
	r3 := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"content":"{\"a\":1}"}`))
	w3 := httptest.NewRecorder()
	h.PutOpenClawJSON(w3, r3)
	if w3.Code != http.StatusOK {
		t.Fatalf("expect 200 got %d body=%s", w3.Code, w3.Body.String())
	}
	b, _ := os.ReadFile(p)
	if string(b) != `{"a":1}` {
		t.Fatalf("write failed: %s", string(b))
	}

	w4 := httptest.NewRecorder()
	h.ListRevisions(w4, httptest.NewRequest(http.MethodGet, "/", nil))
	if w4.Code != http.StatusOK || !strings.Contains(w4.Body.String(), "revision_id") {
		t.Fatalf("list revision failed code=%d body=%s", w4.Code, w4.Body.String())
	}
}
