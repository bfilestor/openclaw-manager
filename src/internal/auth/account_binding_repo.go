package auth

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

var ErrAccountBindingNotFound = errors.New("account binding not found")

type AccountBinding struct {
	UserID    string    `json:"user_id"`
	AccountID string    `json:"account_id"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AccountBindingRepository struct {
	db *sql.DB
}

func NewAccountBindingRepository(db *sql.DB) *AccountBindingRepository {
	return &AccountBindingRepository{db: db}
}

func (r *AccountBindingRepository) GetByUserID(userID string) (*AccountBinding, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrAccountBindingNotFound
	}
	row := r.db.QueryRow(`SELECT user_id, account_id, updated_at FROM user_account_bindings WHERE user_id=?`, userID)
	var item AccountBinding
	var updatedAt string
	if err := row.Scan(&item.UserID, &item.AccountID, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAccountBindingNotFound
		}
		return nil, err
	}
	if ts, err := time.Parse(time.RFC3339, updatedAt); err == nil {
		item.UpdatedAt = ts
	}
	return &item, nil
}

func (r *AccountBindingRepository) Upsert(userID, accountID string) error {
	userID = strings.TrimSpace(userID)
	accountID = strings.TrimSpace(accountID)
	if userID == "" || accountID == "" {
		return errors.New("user_id/account_id required")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`
INSERT INTO user_account_bindings(user_id, account_id, created_at, updated_at)
VALUES(?, ?, ?, ?)
ON CONFLICT(user_id) DO UPDATE SET
  account_id=excluded.account_id,
  updated_at=excluded.updated_at
`, userID, accountID, now, now)
	return err
}

func (r *AccountBindingRepository) Delete(userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil
	}
	_, err := r.db.Exec(`DELETE FROM user_account_bindings WHERE user_id=?`, userID)
	return err
}
