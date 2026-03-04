package skills

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"openclaw-manager/internal/agent"
	"openclaw-manager/internal/middleware"
)

type SkillItem struct {
	Name      string `json:"name"`
	Scope     string `json:"scope"`
	AgentID   string `json:"agent_id,omitempty"`
	SizeBytes int64  `json:"size_bytes"`
	HasMeta   bool   `json:"has_meta"`
}

type Handler struct {
	AgentRepo *agent.Repository
	GlobalDir string
}

func (h *Handler) ListSkills(w http.ResponseWriter, r *http.Request) {
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	if scope == "" {
		scope = "global"
	}
	var base string
	agentID := ""
	switch scope {
	case "global":
		base = h.GlobalDir
		if base == "" {
			home, _ := os.UserHomeDir()
			base = filepath.Join(home, ".openclaw", "skills")
		}
	case "agent":
		agentID = strings.TrimSpace(r.URL.Query().Get("agent_id"))
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
		base = filepath.Join(wp, "skills")
	default:
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"scope": "must be global or agent"}))
		return
	}

	items, err := scanSkills(base, scope, agentID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"skills": items})
}

func scanSkills(base, scope, agentID string) ([]SkillItem, error) {
	st, err := os.Stat(base)
	if err != nil {
		if os.IsNotExist(err) {
			return []SkillItem{}, nil
		}
		return nil, err
	}
	if !st.IsDir() {
		return []SkillItem{}, nil
	}
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, err
	}
	out := make([]SkillItem, 0)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p := filepath.Join(base, e.Name())
		size, _ := dirSize(p)
		_, hasJSON := safeStat(filepath.Join(p, "skill.json"))
		_, hasReadme := safeStat(filepath.Join(p, "README.md"))
		out = append(out, SkillItem{Name: e.Name(), Scope: scope, AgentID: agentID, SizeBytes: size, HasMeta: hasJSON || hasReadme})
	}
	return out, nil
}

func dirSize(dir string) (int64, error) {
	var total int64
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}

func safeStat(path string) (os.FileInfo, bool) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	return st, true
}
