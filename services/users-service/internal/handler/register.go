package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"users-service/internal/dto"
	"users-service/internal/mail"
	"users-service/internal/model"
	"users-service/internal/security"
	"users-service/internal/store"
	"users-service/internal/validation"
)

type RegisterHandler struct {
	Repo *store.UserRepository
}

func NewRegisterHandler(repo *store.UserRepository) *RegisterHandler {
	return &RegisterHandler{Repo: repo}
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate username
	if err := validation.ValidateUsername(req.Username); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate names
	if err := validation.ValidateName(req.FirstName); err != nil {
		http.Error(w, "invalid first name: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := validation.ValidateName(req.LastName); err != nil {
		http.Error(w, "invalid last name: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Check for SQL injection
	if err := validation.CheckSQLInjection(req.FirstName); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}
	if err := validation.CheckSQLInjection(req.LastName); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}
	if err := validation.CheckSQLInjection(req.Username); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	// Check for XSS
	if err := validation.CheckXSS(req.FirstName); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}
	if err := validation.CheckXSS(req.LastName); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	// password confirmation
	if req.Password != req.ConfirmPassword {
		http.Error(w, "passwords do not match", http.StatusBadRequest)
		return
	}

	// password strength validation
	if err := validation.IsStrongPassword(req.Password); err != nil {
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
		PasswordExpiresAt: now.Add(60 * 24 * time.Hour),
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
	// Assuming frontend runs on port 3000
	verificationURL := "http://localhost:3000/verify-email?token=" + encodedToken
	mail.SendVerificationEmail(user.Email, verificationURL)

	// Use output encoding for response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]string{
		"message": "registration successful, verification email sent",
	}
	json.NewEncoder(w).Encode(response)
}
