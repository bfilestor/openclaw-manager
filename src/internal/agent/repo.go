package agent

import (
	"context"
	"encoding/json"
	"errors"
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
	exec      Executor
	validator *storage.PathValidator
	ttl       time.Duration

	mu       sync.Mutex
	cachedAt time.Time
	cached   []Agent
}

func NewRepository(exec Executor, validator *storage.PathValidator) *Repository {
	return &Repository{exec: exec, validator: validator, ttl: 60 * time.Second}
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
	var payload struct {
		Agents []struct {
			ID        string `json:"id"`
			Workspace string `json:"workspace"`
			Bindings  []any  `json:"bindings"`
		} `json:"agents"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	out := make([]Agent, 0, len(payload.Agents))
	for _, it := range payload.Agents {
		if !validAgentID(it.ID) {
			continue
		}
		out = append(out, Agent{AgentID: it.ID, WorkspacePath: it.Workspace, BindingsCount: len(it.Bindings)})
	}
	return out, nil
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
