package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"notifications-service/internal/model"
	"notifications-service/internal/store"
)

type NotificationHandler struct {
	Repo *store.NotificationRepository
}

func NewNotificationHandler(repo *store.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{Repo: repo}
}

func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "userId query parameter is required", http.StatusBadRequest)
		return
	}

	notifications, err := h.Repo.GetByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get notifications: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notifications)
}

// CreateNotification handles POST /notifications - creates a new notification
func (h *NotificationHandler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID    string `json:"userId"`
		Type      string `json:"type"`
		Message   string `json:"message"`
		ContentID string `json:"contentId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.UserID == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		http.Error(w, "type is required", http.StatusBadRequest)
		return
	}
	if req.Message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}

	notification := &model.Notification{
		ID:        uuid.NewString(),
		UserID:    req.UserID,
		Type:      req.Type,
		Message:   req.Message,
		ContentID: req.ContentID,
		Read:      false,
		CreatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.Repo.Create(ctx, notification); err != nil {
		http.Error(w, "failed to create notification: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(notification)
}
