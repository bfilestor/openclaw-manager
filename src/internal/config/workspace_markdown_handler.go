package config

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/storage"
)

type WorkspaceResolver interface {
	GetWorkspacePath(ctx context.Context, agentID string) (string, error)
}

type WorkspaceMarkdownHandler struct {
	Workspaces WorkspaceResolver
	Validator  *storage.PathValidator
	Revisions  *RevisionRepository
}

type workspaceMarkdownSaveReq struct {
	Content string `json:"content"`
}

type workspaceMarkdownFileItem struct {
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modified_at"`
}

func (h *WorkspaceMarkdownHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	agentID, err := workspaceAgentID(r.URL.Path)
	if err != nil {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "invalid"}))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	workspace, err := h.Workspaces.GetWorkspacePath(ctx, agentID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if h.Validator != nil {
		if _, err := h.Validator.Validate(workspace); err != nil {
			middleware.WriteAppError(w, err)
			return
		}
	}
	items := make([]workspaceMarkdownFileItem, 0)
	err = filepath.WalkDir(workspace, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !isMarkdownFile(path) {
			return nil
		}
		rel, err := filepath.Rel(workspace, path)
		if err != nil {
			return err
		}
		st, err := os.Stat(path)
		if err != nil {
			return err
		}
		items = append(items, workspaceMarkdownFileItem{Path: filepath.ToSlash(rel), Size: st.Size(), ModifiedAt: st.ModTime().UTC().Format(time.RFC3339)})
		return nil
	})
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"files": items})
}

func (h *WorkspaceMarkdownHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	agentID, err := workspaceAgentID(r.URL.Path)
	if err != nil {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "invalid"}))
		return
	}
	relPath, fullPath, err := h.resolveMarkdownFilePath(r, agentID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	b, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"path": relPath, "content": ""})
			return
		}
		middleware.WriteAppError(w, err)
		return
	}
	st, _ := os.Stat(fullPath)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"path": relPath, "content": string(b), "size": len(b), "modified_at": st.ModTime().UTC().Format(time.RFC3339)})
}

func (h *WorkspaceMarkdownHandler) PutFile(w http.ResponseWriter, r *http.Request) {
	agentID, err := workspaceAgentID(r.URL.Path)
	if err != nil {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "invalid"}))
		return
	}
	relPath, fullPath, err := h.resolveMarkdownFilePath(r, agentID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	var req workspaceMarkdownSaveReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if len([]byte(req.Content)) > 2*1024*1024 {
		middleware.WriteAppError(w, &middleware.AppError{Code: "CONTENT_TOO_LARGE", Message: "content too large", StatusCode: http.StatusBadRequest})
		return
	}
	if h.Revisions != nil {
		_, _ = h.Revisions.Save("agent_workspace_markdown", revisionTargetID(agentID, relPath), req.Content, "")
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if err := storage.AtomicWriteFile(fullPath, []byte(req.Content), 0o644); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "ok"})
}

func (h *WorkspaceMarkdownHandler) ListRevisions(w http.ResponseWriter, r *http.Request) {
	if h.Revisions == nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_IMPLEMENTED", Message: "revisions disabled", StatusCode: http.StatusNotImplemented})
		return
	}
	agentID, err := workspaceAgentID(r.URL.Path)
	if err != nil {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "invalid"}))
		return
	}
	relPath, _, err := h.resolveMarkdownFilePath(r, agentID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, e := strconv.Atoi(v); e == nil {
			limit = n
		}
	}
	list, err := h.Revisions.List("agent_workspace_markdown", revisionTargetID(agentID, relPath), limit)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"revisions": list})
}

func (h *WorkspaceMarkdownHandler) RestoreRevision(w http.ResponseWriter, r *http.Request) {
	if h.Revisions == nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_IMPLEMENTED", Message: "revisions disabled", StatusCode: http.StatusNotImplemented})
		return
	}
	agentID, err := workspaceAgentID(r.URL.Path)
	if err != nil {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "invalid"}))
		return
	}
	revID := revisionIDFromWorkspacePath(r.URL.Path)
	if revID == "" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"rev_id": "required"}))
		return
	}
	relPath, fullPath, err := h.resolveMarkdownFilePath(r, agentID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	rev, err := h.Revisions.FindByID(revID)
	if err != nil || rev.TargetType != "agent_workspace_markdown" || rev.TargetID != revisionTargetID(agentID, relPath) {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_FOUND", Message: "revision not found", StatusCode: http.StatusNotFound})
		return
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if err := storage.AtomicWriteFile(fullPath, []byte(rev.Content), 0o644); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	_, _ = h.Revisions.Save("agent_workspace_markdown", revisionTargetID(agentID, relPath), rev.Content, "")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "restored"})
}

func (h *WorkspaceMarkdownHandler) DeleteRevision(w http.ResponseWriter, r *http.Request) {
	if h.Revisions == nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_IMPLEMENTED", Message: "revisions disabled", StatusCode: http.StatusNotImplemented})
		return
	}
	agentID, err := workspaceAgentID(r.URL.Path)
	if err != nil {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "invalid"}))
		return
	}
	revID := revisionIDFromWorkspacePath(r.URL.Path)
	if revID == "" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"rev_id": "required"}))
		return
	}
	relPath, _, err := h.resolveMarkdownFilePath(r, agentID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	ok, err := h.Revisions.Delete("agent_workspace_markdown", revisionTargetID(agentID, relPath), revID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if !ok {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_FOUND", Message: "revision not found", StatusCode: http.StatusNotFound})
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "deleted"})
}

func (h *WorkspaceMarkdownHandler) resolveMarkdownFilePath(r *http.Request, agentID string) (string, string, error) {
	relPath := strings.TrimSpace(r.URL.Query().Get("path"))
	if relPath == "" {
		return "", "", middleware.NewValidation(map[string]string{"path": "required"})
	}
	relPath = filepath.ToSlash(filepath.Clean(relPath))
	if strings.HasPrefix(relPath, "/") || strings.HasPrefix(relPath, "../") || relPath == ".." || relPath == "." {
		return "", "", middleware.NewValidation(map[string]string{"path": "invalid"})
	}
	if !isMarkdownFile(relPath) {
		return "", "", middleware.NewValidation(map[string]string{"path": "must be .md file"})
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	workspace, err := h.Workspaces.GetWorkspacePath(ctx, agentID)
	if err != nil {
		return "", "", err
	}
	fullPath := filepath.Join(workspace, filepath.FromSlash(relPath))
	if h.Validator != nil {
		if _, err := h.Validator.Validate(fullPath); err != nil {
			return "", "", err
		}
	}
	return relPath, fullPath, nil
}

func revisionTargetID(agentID, path string) string {
	return agentID + ":" + filepath.ToSlash(path)
}

func workspaceAgentID(path string) (string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := 0; i < len(parts)-2; i++ {
		if parts[i] == "agents" && parts[i+2] == "workspace" {
			id := strings.TrimSpace(parts[i+1])
			if id == "" || strings.Contains(id, "..") || strings.ContainsAny(id, `/\\ `) {
				return "", os.ErrInvalid
			}
			return id, nil
		}
	}
	return "", os.ErrInvalid
}

func revisionIDFromWorkspacePath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "revisions" {
			return parts[i+1]
		}
	}
	return ""
}

func isMarkdownFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(path)), ".md")
}
