package task

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeShellExec struct {
	fn func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func (f fakeShellExec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return f.fn(ctx, name, args...)
}

type fakeExitError struct {
	code int
	msg  string
}

func (e fakeExitError) Error() string {
	if e.msg != "" {
		return e.msg
	}
	return fmt.Sprintf("exit code %d", e.code)
}

func (e fakeExitError) ExitCode() int {
	return e.code
}

func TestShellHandlerExecuteSuccess(t *testing.T) {
	h := NewShellHandler(fakeShellExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name != "openclaw" {
			t.Fatalf("unexpected name: %s", name)
		}
		joined := strings.Join(args, " ")
		if joined != `channels add --channel qqbot --token guoqiang:sjdkdjfdkfjdkf` {
			t.Fatalf("unexpected args: %s", joined)
		}
		return []byte("ok\nline2"), nil
	}})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/shell/execute", strings.NewReader(`{"command":"openclaw channels add --channel qqbot --token \"guoqiang:sjdkdjfdkfjdkf\""}`))
	w := httptest.NewRecorder()
	h.Execute(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expect 200 got %d body=%s", w.Code, w.Body.String())
	}

	var resp executeShellResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.Success || resp.ExitCode != 0 {
		t.Fatalf("expect success exit=0 got success=%v exit=%d body=%s", resp.Success, resp.ExitCode, w.Body.String())
	}
	if !strings.Contains(resp.Output, "line2") {
		t.Fatalf("expect output contains line2, body=%s", w.Body.String())
	}
}

func TestShellHandlerExecuteValidation(t *testing.T) {
	h := NewShellHandler(fakeShellExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		t.Fatal("executor should not be called")
		return nil, nil
	}})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/shell/execute", strings.NewReader(`{"command":"ls -la"}`))
	w := httptest.NewRecorder()
	h.Execute(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expect 400 got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "must start with openclaw") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestShellHandlerExecuteFailedCommand(t *testing.T) {
	h := NewShellHandler(fakeShellExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return []byte("failed output"), fakeExitError{code: 2, msg: "exit status 2"}
	}})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/shell/execute", strings.NewReader(`{"command":"openclaw gateway restart"}`))
	w := httptest.NewRecorder()
	h.Execute(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expect 200 got %d body=%s", w.Code, w.Body.String())
	}

	var resp executeShellResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Success {
		t.Fatalf("expect failed command")
	}
	if resp.ExitCode != 2 {
		t.Fatalf("expect exit code 2 got %d body=%s", resp.ExitCode, w.Body.String())
	}
	if !strings.Contains(resp.Error, "exit status 2") {
		t.Fatalf("unexpected error body=%s", w.Body.String())
	}
}

func TestShellHandlerExecuteTimeout(t *testing.T) {
	h := NewShellHandler(fakeShellExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		<-ctx.Done()
		return []byte("timeout output"), ctx.Err()
	}})
	h.Timeout = 10 * time.Millisecond

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/shell/execute", strings.NewReader(`{"command":"openclaw gateway restart"}`))
	w := httptest.NewRecorder()
	h.Execute(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expect 200 got %d body=%s", w.Code, w.Body.String())
	}
	var resp executeShellResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Success || resp.ExitCode != -1 {
		t.Fatalf("expect timeout failed with exit -1 got success=%v exit=%d body=%s", resp.Success, resp.ExitCode, w.Body.String())
	}
	if resp.Error != "command timeout" {
		t.Fatalf("unexpected timeout error: %s", resp.Error)
	}
}

func TestSplitCommandLine(t *testing.T) {
	args, err := splitCommandLine(`openclaw channels add --channel qqbot --token "abc:def" --name 'bot 2'`)
	if err != nil {
		t.Fatalf("split error: %v", err)
	}
	got := strings.Join(args, "|")
	want := "openclaw|channels|add|--channel|qqbot|--token|abc:def|--name|bot 2"
	if got != want {
		t.Fatalf("split mismatch got=%s want=%s", got, want)
	}
}
