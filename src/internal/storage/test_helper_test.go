package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTestDB(t *testing.T) {
	db := NewTestDB(t)
	_, err := db.SQL.Exec(`INSERT INTO users(user_id, username, password_hash, created_at) VALUES('u1','tester','hash','now')`)
	assert.NoError(t, err)
}
