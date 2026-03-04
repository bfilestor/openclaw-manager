package backup

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"openclaw-manager/internal/storage"
)

func TestRestoreAPI(t *testing.T) {
	db := storage.NewTestDB(t)
	home := t.TempDir()
	_ = os.WriteFile(filepath.Join(home, "openclaw.json"), []byte(`{"x":1}`), 0o644)
	s := &Service{DB: db.SQL, BackupHome: t.TempDir(), OpenclawHome: home, ManagerHome: t.TempDir()}
	id, err := s.Create([]string{"openclaw_json"}, "b", "")
	if err != nil { t.Fatal(err) }
	h := &APIHandler{Service: s, DB: db.SQL}

	w1 := httptest.NewRecorder()
	h.RestoreBackup(w1, httptest.NewRequest(http.MethodPost, "/api/v1/backups/"+id+"/restore", strings.NewReader(`{"dry_run":true}`)))
	if w1.Code != http.StatusOK || !strings.Contains(w1.Body.String(), "will_overwrite") {
		t.Fatalf("dry run restore failed code=%d body=%s", w1.Code, w1.Body.String())
	}

	w2 := httptest.NewRecorder()
	h.RestoreBackup(w2, httptest.NewRequest(http.MethodPost, "/api/v1/backups/"+id+"/restore", strings.NewReader(`{"dry_run":false}`)))
	if w2.Code != http.StatusAccepted {
		t.Fatalf("apply restore should be 202 got %d body=%s", w2.Code, w2.Body.String())
	}
}
