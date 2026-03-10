package agent

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"openclaw-manager/internal/storage"
)

var (
	ErrNotFound       = errors.New("agent not found")
	ErrInvalidAgentID = errors.New("invalid agent id")
)

type Executor interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type Agent struct {
	AgentID       string `json:"agent_id"`
	WorkspacePath string `json:"workspace_path"`
	BindingsCount int    `json:"bindings_count"`
}

type Repository struct {
	exec             Executor
	validator        *storage.PathValidator
	openclawJSONPath string
	ttl              time.Duration

	mu       sync.Mutex
	cachedAt time.Time
	cached   []Agent
}

func NewRepository(exec Executor, validator *storage.PathValidator) *Repository {
	home, _ := os.UserHomeDir()
	return &Repository{
		exec:             exec,
		validator:        validator,
		openclawJSONPath: filepath.Join(home, ".openclaw", "openclaw.json"),
		ttl:              60 * time.Second,
	}
}

func (r *Repository) InvalidateCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cachedAt = time.Time{}
	r.cached = nil
}

func (r *Repository) List(ctx context.Context) ([]Agent, error) {
	r.mu.Lock()
	if time.Since(r.cachedAt) < r.ttl && r.cached != nil {
		out := make([]Agent, len(r.cached))
		copy(out, r.cached)
		r.mu.Unlock()
		return out, nil
	}
	r.mu.Unlock()

	out, err := r.exec.Run(ctx, "openclaw", "agents", "list", "--bindings", "--json")
	if err != nil {
		return nil, err
	}
	agents, err := parseAgentsJSON(out)
	if err != nil {
		return nil, err
	}
	agents = r.fillWorkspacePathFromConfig(agents)

	r.mu.Lock()
	r.cachedAt = time.Now()
	r.cached = make([]Agent, len(agents))
	copy(r.cached, agents)
	r.mu.Unlock()
	return agents, nil
}

func (r *Repository) GetWorkspacePath(ctx context.Context, agentID string) (string, error) {
	if !validAgentID(agentID) {
		return "", ErrInvalidAgentID
	}
	agents, err := r.List(ctx)
	if err != nil {
		return "", err
	}
	for _, a := range agents {
		if a.AgentID == agentID {
			p := filepath.Clean(a.WorkspacePath)
			if r.validator != nil {
				if _, err := r.validator.Validate(p); err != nil {
					return "", err
				}
			}
			return p, nil
		}
	}
	return "", ErrNotFound
}

func parseAgentsJSON(raw []byte) ([]Agent, error) {
	type rawAgent struct {
		ID        string          `json:"id"`
		Workspace string          `json:"workspace"`
		Bindings  json.RawMessage `json:"bindings"`
	}

	var agents []rawAgent
	trimmed := strings.TrimSpace(string(raw))
	if strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal(raw, &agents); err != nil {
			return nil, err
		}
	} else {
		var payload struct {
			Agents []rawAgent `json:"agents"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, err
		}
		agents = payload.Agents
	}

	out := make([]Agent, 0, len(agents))
	for _, it := range agents {
		if !validAgentID(it.ID) {
			continue
		}
		out = append(out, Agent{AgentID: it.ID, WorkspacePath: it.Workspace, BindingsCount: parseBindingsCount(it.Bindings)})
	}
	return out, nil
}

func parseBindingsCount(raw json.RawMessage) int {
	if len(raw) == 0 {
		return 0
	}

	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		if n < 0 {
			return 0
		}
		return n
	}

	var list []any
	if err := json.Unmarshal(raw, &list); err == nil {
		return len(list)
	}

	return 0
}

func validAgentID(id string) bool {
	id = strings.TrimSpace(id)
	if id == "" || len(id) > 64 || strings.Contains(id, "..") || strings.ContainsAny(id, `/\\ `) {
		return false
	}
	for _, ch := range id {
		if !(ch == '_' || ch == '-' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
			return false
		}
	}
	return true
}

func (r *Repository) fillWorkspacePathFromConfig(list []Agent) []Agent {
	workspaceMap := r.loadWorkspaceMapFromOpenClawJSON()
	if len(workspaceMap) == 0 {
		return list
	}
	for i := range list {
		if strings.TrimSpace(list[i].WorkspacePath) != "" {
			continue
		}
		if ws, ok := workspaceMap[list[i].AgentID]; ok {
			list[i].WorkspacePath = ws
		}
	}
	return list
}

func (r *Repository) loadWorkspaceMapFromOpenClawJSON() map[string]string {
	type rawAgent struct {
		ID        string `json:"id"`
		Workspace string `json:"workspace"`
	}
	type openclawConfig struct {
		Agents struct {
			Defaults struct {
				Workspace string `json:"workspace"`
			} `json:"defaults"`
			List []rawAgent `json:"list"`
		} `json:"agents"`
	}

	path := strings.TrimSpace(r.openclawJSONPath)
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var cfg openclawConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}

	defaultWorkspace := strings.TrimSpace(cfg.Agents.Defaults.Workspace)
	workspaceMap := map[string]string{}
	if defaultWorkspace != "" {
		workspaceMap["main"] = defaultWorkspace
	}
	baseDir := filepath.Dir(defaultWorkspace)
	if baseDir == "." {
		baseDir = filepath.Dir(path)
	}
	for _, it := range cfg.Agents.List {
		agentID := strings.TrimSpace(it.ID)
		if !validAgentID(agentID) {
			continue
		}
		workspace := strings.TrimSpace(it.Workspace)
		if workspace == "" {
			if agentID == "main" {
				workspace = defaultWorkspace
			} else if defaultWorkspace != "" {
				workspace = filepath.Join(baseDir, "workspace-"+agentID)
			}
		}
		if workspace != "" {
			workspaceMap[agentID] = workspace
		}
	}
	return workspaceMap
}
