package handler

import (
	"encoding/json"
	"net/http"

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
