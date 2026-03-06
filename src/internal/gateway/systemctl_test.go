package gateway

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeExec struct {
	out []byte
	err error
	fn  func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func (f *fakeExec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	if f.fn != nil {
		return f.fn(ctx, name, args...)
	}
	return f.out, f.err
}

func TestStatusParse(t *testing.T) {
	ex := &fakeExec{out: []byte("ActiveState=active\nSubState=running\nMainPID=123\nFragmentPath=/x.service\n")}
	s := NewSystemctlService(ex)
	st, err := s.Status("openclaw-gateway.service")
	if err != nil {
		t.Fatal(err)
	}
	if st.ActiveState != "active" || st.SubState != "running" || st.MainPID != "123" {
		t.Fatalf("bad status: %+v", st)
	}
}

func TestInvalidServiceName(t *testing.T) {
	s := NewSystemctlService(&fakeExec{})
	if err := s.Start("../evil"); !errors.Is(err, ErrInvalidServiceName) {
		t.Fatalf("expect invalid service, got %v", err)
	}
}

func TestTimeout(t *testing.T) {
	ex := &fakeExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}}
	s := NewSystemctlService(ex)
	s.timeout = 10 * time.Millisecond
	if err := s.Restart("openclaw-gateway.service"); !errors.Is(err, ErrCommandTimeout) {
		t.Fatalf("expect timeout err, got %v", err)
	}
}

func TestDeepStatusNVMWarningTrue(t *testing.T) {
	ex := &fakeExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		switch name {
		case "systemctl":
			return []byte("ActiveState=active\nSubState=running\nMainPID=123\n"), nil
		case "openclaw":
			return []byte("bind=127.0.0.1:18789\nlog_path=/tmp/openclaw/openclaw.log\nnode_path=/home/mixi/.nvm/versions/node/v20/bin/node\n"), nil
		default:
			return nil, errors.New("unexpected command")
		}
	}}

	s := NewSystemctlService(ex)
	st, err := s.DeepStatus("openclaw-gateway.service")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if st == nil || st.Service == nil {
		t.Fatalf("expect service status, got %+v", st)
	}
	if !st.NVMWarning {
		t.Fatalf("expect nvm warning true, got false")
	}
}

func TestDeepStatusNVMWarningFalse(t *testing.T) {
	ex := &fakeExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		switch name {
		case "systemctl":
			return []byte("ActiveState=active\nSubState=running\n"), nil
		case "openclaw":
			return []byte("node_path=/usr/bin/node\n"), nil
		default:
			return nil, errors.New("unexpected command")
		}
	}}

	s := NewSystemctlService(ex)
	st, err := s.DeepStatus("openclaw-gateway.service")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if st.NVMWarning {
		t.Fatalf("expect nvm warning false, got true")
	}
}

func TestDeepStatusOpenclawTimeoutKeepsSystemctlResult(t *testing.T) {
	ex := &fakeExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		switch name {
		case "systemctl":
			return []byte("ActiveState=active\nSubState=running\nMainPID=999\n"), nil
		case "openclaw":
			<-ctx.Done()
			return nil, ctx.Err()
		default:
			return nil, errors.New("unexpected command")
		}
	}}

	s := NewSystemctlService(ex)
	s.timeout = 10 * time.Millisecond
	st, err := s.DeepStatus("openclaw-gateway.service")
	if !errors.Is(err, ErrCommandTimeout) {
		t.Fatalf("expect timeout err, got %v", err)
	}
	if st == nil || st.Service == nil {
		t.Fatalf("expect partial systemctl result, got %+v", st)
	}
	if st.Service.ActiveState != "active" || st.Service.MainPID != "999" {
		t.Fatalf("unexpected systemctl result: %+v", st.Service)
	}
}

func TestDeepStatusParseBindAddress(t *testing.T) {
	ex := &fakeExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		switch name {
		case "systemctl":
			return []byte("ActiveState=active\nSubState=running\n"), nil
		case "openclaw":
			return []byte("bind=127.0.0.1:18789\n"), nil
		default:
			return nil, errors.New("unexpected command")
		}
	}}

	s := NewSystemctlService(ex)
	st, err := s.DeepStatus("openclaw-gateway.service")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if st.BindAddr != "127.0.0.1" || st.Port != 18789 {
		t.Fatalf("unexpected bind parse result: bind=%s port=%d", st.BindAddr, st.Port)
	}
}

func TestDeepStatusParseHumanReadableOutput(t *testing.T) {
	ex := &fakeExec{fn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		switch name {
		case "systemctl":
			return []byte("ActiveState=active\nSubState=running\n"), nil
		case "openclaw":
			return []byte("File logs: /tmp/openclaw/openclaw.log\nCommand: /home/mixi/.nvm/versions/node/v24.14.0/bin/node /x\nGateway: bind=loopback (127.0.0.1), port=18789 (service args)\nListening: 127.0.0.1:18789\n"), nil
		default:
			return nil, errors.New("unexpected command")
		}
	}}

	s := NewSystemctlService(ex)
	st, err := s.DeepStatus("openclaw-gateway.service")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if st.BindAddr != "127.0.0.1" || st.Port != 18789 {
		t.Fatalf("unexpected bind parse result: bind=%s port=%d", st.BindAddr, st.Port)
	}
	if st.LogPath != "/tmp/openclaw/openclaw.log" {
		t.Fatalf("unexpected log path: %s", st.LogPath)
	}
	if st.NodePath == "" || !st.NVMWarning {
		t.Fatalf("unexpected node parse: node=%s nvm=%v", st.NodePath, st.NVMWarning)
	}
}

func TestParseBindAddressPlaceholder(t *testing.T) {
	addr, port := parseBindAddress("-:-")
	if addr != "" || port != 0 {
		t.Fatalf("expected empty placeholder parse, got %q:%d", addr, port)
	}
}
