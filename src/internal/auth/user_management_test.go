package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"openclaw-manager/internal/config"
	"openclaw-manager/internal/storage"
	"openclaw-manager/internal/user"
)

func setupUserMgmt(t *testing.T) *Handler {
	db := storage.NewTestDB(t)
	repo := user.NewRepository(db.SQL)
	pass := NewPasswordService()
	hash, _ := pass.Hash("Pass1234")
	now := time.Now().UTC()
	admin := &user.User{UserID: "admin-1", Username: "admin", PasswordHash: hash, Role: user.RoleAdmin, Status: user.StatusActive, CreatedAt: now}
	viewer := &user.User{UserID: "viewer-1", Username: "viewer", PasswordHash: hash, Role: user.RoleViewer, Status: user.StatusActive, CreatedAt: now.Add(time.Second)}
	if err := repo.Create(admin); err != nil { t.Fatal(err) }
	if err := repo.Create(viewer); err != nil { t.Fatal(err) }
	tr := NewTokenRepository(db.SQL)
	jwtSvc := &JWTService{Secret: []byte("abcdefghijklmnopqrstuvwxyz123456"), AccessTokenTTL: time.Hour, RefreshTokenTTL: 7 * 24 * time.Hour, BlacklistChecker: tr}
	return &Handler{Repo: repo, Pass: pass, Config: &config.Config{}, JWT: jwtSvc, TokenRepo: tr}
}

func withUC(r *http.Request, uid string, role user.Role) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userContextKey, &UserContext{UserID: uid, Role: role}))
}

func TestMeAndChangePassword(t *testing.T) {
	h := setupUserMgmt(t)

	w1 := httptest.NewRecorder()
	h.Me(w1, withUC(httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil), "admin-1", user.RoleAdmin))
	if w1.Code != http.StatusOK || strings.Contains(w1.Body.String(), "password_hash") {
		t.Fatalf("me failed code=%d body=%s", w1.Code, w1.Body.String())
	}

	w2 := httptest.NewRecorder()
	h.ChangeMyPassword(w2, withUC(httptest.NewRequest(http.MethodPut, "/api/v1/users/me/password", strings.NewReader(`{"old_password":"bad","new_password":"NewPass123"}`)), "admin-1", user.RoleAdmin))
	if w2.Code != http.StatusBadRequest { t.Fatalf("expect 400 got %d", w2.Code) }

	w3 := httptest.NewRecorder()
	h.ChangeMyPassword(w3, withUC(httptest.NewRequest(http.MethodPut, "/api/v1/users/me/password", strings.NewReader(`{"old_password":"Pass1234","new_password":"NewPass123"}`)), "admin-1", user.RoleAdmin))
	if w3.Code != http.StatusOK { t.Fatalf("expect 200 got %d", w3.Code) }
}

func TestListUsersAndAdminOnly(t *testing.T) {
	h := setupUserMgmt(t)

	w1 := httptest.NewRecorder()
	h.ListUsers(w1, withUC(httptest.NewRequest(http.MethodGet, "/api/v1/users?limit=5&offset=0", nil), "viewer-1", user.RoleViewer))
	if w1.Code != http.StatusForbidden { t.Fatalf("viewer expect 403 got %d", w1.Code) }

	w2 := httptest.NewRecorder()
	h.ListUsers(w2, withUC(httptest.NewRequest(http.MethodGet, "/api/v1/users?limit=5&offset=0", nil), "admin-1", user.RoleAdmin))
	if w2.Code != http.StatusOK || !strings.Contains(w2.Body.String(), "\"users\"") {
		t.Fatalf("admin list failed code=%d body=%s", w2.Code, w2.Body.String())
	}
	if strings.Contains(w2.Body.String(), "password_hash") {
		t.Fatalf("must not expose password hash: %s", w2.Body.String())
	}
}

func TestUpdateRoleDeleteDisableGuards(t *testing.T) {
	h := setupUserMgmt(t)

	// cannot modify self
	w1 := httptest.NewRecorder()
	h.UpdateUserRole(w1, withUC(httptest.NewRequest(http.MethodPut, "/api/v1/users/admin-1/role", strings.NewReader(`{"role":"Operator"}`)), "admin-1", user.RoleAdmin))
	if w1.Code != http.StatusBadRequest { t.Fatalf("expect 400 got %d", w1.Code) }

	// update other role
	w2 := httptest.NewRecorder()
	h.UpdateUserRole(w2, withUC(httptest.NewRequest(http.MethodPut, "/api/v1/users/viewer-1/role", strings.NewReader(`{"role":"Operator"}`)), "admin-1", user.RoleAdmin))
	if w2.Code != http.StatusOK { t.Fatalf("expect 200 got %d body=%s", w2.Code, w2.Body.String()) }

	// cannot disable self
	w3 := httptest.NewRecorder()
	h.DisableUser(w3, withUC(httptest.NewRequest(http.MethodPost, "/api/v1/users/admin-1/disable", strings.NewReader(`{"disabled":true}`)), "admin-1", user.RoleAdmin))
	if w3.Code != http.StatusBadRequest { t.Fatalf("expect 400 got %d", w3.Code) }

	// disable other
	w4 := httptest.NewRecorder()
	h.DisableUser(w4, withUC(httptest.NewRequest(http.MethodPost, "/api/v1/users/viewer-1/disable", strings.NewReader(`{"disabled":true}`)), "admin-1", user.RoleAdmin))
	if w4.Code != http.StatusOK { t.Fatalf("expect 200 got %d", w4.Code) }

	// cannot delete self
	w5 := httptest.NewRecorder()
	h.DeleteUser(w5, withUC(httptest.NewRequest(http.MethodDelete, "/api/v1/users/admin-1", nil), "admin-1", user.RoleAdmin))
	if w5.Code != http.StatusBadRequest { t.Fatalf("expect 400 got %d", w5.Code) }
}
