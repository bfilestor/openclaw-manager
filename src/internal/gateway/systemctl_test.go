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
