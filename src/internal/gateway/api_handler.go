package gateway

import (
	"encoding/json"
	"net/http"
	"sync"

	"openclaw-manager/internal/middleware"
)

type APIHandler struct {
	Service     *SystemctlService
	ServiceName string

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
