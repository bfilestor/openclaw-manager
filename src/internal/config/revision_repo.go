package config

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

type Revision struct {
	RevisionID string    `json:"revision_id"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	Content    string    `json:"content"`
	SHA256     string    `json:"sha256"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  string    `json:"created_by"`
}

type RevisionRepository struct{ db *sql.DB }

func NewRevisionRepository(db *sql.DB) *RevisionRepository { return &RevisionRepository{db: db} }

func (r *RevisionRepository) Save(targetType, targetID, content, createdBy string) (*Revision, error) {
	now := time.Now().UTC()
	rev := &Revision{
		RevisionID: uuid.NewString(),
		TargetType: targetType,
		TargetID:   targetID,
		Content:    content,
		SHA256:     sha(content),
		CreatedAt:  now,
		CreatedBy:  createdBy,
	}
	var createdByVal any
	if rev.CreatedBy == "" {
		createdByVal = nil
	} else {
		createdByVal = rev.CreatedBy
	}
	_, err := r.db.Exec(`INSERT INTO revisions(revision_id,target_type,target_id,content,sha256,created_at,created_by) VALUES(?,?,?,?,?,?,?)`,
		rev.RevisionID, rev.TargetType, rev.TargetID, rev.Content, rev.SHA256, rev.CreatedAt.Format(time.RFC3339), createdByVal)
	if err != nil {
		return nil, err
	}
	_, _ = r.db.Exec(`DELETE FROM revisions WHERE revision_id IN (
		SELECT revision_id FROM revisions WHERE target_type=? AND ifnull(target_id,'')=ifnull(?, '') ORDER BY created_at DESC LIMIT -1 OFFSET 50
	)`, targetType, targetID)
	return rev, nil
}

func (r *RevisionRepository) List(targetType, targetID string, limit int) ([]Revision, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.Query(`SELECT revision_id,target_type,target_id,content,sha256,created_at,created_by FROM revisions WHERE target_type=? AND ifnull(target_id,'')=ifnull(?, '') ORDER BY created_at DESC LIMIT ?`, targetType, targetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Revision, 0)
	for rows.Next() {
		var rev Revision
		var created string
		var createdBy sql.NullString
		if err := rows.Scan(&rev.RevisionID, &rev.TargetType, &rev.TargetID, &rev.Content, &rev.SHA256, &created, &createdBy); err != nil {
			return nil, err
		}
		if createdBy.Valid {
			rev.CreatedBy = createdBy.String
		}
		rev.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, rev)
	}
	return out, rows.Err()
}

func (r *RevisionRepository) FindByID(id string) (*Revision, error) {
	var rev Revision
	var created string
	var createdBy sql.NullString
	err := r.db.QueryRow(`SELECT revision_id,target_type,target_id,content,sha256,created_at,created_by FROM revisions WHERE revision_id=?`, id).
		Scan(&rev.RevisionID, &rev.TargetType, &rev.TargetID, &rev.Content, &rev.SHA256, &created, &createdBy)
	if err != nil {
		return nil, err
	}
	if createdBy.Valid {
		rev.CreatedBy = createdBy.String
	}
	rev.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return &rev, nil
}

func sha(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
