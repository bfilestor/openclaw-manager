package middleware

import (
	"encoding/json"
	"io"
	"net/http"
)

// BindJSON 统一 JSON 绑定入口：未知字段报错、空体报错。
func BindJSON(r *http.Request, out any) *AppError {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(out); err != nil {
		if err == io.EOF {
			return NewValidation(map[string]string{"body": "required"})
		}
		return NewValidation(map[string]string{"body": err.Error()})
	}

	if dec.More() {
		return NewValidation(map[string]string{"body": "multiple json objects not allowed"})
	}
	return nil
}
