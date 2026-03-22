package entity

import "time"

type User struct {
	Id           uint
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
