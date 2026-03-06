package agent

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/storage"
)

type WorkspaceResolver interface {
	GetWorkspacePath(ctx context.Context, agentID string) (string, error)
}

type ManageHandler struct {
	Exec             Executor
	Workspaces       WorkspaceResolver
	OpenClawJSONPath string
}

type createReq struct {
	AgentID string `json:"agent_id"`
}

func (h *ManageHandler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if !validAgentID(req.AgentID) {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "invalid"}))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	_, err := h.Exec.Run(ctx, "openclaw", "agents", "create", req.AgentID)
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && strings.Contains(strings.ToLower(string(ee.Stderr)), "exist") {
			middleware.WriteAppError(w, &middleware.AppError{Code: "CONFLICT", Message: "agent exists", StatusCode: http.StatusConflict})
			return
		}
		middleware.WriteAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"task_type":"agent.add","status":"PENDING"}`))
}

func (h *ManageHandler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	agentID := lastPart(r.URL.Path)
	if !validAgentID(agentID) {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "invalid"}))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	_, _ = h.Exec.Run(ctx, "openclaw", "agents", "unbind", "--all", agentID)
	if _, err := h.Exec.Run(ctx, "openclaw", "agents", "delete", agentID); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"task_type":"agent.delete","status":"PENDING"}`))
}

type migrateWorkspaceReq struct {
	NewWorkspacePath string `json:"new_workspace_path"`
}

func (h *ManageHandler) MigrateWorkspace(w http.ResponseWriter, r *http.Request) {
	agentID := workspaceMigrateAgentID(r.URL.Path)
	if !validAgentID(agentID) {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "invalid"}))
		return
	}
	if h.Workspaces == nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_IMPLEMENTED", Message: "workspace resolver not configured", StatusCode: http.StatusNotImplemented})
		return
	}

	var req migrateWorkspaceReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	newPath := strings.TrimSpace(req.NewWorkspacePath)
	if newPath == "" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"new_workspace_path": "required"}))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
	defer cancel()
	oldPath, err := h.Workspaces.GetWorkspacePath(ctx, agentID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_FOUND", Message: "agent not found", StatusCode: http.StatusNotFound})
			return
		}
		middleware.WriteAppError(w, err)
		return
	}
	if filepath.Clean(oldPath) == filepath.Clean(newPath) {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"new_workspace_path": "must be different from current workspace"}))
		return
	}
	if err := migrateWorkspaceFiles(oldPath, newPath); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if err := h.updateOpenClawWorkspace(agentID, newPath); err != nil {
		middleware.WriteAppError(w, err)
		return
	}

	restartCtx, restartCancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer restartCancel()
	if _, err := h.Exec.Run(restartCtx, "openclaw", "gateway", "restart"); err != nil {
		middleware.WriteAppError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message":            "workspace migrated",
		"agent_id":           agentID,
		"old_workspace_path": oldPath,
		"new_workspace_path": filepath.Clean(newPath),
		"gateway_restarted":  true,
	})
}

func (h *ManageHandler) updateOpenClawWorkspace(agentID, newPath string) error {
	jsonPath := strings.TrimSpace(h.OpenClawJSONPath)
	if jsonPath == "" {
		home, _ := os.UserHomeDir()
		jsonPath = filepath.Join(home, ".openclaw", "openclaw.json")
	}
	raw, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return &middleware.AppError{Code: "INVALID_JSON", Message: "openclaw.json is invalid", StatusCode: http.StatusBadRequest}
	}

	agentsObj, ok := root["agents"].(map[string]any)
	if !ok {
		return &middleware.AppError{Code: "INVALID_JSON", Message: "openclaw.json missing agents", StatusCode: http.StatusBadRequest}
	}

	if agentID == "main" {
		defaultsObj, ok := agentsObj["defaults"].(map[string]any)
		if !ok {
			defaultsObj = map[string]any{}
			agentsObj["defaults"] = defaultsObj
		}
		defaultsObj["workspace"] = filepath.Clean(newPath)
	}

	found := false
	listAny, ok := agentsObj["list"].([]any)
	if !ok {
		listAny = []any{}
	}
	for i := range listAny {
		item, ok := listAny[i].(map[string]any)
		if !ok {
			continue
		}
		if strings.TrimSpace(asString(item["id"])) != agentID {
			continue
		}
		item["workspace"] = filepath.Clean(newPath)
		found = true
		break
	}
	if !found {
		listAny = append(listAny, map[string]any{"id": agentID, "workspace": filepath.Clean(newPath)})
	}
	agentsObj["list"] = listAny

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	return storage.AtomicWriteFile(jsonPath, out, 0o644)
}

func workspaceMigrateAgentID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := 0; i < len(parts)-3; i++ {
		if parts[i] == "agents" && parts[i+2] == "workspace" && parts[i+3] == "migrate" {
			return parts[i+1]
		}
	}
	return ""
}

func migrateWorkspaceFiles(oldPath, newPath string) error {
	oldPath = filepath.Clean(oldPath)
	newPath = filepath.Clean(newPath)
	st, err := os.Stat(oldPath)
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return &middleware.AppError{Code: "VALIDATION_ERROR", Message: "old workspace is not a directory", StatusCode: http.StatusBadRequest}
	}
	if err := os.MkdirAll(newPath, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(oldPath)
	if err != nil {
		return err
	}
	for _, e := range entries {
		src := filepath.Join(oldPath, e.Name())
		dst := filepath.Join(newPath, e.Name())
		if _, err := os.Stat(dst); err == nil {
			return &middleware.AppError{Code: "CONFLICT", Message: "target already contains: " + e.Name(), StatusCode: http.StatusConflict}
		}
		if err := os.Rename(src, dst); err != nil {
			if !errors.Is(err, syscall.EXDEV) {
				return err
			}
			if err := copyPath(src, dst); err != nil {
				return err
			}
			if err := os.RemoveAll(src); err != nil {
				return err
			}
		}
	}
	_ = os.Remove(oldPath)
	return nil
}

func copyPath(src, dst string) error {
	st, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if st.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(link, dst)
	}
	if st.IsDir() {
		if err := os.MkdirAll(dst, st.Mode().Perm()); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if err := copyPath(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, st.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}
