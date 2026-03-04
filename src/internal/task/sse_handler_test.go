package task

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"openclaw-manager/internal/storage"
)

func TestSSETaskEvents(t *testing.T) {
	db := storage.NewTestDB(t)
	_, _ = db.SQL.Exec(`INSERT INTO users(user_id,username,password_hash,role,status,created_at) VALUES(?,?,?,?,?,?)`, "u1", "u1", "x", "Viewer", "active", time.Now().UTC().Format(time.RFC3339))
	r := NewRepository(db.SQL)
	exit := 0
	tk := &Task{TaskID: uuid.NewString(), TaskType: "x", Status: StatusSucceeded, StdoutTail: "a\nb", StderrTail: "e1", ExitCode: &exit, CreatedBy: "u1", CreatedAt: time.Now().UTC()}
	if err := r.Create(tk); err != nil { t.Fatal(err) }

	h := &SSEHandler{Repo: r}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+tk.TaskID+"/events?token=t", nil)
	h.TaskEvents(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expect 200 got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"seq":1`) || !strings.Contains(body, `"type":"done"`) {
		t.Fatalf("unexpected sse body: %s", body)
	}
}

func TestSSETaskEventsUnauthorizedAndNotFound(t *testing.T) {
	db := storage.NewTestDB(t)
	r := NewRepository(db.SQL)
	h := &SSEHandler{Repo: r}

	w1 := httptest.NewRecorder()
	h.TaskEvents(w1, httptest.NewRequest(http.MethodGet, "/api/v1/tasks/x/events", nil))
	if w1.Code != http.StatusUnauthorized { t.Fatalf("expect 401 got %d", w1.Code) }

	w2 := httptest.NewRecorder()
	h.TaskEvents(w2, httptest.NewRequest(http.MethodGet, "/api/v1/tasks/notfound/events?token=t", nil))
	if w2.Code != http.StatusNotFound { t.Fatalf("expect 404 got %d", w2.Code) }
}
