package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	WriteAppError(w, NewUnauthorized())

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"code":"AUTH_REQUIRED"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestWriteForbiddenWithRequiredRole(t *testing.T) {
	w := httptest.NewRecorder()
	WriteAppError(w, NewForbidden("Operator"))

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"required_role":"Operator"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestWriteValidationFields(t *testing.T) {
	w := httptest.NewRecorder()
	WriteAppError(w, NewValidation(map[string]string{"username": "required"}))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"username":"required"`) {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestWriteInternalForGenericError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteAppError(w, errors.New("db down"))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "db down") {
		t.Fatalf("should not leak internal error, got: %s", w.Body.String())
	}
}
