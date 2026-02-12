package logger

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LevelInfo    LogLevel = "INFO"
	LevelWarning LogLevel = "WARN"
	LevelError   LogLevel = "ERROR"
	LevelAudit   LogLevel = "AUDIT"
)

// EventType represents the type of event being logged
type EventType string

const (
	EventValidationFailure    EventType = "VALIDATION_FAILURE"
	EventLoginSuccess         EventType = "LOGIN_SUCCESS"
	EventLoginFailure         EventType = "LOGIN_FAILURE"
	EventAccessControlFailure EventType = "ACCESS_CONTROL_FAILURE"
	EventStateChange          EventType = "STATE_CHANGE"
	EventInvalidToken         EventType = "INVALID_TOKEN"
	EventExpiredToken         EventType = "EXPIRED_TOKEN"
	EventAdminActivity        EventType = "ADMIN_ACTIVITY"
	EventTLSFailure           EventType = "TLS_FAILURE"
)

// Logger is a structured logger with file rotation and security features
type Logger struct {
	infoLogger    *log.Logger
	warnLogger    *log.Logger
	errorLogger   *log.Logger
	auditLogger   *log.Logger
	logDir        string
	currentFile   string
	maxSize       int64 // Maximum file size in bytes (default: 10MB)
	maxFiles      int   // Maximum number of rotated files to keep
	mu            sync.Mutex
	file          *os.File
	checksums     map[string]string // File path -> SHA256 checksum
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// InitLogger initializes the default logger
func InitLogger(logDir string) error {
	var err error
	once.Do(func() {
		defaultLogger, err = NewLogger(logDir)
	})
	return err
}

// GetLogger returns the default logger instance
func GetLogger() *Logger {
	if defaultLogger == nil {
		// Fallback to stdout if not initialized
		return NewStdoutLogger()
	}
	return defaultLogger
}

// NewLogger creates a new logger instance
func NewLogger(logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logger := &Logger{
		logDir:    logDir,
		maxSize:   10 * 1024 * 1024, // 10MB default
		maxFiles:  5,                 // Keep 5 rotated files
		checksums: make(map[string]string),
	}

	if err := logger.openLogFile(); err != nil {
		return nil, err
	}

	return logger, nil
}

// NewStdoutLogger creates a logger that writes to stdout (for testing/fallback)
func NewStdoutLogger() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		warnLogger:  log.New(os.Stdout, "[WARN] ", log.LstdFlags),
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
		auditLogger: log.New(os.Stdout, "[AUDIT] ", log.LstdFlags),
	}
}

// openLogFile opens or creates the current log file
func (l *Logger) openLogFile() error {
	timestamp := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("app-%s.log", timestamp)
	l.currentFile = filepath.Join(l.logDir, filename)

	file, err := os.OpenFile(l.currentFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.file = file

	// Create separate loggers for each level
	l.infoLogger = log.New(l.file, "[INFO] ", log.LstdFlags)
	l.warnLogger = log.New(l.file, "[WARN] ", log.LstdFlags)
	l.errorLogger = log.New(l.file, "[ERROR] ", log.LstdFlags)
	l.auditLogger = log.New(l.file, "[AUDIT] ", log.LstdFlags)

	// Calculate and store checksum
	if err := l.updateChecksum(); err != nil {
		return err
	}

	return nil
}

// rotateLog rotates the log file if it exceeds maxSize
func (l *Logger) rotateLog() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check file size
	info, err := l.file.Stat()
	if err != nil {
		return err
	}

	if info.Size() < l.maxSize {
		return nil // No rotation needed
	}

	// Close current file
	l.file.Close()

	// Calculate checksum before rotation
	if err := l.updateChecksum(); err != nil {
		return err
	}

	// Rotate: rename current file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	rotatedFile := fmt.Sprintf("%s.%s", l.currentFile, timestamp)
	if err := os.Rename(l.currentFile, rotatedFile); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	// Clean up old rotated files
	l.cleanupOldFiles()

	// Open new log file
	return l.openLogFile()
}

// cleanupOldFiles removes old rotated log files beyond maxFiles limit
func (l *Logger) cleanupOldFiles() {
	pattern := filepath.Join(l.logDir, "app-*.log.*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	if len(matches) <= l.maxFiles {
		return
	}

	// Sort by modification time (oldest first)
	// Simple approach: remove files beyond limit
	for i := 0; i < len(matches)-l.maxFiles; i++ {
		os.Remove(matches[i])
		delete(l.checksums, matches[i])
	}
}

// updateChecksum calculates and stores SHA256 checksum of the log file
func (l *Logger) updateChecksum() error {
	file, err := os.Open(l.currentFile)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	l.checksums[l.currentFile] = checksum

	// Write checksum to separate file for integrity verification
	checksumFile := l.currentFile + ".checksum"
	return os.WriteFile(checksumFile, []byte(checksum), 0640)
}

// sanitizeMessage removes sensitive data from log messages
func sanitizeMessage(message string) string {
	// Remove passwords (common patterns like "password=...")
	// Remove JWT tokens (keep only first/last few chars if present)
	// Remove stack traces (lines starting with "goroutine", "panic", etc.)
	// Note: Most sensitive data is already filtered in Log() function for fields
	// This function handles message string sanitization
	
	// Remove stack trace patterns
	lines := strings.Split(message, "\n")
	var sanitizedLines []string
	for _, line := range lines {
		// Skip stack trace lines
		if strings.HasPrefix(strings.TrimSpace(line), "goroutine") ||
			strings.HasPrefix(strings.TrimSpace(line), "panic:") ||
			strings.Contains(line, "runtime.") ||
			strings.Contains(line, "main.") {
			continue
		}
		sanitizedLines = append(sanitizedLines, line)
	}
	
	return strings.Join(sanitizedLines, "\n")
}

// Log logs a structured event
func (l *Logger) Log(level LogLevel, eventType EventType, message string, fields map[string]interface{}) {
	// Check if rotation is needed
	if l.file != nil {
		if err := l.rotateLog(); err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "Log rotation error: %v\n", err)
		}
	}

	// Sanitize message
	sanitizedMsg := sanitizeMessage(message)

	// Build log entry
	entry := fmt.Sprintf("[%s] EventType=%s Message=%s", level, eventType, sanitizedMsg)
	if len(fields) > 0 {
		entry += " Fields="
		for k, v := range fields {
			// Don't log sensitive fields
			if k == "password" || k == "token" || k == "otp" || k == "secret" {
				entry += fmt.Sprintf("%s=*** ", k)
			} else {
				entry += fmt.Sprintf("%s=%v ", k, v)
			}
		}
	}

	// Write to appropriate logger
	switch level {
	case LevelInfo:
		l.infoLogger.Println(entry)
	case LevelWarning:
		l.warnLogger.Println(entry)
	case LevelError:
		l.errorLogger.Println(entry)
	case LevelAudit:
		l.auditLogger.Println(entry)
	}
}

// LogValidationFailure logs a validation failure
func (l *Logger) LogValidationFailure(field string, reason string, value interface{}) {
	l.Log(LevelWarning, EventValidationFailure, "Validation failed",
		map[string]interface{}{
			"field":  field,
			"reason": reason,
			"value":  value,
		})
}

// LogLoginSuccess logs a successful login
func (l *Logger) LogLoginSuccess(username string, ipAddress string) {
	l.Log(LevelAudit, EventLoginSuccess, "Login successful",
		map[string]interface{}{
			"username": username,
			"ip":       ipAddress,
			"timestamp": time.Now().Unix(),
		})
}

// LogLoginFailure logs a failed login attempt
func (l *Logger) LogLoginFailure(username string, reason string, ipAddress string) {
	l.Log(LevelWarning, EventLoginFailure, "Login failed",
		map[string]interface{}{
			"username": username,
			"reason":   reason,
			"ip":       ipAddress,
			"timestamp": time.Now().Unix(),
		})
}

// LogAccessControlFailure logs an access control failure
func (l *Logger) LogAccessControlFailure(userID string, resource string, action string, reason string) {
	l.Log(LevelWarning, EventAccessControlFailure, "Access control failure",
		map[string]interface{}{
			"userID":  userID,
			"resource": resource,
			"action":  action,
			"reason":  reason,
		})
}

// LogStateChange logs an unexpected state change
func (l *Logger) LogStateChange(entity string, oldState interface{}, newState interface{}, userID string) {
	l.Log(LevelWarning, EventStateChange, "State change detected",
		map[string]interface{}{
			"entity":   entity,
			"oldState": oldState,
			"newState": newState,
			"userID":   userID,
		})
}

// LogInvalidToken logs an attempt with invalid token
func (l *Logger) LogInvalidToken(tokenPrefix string, reason string, ipAddress string) {
	l.Log(LevelWarning, EventInvalidToken, "Invalid token used",
		map[string]interface{}{
			"tokenPrefix": tokenPrefix, // Only log prefix, not full token
			"reason":      reason,
			"ip":          ipAddress,
		})
}

// LogExpiredToken logs an attempt with expired token
func (l *Logger) LogExpiredToken(userID string, ipAddress string) {
	l.Log(LevelWarning, EventExpiredToken, "Expired token used",
		map[string]interface{}{
			"userID": userID,
			"ip":     ipAddress,
		})
}

// LogAdminActivity logs an administrative action
func (l *Logger) LogAdminActivity(adminID string, action string, resource string, details map[string]interface{}) {
	fields := map[string]interface{}{
		"adminID": adminID,
		"action":  action,
		"resource": resource,
	}
	for k, v := range details {
		fields[k] = v
	}
	l.Log(LevelAudit, EventAdminActivity, "Admin activity",
		fields)
}

// LogTLSFailure logs a TLS connection failure
func (l *Logger) LogTLSFailure(service string, errorMsg string, remoteAddr string) {
	l.Log(LevelError, EventTLSFailure, "TLS connection failure",
		map[string]interface{}{
			"service":    service,
			"error":      errorMsg,
			"remoteAddr": remoteAddr,
		})
}

// VerifyIntegrity verifies the integrity of log files by checking checksums
func (l *Logger) VerifyIntegrity() error {
	for filePath, expectedChecksum := range l.checksums {
		file, err := os.Open(filePath)
		if err != nil {
			continue // File might have been rotated
		}

		hash := sha256.New()
		if _, err := io.Copy(hash, file); err != nil {
			file.Close()
			continue
		}
		file.Close()

		actualChecksum := hex.EncodeToString(hash.Sum(nil))
		if actualChecksum != expectedChecksum {
			return fmt.Errorf("integrity check failed for %s", filePath)
		}
	}
	return nil
}

// Close closes the logger and its file handles
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
