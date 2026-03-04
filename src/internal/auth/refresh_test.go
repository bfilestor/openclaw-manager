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

func setupRefresh(t *testing.T) (*Handler, string) {
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

	tokenID, raw, err := jwtSvc.SignRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := tr.Save(&RefreshToken{TokenID: tokenID, UserID: u.UserID, TokenHash: HashToken(raw), ExpiresAt: now.Add(jwtSvc.RefreshTokenTTL), CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	return h, raw
}

func TestRefreshSuccess(t *testing.T) {
	h, raw := setupRefresh(t)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	r.AddCookie(&http.Cookie{Name: "refresh_token", Value: raw})
	w := httptest.NewRecorder()

	h.Refresh(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expect 200 got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "access_token") {
		t.Fatalf("response should include access_token: %s", w.Body.String())
	}
}

func TestRefreshInvalidRevokedExpiredAndDisabled(t *testing.T) {
	h, _ := setupRefresh(t)

	w1 := httptest.NewRecorder()
	h.Refresh(w1, httptest.NewRequest(http.MethodPost, "/", nil))
	if w1.Code != http.StatusUnauthorized { t.Fatalf("missing cookie expect 401 got %d", w1.Code) }

	// invalid
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodPost, "/", nil)
	r2.AddCookie(&http.Cookie{Name: "refresh_token", Value: "fake"})
	h.Refresh(w2, r2)
	if w2.Code != http.StatusUnauthorized { t.Fatalf("invalid expect 401 got %d", w2.Code) }

	// revoked
	tokenID, raw, _ := h.JWT.SignRefreshToken()
	now := time.Now().UTC()
	_ = h.TokenRepo.Save(&RefreshToken{TokenID: tokenID, UserID: "u1", TokenHash: HashToken(raw), ExpiresAt: now.Add(time.Hour), CreatedAt: now, Revoked: true})
	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest(http.MethodPost, "/", nil)
	r3.AddCookie(&http.Cookie{Name: "refresh_token", Value: raw})
	h.Refresh(w3, r3)
	if w3.Code != http.StatusUnauthorized { t.Fatalf("revoked expect 401 got %d", w3.Code) }

	// expired
	tokenID2, raw2, _ := h.JWT.SignRefreshToken()
	_ = h.TokenRepo.Save(&RefreshToken{TokenID: tokenID2, UserID: "u1", TokenHash: HashToken(raw2), ExpiresAt: now.Add(-time.Hour), CreatedAt: now})
	w4 := httptest.NewRecorder()
	r4 := httptest.NewRequest(http.MethodPost, "/", nil)
	r4.AddCookie(&http.Cookie{Name: "refresh_token", Value: raw2})
	h.Refresh(w4, r4)
	if w4.Code != http.StatusUnauthorized { t.Fatalf("expired expect 401 got %d", w4.Code) }

	// disabled user
	tokenID3, raw3, _ := h.JWT.SignRefreshToken()
	_ = h.TokenRepo.Save(&RefreshToken{TokenID: tokenID3, UserID: "u1", TokenHash: HashToken(raw3), ExpiresAt: now.Add(time.Hour), CreatedAt: now})
	u, _ := h.Repo.FindByID("u1")
	u.Status = user.StatusDisabled
	tm := time.Now().UTC()
	u.UpdatedAt = &tm
	_ = h.Repo.Update(u)
	w5 := httptest.NewRecorder()
	r5 := httptest.NewRequest(http.MethodPost, "/", nil)
	r5.AddCookie(&http.Cookie{Name: "refresh_token", Value: raw3})
	h.Refresh(w5, r5)
	if w5.Code != http.StatusForbidden { t.Fatalf("disabled expect 403 got %d", w5.Code) }
}
