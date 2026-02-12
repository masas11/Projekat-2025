package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"users-service/config"
	"users-service/internal/dto"
	"users-service/internal/logger"
	"users-service/internal/mail"
	"users-service/internal/model"
	"users-service/internal/security"
	"users-service/internal/store"
	"users-service/internal/validation"
)

type RegisterHandler struct {
	Repo   *store.UserRepository
	Config *config.Config
	Logger *logger.Logger
}

func NewRegisterHandler(repo *store.UserRepository, cfg *config.Config, log *logger.Logger) *RegisterHandler {
	return &RegisterHandler{Repo: repo, Config: cfg, Logger: log}
}

func (h *RegisterHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Sanitize inputs
	req.FirstName = validation.SanitizeString(req.FirstName)
	req.LastName = validation.SanitizeString(req.LastName)
	req.Email = validation.SanitizeString(req.Email)
	req.Username = validation.SanitizeString(req.Username)

	// Validate email
	if err := validation.ValidateEmail(req.Email); err != nil {
		if h.Logger != nil {
			h.Logger.LogValidationFailure("email", err.Error(), req.Email)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate username
	if err := validation.ValidateUsername(req.Username); err != nil {
		if h.Logger != nil {
			h.Logger.LogValidationFailure("username", err.Error(), req.Username)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate names
	if err := validation.ValidateName(req.FirstName); err != nil {
		if h.Logger != nil {
			h.Logger.LogValidationFailure("firstName", err.Error(), req.FirstName)
		}
		http.Error(w, "invalid first name: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := validation.ValidateName(req.LastName); err != nil {
		if h.Logger != nil {
			h.Logger.LogValidationFailure("lastName", err.Error(), req.LastName)
		}
		http.Error(w, "invalid last name: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Check for SQL injection
	if err := validation.CheckSQLInjection(req.FirstName); err != nil {
		if h.Logger != nil {
			h.Logger.LogValidationFailure("firstName", "SQL injection attempt detected", "")
		}
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}
	if err := validation.CheckSQLInjection(req.LastName); err != nil {
		if h.Logger != nil {
			h.Logger.LogValidationFailure("lastName", "SQL injection attempt detected", "")
		}
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}
	if err := validation.CheckSQLInjection(req.Username); err != nil {
		if h.Logger != nil {
			h.Logger.LogValidationFailure("username", "SQL injection attempt detected", "")
		}
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}
	if err := validation.CheckSQLInjection(req.Email); err != nil {
		if h.Logger != nil {
			h.Logger.LogValidationFailure("email", "SQL injection attempt detected", "")
		}
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	// Check for XSS
	if err := validation.CheckXSS(req.FirstName); err != nil {
		if h.Logger != nil {
			h.Logger.LogValidationFailure("firstName", "XSS attempt detected", "")
		}
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}
	if err := validation.CheckXSS(req.LastName); err != nil {
		if h.Logger != nil {
			h.Logger.LogValidationFailure("lastName", "XSS attempt detected", "")
		}
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}
	if err := validation.CheckXSS(req.Email); err != nil {
		if h.Logger != nil {
			h.Logger.LogValidationFailure("email", "XSS attempt detected", "")
		}
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	// password confirmation
	if req.Password != req.ConfirmPassword {
		if h.Logger != nil {
			h.Logger.LogValidationFailure("password", "passwords do not match", "")
		}
		http.Error(w, "passwords do not match", http.StatusBadRequest)
		return
	}

	// password strength validation
	if err := validation.IsStrongPassword(req.Password); err != nil {
		if h.Logger != nil {
			h.Logger.LogValidationFailure("password", err.Error(), "")
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// hash lozinke
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	now := time.Now()

	user := &model.User{
		ID:                uuid.NewString(),
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		Email:             req.Email,
		Username:          req.Username,
		PasswordHash:      string(hash),
		Role:              "USER",
		Verified:          false, // User must verify email first
		PasswordChangedAt: now,
		PasswordExpiresAt: now.Add(time.Duration(h.Config.PasswordExpirationDays) * 24 * time.Hour),
		CreatedAt:         now,
	}

	ctx := r.Context()
	if err := h.Repo.Create(ctx, user); err != nil {
		if err == store.ErrUserExists {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "failed to create user", http.StatusInternalServerError)
		}
		return
	}

	// Generate verification token
	verificationToken, err := security.GenerateVerificationToken()
	if err != nil {
		http.Error(w, "failed to generate verification token", http.StatusInternalServerError)
		return
	}

	// Store verification token
	if err := h.Repo.SetVerificationToken(ctx, user.Email, verificationToken); err != nil {
		http.Error(w, "failed to store verification token", http.StatusInternalServerError)
		return
	}

	// Send verification email (URL encode token to handle special characters)
	// Note: Link points to frontend, which will call the API
	encodedToken := url.QueryEscape(verificationToken)
	verificationURL := h.Config.FrontendURL + "/verify-email?token=" + encodedToken
	mail.SendVerificationEmail(user.Email, verificationURL)

	// Use output encoding for response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	
	// Check if SMTP is configured - if not, include verification link in response for development
	response := map[string]interface{}{
		"message": "registration successful, verification email sent",
	}
	
	// If SMTP is not configured, include verification link in response (development only)
	if h.Config.SMTPHost == "" || h.Config.SMTPUsername == "" || h.Config.SMTPPassword == "" {
		response["verificationLink"] = verificationURL
		response["note"] = "SMTP not configured - verification link included in response. Check logs for details."
	}
	
	json.NewEncoder(w).Encode(response)
}
