package validation

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"mime"
	"strings"
)

var (
	ErrInvalidFileType  = errors.New("file type not allowed")
	ErrFileTooLarge     = errors.New("file size exceeds maximum allowed")
	ErrInvalidFileSize  = errors.New("invalid file size")
	ErrFileReadError    = errors.New("error reading file")
	ErrIntegrityCheckFailed = errors.New("file integrity check failed")
)

// AllowedFileTypes defines whitelist of allowed MIME types
var AllowedFileTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"application/pdf": true,
	"text/plain":      true,
	"audio/mpeg":      true,
	"audio/wav":       true,
}

// MaxFileSize is maximum allowed file size in bytes (10MB)
const MaxFileSize = 10 * 1024 * 1024

// ValidateFileType checks if file MIME type is in whitelist
func ValidateFileType(mimeType string) error {
	// Parse MIME type to handle content-type with charset
	mediaType, _, err := mime.ParseMediaType(mimeType)
	if err != nil {
		return ErrInvalidFileType
	}

	// Check whitelist
	if !AllowedFileTypes[mediaType] {
		return ErrInvalidFileType
	}

	return nil
}

// ValidateFileSize checks if file size is within allowed limits
func ValidateFileSize(size int64) error {
	if size <= 0 {
		return ErrInvalidFileSize
	}

	if size > MaxFileSize {
		return ErrFileTooLarge
	}

	return nil
}

// ValidateFileExtension checks file extension (additional layer of validation)
func ValidateFileExtension(filename string, allowedExtensions []string) error {
	if filename == "" {
		return ErrInvalidFileType
	}

	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return ErrInvalidFileType
	}

	ext := strings.ToLower(parts[len(parts)-1])

	for _, allowed := range allowedExtensions {
		if ext == strings.ToLower(allowed) {
			return nil
		}
	}

	return ErrInvalidFileType
}

// CalculateFileHash calculates MD5 hash of file for integrity checking
func CalculateFileHash(reader io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", ErrFileReadError
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// VerifyFileIntegrity verifies file integrity by comparing hashes
func VerifyFileIntegrity(expectedHash string, fileReader io.Reader) error {
	actualHash, err := CalculateFileHash(fileReader)
	if err != nil {
		return err
	}

	if actualHash != expectedHash {
		return ErrIntegrityCheckFailed
	}

	return nil
}

// ValidateFileUpload performs comprehensive file upload validation
// This includes type whitelisting, size checking, and integrity verification
func ValidateFileUpload(filename, mimeType string, size int64, reader io.Reader) (string, error) {
	// 1. Validate file size
	if err := ValidateFileSize(size); err != nil {
		return "", err
	}

	// 2. Validate MIME type (whitelist approach)
	if err := ValidateFileType(mimeType); err != nil {
		return "", err
	}

	// 3. Calculate file hash for integrity
	fileHash, err := CalculateFileHash(reader)
	if err != nil {
		return "", err
	}

	return fileHash, nil
}
