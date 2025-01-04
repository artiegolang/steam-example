package domain

import "time"

type User struct {
	ID        int64      `json:"id"`
	Username  string     `json:"username"`
	Password  string     `json:"-"` // Пароль не возвращаем в JSON
	Email     string     `json:"email"`
	Role      UserRole   `json:"role"`
	Status    UserStatus `json:"status"`
	Balance   int64      `json:"balance"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// UserRole представляет роль пользователя в системе
type UserRole string

const (
	RoleAdmin UserRole = "ADMIN"
	RoleUser  UserRole = "USER"
)

// UserStatus представляет статус пользователя
type UserStatus string

const (
	StatusActive  UserStatus = "ACTIVE"
	StatusBlocked UserStatus = "BLOCKED"
	StatusDeleted UserStatus = "DELETED"
)
