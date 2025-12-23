package validation

import (
	"errors"
	"unicode"
)

var ErrWeakPassword = errors.New(
	"password must be at least 8 characters long and contain one uppercase letter and one number",
)

// IsStrongPassword proverava da li lozinka ispunjava osnovne sigurnosne kriterijume
func IsStrongPassword(password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}

	hasUpper := false
	hasNumber := false

	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsDigit(r) {
			hasNumber = true
		}
	}

	if !hasUpper || !hasNumber {
		return ErrWeakPassword
	}

	return nil
}
