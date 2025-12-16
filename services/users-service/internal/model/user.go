package model

import "time"

type User struct {
	ID           string    `json:"id"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Verified     bool      `json:"verified"`
	CreatedAt    time.Time `json:"createdAt"`
}
