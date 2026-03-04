package task

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"openclaw-manager/internal/storage"
)

func TestTaskRepositoryCRUD(t *testing.T) {
	db := storage.NewTestDB(t)
	r := NewRepository(db.SQL)
	now := time.Now().UTC().Format(time.RFC3339)
	_, _ = db.SQL.Exec(`INSERT INTO users(user_id,username,password_hash,role,status,created_at) VALUES(?,?,?,?,?,?)`, "u1", "u1", "x", "Viewer", "active", now)

	tk := &Task{TaskID: uuid.NewString(), TaskType: "gateway.start", Status: StatusPending, CreatedBy: "u1", CreatedAt: time.Now().UTC()}
	if err := r.Create(tk); err != nil {
		t.Fatal(err)
	}
	found, err := r.FindByID(tk.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if found.Status != StatusPending {
		t.Fatalf("expect pending got %s", found.Status)
	}

	if err := r.UpdateStatus(tk.TaskID, StatusRunning); err != nil {
		t.Fatal(err)
	}
	found, _ = r.FindByID(tk.TaskID)
	if found.StartedAt == nil {
		t.Fatalf("started_at should be set")
	}

	exit := 0
	if err := r.UpdateResult(tk.TaskID, &exit, "ok", "", "/tmp/x.log"); err != nil {
		t.Fatal(err)
	}
	if err := r.UpdateStatus(tk.TaskID, StatusSucceeded); err != nil {
		t.Fatal(err)
	}
	found, _ = r.FindByID(tk.TaskID)
	if found.FinishedAt == nil || found.ExitCode == nil || *found.ExitCode != 0 {
		t.Fatalf("unexpected finished task: %+v", found)
	}
}

func TestTaskRepositoryListFilter(t *testing.T) {
	db := storage.NewTestDB(t)
	r := NewRepository(db.SQL)
	now := time.Now().UTC().Format(time.RFC3339)
	_, _ = db.SQL.Exec(`INSERT INTO users(user_id,username,password_hash,role,status,created_at) VALUES(?,?,?,?,?,?)`, "u1", "u1", "x", "Viewer", "active", now)
	_, _ = db.SQL.Exec(`INSERT INTO users(user_id,username,password_hash,role,status,created_at) VALUES(?,?,?,?,?,?)`, "u2", "u2", "x", "Viewer", "active", now)
	for i := 0; i < 4; i++ {
		status := StatusPending
		uid := "u1"
		if i%2 == 0 {
			status = StatusRunning
			uid = "u2"
		}
		_ = r.Create(&Task{TaskID: uuid.NewString(), TaskType: "t", Status: status, CreatedBy: uid, CreatedAt: time.Now().UTC().Add(time.Duration(i) * time.Second)})
	}
	list, total, err := r.List(ListFilter{Status: StatusRunning, CreatedBy: "u2", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("expect 2, total=%d len=%d", total, len(list))
	}
}
