package user

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("user not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(u *User) error {
	_, err := r.db.Exec(`
INSERT INTO users(user_id, username, password_hash, role, status, created_at, last_login_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		u.UserID, u.Username, u.PasswordHash, string(u.Role), string(u.Status),
		u.CreatedAt.Format(time.RFC3339), toNullTime(u.LastLoginAt), toNullTime(u.UpdatedAt),
	)
	return err
}

func (r *Repository) FindByID(id string) (*User, error) {
	return r.findOne(`SELECT user_id, username, password_hash, role, status, created_at, last_login_at, updated_at FROM users WHERE user_id=?`, id)
}

func (r *Repository) FindByUsername(username string) (*User, error) {
	return r.findOne(`SELECT user_id, username, password_hash, role, status, created_at, last_login_at, updated_at FROM users WHERE username=?`, username)
}

func (r *Repository) Update(u *User) error {
	res, err := r.db.Exec(`
UPDATE users SET username=?, password_hash=?, role=?, status=?, last_login_at=?, updated_at=? WHERE user_id=?`,
		u.Username, u.PasswordHash, string(u.Role), string(u.Status), toNullTime(u.LastLoginAt), toNullTime(u.UpdatedAt), u.UserID,
	)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM users WHERE user_id=?`, id)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) List(offset, limit int) ([]*User, int, error) {
	var total int
	if err := r.db.QueryRow(`SELECT COUNT(1) FROM users`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(`
SELECT user_id, username, password_hash, role, status, created_at, last_login_at, updated_at
FROM users ORDER BY created_at ASC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]*User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, u)
	}
	return out, total, rows.Err()
}

func (r *Repository) Count() (int, error) {
	var c int
	err := r.db.QueryRow(`SELECT COUNT(1) FROM users`).Scan(&c)
	return c, err
}

func (r *Repository) ExistsAdmin() (bool, error) {
	var c int
	err := r.db.QueryRow(`SELECT COUNT(1) FROM users WHERE role=?`, string(RoleAdmin)).Scan(&c)
	return c > 0, err
}

func (r *Repository) CountByRole(role Role) (int, error) {
	var c int
	err := r.db.QueryRow(`SELECT COUNT(1) FROM users WHERE role=?`, string(role)).Scan(&c)
	return c, err
}

func (r *Repository) FindFirstRegistered() (*User, error) {
	row := r.db.QueryRow(`
SELECT user_id, username, password_hash, role, status, created_at, last_login_at, updated_at
FROM users
ORDER BY created_at ASC, user_id ASC
LIMIT 1`)
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *Repository) findOne(q string, arg any) (*User, error) {
	row := r.db.QueryRow(q, arg)
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

type scanner interface{ Scan(dest ...any) error }

func scanUser(s scanner) (*User, error) {
	var (
		u                       User
		role, status, createdAt string
		lastLoginAt, updatedAt  sql.NullString
	)
	if err := s.Scan(&u.UserID, &u.Username, &u.PasswordHash, &role, &status, &createdAt, &lastLoginAt, &updatedAt); err != nil {
		return nil, err
	}
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	u.CreatedAt = t
	u.Role = Role(role)
	u.Status = Status(status)
	if lastLoginAt.Valid {
		lt, err := time.Parse(time.RFC3339, lastLoginAt.String)
		if err == nil {
			u.LastLoginAt = &lt
		}
	}
	if updatedAt.Valid {
		ut, err := time.Parse(time.RFC3339, updatedAt.String)
		if err == nil {
			u.UpdatedAt = &ut
		}
	}
	return &u, nil
}

func toNullTime(t *time.Time) sql.NullString {
	if t == nil {
		return sql.NullString{}
	}
	return sql.NullString{Valid: true, String: t.Format(time.RFC3339)}
}
