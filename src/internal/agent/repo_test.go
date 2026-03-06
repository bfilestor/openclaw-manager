package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"openclaw-manager/internal/storage"
)

type fakeExec struct {
	out   []byte
	calls int
	err   error
}

func (f *fakeExec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.out, nil
}

func TestRepoListAndWorkspace(t *testing.T) {
	v, _ := storage.NewPathValidator([]string{"/tmp", "/home/mixi/.openclaw"})
	ex := &fakeExec{out: []byte(`{"agents":[{"id":"a1","workspace":"/tmp/w1","bindings":[1,2]},{"id":"a2","workspace":"/tmp/w2","bindings":[]}]}`)}
	r := NewRepository(ex, v)
	list, err := r.List(context.Background())
	if err != nil || len(list) != 2 || list[0].BindingsCount != 2 {
		t.Fatalf("bad list err=%v list=%+v", err, list)
	}
	p, err := r.GetWorkspacePath(context.Background(), "a1")
	if err != nil || p != "/tmp/w1" {
		t.Fatalf("bad workspace p=%s err=%v", p, err)
	}
}

func TestRepoTTLCacheAndInvalidID(t *testing.T) {
	v, _ := storage.NewPathValidator([]string{"/tmp"})
	ex := &fakeExec{out: []byte(`{"agents":[{"id":"a1","workspace":"/tmp/w1","bindings":[]}]}`)}
	r := NewRepository(ex, v)
	r.ttl = time.Hour
	_, _ = r.List(context.Background())
	_, _ = r.List(context.Background())
	if ex.calls != 1 {
		t.Fatalf("expected 1 call got %d", ex.calls)
	}
	if _, err := r.GetWorkspacePath(context.Background(), "../x"); !errors.Is(err, ErrInvalidAgentID) {
		t.Fatalf("expect invalid id err got %v", err)
	}
}

func TestRepoListNewCLIJSONShape(t *testing.T) {
	v, _ := storage.NewPathValidator([]string{"/tmp"})
	ex := &fakeExec{out: []byte(`[{"id":"a1","workspace":"/tmp/w1","bindings":3},{"id":"a2","workspace":"/tmp/w2","bindings":0}]`)}
	r := NewRepository(ex, v)
	list, err := r.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 agents got %d", len(list))
	}
	if list[0].BindingsCount != 3 || list[1].BindingsCount != 0 {
		t.Fatalf("unexpected bindings count: %+v", list)
	}
}
