package auth

import (
	"database/sql"
	"time"
)

type SystemSettingsRepository struct {
	db *sql.DB
}

func NewSystemSettingsRepository(db *sql.DB) *SystemSettingsRepository {
	return &SystemSettingsRepository{db: db}
}

func (r *SystemSettingsRepository) IsPublicRegistrationEnabled() (bool, error) {
	var raw string
	err := r.db.QueryRow(`SELECT value FROM system_settings WHERE key='public_registration'`).Scan(&raw)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return raw == "true" || raw == "1", nil
}

func (r *SystemSettingsRepository) SetPublicRegistrationEnabled(enabled bool) error {
	value := "false"
	if enabled {
		value = "true"
	}
	_, err := r.db.Exec(`
INSERT INTO system_settings(key, value, updated_at)
VALUES('public_registration', ?, ?)
ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at
`, value, time.Now().UTC().Format(time.RFC3339))
	return err
}
