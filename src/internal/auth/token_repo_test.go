package auth

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"openclaw-manager/internal/storage"
)

func seedUser(t *testing.T, db *sql.DB, id string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO users(user_id, username, password_hash, role, status, created_at) VALUES(?,?,?,?,?,?)`, id, id, "hash", "Viewer", "active", time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		t.Fatal(err)
	}
}

func TestTokenRepositoryFlow(t *testing.T) {
	db := storage.NewTestDB(t)
	seedUser(t, db.SQL, "u1")
	r := NewTokenRepository(db.SQL)
	raw := "refresh-token-abc"
	h := HashToken(raw)
	rt := &RefreshToken{TokenID: "t1", UserID: "u1", TokenHash: h, ExpiresAt: time.Now().UTC().Add(time.Hour), CreatedAt: time.Now().UTC()}
	if err := r.Save(rt); err != nil {
		t.Fatal(err)
	}
	got, err := r.FindByHash(h)
	if err != nil || got.TokenHash != h {
		t.Fatalf("find by hash failed err=%v got=%+v", err, got)
	}
	if err := r.Revoke("t1"); err != nil { t.Fatal(err) }
	got, _ = r.FindByHash(h)
	if !got.Revoked { t.Fatal("expect revoked") }
	if err := r.RevokeAllByUser("u1"); err != nil { t.Fatal(err) }
	if err := r.DeleteExpired(time.Now().UTC().Add(2 * time.Hour)); err != nil { t.Fatal(err) }
	if _, err := r.FindByHash(h); !errors.Is(err, sql.ErrNoRows) { t.Fatalf("expected no rows got %v", err) }
}

func TestBlacklistFlow(t *testing.T) {
	db := storage.NewTestDB(t)
	r := NewTokenRepository(db.SQL)
	exp := time.Now().UTC().Add(time.Hour)
	if err := r.AddBlacklist("j1", exp); err != nil { t.Fatal(err) }
	yes, err := r.ExistsJTI("j1")
	if err != nil || !yes { t.Fatalf("exists mismatch err=%v yes=%v", err, yes) }
	if err := r.CleanExpiredBlacklist(time.Now().UTC().Add(2 * time.Hour)); err != nil { t.Fatal(err) }
	yes, _ = r.ExistsJTI("j1")
	if yes { t.Fatal("expected cleaned") }
}
