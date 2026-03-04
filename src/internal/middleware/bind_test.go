package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"
)

type loginReq struct {
	Username string `json:"username"`
}

func TestBindJSONOK(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"username":"alice"}`))
	var req loginReq
	if err := BindJSON(r, &req); err != nil {
		t.Fatalf("bind failed: %v", err)
	}
	if req.Username != "alice" {
		t.Fatalf("username mismatch: %s", req.Username)
	}
}

func TestBindJSONEmptyBody(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(""))
	var req loginReq
	err := BindJSON(r, &req)
	if err == nil || err.Code != CodeValidationError {
		t.Fatalf("expected validation error, got: %+v", err)
	}
}

func TestBindJSONUnknownField(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"username":"alice","x":1}`))
	var req loginReq
	err := BindJSON(r, &req)
	if err == nil || err.Code != CodeValidationError {
		t.Fatalf("expected validation error, got: %+v", err)
	}
}
