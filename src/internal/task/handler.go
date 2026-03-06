package task

import (
	"encoding/json"
	"net/http"
	"strings"

	"openclaw-manager/internal/auth"
	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/user"
)

type Handler struct {
	Repo *Repository
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	uc, ok := auth.GetUserContext(r.Context())
	if !ok {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	f := ListFilter{
		TaskType: r.URL.Query().Get("type"),
	}
	if s := strings.TrimSpace(r.URL.Query().Get("status")); s != "" {
		f.Status = Status(strings.ToUpper(s))
	}
	if uc.Role != user.RoleAdmin {
		f.CreatedBy = uc.UserID
	}
	list, total, err := h.Repo.List(f)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"tasks": list, "total": total})
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	uc, ok := auth.GetUserContext(r.Context())
	if !ok {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	taskID := lastPart(r.URL.Path)
	t, err := h.Repo.FindByID(taskID)
	if err != nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_FOUND", Message: "task not found", StatusCode: http.StatusNotFound})
		return
	}
	if uc.Role != user.RoleAdmin && t.CreatedBy != uc.UserID {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_FOUND", Message: "task not found", StatusCode: http.StatusNotFound})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(t)
}

func (h *Handler) CancelTask(w http.ResponseWriter, r *http.Request) {
	uc, ok := auth.GetUserContext(r.Context())
	if !ok {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	if uc.Role != user.RoleOperator && uc.Role != user.RoleAdmin {
		middleware.WriteAppError(w, middleware.NewForbidden(string(user.RoleOperator)))
		return
	}
	taskID := taskIDFromCancelPath(r.URL.Path)
	if taskID == "" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"task_id": "required"}))
		return
	}
	t, err := h.Repo.FindByID(taskID)
	if err != nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_FOUND", Message: "task not found", StatusCode: http.StatusNotFound})
		return
	}
	if t.Status != StatusPending {
		middleware.WriteAppError(w, &middleware.AppError{Code: "BAD_REQUEST", Message: "only pending task can be canceled", StatusCode: http.StatusBadRequest})
		return
	}
	if err := h.Repo.UpdateStatus(taskID, StatusCanceled); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"canceled"}`))
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	uc, ok := auth.GetUserContext(r.Context())
	if !ok {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	taskID := lastPart(r.URL.Path)
	if strings.TrimSpace(taskID) == "" || taskID == "tasks" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"task_id": "required"}))
		return
	}
	t, err := h.Repo.FindByID(taskID)
	if err != nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_FOUND", Message: "task not found", StatusCode: http.StatusNotFound})
		return
	}
	if uc.Role != user.RoleAdmin && t.CreatedBy != uc.UserID {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_FOUND", Message: "task not found", StatusCode: http.StatusNotFound})
		return
	}
	if err := h.Repo.Delete(taskID); err != nil {
		if err == ErrNotFound {
			middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_FOUND", Message: "task not found", StatusCode: http.StatusNotFound})
			return
		}
		middleware.WriteAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"deleted"}`))
}

func (h *Handler) ClearTasks(w http.ResponseWriter, r *http.Request) {
	uc, ok := auth.GetUserContext(r.Context())
	if !ok {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	var (
		deleted int64
		err     error
	)
	if uc.Role == user.RoleAdmin {
		deleted, err = h.Repo.ClearAll()
	} else {
		deleted, err = h.Repo.ClearByCreatedBy(uc.UserID)
	}
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "cleared", "deleted": deleted})
}

func lastPart(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func taskIDFromCancelPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := 0; i < len(parts)-2; i++ {
		if parts[i] == "tasks" && parts[i+2] == "cancel" {
			return parts[i+1]
		}
	}
	return ""
}
