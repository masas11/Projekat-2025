package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"users-service/internal/security"
	"users-service/internal/store"
)

type VerificationHandler struct {
	Repo *store.UserRepository
}

func NewVerificationHandler(repo *store.UserRepository) *VerificationHandler {
	return &VerificationHandler{Repo: repo}
}

// VerifyEmail verifies user's email using token from registration
func (h *VerificationHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	log.Printf("[VERIFY EMAIL] Received token from query: %s (length: %d)", token, len(token))
	log.Printf("[VERIFY EMAIL] Full URL: %s", r.URL.String())
	
	if token == "" {
		log.Printf("[VERIFY EMAIL] ERROR: Token is empty")
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	entry, ok := h.Repo.GetVerificationToken(ctx, token)
	log.Printf("[VERIFY EMAIL] Token lookup result: found=%v, expired=%v", ok, !ok || security.IsVerificationTokenExpired(entry))
	
	if !ok || security.IsVerificationTokenExpired(entry) {
		http.Error(w, "invalid or expired verification token", http.StatusUnauthorized)
		return
	}

	// Get user by email
	user, err := h.Repo.GetByEmail(ctx, entry.Email)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Verify user
	user.Verified = true
	if err := h.Repo.Update(ctx, user); err != nil {
		http.Error(w, "failed to verify user", http.StatusInternalServerError)
		return
	}

	// Delete used verification token
	h.Repo.DeleteVerificationToken(ctx, token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "email verified successfully",
	})
}
