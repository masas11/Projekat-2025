package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"users-service/config"
	"users-service/internal/dto"
	"users-service/internal/logger"
	"users-service/internal/mail"
	"users-service/internal/security"
	"users-service/internal/store"
)

type LoginHandler struct {
	Repo   *store.UserRepository
	Config *config.Config
	Logger *logger.Logger
}

func NewLoginHandler(repo *store.UserRepository, cfg *config.Config, log *logger.Logger) *LoginHandler {
	return &LoginHandler{
		Repo:   repo,
		Config: cfg,
		Logger: log,
	}
}

func (h *LoginHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.LoginRequest
	json.NewDecoder(r.Body).Decode(&req)

	ctx := r.Context()
	ipAddress := getClientIP(r)
	user, err := h.Repo.GetByUsername(ctx, req.Username)
	if err != nil {
		if h.Logger != nil {
			h.Logger.LogLoginFailure(req.Username, "user not found", ipAddress)
		}
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check if email is verified
	if !user.Verified {
		if h.Logger != nil {
			h.Logger.LogLoginFailure(req.Username, "email not verified", ipAddress)
		}
		http.Error(w, "email not verified", http.StatusForbidden)
		return
	}

	if time.Now().Before(user.LockedUntil) {
		if h.Logger != nil {
			h.Logger.LogLoginFailure(req.Username, "account locked", ipAddress)
		}
		http.Error(w, "account locked", http.StatusForbidden)
		return
	}

	if time.Now().After(user.PasswordExpiresAt) {
		if h.Logger != nil {
			h.Logger.LogLoginFailure(req.Username, "password expired", ipAddress)
		}
		http.Error(w, "password expired", http.StatusForbidden)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		user.FailedLoginAttempts++
		if user.FailedLoginAttempts >= 5 {
			user.LockedUntil = time.Now().Add(15 * time.Minute)
		}
		// Update failed login attempts
		h.Repo.Update(ctx, user)
		if h.Logger != nil {
			h.Logger.LogLoginFailure(req.Username, "invalid password", ipAddress)
		}
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	user.FailedLoginAttempts = 0
	h.Repo.Update(ctx, user)

	otp, _ := security.GenerateOTP()
	h.Repo.SetOTP(ctx, user.Username, otp)
	mail.SendOTP(user.Email, otp)

	if h.Logger != nil {
		h.Logger.Log(logger.LevelInfo, logger.EventLoginSuccess, "OTP requested successfully",
			map[string]interface{}{
				"username": user.Username,
				"ip":       ipAddress,
			})
	}

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

	ctx := r.Context()
	ipAddress := getClientIP(r)
	entry, ok := h.Repo.GetOTP(ctx, req.Username)
	if !ok || security.IsExpired(entry) || entry.Code != req.OTP {
		if h.Logger != nil {
			reason := "invalid OTP"
			if !ok {
				reason = "OTP not found"
			} else if security.IsExpired(entry) {
				reason = "OTP expired"
			}
			h.Logger.LogLoginFailure(req.Username, reason, ipAddress)
		}
		http.Error(w, "invalid OTP", http.StatusUnauthorized)
		return
	}

	// Get user to generate token
	user, err := h.Repo.GetByUsername(ctx, req.Username)
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

	h.Repo.DeleteOTP(ctx, req.Username)

	// Log successful login
	if h.Logger != nil {
		h.Logger.LogLoginSuccess(user.Username, ipAddress)
	}

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

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	// Fallback to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// Logout handles user logout (mainly for audit/logging purposes since JWT is stateless)
func (h *LoginHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// With stateless JWT, logout is primarily a client-side operation
	// This endpoint can be used for audit logging or token blacklisting in the future
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "logged out successfully",
	})
}
