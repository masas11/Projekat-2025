package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"users-service/config"
	"users-service/internal/dto"
	"users-service/internal/mail"
	"users-service/internal/security"
	"users-service/internal/store"
)

type MagicLinkHandler struct {
	Repo   *store.UserRepository
	Config *config.Config
}

func NewMagicLinkHandler(repo *store.UserRepository, cfg *config.Config) *MagicLinkHandler {
	return &MagicLinkHandler{
		Repo:   repo,
		Config: cfg,
	}
}

// RequestMagicLink generates and sends magic link to user's email
func (h *MagicLinkHandler) RequestMagicLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.MagicLinkRequest
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
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "if email exists, magic link has been sent",
		})
		return
	}

	// Generate magic link token
	token, err := security.GenerateMagicLinkToken()
	if err != nil {
		http.Error(w, "failed to generate magic link", http.StatusInternalServerError)
		return
	}

	// Store magic link
	if err := h.Repo.SetMagicLink(ctx, user.Email, token); err != nil {
		http.Error(w, "failed to store magic link", http.StatusInternalServerError)
		return
	}

	// Send magic link via email - URL encode token to handle special characters
	// Note: Link points to frontend, which will call the API
	encodedToken := url.QueryEscape(token)
	magicLinkURL := h.Config.FrontendURL + "/verify-magic-link?token=" + encodedToken
	mail.SendMagicLink(user.Email, magicLinkURL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "if email exists, magic link has been sent",
	})
}

// VerifyMagicLink verifies the magic link token and logs in the user
func (h *MagicLinkHandler) VerifyMagicLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	entry, ok := h.Repo.GetMagicLink(ctx, token)
	if !ok || security.IsMagicLinkExpired(entry) {
		http.Error(w, "invalid or expired magic link", http.StatusUnauthorized)
		return
	}

	// Get user by email
	user, err := h.Repo.GetByEmail(ctx, entry.Email)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	// Check if account is locked
	if time.Now().Before(user.LockedUntil) {
		http.Error(w, "account locked", http.StatusForbidden)
		return
	}

	// Check if password expired
	if time.Now().After(user.PasswordExpiresAt) {
		http.Error(w, "password expired", http.StatusForbidden)
		return
	}

	// Generate JWT token
	jwtToken, err := security.GenerateToken(user.ID, user.Username, user.Role, h.Config.JWTSecret)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	// Delete used magic link
	h.Repo.DeleteMagicLink(ctx, token)

	// Return token and user info
	response := dto.LoginResponse{
		Token:     jwtToken,
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
