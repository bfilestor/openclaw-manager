package auth

import (
	"crypto/subtle"
	"net/http"
	"time"

	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/user"
)

type resetPwdAdminResp struct {
	Username string `json:"username"`
}

type resetPwdReq struct {
	SuperToken  string `json:"super_token"`
	NewPassword string `json:"new_password"`
}

func (h *Handler) GetResetPasswordAdmin(w http.ResponseWriter, r *http.Request) {
	superToken := r.URL.Query().Get("super_token")
	if !h.validSuperToken(superToken) {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	admin, err := h.Repo.FindFirstRegistered()
	if err != nil {
		if err == user.ErrNotFound {
			middleware.WriteAppError(w, &middleware.AppError{Code: "ADMIN_NOT_FOUND", Message: "admin not found", StatusCode: http.StatusNotFound})
			return
		}
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"username":"` + admin.Username + `"}`))
}

func (h *Handler) ResetFirstAdminPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPwdReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if !h.validSuperToken(req.SuperToken) {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	if len(req.NewPassword) > 128 {
		middleware.WriteAppError(w, &middleware.AppError{Code: "PASSWORD_TOO_LONG", Message: "password too long", StatusCode: http.StatusBadRequest})
		return
	}
	if err := h.Pass.ValidateStrength(req.NewPassword); err != nil {
		code := "PASSWORD_WEAK"
		if err == ErrPasswordTooShort {
			code = "PASSWORD_TOO_SHORT"
		}
		middleware.WriteAppError(w, &middleware.AppError{Code: code, Message: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	admin, err := h.Repo.FindFirstRegistered()
	if err != nil {
		if err == user.ErrNotFound {
			middleware.WriteAppError(w, &middleware.AppError{Code: "ADMIN_NOT_FOUND", Message: "admin not found", StatusCode: http.StatusNotFound})
			return
		}
		middleware.WriteAppError(w, err)
		return
	}
	hash, err := h.Pass.Hash(req.NewPassword)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	now := time.Now().UTC()
	admin.PasswordHash = hash
	admin.UpdatedAt = &now
	if err := h.Repo.Update(admin); err != nil {
		middleware.WriteAppError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"password reset"}`))
}

func (h *Handler) validSuperToken(token string) bool {
	if h == nil || h.Config == nil {
		return false
	}
	expect := h.Config.Auth.JWTSecret
	if expect == "" || token == "" {
		return false
	}
	if len(expect) != len(token) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expect), []byte(token)) == 1
}
