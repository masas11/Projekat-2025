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

	ctx := r.Context()
	if err := h.Repo.Create(ctx, user); err != nil {
		if err == store.ErrUserExists {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "failed to create user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "registration successful, verification email sent",
	})
}
