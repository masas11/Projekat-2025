package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/crypto/bcrypt"

	"users-service/config"
	"users-service/internal/dto"
	"users-service/internal/mail"
	"users-service/internal/security"
	"users-service/internal/store"
	"users-service/internal/validation"
)

type PasswordHandler struct {
	Repo   *store.UserRepository
	Config *config.Config
}

func NewPasswordHandler(repo *store.UserRepository, cfg *config.Config) *PasswordHandler {
	return &PasswordHandler{
		Repo:   repo,
		Config: cfg,
	}
}

// CHANGE PASSWORD (must be 1 day old)
func (h *PasswordHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate password strength
	if err := validation.IsStrongPassword(req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	user, err := h.Repo.GetByUsername(ctx, req.Username)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if time.Since(user.PasswordChangedAt) < 24*time.Hour {
		http.Error(w, "password too new", http.StatusForbidden)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)) != nil {
		http.Error(w, "wrong password", http.StatusUnauthorized)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	user.PasswordHash = string(hash)
	user.PasswordChangedAt = time.Now()
	user.PasswordExpiresAt = time.Now().Add(time.Duration(h.Config.PasswordExpirationDays) * 24 * time.Hour)

	if err := h.Repo.Update(ctx, user); err != nil {
		http.Error(w, "failed to update password", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "password changed successfully",
	})
}

// RequestPasswordReset sends password reset token to user's email
func (h *PasswordHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	user, err := h.Repo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists or not (security best practice)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "if email exists, password reset link has been sent",
		})
		return
	}

	// Generate password reset token
	token, err := security.GeneratePasswordResetToken()
	if err != nil {
		http.Error(w, "failed to generate reset token", http.StatusInternalServerError)
		return
	}

	// Store password reset token
	if err := h.Repo.SetPasswordResetToken(ctx, user.Email, token); err != nil {
		http.Error(w, "failed to store reset token", http.StatusInternalServerError)
		return
	}

	// Send password reset email (URL encode token to handle special characters)
	// Note: Link points to frontend, which will call the API
	encodedToken := url.QueryEscape(token)
	resetURL := h.Config.FrontendURL + "/reset-password?token=" + encodedToken
	mail.SendPasswordResetEmail(user.Email, resetURL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "if email exists, password reset link has been sent",
	})
}

// ResetPassword resets password using token from email
func (h *PasswordHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	// Validate password strength
	if err := validation.IsStrongPassword(req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	entry, ok := h.Repo.GetPasswordResetToken(ctx, req.Token)
	if !ok || security.IsPasswordResetTokenExpired(entry) {
		http.Error(w, "invalid or expired reset token", http.StatusUnauthorized)
		return
	}

	// Get user by email
	user, err := h.Repo.GetByEmail(ctx, entry.Email)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Reset password
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	user.PasswordHash = string(hash)
	user.PasswordChangedAt = time.Now()
	user.PasswordExpiresAt = time.Now().Add(time.Duration(h.Config.PasswordExpirationDays) * 24 * time.Hour)

	if err := h.Repo.Update(ctx, user); err != nil {
		http.Error(w, "failed to reset password", http.StatusInternalServerError)
		return
	}

	// Delete used reset token
	h.Repo.DeletePasswordResetToken(ctx, req.Token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "password reset successfully",
	})
}
