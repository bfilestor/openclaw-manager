package auth

import (
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"

	"openclaw-manager/internal/config"
	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/user"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

type Handler struct {
	Repo    *user.Repository
	Pass    *PasswordService
	Config  *config.Config
}

type registerReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if h.Config != nil && !h.Config.Auth.PublicRegister {
		middleware.WriteAppError(w, &middleware.AppError{Code: "REGISTRATION_DISABLED", Message: "registration disabled", StatusCode: http.StatusForbidden})
		return
	}

	var req registerReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if !usernameRe.MatchString(req.Username) {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"username": "must be 3-32 letters/numbers/_"}))
		return
	}
	if len(req.Password) > 128 {
		middleware.WriteAppError(w, &middleware.AppError{Code: "PASSWORD_TOO_LONG", Message: "password too long", StatusCode: http.StatusBadRequest})
		return
	}
	if err := h.Pass.ValidateStrength(req.Password); err != nil {
		code := "PASSWORD_WEAK"
		if errors.Is(err, ErrPasswordTooShort) {
			code = "PASSWORD_TOO_SHORT"
		}
		middleware.WriteAppError(w, &middleware.AppError{Code: code, Message: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}

	if _, err := h.Repo.FindByUsername(req.Username); err == nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "USERNAME_EXISTS", Message: "username exists", StatusCode: http.StatusConflict})
		return
	} else if !errors.Is(err, user.ErrNotFound) {
		middleware.WriteAppError(w, err)
		return
	}

	role := user.RoleViewer
	cnt, err := h.Repo.Count()
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if cnt == 0 {
		role = user.RoleAdmin
	}
	hash, err := h.Pass.Hash(req.Password)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	now := time.Now().UTC()
	u := &user.User{
		UserID:       uuid.NewString(),
		Username:     req.Username,
		PasswordHash: hash,
		Role:         role,
		Status:       user.StatusActive,
		CreatedAt:    now,
	}
	if err := h.Repo.Create(u); err != nil {
		middleware.WriteAppError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(`{"user_id":"` + u.UserID + `","username":"` + u.Username + `","role":"` + string(u.Role) + `","created_at":"` + u.CreatedAt.Format(time.RFC3339) + `"}`))
}
