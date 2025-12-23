package model

import "time"

type User struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Username  string `json:"username"`

	PasswordHash string `json:"-"`
	Role         string `json:"role"`
	Verified     bool   `json:"verified"`

	PasswordChangedAt   time.Time `json:"-"`
	PasswordExpiresAt   time.Time `json:"-"`
	FailedLoginAttempts int       `json:"-"`
	LockedUntil         time.Time `json:"-"`

	CreatedAt time.Time `json:"createdAt"`
}
