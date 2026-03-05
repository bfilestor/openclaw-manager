package skills

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
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
	var bases []string
	agentID := ""
	switch scope {
	case "global":
		bases = globalSkillBases(h.GlobalDir)
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
		bases = []string{filepath.Join(wp, "skills")}
	default:
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"scope": "must be global or agent"}))
		return
	}

	items, err := scanSkillsMany(bases, scope, agentID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"skills": items})
}

func scanSkillsMany(bases []string, scope, agentID string) ([]SkillItem, error) {
	seen := map[string]struct{}{}
	out := make([]SkillItem, 0)
	for _, base := range bases {
		items, err := scanSkills(base, scope, agentID)
		if err != nil {
			return nil, err
		}
		for _, it := range items {
			if _, ok := seen[it.Name]; ok {
				continue
			}
			seen[it.Name] = struct{}{}
			out = append(out, it)
		}
	}
	return out, nil
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
		_, hasSkillMD := safeStat(filepath.Join(p, "SKILL.md"))
		out = append(out, SkillItem{Name: e.Name(), Scope: scope, AgentID: agentID, SizeBytes: size, HasMeta: hasJSON || hasReadme || hasSkillMD})
	}
	return out, nil
}

func globalSkillBases(configured string) []string {
	bases := make([]string, 0, 4)
	seen := map[string]struct{}{}
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}
		c := filepath.Clean(p)
		if _, ok := seen[c]; ok {
			return
		}
		seen[c] = struct{}{}
		bases = append(bases, c)
	}

	add(configured)
	if home, err := os.UserHomeDir(); err == nil {
		add(filepath.Join(home, ".openclaw", "skills"))
	}

	if bin, err := exec.LookPath("openclaw"); err == nil {
		if realBin, err2 := filepath.EvalSymlinks(bin); err2 == nil {
			add(filepath.Join(filepath.Dir(realBin), "..", "lib", "node_modules", "openclaw", "skills"))
		}
	}
	add("/usr/local/lib/node_modules/openclaw/skills")
	add("/usr/lib/node_modules/openclaw/skills")

	return bases
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
