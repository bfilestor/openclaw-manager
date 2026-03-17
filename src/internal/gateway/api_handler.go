package gateway

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"

	cfg "openclaw-manager/internal/config"
	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/storage"
)

type APIHandler struct {
	Service          *SystemctlService
	UpdateService    *UpdateService
	Revisions        *cfg.RevisionRepository
	OpenClawJSONPath string
	ServiceName      string

	mu            sync.Mutex
	runningTaskID string
}

func (h *APIHandler) Status(w http.ResponseWriter, r *http.Request) {
	st, err := h.Service.DeepStatus(h.serviceName())
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(st)
}

func (h *APIHandler) Start(w http.ResponseWriter, r *http.Request)   { h.doAction(w, "start") }
func (h *APIHandler) Stop(w http.ResponseWriter, r *http.Request)    { h.doAction(w, "stop") }
func (h *APIHandler) Restart(w http.ResponseWriter, r *http.Request) { h.doAction(w, "restart") }

func (h *APIHandler) VersionStatus(w http.ResponseWriter, _ *http.Request) {
	svc := h.UpdateService
	if svc == nil {
		svc = NewUpdateService(OSExecutor{})
	}
	status, err := svc.VersionStatus()
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
}

func (h *APIHandler) Upgrade(w http.ResponseWriter, _ *http.Request) {
	h.mu.Lock()
	if h.runningTaskID != "" {
		rid := h.runningTaskID
		h.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "TASK_CONFLICT", "running_task_id": rid})
		return
	}
	taskID := "upgrade-task"
	h.runningTaskID = taskID
	h.mu.Unlock()

	svc := h.UpdateService
	if svc == nil {
		svc = NewUpdateService(OSExecutor{})
	}

	backupRevisionID := ""
	if h.Revisions != nil && h.OpenClawJSONPath != "" {
		if raw, readErr := os.ReadFile(h.OpenClawJSONPath); readErr == nil {
			if rev, saveErr := h.Revisions.Save("openclaw_json", "", string(raw), "system-upgrade-backup"); saveErr == nil && rev != nil {
				backupRevisionID = rev.RevisionID
			}
		}
	}

	result, err := svc.Upgrade()

	h.mu.Lock()
	h.runningTaskID = ""
	h.mu.Unlock()

	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := map[string]any{"task_id": taskID, "status": "SUCCEEDED", "result": result}
	if backupRevisionID != "" {
		resp["openclaw_json_backup_revision_id"] = backupRevisionID
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) Rollback(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	if h.runningTaskID != "" {
		rid := h.runningTaskID
		h.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "TASK_CONFLICT", "running_task_id": rid})
		return
	}
	taskID := "rollback-task"
	h.runningTaskID = taskID
	h.mu.Unlock()

	var req struct {
		Version string `json:"version"`
	}
	if err := middleware.BindJSON(r, &req); err != nil {
		h.mu.Lock()
		h.runningTaskID = ""
		h.mu.Unlock()
		middleware.WriteAppError(w, err)
		return
	}

	svc := h.UpdateService
	if svc == nil {
		svc = NewUpdateService(OSExecutor{})
	}
	result, err := svc.RollbackTo(req.Version)

	h.mu.Lock()
	h.runningTaskID = ""
	h.mu.Unlock()

	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}

	restoredRevisionID := ""
	if h.Revisions != nil && h.OpenClawJSONPath != "" {
		if list, listErr := h.Revisions.List("openclaw_json", "", 1); listErr == nil && len(list) > 0 {
			if writeErr := storage.AtomicWriteFile(h.OpenClawJSONPath, []byte(list[0].Content), 0o644); writeErr == nil {
				restoredRevisionID = list[0].RevisionID
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := map[string]any{"task_id": taskID, "status": "SUCCEEDED", "result": result}
	if restoredRevisionID != "" {
		resp["openclaw_json_restored_revision_id"] = restoredRevisionID
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) doAction(w http.ResponseWriter, action string) {
	h.mu.Lock()
	if h.runningTaskID != "" {
		rid := h.runningTaskID
		h.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "TASK_CONFLICT", "running_task_id": rid})
		return
	}
	taskID := action + "-task"
	h.runningTaskID = taskID
	h.mu.Unlock()

	var err error
	switch action {
	case "start":
		err = h.Service.Start(h.serviceName())
	case "stop":
		err = h.Service.Stop(h.serviceName())
	case "restart":
		err = h.Service.Restart(h.serviceName())
	}

	h.mu.Lock()
	h.runningTaskID = ""
	h.mu.Unlock()

	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{"task_id": taskID, "status": "PENDING"})
}

func (h *APIHandler) serviceName() string {
	if h.ServiceName == "" {
		return "openclaw-gateway.service"
	}
	return h.ServiceName
}
