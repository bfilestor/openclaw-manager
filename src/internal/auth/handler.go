package auth

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"openclaw-manager/internal/config"
	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/user"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

type Handler struct {
	Repo         *user.Repository
	Pass         *PasswordService
	Config       *config.Config
	JWT          *JWTService
	TokenRepo    *TokenRepository
	AccountBinds *AccountBindingRepository
	Settings     *SystemSettingsRepository
}

type registerReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func (h *Handler) PublicRegistrationStatus(w http.ResponseWriter, _ *http.Request) {
	enabled, _ := h.publicRegistrationEnabled()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if enabled {
		_, _ = w.Write([]byte(`{"public_registration":true}`))
		return
	}
	_, _ = w.Write([]byte(`{"public_registration":false}`))
}

func (h *Handler) GetSystemSettings(w http.ResponseWriter, r *http.Request) {
	uc, ok := GetUserContext(r.Context())
	if !ok {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	if roleWeight(uc.Role) < roleWeight(user.RoleAdmin) {
		middleware.WriteAppError(w, middleware.NewForbidden(string(user.RoleAdmin)))
		return
	}
	enabled, err := h.publicRegistrationEnabled()
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"public_registration":` + boolToJSON(enabled) + `}`))
}

func (h *Handler) PutSystemSettings(w http.ResponseWriter, r *http.Request) {
	uc, ok := GetUserContext(r.Context())
	if !ok {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	if roleWeight(uc.Role) < roleWeight(user.RoleAdmin) {
		middleware.WriteAppError(w, middleware.NewForbidden(string(user.RoleAdmin)))
		return
	}
	var req struct {
		PublicRegistration *bool `json:"public_registration"`
	}
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if req.PublicRegistration == nil {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"public_registration": "required"}))
		return
	}
	if h.Settings != nil {
		if err := h.Settings.SetPublicRegistrationEnabled(*req.PublicRegistration); err != nil {
			middleware.WriteAppError(w, err)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"public_registration":` + boolToJSON(*req.PublicRegistration) + `}`))
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	enabled, err := h.publicRegistrationEnabled()
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if !enabled {
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

	role := user.RoleUser
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

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}

	u, err := h.Repo.FindByUsername(req.Username)
	if err != nil || !h.Pass.Verify(req.Password, u.PasswordHash) {
		middleware.WriteAppError(w, &middleware.AppError{Code: "INVALID_CREDENTIALS", Message: "invalid credentials", StatusCode: http.StatusUnauthorized})
		return
	}
	if u.Status == user.StatusDisabled {
		middleware.WriteAppError(w, &middleware.AppError{Code: "ACCOUNT_DISABLED", Message: "account disabled", StatusCode: http.StatusForbidden})
		return
	}

	access, _, err := h.JWT.SignAccessToken(u.UserID, string(u.Role))
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	tokenID, rawRefresh, err := h.JWT.SignRefreshToken()
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	now := time.Now().UTC()
	exp := now.Add(h.JWT.RefreshTokenTTL)
	if err := h.TokenRepo.Save(&RefreshToken{TokenID: tokenID, UserID: u.UserID, TokenHash: HashToken(rawRefresh), ExpiresAt: exp, CreatedAt: now}); err != nil {
		middleware.WriteAppError(w, err)
		return
	}

	u.LastLoginAt = &now
	u.UpdatedAt = &now
	_ = h.Repo.Update(u)

	secureCookie := r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    rawRefresh,
		Path:     "/",
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteStrictMode,
		Expires:  exp,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"access_token":"` + access + `","expires_in":900,"token_type":"Bearer","user":{"user_id":"` + u.UserID + `","username":"` + u.Username + `","role":"` + string(u.Role) + `"}}`))
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}

	t, err := h.TokenRepo.FindByHash(HashToken(cookie.Value))
	if err != nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "TOKEN_INVALID", Message: "token invalid", StatusCode: http.StatusUnauthorized})
		return
	}
	if t.Revoked {
		middleware.WriteAppError(w, &middleware.AppError{Code: "TOKEN_REVOKED", Message: "token revoked", StatusCode: http.StatusUnauthorized})
		return
	}
	now := time.Now().UTC()
	if now.After(t.ExpiresAt) {
		middleware.WriteAppError(w, &middleware.AppError{Code: "TOKEN_EXPIRED", Message: "token expired", StatusCode: http.StatusUnauthorized})
		return
	}

	u, err := h.Repo.FindByID(t.UserID)
	if err != nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: "TOKEN_INVALID", Message: "token invalid", StatusCode: http.StatusUnauthorized})
		return
	}
	if u.Status == user.StatusDisabled {
		middleware.WriteAppError(w, &middleware.AppError{Code: "ACCOUNT_DISABLED", Message: "account disabled", StatusCode: http.StatusForbidden})
		return
	}

	access, _, err := h.JWT.SignAccessToken(u.UserID, string(u.Role))
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"access_token":"` + access + `","expires_in":900,"token_type":"Bearer"}`))
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token, err := extractBearerToken(r.Header.Get("Authorization"))
	if err != nil {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	claims, err := h.JWT.VerifyAccessToken(token)
	if err != nil {
		middleware.WriteAppError(w, middleware.NewUnauthorized())
		return
	}
	if claims.ID != "" {
		_ = h.TokenRepo.AddBlacklist(claims.ID, claims.ExpiresAt.Time)
	}

	if cookie, err := r.Cookie("refresh_token"); err == nil && cookie.Value != "" {
		if rt, findErr := h.TokenRepo.FindByHash(HashToken(cookie.Value)); findErr == nil {
			_ = h.TokenRepo.Revoke(rt.TokenID)
		}
	}

	secureCookie := r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"logged out"}`))
}

func (h *Handler) publicRegistrationEnabled() (bool, error) {
	if h.Settings != nil {
		return h.Settings.IsPublicRegistrationEnabled()
	}
	enabled := true
	if h.Config != nil {
		enabled = h.Config.Auth.PublicRegister
	}
	return enabled, nil
}

func boolToJSON(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
