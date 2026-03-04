package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"openclaw-manager/internal/middleware"
)

type Binding struct {
	AgentID string `json:"agent_id"`
	Channel string `json:"channel"`
	Account string `json:"account"`
	Peer    string `json:"peer"`
}

type BindingApplyReq struct {
	Add    []Binding `json:"add"`
	Remove []Binding `json:"remove"`
}

type BindingHandler struct {
	Exec Executor
}

func NewBindingHandler(exec Executor) *BindingHandler {
	return &BindingHandler{Exec: exec}
}

func (h *BindingHandler) ListBindings(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	out, err := h.Exec.Run(ctx, "openclaw", "agents", "list", "--bindings", "--json")
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out)
}

func (h *BindingHandler) ApplyBindings(w http.ResponseWriter, r *http.Request) {
	var req BindingApplyReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if len(req.Add) == 0 && len(req.Remove) == 0 {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"add/remove": "at least one action required"}))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	failed := 0
	for _, b := range req.Add {
		if strings.TrimSpace(b.AgentID) == "" || strings.TrimSpace(b.Peer) == "" {
			failed++
			continue
		}
		_, err := h.Exec.Run(ctx, "openclaw", "agents", "bind", b.AgentID, b.Channel, b.Account, b.Peer)
		if err != nil {
			failed++
		}
	}
	for _, b := range req.Remove {
		if strings.TrimSpace(b.AgentID) == "" || strings.TrimSpace(b.Peer) == "" {
			failed++
			continue
		}
		_, err := h.Exec.Run(ctx, "openclaw", "agents", "unbind", b.AgentID, b.Channel, b.Account, b.Peer)
		if err != nil {
			failed++
		}
	}
	status := "SUCCEEDED"
	if failed > 0 {
		status = "FAILED"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{"task_type": "binding.apply", "status": status, "failed": failed})
}
