package config

import (
	"testing"

	"openclaw-manager/internal/storage"
)

func TestRevisionRepoSaveListFindAndTrim(t *testing.T) {
	db := storage.NewTestDB(t)
	r := NewRevisionRepository(db.SQL)

	for i := 0; i < 51; i++ {
		_, err := r.Save("openclaw_json", "", "{\"v\":"+string(rune('a'+(i%26)))+"}", "")
		if err != nil {
			t.Fatal(err)
		}
	}
	list, err := r.List("openclaw_json", "", 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 50 {
		t.Fatalf("expect 50 got %d", len(list))
	}
	one, err := r.FindByID(list[0].RevisionID)
	if err != nil || one.SHA256 == "" {
		t.Fatalf("find by id failed err=%v rev=%+v", err, one)
	}

	deleted, err := r.Delete("openclaw_json", "", one.RevisionID)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if !deleted {
		t.Fatalf("expect deleted=true for %s", one.RevisionID)
	}
	if _, err := r.FindByID(one.RevisionID); err == nil {
		t.Fatalf("expect not found after delete: %s", one.RevisionID)
	}
	deleted, err = r.Delete("openclaw_json", "", one.RevisionID)
	if err != nil {
		t.Fatalf("delete missing failed: %v", err)
	}
	if deleted {
		t.Fatalf("expect deleted=false for missing id %s", one.RevisionID)
	}
}
