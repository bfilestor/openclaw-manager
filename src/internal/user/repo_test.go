package user

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"openclaw-manager/internal/storage"
)

func mkUser(id, name string, role Role) *User {
	return &User{
		UserID:       id,
		Username:     name,
		PasswordHash: "hash",
		Role:         role,
		Status:       StatusActive,
		CreatedAt:    time.Now().UTC(),
	}
}

func TestRepositoryCRUDAndQueries(t *testing.T) {
	db := storage.NewTestDB(t)
	r := NewRepository(db.SQL)

	u1 := mkUser("u1", "alice", RoleAdmin)
	u2 := mkUser("u2", "bob", RoleViewer)
	if err := r.Create(u1); err != nil {
		t.Fatalf("create u1: %v", err)
	}
	if err := r.Create(u2); err != nil {
		t.Fatalf("create u2: %v", err)
	}

	if err := r.Create(mkUser("u3", "alice", RoleViewer)); err == nil {
		t.Fatal("expect duplicate username error")
	}

	got, err := r.FindByUsername("alice")
	if err != nil || got.UserID != "u1" {
		t.Fatalf("find by username failed: %v got=%+v", err, got)
	}
	if _, err := r.FindByUsername("none"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expect ErrNotFound, got %v", err)
	}

	count, err := r.Count()
	if err != nil || count != 2 {
		t.Fatalf("count mismatch err=%v count=%d", err, count)
	}

	exists, err := r.ExistsAdmin()
	if err != nil || !exists {
		t.Fatalf("exists admin mismatch err=%v exists=%v", err, exists)
	}
	adminCount, err := r.CountByRole(RoleAdmin)
	if err != nil || adminCount != 1 {
		t.Fatalf("count by role mismatch err=%v count=%d", err, adminCount)
	}

	list, total, err := r.List(1, 2)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if total != 2 || len(list) != 1 {
		t.Fatalf("list mismatch total=%d len=%d", total, len(list))
	}

	u2.Role = RoleOperator
	now := time.Now().UTC()
	u2.UpdatedAt = &now
	if err := r.Update(u2); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if err := r.Delete("u2"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if err := r.Delete("u2"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestExistsAdminFalse(t *testing.T) {
	db := storage.NewTestDB(t)
	r := NewRepository(db.SQL)
	for i := 0; i < 2; i++ {
		u := mkUser("u"+fmt.Sprint(i), fmt.Sprintf("u%d", i), RoleViewer)
		if err := r.Create(u); err != nil {
			t.Fatal(err)
		}
	}
	exists, err := r.ExistsAdmin()
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("expected false")
	}
}
