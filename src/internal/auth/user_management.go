package auth

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/user"
)

type changePasswordReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type changeRoleReq struct {
	Role string `json:"role"`
}

type disableReq struct {
	Disabled bool `json:"disabled"`
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	uc, ok := GetUserContext(r.Context())
	if !ok {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	u, err := h.Repo.FindByID(uc.UserID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	writeUserJSON(w, http.StatusOK, u)
}

func (h *Handler) ChangeMyPassword(w http.ResponseWriter, r *http.Request) {
	uc, ok := GetUserContext(r.Context())
	if !ok {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	var req changePasswordReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	u, err := h.Repo.FindByID(uc.UserID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if !h.Pass.Verify(req.OldPassword, u.PasswordHash) {
		middleware.WriteAppError(w, &middleware.AppError{Code: "INVALID_PASSWORD", Message: "invalid password", StatusCode: http.StatusBadRequest})
		return
	}
	if len(req.NewPassword) > 128 {
		middleware.WriteAppError(w, &middleware.AppError{Code: "PASSWORD_TOO_LONG", Message: "password too long", StatusCode: http.StatusBadRequest})
		return
	}
	if err := h.Pass.ValidateStrength(req.NewPassword); err != nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "PASSWORD_WEAK", Message: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	hash, err := h.Pass.Hash(req.NewPassword)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	now := time.Now().UTC()
	u.PasswordHash = hash
	u.UpdatedAt = &now
	if err := h.Repo.Update(u); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"password updated"}`))
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	uc, ok := GetUserContext(r.Context())
	if !ok {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	if roleWeight(uc.Role) < roleWeight(user.RoleAdmin) {
		middleware.WriteAppError(w, middleware.NewForbidden(string(user.RoleAdmin)))
		return
	}
	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	users, total, err := h.Repo.List(offset, limit)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	respUsers := make([]map[string]any, 0, len(users))
	for _, u := range users {
		respUsers = append(respUsers, publicUser(u))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"users": respUsers, "total": total})
}

func (h *Handler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	uc, ok := GetUserContext(r.Context())
	if !ok {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	if roleWeight(uc.Role) < roleWeight(user.RoleAdmin) {
		middleware.WriteAppError(w, middleware.NewForbidden(string(user.RoleAdmin)))
		return
	}
	targetID := extractUserID(r.URL.Path, "role")
	if targetID == "" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"user_id": "required"}))
		return
	}
	if targetID == uc.UserID {
		middleware.WriteAppError(w, &middleware.AppError{Code: "CANNOT_MODIFY_SELF", Message: "cannot modify self", StatusCode: http.StatusBadRequest})
		return
	}
	var req changeRoleReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	newRole := user.Role(req.Role)
	if newRole != user.RoleViewer && newRole != user.RoleOperator && newRole != user.RoleAdmin {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"role": "invalid role"}))
		return
	}
	target, err := h.Repo.FindByID(targetID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if target.Role == user.RoleAdmin && newRole != user.RoleAdmin && adminCount(h) <= 1 {
		middleware.WriteAppError(w, &middleware.AppError{Code: "LAST_ADMIN_PROTECTED", Message: "at least one admin required", StatusCode: http.StatusBadRequest})
		return
	}
	now := time.Now().UTC()
	target.Role = newRole
	target.UpdatedAt = &now
	if err := h.Repo.Update(target); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	writeUserJSON(w, http.StatusOK, target)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	uc, ok := GetUserContext(r.Context())
	if !ok {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	if roleWeight(uc.Role) < roleWeight(user.RoleAdmin) {
		middleware.WriteAppError(w, middleware.NewForbidden(string(user.RoleAdmin)))
		return
	}
	targetID := extractUserID(r.URL.Path, "")
	if targetID == uc.UserID {
		middleware.WriteAppError(w, &middleware.AppError{Code: "CANNOT_DELETE_SELF", Message: "cannot delete self", StatusCode: http.StatusBadRequest})
		return
	}
	target, err := h.Repo.FindByID(targetID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if target.Role == user.RoleAdmin && adminCount(h) <= 1 {
		middleware.WriteAppError(w, &middleware.AppError{Code: "LAST_ADMIN_PROTECTED", Message: "at least one admin required", StatusCode: http.StatusBadRequest})
		return
	}
	if err := h.Repo.Delete(targetID); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"deleted"}`))
}

func (h *Handler) DisableUser(w http.ResponseWriter, r *http.Request) {
	uc, ok := GetUserContext(r.Context())
	if !ok {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	if roleWeight(uc.Role) < roleWeight(user.RoleAdmin) {
		middleware.WriteAppError(w, middleware.NewForbidden(string(user.RoleAdmin)))
		return
	}
	targetID := extractUserID(r.URL.Path, "disable")
	if targetID == uc.UserID {
		middleware.WriteAppError(w, &middleware.AppError{Code: "CANNOT_DISABLE_SELF", Message: "cannot disable self", StatusCode: http.StatusBadRequest})
		return
	}
	var req disableReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	target, err := h.Repo.FindByID(targetID)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if target.Role == user.RoleAdmin && req.Disabled && adminCount(h) <= 1 {
		middleware.WriteAppError(w, &middleware.AppError{Code: "LAST_ADMIN_PROTECTED", Message: "at least one admin required", StatusCode: http.StatusBadRequest})
		return
	}
	now := time.Now().UTC()
	if req.Disabled {
		target.Status = user.StatusDisabled
	} else {
		target.Status = user.StatusActive
	}
	target.UpdatedAt = &now
	if err := h.Repo.Update(target); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	writeUserJSON(w, http.StatusOK, target)
}

func writeUserJSON(w http.ResponseWriter, status int, u *user.User) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(publicUser(u))
}

func publicUser(u *user.User) map[string]any {
	out := map[string]any{
		"user_id":    u.UserID,
		"username":   u.Username,
		"role":       string(u.Role),
		"status":     string(u.Status),
		"created_at": u.CreatedAt.Format(time.RFC3339),
	}
	if u.LastLoginAt != nil {
		out["last_login_at"] = u.LastLoginAt.Format(time.RFC3339)
	}
	if u.UpdatedAt != nil {
		out["updated_at"] = u.UpdatedAt.Format(time.RFC3339)
	}
	return out
}

func adminCount(h *Handler) int {
	list, _, err := h.Repo.List(0, 10000)
	if err != nil {
		return 0
	}
	c := 0
	for _, u := range list {
		if u.Role == user.RoleAdmin && u.Status == user.StatusActive {
			c++
		}
	}
	return c
}

func extractUserID(path string, tail string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	if tail == "" {
		return parts[len(parts)-1]
	}
	if len(parts) < 2 {
		return ""
	}
	if parts[len(parts)-1] != tail {
		return ""
	}
	return parts[len(parts)-2]
}
