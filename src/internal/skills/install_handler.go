package skills

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"openclaw-manager/internal/agent"
	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/storage"
)

type InstallHandler struct {
	GlobalDir string
	AgentRepo *agent.Repository
	Validator *storage.PathValidator
}

func (h *InstallHandler) InstallSkill(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "FILE_TOO_LARGE", Message: "file too large", StatusCode: http.StatusBadRequest})
		return
	}
	scope := strings.ToLower(strings.TrimSpace(r.FormValue("scope")))
	if scope == "" {
		scope = "global"
	}
	if scope != "global" && scope != "agent" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"scope": "must be global or agent"}))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"file": "required"}))
		return
	}
	defer file.Close()

	name := strings.TrimSpace(r.FormValue("skill_name"))
	if name == "" {
		name = inferSkillName(header)
	}
	if name == "" || strings.Contains(name, "..") || strings.ContainsAny(name, `/\\`) {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"skill_name": "invalid"}))
		return
	}
	if !isSupportedArchive(header.Filename) {
		middleware.WriteAppError(w, &middleware.AppError{Code: "UNSUPPORTED_FORMAT", Message: "only .zip or .tar.gz supported", StatusCode: http.StatusBadRequest})
		return
	}

	base := h.GlobalDir
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".openclaw", "skills")
	}
	if scope == "agent" {
		agentID := strings.TrimSpace(r.FormValue("agent_id"))
		if agentID == "" {
			middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"agent_id": "required when scope=agent"}))
			return
		}
		if h.AgentRepo == nil {
			middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_IMPLEMENTED", Message: "agent scope install not configured", StatusCode: http.StatusNotImplemented})
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
	}
	target := filepath.Join(base, name)
	if h.Validator != nil {
		if _, err := h.Validator.Validate(filepath.Dir(target)); err != nil {
			middleware.WriteAppError(w, err)
			return
		}
	}
	if _, err := os.Stat(target); err == nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "SKILL_EXISTS", Message: "skill exists", StatusCode: http.StatusConflict})
		return
	}

	tmpDir, err := os.MkdirTemp("", "skill-install-*")
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	defer os.RemoveAll(tmpDir)
	archive := filepath.Join(tmpDir, header.Filename)
	if err := saveUpload(file, archive); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if err := storage.SafeExtract(archive, target); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"task_type":"skills.install","status":"PENDING","name":%q,"scope":%q}`, name, scope)))
}

func isSupportedArchive(filename string) bool {
	name := strings.ToLower(strings.TrimSpace(filename))
	return strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".tar.gz")
}

func saveUpload(src multipart.File, dst string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, src)
	return err
}

func inferSkillName(h *multipart.FileHeader) string {
	n := strings.ToLower(h.Filename)
	n = strings.TrimSuffix(n, ".tar.gz")
	n = strings.TrimSuffix(n, ".zip")
	return n
}
