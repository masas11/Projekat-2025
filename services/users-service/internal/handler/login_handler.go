package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"users-service/config"
	"users-service/internal/dto"
	"users-service/internal/mail"
	"users-service/internal/security"
	"users-service/internal/store"
)

type LoginHandler struct {
	Store  *store.UserStore
	Config *config.Config
}

func NewLoginHandler(s *store.UserStore, cfg *config.Config) *LoginHandler {
	return &LoginHandler{
		Store:  s,
		Config: cfg,
	}
}

func (h *LoginHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.LoginRequest
	json.NewDecoder(r.Body).Decode(&req)

	user, err := h.Store.GetByUsername(req.Username)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if time.Now().Before(user.LockedUntil) {
		http.Error(w, "account locked", http.StatusForbidden)
		return
	}

	if time.Now().After(user.PasswordExpiresAt) {
		http.Error(w, "password expired", http.StatusForbidden)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		user.FailedLoginAttempts++
		if user.FailedLoginAttempts >= 5 {
			user.LockedUntil = time.Now().Add(15 * time.Minute)
		}
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	user.FailedLoginAttempts = 0

	otp, _ := security.GenerateOTP()
	h.Store.SetOTP(user.Username, otp)
	mail.SendOTP(user.Email, otp)

	w.WriteHeader(http.StatusOK)
}

func (h *LoginHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.OTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	entry, ok := h.Store.GetOTP(req.Username)
	if !ok || security.IsExpired(entry) || entry.Code != req.OTP {
		http.Error(w, "invalid OTP", http.StatusUnauthorized)
		return
	}

	// Get user to generate token
	user, err := h.Store.GetByUsername(req.Username)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := security.GenerateToken(user.ID, user.Username, user.Role, h.Config.JWTSecret)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	h.Store.DeleteOTP(req.Username)

	// Return token and user info
	response := dto.LoginResponse{
		Token:     token,
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
