package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"openclaw-manager/internal/storage"
)

func setupEngineRepo(t *testing.T) *Repository {
	db := storage.NewTestDB(t)
	_, _ = db.SQL.Exec(`INSERT INTO users(user_id,username,password_hash,role,status,created_at) VALUES(?,?,?,?,?,?)`, "u1", "u1", "x", "Viewer", "active", time.Now().UTC().Format(time.RFC3339))
	return NewRepository(db.SQL)
}

func TestEngineRunSuccessAndFailAndTimeout(t *testing.T) {
	r := setupEngineRepo(t)
	e := NewEngine(r, 2)
	e.Register("ok", time.Second, func(ctx context.Context, task *Task) (string, string, int, error) {
		return "done", "", 0, nil
	})
	e.Register("bad", time.Second, func(ctx context.Context, task *Task) (string, string, int, error) {
		return "", "boom", 1, errors.New("boom")
	})
	e.Register("slow", 10*time.Millisecond, func(ctx context.Context, task *Task) (string, string, int, error) {
		<-ctx.Done()
		return "", "timeout", -1, ctx.Err()
	})
	e.Start()
	defer e.Stop()

	mk := func(tp string) string {
		id := uuid.NewString()
		_ = r.Create(&Task{TaskID: id, TaskType: tp, Status: StatusPending, CreatedBy: "u1", CreatedAt: time.Now().UTC()})
		e.Enqueue(id)
		return id
	}
	id1 := mk("ok")
	id2 := mk("bad")
	id3 := mk("slow")
	time.Sleep(80 * time.Millisecond)

	t1, _ := r.FindByID(id1)
	t2, _ := r.FindByID(id2)
	t3, _ := r.FindByID(id3)
	if t1.Status != StatusSucceeded { t.Fatalf("ok should succeed: %+v", t1) }
	if t2.Status != StatusFailed { t.Fatalf("bad should fail: %+v", t2) }
	if t3.Status != StatusFailed || t3.ExitCode == nil || *t3.ExitCode != -1 { t.Fatalf("slow should timeout fail: %+v", t3) }
}

func TestEngineGatewayMutexAndCancel(t *testing.T) {
	r := setupEngineRepo(t)
	e := NewEngine(r, 2)
	e.Register("gateway.start", 200*time.Millisecond, func(ctx context.Context, task *Task) (string, string, int, error) {
		time.Sleep(80 * time.Millisecond)
		return "", "", 0, nil
	})
	e.Start()
	defer e.Stop()

	id1 := uuid.NewString()
	id2 := uuid.NewString()
	id3 := uuid.NewString()
	_ = r.Create(&Task{TaskID: id1, TaskType: "gateway.start", Status: StatusPending, CreatedBy: "u1", CreatedAt: time.Now().UTC()})
	_ = r.Create(&Task{TaskID: id2, TaskType: "gateway.start", Status: StatusPending, CreatedBy: "u1", CreatedAt: time.Now().UTC()})
	_ = r.Create(&Task{TaskID: id3, TaskType: "gateway.start", Status: StatusPending, CreatedBy: "u1", CreatedAt: time.Now().UTC()})
	if err := e.Cancel(id3); err != nil { t.Fatal(err) }
	e.Enqueue(id1)
	e.Enqueue(id2)
	e.Enqueue(id3)
	time.Sleep(150 * time.Millisecond)

	t1, _ := r.FindByID(id1)
	t2, _ := r.FindByID(id2)
	t3, _ := r.FindByID(id3)
	if !((t1.Status == StatusSucceeded && t2.Status == StatusFailed) || (t1.Status == StatusFailed && t2.Status == StatusSucceeded)) {
		t.Fatalf("one should succeed and one should conflict fail: t1=%+v t2=%+v", t1, t2)
	}
	if t3.Status != StatusCanceled { t.Fatalf("id3 should canceled: %+v", t3) }
}
