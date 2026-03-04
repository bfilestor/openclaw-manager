package skills

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/storage"
)

type InstallHandler struct {
	GlobalDir string
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
	if scope != "global" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"scope": "only global implemented"}))
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
	base := h.GlobalDir
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".openclaw", "skills")
	}
	target := filepath.Join(base, name)
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
	_, _ = w.Write([]byte(fmt.Sprintf(`{"task_type":"skills.install","status":"PENDING","name":%q}`, name)))
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
