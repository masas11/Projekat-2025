package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"users-service/internal/dto"
	"users-service/internal/model"
	"users-service/internal/store"
	"users-service/internal/validation"
)

type RegisterHandler struct {
	Store *store.UserStore
}

func NewRegisterHandler(s *store.UserStore) *RegisterHandler {
	return &RegisterHandler{Store: s}
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

	// basic required fields validation
	if req.FirstName == "" || req.LastName == "" ||
		req.Email == "" || req.Username == "" ||
		req.Password == "" || req.ConfirmPassword == "" {
		http.Error(w, "all fields are required", http.StatusBadRequest)
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
		Verified:          true,
		PasswordChangedAt: now,
		PasswordExpiresAt: now.Add(60 * 24 * time.Hour),
		CreatedAt:         now,
	}

	if err := h.Store.AddUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "registration successful, verification email sent",
	})
}
