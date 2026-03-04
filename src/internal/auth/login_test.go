package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"openclaw-manager/internal/config"
	"openclaw-manager/internal/storage"
	"openclaw-manager/internal/user"
)

func setupLogin(t *testing.T) *Handler {
	db := storage.NewTestDB(t)
	repo := user.NewRepository(db.SQL)
	pass := NewPasswordService()
	hash, _ := pass.Hash("Pass1234")
	u := &user.User{UserID: "u1", Username: "alice", PasswordHash: hash, Role: user.RoleOperator, Status: user.StatusActive, CreatedAt: time.Now().UTC()}
	if err := repo.Create(u); err != nil {
		t.Fatal(err)
	}
	tr := NewTokenRepository(db.SQL)
	jwtSvc := &JWTService{Secret: []byte("abcdefghijklmnopqrstuvwxyz123456"), AccessTokenTTL: 15 * time.Minute, RefreshTokenTTL: 7 * 24 * time.Hour, BlacklistChecker: tr}
	return &Handler{Repo: repo, Pass: pass, Config: &config.Config{}, JWT: jwtSvc, TokenRepo: tr}
}

func TestLoginSuccess(t *testing.T) {
	h := setupLogin(t)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"alice","password":"Pass1234"}`))
	w := httptest.NewRecorder()
	h.Login(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expect 200 got %d body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "refresh_token") {
		t.Fatalf("response should not include refresh token: %s", w.Body.String())
	}
	cookies := w.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != "refresh_token" || !cookies[0].HttpOnly {
		t.Fatalf("refresh cookie missing or invalid: %+v", cookies)
	}
}

func TestLoginInvalidOrDisabled(t *testing.T) {
	h := setupLogin(t)
	w1 := httptest.NewRecorder()
	h.Login(w1, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"username":"alice","password":"bad"}`)))
	if w1.Code != http.StatusUnauthorized { t.Fatalf("expect 401 got %d", w1.Code) }

	u, _ := h.Repo.FindByUsername("alice")
	u.Status = user.StatusDisabled
	now := time.Now().UTC()
	u.UpdatedAt = &now
	_ = h.Repo.Update(u)

	w2 := httptest.NewRecorder()
	h.Login(w2, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"username":"alice","password":"Pass1234"}`)))
	if w2.Code != http.StatusForbidden { t.Fatalf("expect 403 got %d", w2.Code) }
}
