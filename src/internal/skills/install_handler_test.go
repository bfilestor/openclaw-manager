package skills

import (
	"archive/zip"
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"openclaw-manager/internal/agent"
)

func makeZip(t *testing.T) []byte {
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	f, _ := zw.Create("README.md")
	_, _ = f.Write([]byte("hello"))
	_ = zw.Close()
	return buf.Bytes()
}

type iexec struct{}

func (iexec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return []byte(`{"agents":[{"id":"a1","workspace":"/tmp/agent-install","bindings":[]}]}`), nil
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

func TestInstallSkillAgentScope(t *testing.T) {
	t.Skip("temporarily disabled due to environment-sensitive conflict")
	_ = os.MkdirAll("/tmp/agent-install/skills", 0o755)
	h := &InstallHandler{AgentRepo: agent.NewRepository(iexec{}, nil)}

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	_ = mw.WriteField("scope", "agent")
	_ = mw.WriteField("agent_id", "a1")
	_ = mw.WriteField("skill_name", "s2")
	fw, _ := mw.CreateFormFile("file", "s2.zip")
	_, _ = fw.Write(makeZip(t))
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/install", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	h.InstallSkill(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expect 202 got %d body=%s", w.Code, w.Body.String())
	}
	if _, err := os.Stat("/tmp/agent-install/skills/s2/README.md"); err != nil {
		t.Fatalf("agent skill not installed: %v", err)
	}
}

func TestInstallSkillUnsupportedFormat(t *testing.T) {
	d := t.TempDir()
	h := &InstallHandler{GlobalDir: d}

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	_ = mw.WriteField("scope", "global")
	_ = mw.WriteField("skill_name", "bad")
	fw, _ := mw.CreateFormFile("file", "bad.rar")
	_, _ = fw.Write([]byte("not archive"))
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/install", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	h.InstallSkill(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expect 400 got %d body=%s", w.Code, w.Body.String())
	}
}
