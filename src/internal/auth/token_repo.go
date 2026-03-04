package auth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"
)

type RefreshToken struct {
	TokenID   string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}

type TokenRepository struct{ db *sql.DB }

func NewTokenRepository(db *sql.DB) *TokenRepository { return &TokenRepository{db: db} }

func HashToken(raw string) string {
	s := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(s[:])
}

func (r *TokenRepository) Save(t *RefreshToken) error {
	_, err := r.db.Exec(`INSERT INTO refresh_tokens(token_id,user_id,token_hash,expires_at,revoked,created_at,user_agent,ip_address) VALUES(?,?,?,?,?,?,?,?)`,
		t.TokenID, t.UserID, t.TokenHash, t.ExpiresAt.Format(time.RFC3339), boolToInt(t.Revoked), t.CreatedAt.Format(time.RFC3339), "", "")
	return err
}

func (r *TokenRepository) FindByHash(hash string) (*RefreshToken, error) {
	var t RefreshToken
	var exp, created string
	var revoked int
	err := r.db.QueryRow(`SELECT token_id,user_id,token_hash,expires_at,revoked,created_at FROM refresh_tokens WHERE token_hash=?`, hash).
		Scan(&t.TokenID, &t.UserID, &t.TokenHash, &exp, &revoked, &created)
	if err != nil {
		return nil, err
	}
	t.ExpiresAt, _ = time.Parse(time.RFC3339, exp)
	t.CreatedAt, _ = time.Parse(time.RFC3339, created)
	t.Revoked = revoked == 1
	return &t, nil
}

func (r *TokenRepository) Revoke(tokenID string) error {
	_, err := r.db.Exec(`UPDATE refresh_tokens SET revoked=1 WHERE token_id=?`, tokenID)
	return err
}

func (r *TokenRepository) RevokeAllByUser(userID string) error {
	_, err := r.db.Exec(`UPDATE refresh_tokens SET revoked=1 WHERE user_id=?`, userID)
	return err
}

func (r *TokenRepository) DeleteExpired(now time.Time) error {
	_, err := r.db.Exec(`DELETE FROM refresh_tokens WHERE expires_at < ?`, now.Format(time.RFC3339))
	return err
}

func (r *TokenRepository) AddBlacklist(jti string, expiresAt time.Time) error {
	_, err := r.db.Exec(`INSERT OR REPLACE INTO token_blacklist(jti,expires_at,created_at) VALUES(?,?,?)`, jti, expiresAt.Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339))
	return err
}

func (r *TokenRepository) ExistsJTI(jti string) (bool, error) {
	var c int
	err := r.db.QueryRow(`SELECT COUNT(1) FROM token_blacklist WHERE jti=?`, jti).Scan(&c)
	return c > 0, err
}

func (r *TokenRepository) CleanExpiredBlacklist(now time.Time) error {
	_, err := r.db.Exec(`DELETE FROM token_blacklist WHERE expires_at < ?`, now.Format(time.RFC3339))
	return err
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
