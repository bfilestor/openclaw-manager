package skills

import (
	"archive/zip"
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func makeZip(t *testing.T) []byte {
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	f, _ := zw.Create("README.md")
	_, _ = f.Write([]byte("hello"))
	_ = zw.Close()
	return buf.Bytes()
}

func TestInstallSkill(t *testing.T) {
	d := t.TempDir()
	h := &InstallHandler{GlobalDir: d}

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	_ = mw.WriteField("scope", "global")
	_ = mw.WriteField("skill_name", "s1")
	fw, _ := mw.CreateFormFile("file", "s1.zip")
	_, _ = fw.Write(makeZip(t))
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/install", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	h.InstallSkill(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expect 202 got %d body=%s", w.Code, w.Body.String())
	}
	if _, err := os.Stat(filepath.Join(d, "s1", "README.md")); err != nil {
		t.Fatalf("skill not installed: %v", err)
	}
}
