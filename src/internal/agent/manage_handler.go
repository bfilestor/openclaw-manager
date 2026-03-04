package agent

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"openclaw-manager/internal/middleware"
)

type ManageHandler struct {
	Exec Executor
}

type createReq struct {
	AgentID string `json:"agent_id"`
}

func (h *ManageHandler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if !validAgentID(req.AgentID) {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "invalid"}))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	_, err := h.Exec.Run(ctx, "openclaw", "agents", "create", req.AgentID)
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && strings.Contains(strings.ToLower(string(ee.Stderr)), "exist") {
			middleware.WriteAppError(w, &middleware.AppError{Code: "CONFLICT", Message: "agent exists", StatusCode: http.StatusConflict})
			return
		}
		middleware.WriteAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"task_type":"agent.add","status":"PENDING"}`))
}

func (h *ManageHandler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	agentID := lastPart(r.URL.Path)
	if !validAgentID(agentID) {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "invalid"}))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	_, _ = h.Exec.Run(ctx, "openclaw", "agents", "unbind", "--all", agentID)
	if _, err := h.Exec.Run(ctx, "openclaw", "agents", "delete", agentID); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"task_type":"agent.delete","status":"PENDING"}`))
}
