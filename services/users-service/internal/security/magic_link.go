package security

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

type MagicLinkEntry struct {
	Token     string
	Email     string
	ExpiresAt time.Time
}

// VerificationTokenEntry represents an email verification token
type VerificationTokenEntry struct {
	Token     string
	Email     string
	ExpiresAt time.Time
}

// PasswordResetTokenEntry represents a password reset token
type PasswordResetTokenEntry struct {
	Token     string
	Email     string
	ExpiresAt time.Time
}

// GenerateMagicLinkToken generates a secure random token for magic link
func GenerateMagicLinkToken() (string, error) {
	return GenerateSecureToken()
}

// GenerateVerificationToken generates a secure random token for email verification
func GenerateVerificationToken() (string, error) {
	return GenerateSecureToken()
}

// GeneratePasswordResetToken generates a secure random token for password reset
func GeneratePasswordResetToken() (string, error) {
	return GenerateSecureToken()
}

// GenerateSecureToken generates a secure random token
func GenerateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// IsExpired checks if magic link token is expired
func IsMagicLinkExpired(e MagicLinkEntry) bool {
	return time.Now().After(e.ExpiresAt)
}

// IsVerificationTokenExpired checks if verification token is expired
func IsVerificationTokenExpired(e VerificationTokenEntry) bool {
	return time.Now().After(e.ExpiresAt)
}

// IsPasswordResetTokenExpired checks if password reset token is expired
func IsPasswordResetTokenExpired(e PasswordResetTokenEntry) bool {
	return time.Now().After(e.ExpiresAt)
}
