package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"openclaw-manager/internal/config"
	"openclaw-manager/internal/storage"
	"openclaw-manager/internal/user"
)

func setupRegisterHandler(t *testing.T, public bool) *Handler {
	db := storage.NewTestDB(t)
	return &Handler{
		Repo:   user.NewRepository(db.SQL),
		Pass:   NewPasswordService(),
		Config: &config.Config{Auth: config.AuthConfig{PublicRegister: public}},
	}
}

func doRegister(h *Handler, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.Register(w, r)
	return w
}

func TestRegisterFirstAdminThenUser(t *testing.T) {
	h := setupRegisterHandler(t, true)
	w1 := doRegister(h, `{"username":"alice","password":"Pass1234"}`)
	if w1.Code != http.StatusCreated {
		t.Fatalf("expect 201 got %d body=%s", w1.Code, w1.Body.String())
	}
	if !strings.Contains(w1.Body.String(), `"role":"Admin"`) {
		t.Fatalf("first user should be admin: %s", w1.Body.String())
	}
	w2 := doRegister(h, `{"username":"bob","password":"Pass1234"}`)
	if !strings.Contains(w2.Body.String(), `"role":"User"`) {
		t.Fatalf("second user should be user: %s", w2.Body.String())
	}
}

func TestRegisterValidationAndConflict(t *testing.T) {
	h := setupRegisterHandler(t, true)
	if code := doRegister(h, `{"username":"ab","password":"Pass1234"}`).Code; code != http.StatusBadRequest {
		t.Fatalf("short username expect 400 got %d", code)
	}
	if code := doRegister(h, `{"username":"alice!","password":"Pass1234"}`).Code; code != http.StatusBadRequest {
		t.Fatalf("bad username expect 400 got %d", code)
	}
	if code := doRegister(h, `{"username":"alice","password":"12345678"}`).Code; code != http.StatusBadRequest {
		t.Fatalf("weak pwd expect 400 got %d", code)
	}
	_ = doRegister(h, `{"username":"alice","password":"Pass1234"}`)
	if code := doRegister(h, `{"username":"alice","password":"Pass1234"}`).Code; code != http.StatusConflict {
		t.Fatalf("duplicate expect 409 got %d", code)
	}
	if code := doRegister(h, `{}`).Code; code != http.StatusBadRequest {
		t.Fatalf("empty body expect 400 got %d", code)
	}
	longPwd := strings.Repeat("a", 1001)
	if code := doRegister(h, `{"username":"u1000","password":"`+longPwd+`"}`).Code; code != http.StatusBadRequest {
		t.Fatalf("long password expect 400 got %d", code)
	}
}

func TestRegisterDisabled(t *testing.T) {
	h := setupRegisterHandler(t, false)
	w := doRegister(h, `{"username":"alice","password":"Pass1234"}`)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expect 403 got %d", w.Code)
	}
	var m map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &m)
	if m["code"] != "REGISTRATION_DISABLED" {
		t.Fatalf("unexpected code: %v", m["code"])
	}
}
