package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"openclaw-manager/internal/middleware"
)

type Handler struct {
	Repo *Repository
}

func (h *Handler) ListAgents(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	list, err := h.Repo.List(ctx)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"agents": list})
}

func (h *Handler) GetAgent(w http.ResponseWriter, r *http.Request) {
	agentID := lastPart(r.URL.Path)
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	list, err := h.Repo.List(ctx)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	for _, a := range list {
		if a.AgentID == agentID {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(a)
			return
		}
	}
	middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_FOUND", Message: "agent not found", StatusCode: http.StatusNotFound})
}

func lastPart(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
