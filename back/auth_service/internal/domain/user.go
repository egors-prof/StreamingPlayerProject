package domain

import "time"

type User struct {
	ID        int
	FullName  string
	Username  string
	Password  string
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Role string

const (
	RoleUser  = "USER"
	RoleAdmin = "ADMIN"
)
