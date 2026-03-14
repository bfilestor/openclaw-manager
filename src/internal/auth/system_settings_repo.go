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
	return r.getBoolSetting("public_registration", true)
}

func (r *SystemSettingsRepository) SetPublicRegistrationEnabled(enabled bool) error {
	return r.setBoolSetting("public_registration", enabled)
}

func (r *SystemSettingsRepository) IsLobsterGuardianEnabled() (bool, error) {
	return r.getBoolSetting("lobster_guardian", false)
}

func (r *SystemSettingsRepository) SetLobsterGuardianEnabled(enabled bool) error {
	return r.setBoolSetting("lobster_guardian", enabled)
}

func (r *SystemSettingsRepository) getBoolSetting(key string, defaultVal bool) (bool, error) {
	var raw string
	err := r.db.QueryRow(`SELECT value FROM system_settings WHERE key=?`, key).Scan(&raw)
	if err == sql.ErrNoRows {
		return defaultVal, nil
	}
	if err != nil {
		return false, err
	}
	return raw == "true" || raw == "1", nil
}

func (r *SystemSettingsRepository) setBoolSetting(key string, enabled bool) error {
	value := "false"
	if enabled {
		value = "true"
	}
	_, err := r.db.Exec(`
INSERT INTO system_settings(key, value, updated_at)
VALUES(?, ?, ?)
ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at
`, key, value, time.Now().UTC().Format(time.RFC3339))
	return err
}
