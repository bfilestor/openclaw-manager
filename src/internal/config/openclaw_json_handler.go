package config

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/storage"
)

type OpenClawJSONHandler struct {
	FilePath   string
	Validator  *storage.PathValidator
	Revisions  *RevisionRepository
}

type writeReq struct {
	Content string `json:"content"`
}

func (h *OpenClawJSONHandler) GetOpenClawJSON(w http.ResponseWriter, r *http.Request) {
	if err := h.checkPath(); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	b, err := os.ReadFile(h.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"content":null}`))
			return
		}
		middleware.WriteAppError(w, err)
		return
	}
	st, _ := os.Stat(h.FilePath)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"content": string(b), "size": len(b), "modified_at": st.ModTime().UTC().Format(time.RFC3339)})
}

func (h *OpenClawJSONHandler) PutOpenClawJSON(w http.ResponseWriter, r *http.Request) {
	if err := h.checkPath(); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	var req writeReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"content": "required"}))
		return
	}
	var tmp any
	if err := json.Unmarshal([]byte(req.Content), &tmp); err != nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "INVALID_JSON", Message: "invalid json", StatusCode: http.StatusBadRequest})
		return
	}
	createdBy := ""
	if h.Revisions != nil {
		if _, err := h.Revisions.Save("openclaw_json", "", req.Content, createdBy); err != nil {
			middleware.WriteAppError(w, err)
			return
		}
	}
	if err := storage.AtomicWriteFile(h.FilePath, []byte(req.Content), 0o644); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"ok"}`))
}

func (h *OpenClawJSONHandler) ListRevisions(w http.ResponseWriter, r *http.Request) {
	if h.Revisions == nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_IMPLEMENTED", Message: "revisions disabled", StatusCode: http.StatusNotImplemented})
		return
	}
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	list, err := h.Revisions.List("openclaw_json", "", limit)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"revisions": list})
}

func (h *OpenClawJSONHandler) RestoreRevision(w http.ResponseWriter, r *http.Request) {
	if h.Revisions == nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_IMPLEMENTED", Message: "revisions disabled", StatusCode: http.StatusNotImplemented})
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"rev_id": "required"}))
		return
	}
	revID := parts[len(parts)-2]
	rev, err := h.Revisions.FindByID(revID)
	if err != nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "NOT_FOUND", Message: "revision not found", StatusCode: http.StatusNotFound})
		return
	}
	if err := storage.AtomicWriteFile(h.FilePath, []byte(rev.Content), 0o644); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	createdBy := ""
	_, _ = h.Revisions.Save("openclaw_json", "", rev.Content, createdBy)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"restored"}`))
}

func (h *OpenClawJSONHandler) checkPath() error {
	if h.Validator == nil {
		return nil
	}
	_, err := h.Validator.Validate(h.FilePath)
	return err
}
