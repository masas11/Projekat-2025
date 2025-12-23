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
	Store *store.UserStore
}

func NewPasswordHandler(store *store.UserStore) *PasswordHandler {
	return &PasswordHandler{Store: store}
}

// CHANGE PASSWORD (must be 1 day old)
func (h *PasswordHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ChangePasswordRequest
	json.NewDecoder(r.Body).Decode(&req)

	user, err := h.Store.GetByUsername(req.Username)
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

	w.WriteHeader(http.StatusOK)
}

// RESET PASSWORD (email link simulated)
func (h *PasswordHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ResetPasswordRequest
	json.NewDecoder(r.Body).Decode(&req)

	user, err := h.Store.GetByUsername(req.Username)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	user.PasswordHash = string(hash)
	user.PasswordChangedAt = time.Now()
	user.PasswordExpiresAt = time.Now().Add(60 * 24 * time.Hour)

	w.WriteHeader(http.StatusOK)
}
