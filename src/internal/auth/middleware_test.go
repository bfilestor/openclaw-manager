package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"openclaw-manager/internal/user"
)

func TestAuthMiddlewareAndRequireRole(t *testing.T) {
	tr := &mockBlacklist{}
	jwtSvc := &JWTService{Secret: []byte("abcdefghijklmnopqrstuvwxyz123456"), AccessTokenTTL: time.Hour, RefreshTokenTTL: time.Hour, BlacklistChecker: tr}
	token, _, err := jwtSvc.SignAccessToken("u1", string(user.RoleViewer))
	if err != nil {
		t.Fatal(err)
	}

	called := false
	h := AuthMiddleware(jwtSvc)(RequireRole(user.RoleViewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		uc, ok := GetUserContext(r.Context())
		if !ok || uc.UserID != "u1" || uc.Role != user.RoleViewer {
			t.Fatalf("bad user context: %+v", uc)
		}
		w.WriteHeader(http.StatusOK)
	})))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if !called || w.Code != http.StatusOK {
		t.Fatalf("expect called and 200, called=%v code=%d", called, w.Code)
	}
}

func TestAuthMiddlewareFailures(t *testing.T) {
	tr := &mockBlacklist{revoked: map[string]bool{}}
	jwtSvc := &JWTService{Secret: []byte("abcdefghijklmnopqrstuvwxyz123456"), AccessTokenTTL: time.Hour, RefreshTokenTTL: time.Hour, BlacklistChecker: tr}
	h := AuthMiddleware(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))

	// missing header
	w1 := httptest.NewRecorder()
	h.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/", nil))
	if w1.Code != http.StatusUnauthorized { t.Fatalf("expect 401 got %d", w1.Code) }

	// bad format
	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	r2.Header.Set("Authorization", "Token abc")
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, r2)
	if w2.Code != http.StatusUnauthorized { t.Fatalf("expect 401 got %d", w2.Code) }

	// revoked
	token, jti, _ := jwtSvc.SignAccessToken("u1", string(user.RoleViewer))
	tr.revoked[jti] = true
	r3 := httptest.NewRequest(http.MethodGet, "/", nil)
	r3.Header.Set("Authorization", "Bearer "+token)
	w3 := httptest.NewRecorder()
	h.ServeHTTP(w3, r3)
	if w3.Code != http.StatusUnauthorized { t.Fatalf("expect 401 got %d", w3.Code) }
}

func TestRequireRole(t *testing.T) {
	base := RequireRole(user.RoleOperator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// missing user context
	w1 := httptest.NewRecorder()
	base.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/", nil))
	if w1.Code != http.StatusUnauthorized { t.Fatalf("missing ctx expect 401 got %d", w1.Code) }

	// viewer forbidden
	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	r2 = r2.WithContext(context.WithValue(r2.Context(), userContextKey, &UserContext{UserID: "u1", Role: user.RoleViewer}))
	w2 := httptest.NewRecorder()
	base.ServeHTTP(w2, r2)
	if w2.Code != http.StatusForbidden { t.Fatalf("viewer expect 403 got %d", w2.Code) }

	// admin pass
	r3 := httptest.NewRequest(http.MethodGet, "/", nil)
	r3 = r3.WithContext(context.WithValue(r3.Context(), userContextKey, &UserContext{UserID: "u2", Role: user.RoleAdmin}))
	w3 := httptest.NewRecorder()
	base.ServeHTTP(w3, r3)
	if w3.Code != http.StatusOK { t.Fatalf("admin expect 200 got %d", w3.Code) }
}

type mockBlacklist struct { revoked map[string]bool }

func (m *mockBlacklist) ExistsJTI(jti string) (bool, error) {
	if m.revoked == nil { m.revoked = map[string]bool{} }
	return m.revoked[jti], nil
}
