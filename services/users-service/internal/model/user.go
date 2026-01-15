package model

import "time"

type User struct {
	ID        string `json:"id" bson:"_id"`
	FirstName string `json:"firstName" bson:"firstName"`
	LastName  string `json:"lastName" bson:"lastName"`
	Email     string `json:"email" bson:"email"`
	Username  string `json:"username" bson:"username"`

	PasswordHash string `json:"-" bson:"passwordHash"`
	Role         string `json:"role" bson:"role"`
	Verified     bool   `json:"verified" bson:"verified"`

	PasswordChangedAt   time.Time `json:"-" bson:"passwordChangedAt"`
	PasswordExpiresAt   time.Time `json:"-" bson:"passwordExpiresAt"`
	FailedLoginAttempts int       `json:"-" bson:"failedLoginAttempts"`
	LockedUntil         time.Time `json:"-" bson:"lockedUntil"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}
