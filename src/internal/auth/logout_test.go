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

func setupLogout(t *testing.T) (*Handler, string, string, string) {
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
	h := &Handler{Repo: repo, Pass: pass, Config: &config.Config{}, JWT: jwtSvc, TokenRepo: tr}

	access, jti, err := jwtSvc.SignAccessToken(u.UserID, string(u.Role))
	if err != nil {
		t.Fatal(err)
	}
	tokenID, raw, err := jwtSvc.SignRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := tr.Save(&RefreshToken{TokenID: tokenID, UserID: u.UserID, TokenHash: HashToken(raw), ExpiresAt: now.Add(jwtSvc.RefreshTokenTTL), CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	return h, access, raw, jti
}

func TestLogoutSuccess(t *testing.T) {
	h, access, raw, jti := setupLogout(t)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	r.Header.Set("Authorization", "Bearer "+access)
	r.AddCookie(&http.Cookie{Name: "refresh_token", Value: raw})
	w := httptest.NewRecorder()

	h.Logout(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expect 200 got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "logged out") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
	cookies := w.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != "refresh_token" || !cookies[0].Expires.Before(time.Now().Add(2*time.Second)) {
		t.Fatalf("refresh cookie not cleared: %+v", cookies)
	}

	revoked, err := h.TokenRepo.ExistsJTI(jti)
	if err != nil || !revoked {
		t.Fatalf("expect jti revoked, revoked=%v err=%v", revoked, err)
	}

	rt, err := h.TokenRepo.FindByHash(HashToken(raw))
	if err != nil || !rt.Revoked {
		t.Fatalf("expect refresh revoked, token=%+v err=%v", rt, err)
	}
}

func TestLogoutWithoutToken(t *testing.T) {
	h, _, _, _ := setupLogout(t)
	w := httptest.NewRecorder()
	h.Logout(w, httptest.NewRequest(http.MethodPost, "/", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expect 401 got %d", w.Code)
	}
}
