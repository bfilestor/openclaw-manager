package user

import "time"

type Role string

const (
	RoleUser     Role = "User"
	RoleViewer   Role = "Viewer"
	RoleOperator Role = "Operator"
	RoleAdmin    Role = "Admin"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

type User struct {
	UserID       string
	Username     string
	PasswordHash string
	Role         Role
	Status       Status
	CreatedAt    time.Time
	LastLoginAt  *time.Time
	UpdatedAt    *time.Time
}
