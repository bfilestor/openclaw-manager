package backup

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"openclaw-manager/internal/auth"
	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/task"

	"github.com/google/uuid"
)

type APIHandler struct {
	Service *Service
	PlanSvc *PlanService
	DB      *sql.DB
}

type createReq struct {
	Label string   `json:"label"`
	Scope []string `json:"scope"`
}

func (h *APIHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if len(req.Scope) == 0 {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"scope": "required"}))
		return
	}
	createdBy := currentUserID(r)
	taskRepo := task.NewRepository(h.DB)
	taskID := "backup-create-" + uuid.NewString()
	reqJSON := fmt.Sprintf(`{"label":%q,"scope":%s}`, req.Label, mustJSON(req.Scope))
	now := time.Now().UTC()
	_ = taskRepo.Create(&task.Task{
		TaskID:      taskID,
		TaskType:    "backup.create",
		Status:      task.StatusRunning,
		RequestJSON: reqJSON,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		StartedAt:   &now,
	})

	id, err := h.Service.Create(req.Scope, req.Label, createdBy)
	if err != nil {
		exitCode := 1
		_ = taskRepo.UpdateResult(taskID, &exitCode, "", err.Error(), "")
		_ = taskRepo.UpdateStatus(taskID, task.StatusFailed)
		middleware.WriteAppError(w, err)
		return
	}
	_ = taskRepo.UpdateResult(taskID, nil, fmt.Sprintf("backup_id=%s", id), "", "")
	_ = taskRepo.UpdateStatus(taskID, task.StatusSucceeded)
	if h.PlanSvc != nil {
		_ = h.PlanSvc.SavePreference(createdBy, req.Label, req.Scope)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"task_id":"` + taskID + `","backup_id":"` + id + `","status":"PENDING"}`))
}

func (h *APIHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`SELECT backup_id,label,size_bytes,sha256,created_at FROM backups ORDER BY created_at DESC`)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	defer rows.Close()
	list := make([]map[string]any, 0)
	for rows.Next() {
		var id, label, sha, created string
		var size int64
		if err := rows.Scan(&id, &label, &size, &sha, &created); err != nil {
			middleware.WriteAppError(w, err)
			return
		}
		list = append(list, map[string]any{"backup_id": id, "label": label, "size_bytes": size, "sha256": sha, "created_at": created})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"backups": list})
}

func (h *APIHandler) GetBackup(w http.ResponseWriter, r *http.Request) {
	id := lastPart(r.URL.Path)
	var manifest string
	err := h.DB.QueryRow(`SELECT scope_json FROM backups WHERE backup_id=?`, id).Scan(&manifest)
	if err != nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_FOUND", Message: "backup not found", StatusCode: http.StatusNotFound})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(manifest))
}

func (h *APIHandler) DownloadBackup(w http.ResponseWriter, r *http.Request) {
	id := backupIDFromDownloadPath(r.URL.Path)
	if id == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	p := filepath.Join(h.Service.BackupHome, id+".tar.gz")
	w.Header().Set("Content-Type", "application/gzip")
	http.ServeFile(w, r, p)
}

func (h *APIHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	id := backupIDFromRestorePath(r.URL.Path)
	if id == "" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"backup_id": "required"}))
		return
	}
	type restoreReq struct {
		DryRun         *bool `json:"dry_run"`
		RestartGateway bool  `json:"restart_gateway"`
	}
	var req restoreReq
	_ = middleware.BindJSON(r, &req)
	dry := true
	if req.DryRun != nil {
		dry = *req.DryRun
	}
	createdBy := currentUserID(r)
	report, err := h.Service.Restore(id, dry, req.RestartGateway, createdBy)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if dry {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(report)
		return
	}

	taskRepo := task.NewRepository(h.DB)
	taskID := "backup-restore-" + uuid.NewString()
	reqJSON := fmt.Sprintf(`{"backup_id":%q,"dry_run":false,"restart_gateway":%t}`, id, req.RestartGateway)
	now := time.Now().UTC()
	_ = taskRepo.Create(&task.Task{
		TaskID:      taskID,
		TaskType:    "backup.restore",
		Status:      task.StatusSucceeded,
		RequestJSON: reqJSON,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		StartedAt:   &now,
		FinishedAt:  &now,
		StdoutTail:  fmt.Sprintf("restored backup_id=%s", id),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"task_id":"` + taskID + `","task_type":"backup.restore","status":"PENDING"}`))
}

func (h *APIHandler) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	id := lastPart(r.URL.Path)
	res, err := h.DB.Exec(`DELETE FROM backups WHERE backup_id=?`, id)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_FOUND", Message: "backup not found", StatusCode: http.StatusNotFound})
		return
	}
	_ = removeIfExists(filepath.Join(h.Service.BackupHome, id+".tar.gz"))
	_ = removeIfExists(filepath.Join(h.Service.BackupHome, id+".manifest.json"))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"deleted"}`))
}

func backupIDFromRestorePath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "backups" && i+2 < len(parts) && parts[i+2] == "restore" {
			return parts[i+1]
		}
	}
	return ""
}

func backupIDFromDownloadPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "backups" && i+2 < len(parts) && parts[i+2] == "download" {
			return parts[i+1]
		}
	}
	return ""
}

func lastPart(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func removeIfExists(path string) error {
	if _, err := os.Stat(path); err == nil {
		return os.Remove(path)
	}
	return nil
}

func currentUserID(r *http.Request) string {
	uc, ok := auth.GetUserContext(r.Context())
	if !ok || uc == nil {
		return ""
	}
	return uc.UserID
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(b)
}
