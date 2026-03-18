//go:build linux

package platform

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeExec struct {
	fn func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func (f fakeExec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return f.fn(ctx, name, args...)
}

func TestLinuxServiceControllerStatus(t *testing.T) {
	c := NewDefaultServiceController(fakeExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name != "systemctl" {
			t.Fatalf("unexpected cmd: %s", name)
		}
		return []byte("ActiveState=active\nSubState=running\nMainPID=123\n"), nil
	}})

	st, err := c.Status("openclaw-gateway.service")
	if err != nil {
		t.Fatalf("status error: %v", err)
	}
	if st["ActiveState"] != "active" || st["SubState"] != "running" {
		t.Fatalf("unexpected status: %#v", st)
	}
}

func TestLinuxServiceControllerActionError(t *testing.T) {
	c := NewDefaultServiceController(fakeExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return []byte("boom"), errors.New("failed")
	}})
	if err := c.Start("openclaw-gateway.service"); err == nil || !strings.Contains(err.Error(), "systemctl start failed") {
		t.Fatalf("expected wrapped start error, got %v", err)
	}
}
