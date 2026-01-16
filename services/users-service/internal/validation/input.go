package validation

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidUsername = errors.New("username must be 3-20 characters and contain only letters, numbers, and underscores")
	ErrInvalidName     = errors.New("name must contain only letters and spaces")
	ErrInvalidLength   = errors.New("input length exceeds maximum allowed")
	ErrSpecialChars    = errors.New("input contains invalid special characters")
	ErrSQLInjection    = errors.New("input contains potentially dangerous SQL characters")
	ErrXSS             = errors.New("input contains potentially dangerous XSS characters")
)

// Email validation using regex
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	// Basic email regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	// Boundary check - email should not exceed 254 characters (RFC 5321)
	if len(email) > 254 {
		return ErrInvalidLength
	}

	return nil
}

// Username validation - whitelist approach
func ValidateUsername(username string) error {
	if username == "" {
		return errors.New("username is required")
	}

	// Boundary check
	if len(username) < 3 || len(username) > 20 {
		return ErrInvalidUsername
	}

	// Whitelist: only letters, numbers, and underscores
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(username) {
		return ErrInvalidUsername
	}

	return nil
}

// Name validation - only letters and spaces
func ValidateName(name string) error {
	if name == "" {
		return errors.New("name is required")
	}

	// Boundary check
	if len(name) > 100 {
		return ErrInvalidLength
	}

	// Whitelist: only letters and spaces
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsSpace(r) {
			return ErrInvalidName
		}
	}

	return nil
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newline and tab
	var result strings.Builder
	for _, r := range input {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			continue
		}
		result.WriteRune(r)
	}

	return result.String()
}

// CheckSQLInjection attempts to detect SQL injection patterns
func CheckSQLInjection(input string) error {
	// Common SQL injection patterns
	sqlPatterns := []string{
		"' OR '1'='1",
		"' OR '1'='1'--",
		"'; DROP TABLE",
		"UNION SELECT",
		"'; INSERT INTO",
		"'; UPDATE",
		"'; DELETE FROM",
	}

	inputLower := strings.ToLower(input)
	for _, pattern := range sqlPatterns {
		if strings.Contains(inputLower, strings.ToLower(pattern)) {
			return ErrSQLInjection
		}
	}

	return nil
}

// CheckXSS attempts to detect XSS patterns
func CheckXSS(input string) error {
	// Common XSS patterns
	xssPatterns := []string{
		"<script",
		"</script>",
		"javascript:",
		"onerror=",
		"onload=",
		"onclick=",
		"<iframe",
		"<img",
		"<svg",
	}

	inputLower := strings.ToLower(input)
	for _, pattern := range xssPatterns {
		if strings.Contains(inputLower, pattern) {
			return ErrXSS
		}
	}

	return nil
}

// ValidateNumeric checks if a string is a valid number within bounds
func ValidateNumeric(value string, min, max int) error {
	if value == "" {
		return errors.New("numeric value is required")
	}

	// Check if all characters are digits
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return errors.New("value must be numeric")
		}
	}

	// Parse and check bounds
	var num int
	for _, r := range value {
		num = num*10 + int(r-'0')
		if num > max {
			return errors.New("value exceeds maximum")
		}
	}

	if num < min {
		return errors.New("value below minimum")
	}

	return nil
}

// ValidateStringLength checks string length boundaries
func ValidateStringLength(input string, min, max int) error {
	length := len(input)
	if length < min {
		return errors.New("input too short")
	}
	if length > max {
		return ErrInvalidLength
	}
	return nil
}
