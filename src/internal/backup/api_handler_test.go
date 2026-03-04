package backup

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"openclaw-manager/internal/storage"
)

func TestBackupAPIFlow(t *testing.T) {
	db := storage.NewTestDB(t)
	s := &Service{DB: db.SQL, BackupHome: t.TempDir(), OpenclawHome: t.TempDir(), ManagerHome: t.TempDir()}
	h := &APIHandler{Service: s, DB: db.SQL}

	w1 := httptest.NewRecorder()
	h.CreateBackup(w1, httptest.NewRequest(http.MethodPost, "/api/v1/backups", strings.NewReader(`{"label":"l1","scope":["openclaw_json"]}`)))
	if w1.Code != http.StatusAccepted { t.Fatalf("create expect 202 got %d body=%s", w1.Code, w1.Body.String()) }

	w2 := httptest.NewRecorder()
	h.ListBackups(w2, httptest.NewRequest(http.MethodGet, "/api/v1/backups", nil))
	if w2.Code != http.StatusOK || !strings.Contains(w2.Body.String(), "backups") {
		t.Fatalf("list failed code=%d body=%s", w2.Code, w2.Body.String())
	}
}
