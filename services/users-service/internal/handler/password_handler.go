package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"users-service/internal/dto"
	"users-service/internal/store"
)

type PasswordHandler struct {
	Repo *store.UserRepository
}

func NewPasswordHandler(repo *store.UserRepository) *PasswordHandler {
	return &PasswordHandler{Repo: repo}
}

// CHANGE PASSWORD (must be 1 day old)
func (h *PasswordHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ChangePasswordRequest
	json.NewDecoder(r.Body).Decode(&req)

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
	user.PasswordExpiresAt = time.Now().Add(60 * 24 * time.Hour)

	if err := h.Repo.Update(ctx, user); err != nil {
		http.Error(w, "failed to update password", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RESET PASSWORD (email link simulated)
func (h *PasswordHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ResetPasswordRequest
	json.NewDecoder(r.Body).Decode(&req)

	ctx := r.Context()
	user, err := h.Repo.GetByUsername(ctx, req.Username)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	user.PasswordHash = string(hash)
	user.PasswordChangedAt = time.Now()
	user.PasswordExpiresAt = time.Now().Add(60 * 24 * time.Hour)

	if err := h.Repo.Update(ctx, user); err != nil {
		http.Error(w, "failed to reset password", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
