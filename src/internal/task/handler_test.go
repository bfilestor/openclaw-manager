package task

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"openclaw-manager/internal/auth"
	"openclaw-manager/internal/storage"
	"openclaw-manager/internal/user"
)

func setupTaskHandler(t *testing.T) (*Handler, *Repository, string, string) {
	db := storage.NewTestDB(t)
	r := NewRepository(db.SQL)
	u1 := "u1"
	u2 := "u2"
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.SQL.Exec(`INSERT INTO users(user_id,username,password_hash,role,status,created_at) VALUES(?,?,?,?,?,?)`, u1, "u1", "x", "Viewer", "active", now); err != nil { t.Fatal(err) }
	if _, err := db.SQL.Exec(`INSERT INTO users(user_id,username,password_hash,role,status,created_at) VALUES(?,?,?,?,?,?)`, u2, "u2", "x", "Viewer", "active", now); err != nil { t.Fatal(err) }
	if err := r.Create(&Task{TaskID: uuid.NewString(), TaskType: "a", Status: StatusPending, CreatedBy: u1, CreatedAt: time.Now().UTC()}); err != nil { t.Fatal(err) }
	if err := r.Create(&Task{TaskID: uuid.NewString(), TaskType: "b", Status: StatusRunning, CreatedBy: u2, CreatedAt: time.Now().UTC().Add(time.Second)}); err != nil { t.Fatal(err) }
	return &Handler{Repo: r}, r, u1, u2
}

func withUC(r *http.Request, uid string, role user.Role) *http.Request {
	ctx := auth.WithUserContext(r.Context(), &auth.UserContext{UserID: uid, Role: role})
	return r.WithContext(ctx)
}

func TestListTasksViewerOnlyOwnAndAdminAll(t *testing.T) {
	h, _, u1, _ := setupTaskHandler(t)

	w1 := httptest.NewRecorder()
	h.ListTasks(w1, withUC(httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil), u1, user.RoleViewer))
	if w1.Code != http.StatusOK || strings.Count(w1.Body.String(), "task_id") != 1 {
		t.Fatalf("viewer should only see own tasks, code=%d body=%s", w1.Code, w1.Body.String())
	}

	w2 := httptest.NewRecorder()
	h.ListTasks(w2, withUC(httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil), "admin", user.RoleAdmin))
	if w2.Code != http.StatusOK || strings.Count(w2.Body.String(), "task_id") != 2 {
		t.Fatalf("admin should see all tasks, code=%d body=%s", w2.Code, w2.Body.String())
	}
}

func TestGetTaskAndCancel(t *testing.T) {
	h, r, u1, _ := setupTaskHandler(t)
	list, _, _ := r.List(ListFilter{CreatedBy: u1})
	taskID := list[0].TaskID

	w1 := httptest.NewRecorder()
	h.GetTask(w1, withUC(httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+taskID, nil), u1, user.RoleViewer))
	if w1.Code != http.StatusOK {
		t.Fatalf("expect 200 got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	h.CancelTask(w2, withUC(httptest.NewRequest(http.MethodPost, "/api/v1/tasks/"+taskID+"/cancel", nil), u1, user.RoleViewer))
	if w2.Code != http.StatusForbidden {
		t.Fatalf("viewer cancel expect 403 got %d", w2.Code)
	}

	w2b := httptest.NewRecorder()
	h.CancelTask(w2b, withUC(httptest.NewRequest(http.MethodPost, "/api/v1/tasks/cancel", nil), "op", user.RoleOperator))
	if w2b.Code != http.StatusBadRequest {
		t.Fatalf("invalid cancel path expect 400 got %d", w2b.Code)
	}

	w3 := httptest.NewRecorder()
	h.CancelTask(w3, withUC(httptest.NewRequest(http.MethodPost, "/api/v1/tasks/"+taskID+"/cancel", nil), "op", user.RoleOperator))
	if w3.Code != http.StatusOK {
		t.Fatalf("operator cancel expect 200 got %d body=%s", w3.Code, w3.Body.String())
	}

	tk, _ := r.FindByID(taskID)
	if tk.Status != StatusCanceled {
		t.Fatalf("expect canceled got %s", tk.Status)
	}
}
