package config

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"openclaw-manager/internal/agent"
	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/storage"
)

type IdentityHandler struct {
	AgentRepo *agent.Repository
	Revisions *RevisionRepository
	Validator *storage.PathValidator
}

type identityReq struct {
	Content string `json:"content"`
}

func (h *IdentityHandler) GetIdentity(w http.ResponseWriter, r *http.Request) {
	agentID, err := extractAgentIDFromIdentityPath(r.URL.Path)
	if err != nil {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "invalid"}))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	wp, err := h.AgentRepo.GetWorkspacePath(ctx, agentID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	p := filepath.Join(wp, "IDENTITY.md")
	if h.Validator != nil {
		if _, err := h.Validator.Validate(p); err != nil {
			middleware.WriteAppError(w, err)
			return
		}
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"content":""}`))
			return
		}
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"content": string(b)})
}

func (h *IdentityHandler) PutIdentity(w http.ResponseWriter, r *http.Request) {
	agentID, err := extractAgentIDFromIdentityPath(r.URL.Path)
	if err != nil {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "invalid"}))
		return
	}
	var req identityReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if len([]byte(req.Content)) > 1024*1024 {
		middleware.WriteAppError(w, &middleware.AppError{Code: "CONTENT_TOO_LARGE", Message: "content too large", StatusCode: http.StatusBadRequest})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	wp, err := h.AgentRepo.GetWorkspacePath(ctx, agentID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	p := filepath.Join(wp, "IDENTITY.md")
	if h.Validator != nil {
		if _, err := h.Validator.Validate(p); err != nil {
			middleware.WriteAppError(w, err)
			return
		}
	}
	if h.Revisions != nil {
		_, _ = h.Revisions.Save("agent_identity", agentID, req.Content, "")
	}
	if err := storage.AtomicWriteFile(p, []byte(req.Content), 0o644); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"ok"}`))
}

func (h *IdentityHandler) ListIdentityRevisions(w http.ResponseWriter, r *http.Request) {
	agentID, err := extractAgentIDFromIdentityPath(r.URL.Path)
	if err != nil {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "invalid"}))
		return
	}
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, e := strconv.Atoi(v); e == nil {
			limit = n
		}
	}
	list, err := h.Revisions.List("agent_identity", agentID, limit)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"revisions": list})
}

func extractAgentIDFromIdentityPath(path string) (string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// /api/v1/agents/{id}/identity[/...]
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "agents" && i+1 < len(parts) {
			id := parts[i+1]
			if id == "" || strings.Contains(id, "..") || strings.ContainsAny(id, `/\\`) {
				return "", os.ErrInvalid
			}
			return id, nil
		}
	}
	return "", os.ErrInvalid
}
