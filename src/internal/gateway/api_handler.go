package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"

	"openclaw-manager/internal/auth"
	cfg "openclaw-manager/internal/config"
	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/storage"
	"openclaw-manager/internal/task"
)

type APIHandler struct {
	Service          *SystemctlService
	UpdateService    *UpdateService
	Revisions        *cfg.RevisionRepository
	TaskRepo         *task.Repository
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

func (h *APIHandler) Upgrade(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.beginTask("upgrade")
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "TASK_CONFLICT", "running_task_id": h.getRunningTaskID()})
		return
	}
	createdBy := taskCreatedBy(r)
	if h.TaskRepo != nil {
		reqJSON, _ := json.Marshal(map[string]any{"action": "upgrade"})
		_ = h.TaskRepo.Create(&task.Task{TaskID: taskID, TaskType: "gateway.upgrade", Status: task.StatusPending, RequestJSON: string(reqJSON), CreatedBy: createdBy})
	}
	go h.runUpgradeTask(taskID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{"task_id": taskID, "status": "PENDING"})
}

func (h *APIHandler) Rollback(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Version string `json:"version"`
	}
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if strings.TrimSpace(req.Version) == "" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"version": "required"}))
		return
	}

	taskID, ok := h.beginTask("rollback")
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "TASK_CONFLICT", "running_task_id": h.getRunningTaskID()})
		return
	}
	createdBy := taskCreatedBy(r)
	if h.TaskRepo != nil {
		reqJSON, _ := json.Marshal(map[string]any{"action": "rollback", "version": req.Version})
		_ = h.TaskRepo.Create(&task.Task{TaskID: taskID, TaskType: "gateway.rollback", Status: task.StatusPending, RequestJSON: string(reqJSON), CreatedBy: createdBy})
	}
	go h.runRollbackTask(taskID, req.Version)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{"task_id": taskID, "status": "PENDING"})
}

func (h *APIHandler) runUpgradeTask(taskID string) {
	defer h.finishTask()
	svc := h.UpdateService
	if svc == nil {
		svc = NewUpdateService(OSExecutor{})
	}
	logLines := []string{"[upgrade] task started"}
	if h.TaskRepo != nil {
		_ = h.TaskRepo.UpdateStatus(taskID, task.StatusRunning)
	}

	backupRevisionID := ""
	if h.Revisions != nil && h.OpenClawJSONPath != "" {
		if raw, readErr := os.ReadFile(h.OpenClawJSONPath); readErr == nil {
			if rev, saveErr := h.Revisions.Save("openclaw_json", "", string(raw), "system-upgrade-backup"); saveErr == nil && rev != nil {
				backupRevisionID = rev.RevisionID
				logLines = append(logLines, "[openclaw.json backup revision] "+backupRevisionID)
			}
		}
	}

	logLines = append(logLines, "[upgrade] running openclaw update --yes --json")
	result, err := svc.Upgrade()
	if err != nil {
		stderr := err.Error()
		if backupRevisionID != "" {
			stderr += "\nrollback target revision available: " + backupRevisionID
		}
		if h.TaskRepo != nil {
			exitCode := 1
			_ = h.TaskRepo.UpdateResult(taskID, &exitCode, strings.Join(logLines, "\n"), stderr, "")
			_ = h.TaskRepo.UpdateStatus(taskID, task.StatusFailed)
		}
		return
	}

	resultText := prettyJSON(result)
	if h.TaskRepo != nil {
		exitCode := 0
		_ = h.TaskRepo.UpdateResult(taskID, &exitCode, strings.Join(append(logLines, resultText), "\n"), "", "")
		_ = h.TaskRepo.UpdateStatus(taskID, task.StatusSucceeded)
	}
}

func (h *APIHandler) runRollbackTask(taskID, version string) {
	defer h.finishTask()
	svc := h.UpdateService
	if svc == nil {
		svc = NewUpdateService(OSExecutor{})
	}
	logLines := []string{"[rollback] task started", "[rollback] target version: " + version}
	if h.TaskRepo != nil {
		_ = h.TaskRepo.UpdateStatus(taskID, task.StatusRunning)
	}

	result, err := svc.RollbackTo(version)
	if err != nil {
		if h.TaskRepo != nil {
			exitCode := 1
			_ = h.TaskRepo.UpdateResult(taskID, &exitCode, strings.Join(logLines, "\n"), err.Error(), "")
			_ = h.TaskRepo.UpdateStatus(taskID, task.StatusFailed)
		}
		return
	}

	restoredRevisionID := ""
	if h.Revisions != nil && h.OpenClawJSONPath != "" {
		if list, listErr := h.Revisions.List("openclaw_json", "", 1); listErr == nil && len(list) > 0 {
			if writeErr := storage.AtomicWriteFile(h.OpenClawJSONPath, []byte(list[0].Content), 0o644); writeErr == nil {
				restoredRevisionID = list[0].RevisionID
				logLines = append(logLines, "[openclaw.json restored revision] "+restoredRevisionID)
			}
		}
	}

	if h.TaskRepo != nil {
		exitCode := 0
		_ = h.TaskRepo.UpdateResult(taskID, &exitCode, strings.Join(append(logLines, prettyJSON(result)), "\n"), "", "")
		_ = h.TaskRepo.UpdateStatus(taskID, task.StatusSucceeded)
	}
}

func (h *APIHandler) beginTask(prefix string) (string, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.runningTaskID != "" {
		return "", false
	}
	taskID := fmt.Sprintf("%s-%s", prefix, uuid.NewString())
	h.runningTaskID = taskID
	return taskID, true
}

func (h *APIHandler) finishTask() {
	h.mu.Lock()
	h.runningTaskID = ""
	h.mu.Unlock()
}

func (h *APIHandler) getRunningTaskID() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.runningTaskID
}

func taskCreatedBy(r *http.Request) string {
	uc, ok := auth.GetUserContext(r.Context())
	if !ok {
		return ""
	}
	return uc.UserID
}

func prettyJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
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
