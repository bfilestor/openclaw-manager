package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"openclaw-manager/internal/middleware"
)

type SessionStatusHandler struct {
	Exec Executor
}

type statusAgent struct {
	ID              string `json:"id"`
	WorkspaceDir    string `json:"workspaceDir"`
	SessionsCount   int    `json:"sessionsCount"`
	LastUpdatedAt   int64  `json:"lastUpdatedAt"`
	LastActiveAgeMs int64  `json:"lastActiveAgeMs"`
	BootstrapPending bool  `json:"bootstrapPending"`
}

type statusPayload struct {
	Agents struct {
		Agents []statusAgent `json:"agents"`
	} `json:"agents"`
}

func (h *SessionStatusHandler) ListAgentSessions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	out, err := h.Exec.Run(ctx, "openclaw", "status", "--json")
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}

	var payload statusPayload
	if err := json.Unmarshal(out, &payload); err != nil {
		middleware.WriteAppError(w, err)
		return
	}

	nowMs := time.Now().UnixMilli()
	agentFilter := strings.TrimSpace(r.URL.Query().Get("agentId"))
	rows := make([]map[string]any, 0, len(payload.Agents.Agents))
	for _, a := range payload.Agents.Agents {
		if agentFilter != "" && a.ID != agentFilter {
			continue
		}
		if !validAgentID(a.ID) {
			continue
		}
		status := deriveAgentStatus(a, nowMs)
		createdAt := ""
		if a.LastUpdatedAt > 0 {
			createdAt = msToRFC3339(a.LastUpdatedAt)
		}
		rows = append(rows, map[string]any{
			"id":           "agent:" + a.ID,
			"agentId":      a.ID,
			"status":       status,
			"createdAt":    createdAt,
			"lastActivity": createdAt,
			"workspace":    a.WorkspaceDir,
			"sessionsCount": a.SessionsCount,
			"lastActiveAgeMs": a.LastActiveAgeMs,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"sessions": rows,
		"source":   "openclaw status --json",
	})
}

func deriveAgentStatus(a statusAgent, nowMs int64) string {
	if a.BootstrapPending {
		return "waiting"
	}
	if a.LastUpdatedAt <= 0 {
		return "waiting"
	}
	age := a.LastActiveAgeMs
	if age <= 0 && nowMs > a.LastUpdatedAt {
		age = nowMs - a.LastUpdatedAt
	}
	if age <= 2*60*1000 {
		return "running"
	}
	if age <= 30*60*1000 {
		return "waiting"
	}
	return "completed"
}

func msToRFC3339(ts int64) string {
	sec := ts / 1000
	nsec := (ts % 1000) * int64(time.Millisecond)
	return time.Unix(sec, nsec).UTC().Format(time.RFC3339)
}

