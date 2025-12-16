package validation

import (
	"errors"
	"regexp"
)

var (
	ErrWeakPassword = errors.New("password must be at least 8 characters long and contain one uppercase letter and one number")
)

var (
	upperCaseRegex = regexp.MustCompile(`[A-Z]`)
	numberRegex    = regexp.MustCompile(`[0-9]`)
)

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}

	if !upperCaseRegex.MatchString(password) {
		return ErrWeakPassword
	}

	if !numberRegex.MatchString(password) {
		return ErrWeakPassword
	}

	return nil
}
