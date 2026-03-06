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

	agents, err := parseBindingAgentsJSON(out)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"agents": agents})
}

func parseBindingAgentsJSON(raw []byte) ([]map[string]any, error) {
	trimmed := strings.TrimSpace(string(raw))

	if strings.HasPrefix(trimmed, "[") {
		var list []struct {
			ID       string          `json:"id"`
			Bindings json.RawMessage `json:"bindings"`
		}
		if err := json.Unmarshal(raw, &list); err != nil {
			return nil, err
		}
		out := make([]map[string]any, 0, len(list))
		for _, it := range list {
			if !validAgentID(it.ID) {
				continue
			}
			out = append(out, map[string]any{"id": it.ID, "bindings": normalizeBindings(it.Bindings)})
		}
		return out, nil
	}

	var payload struct {
		Agents []map[string]any `json:"agents"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(payload.Agents))
	for _, it := range payload.Agents {
		id, _ := it["id"].(string)
		if !validAgentID(id) {
			continue
		}
		if b, ok := it["bindings"]; ok {
			it["bindings"] = normalizeBindingsAny(b)
		} else {
			it["bindings"] = []any{}
		}
		out = append(out, it)
	}
	return out, nil
}

func normalizeBindings(raw json.RawMessage) any {
	if len(raw) == 0 {
		return []any{}
	}

	var list []any
	if err := json.Unmarshal(raw, &list); err == nil {
		return list
	}

	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		if n <= 0 {
			return []any{}
		}
		return map[string]any{"count": n}
	}

	return []any{}
}

func normalizeBindingsAny(v any) any {
	switch t := v.(type) {
	case []any:
		return t
	case float64:
		n := int(t)
		if n <= 0 {
			return []any{}
		}
		return map[string]any{"count": n}
	default:
		return []any{}
	}
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
