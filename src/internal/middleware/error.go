package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
)

const (
	CodeAuthRequired     = "AUTH_REQUIRED"
	CodePermissionDenied = "PERMISSION_DENIED"
	CodeNotFound         = "NOT_FOUND"
	CodeValidationError  = "VALIDATION_ERROR"
	CodeConflict         = "CONFLICT"
	CodeInternalError    = "INTERNAL_ERROR"
)

type AppError struct {
	Code         string            `json:"code"`
	Message      string            `json:"message"`
	StatusCode   int               `json:"status_code"`
	RequiredRole string            `json:"required_role,omitempty"`
	Fields       map[string]string `json:"fields,omitempty"`
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func WriteAppError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var appErr *AppError
	if errors.As(err, &appErr) {
		w.WriteHeader(appErr.StatusCode)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":         appErr.Message,
			"code":          appErr.Code,
			"required_role": appErr.RequiredRole,
			"fields":        appErr.Fields,
		})
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":  "internal server error",
		"code":   CodeInternalError,
		"detail": errString(err),
		"where":  callerLocation(2),
	})
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func callerLocation(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

func NewUnauthorized() *AppError {
	return &AppError{Code: CodeAuthRequired, Message: "unauthorized", StatusCode: http.StatusUnauthorized}
}

func NewForbidden(requiredRole string) *AppError {
	return &AppError{Code: CodePermissionDenied, Message: "forbidden", StatusCode: http.StatusForbidden, RequiredRole: requiredRole}
}

func NewValidation(fields map[string]string) *AppError {
	return &AppError{Code: CodeValidationError, Message: "validation failed", StatusCode: http.StatusBadRequest, Fields: fields}
}
