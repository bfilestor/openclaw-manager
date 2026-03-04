package skills

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"openclaw-manager/internal/agent"
	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/storage"
)

type DeleteHandler struct {
	AgentRepo  *agent.Repository
	GlobalDir  string
	Validator  *storage.PathValidator
}

func (h *DeleteHandler) DeleteSkill(w http.ResponseWriter, r *http.Request) {
	name := lastPart(r.URL.Path)
	if strings.TrimSpace(name) == "" || strings.Contains(name, "..") || strings.ContainsAny(name, `/\\`) {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"name": "invalid"}))
		return
	}
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	if scope == "" {
		scope = "global"
	}
	var skillPath string
	switch scope {
	case "global":
		base := h.GlobalDir
		if base == "" {
			home, _ := os.UserHomeDir()
			base = filepath.Join(home, ".openclaw", "skills")
		}
		skillPath = filepath.Join(base, name)
	case "agent":
		agentID := strings.TrimSpace(r.URL.Query().Get("agent_id"))
		if agentID == "" {
			middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "required when scope=agent"}))
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		wp, err := h.AgentRepo.GetWorkspacePath(ctx, agentID)
		if err != nil {
			middleware.WriteAppError(w, err)
			return
		}
		skillPath = filepath.Join(wp, "skills", name)
	default:
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"scope": "must be global or agent"}))
		return
	}
	if h.Validator != nil {
		if _, err := h.Validator.Validate(skillPath); err != nil {
			middleware.WriteAppError(w, err)
			return
		}
	}
	if _, err := os.Stat(skillPath); err != nil {
		if os.IsNotExist(err) {
			middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_FOUND", Message: "skill not found", StatusCode: http.StatusNotFound})
			return
		}
		middleware.WriteAppError(w, err)
		return
	}
	if err := os.RemoveAll(skillPath); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"SUCCEEDED"}`))
}

func lastPart(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
