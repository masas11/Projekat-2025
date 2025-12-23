package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"users-service/internal/dto"
	"users-service/internal/mail"
	"users-service/internal/security"
	"users-service/internal/store"
)

type LoginHandler struct {
	Store *store.UserStore
}

func NewLoginHandler(s *store.UserStore) *LoginHandler {
	return &LoginHandler{Store: s}
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
	var req dto.OTPRequest
	json.NewDecoder(r.Body).Decode(&req)

	entry, ok := h.Store.GetOTP(req.Username)
	if !ok || security.IsExpired(entry) || entry.Code != req.OTP {
		http.Error(w, "invalid OTP", http.StatusUnauthorized)
		return
	}

	h.Store.DeleteOTP(req.Username)
	w.WriteHeader(http.StatusOK)
}
