package config

import (
	"fmt"
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

	list, err := rev.List("openclaw_json", "", 10)
	if err != nil || len(list) == 0 {
		t.Fatalf("expect revisions, err=%v len=%d", err, len(list))
	}
	revID := list[0].RevisionID

	restoreReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/config/openclaw/revisions/%s/restore", revID), nil)
	w5 := httptest.NewRecorder()
	h.RestoreRevision(w5, restoreReq)
	if w5.Code != http.StatusOK || !strings.Contains(w5.Body.String(), "restored") {
		t.Fatalf("restore failed code=%d body=%s", w5.Code, w5.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/config/openclaw/revisions/%s", revID), nil)
	w6 := httptest.NewRecorder()
	h.DeleteRevision(w6, deleteReq)
	if w6.Code != http.StatusOK || !strings.Contains(w6.Body.String(), "deleted") {
		t.Fatalf("delete failed code=%d body=%s", w6.Code, w6.Body.String())
	}

	w7 := httptest.NewRecorder()
	h.DeleteRevision(w7, deleteReq)
	if w7.Code != http.StatusNotFound {
		t.Fatalf("delete missing should 404, got code=%d body=%s", w7.Code, w7.Body.String())
	}
}

func TestOpenClawJSONPutSavesPreviousRevisionOnUpdate(t *testing.T) {
	db := storage.NewTestDB(t)
	rev := NewRevisionRepository(db.SQL)
	d := t.TempDir()
	p := filepath.Join(d, "openclaw.json")
	if err := os.WriteFile(p, []byte(`{"old":1}`), 0o644); err != nil {
		t.Fatalf("seed file failed: %v", err)
	}
	v, _ := storage.NewPathValidator([]string{d})
	h := &OpenClawJSONHandler{FilePath: p, Validator: v, Revisions: rev}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"content":"{\"new\":2}"}`))
	h.PutOpenClawJSON(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expect 200 got %d body=%s", w.Code, w.Body.String())
	}

	list, err := rev.List("openclaw_json", "", 10)
	if err != nil {
		t.Fatalf("list revisions failed: %v", err)
	}
	if len(list) == 0 {
		t.Fatalf("expect at least one revision")
	}
	if got := strings.TrimSpace(list[0].Content); got != `{"old":1}` {
		t.Fatalf("expect previous content saved, got: %s", got)
	}
}
