package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"analytics-service/internal/model"
	"analytics-service/internal/store"
)

type ActivityHandler struct {
	ActivityStore *store.ActivityStore
}

func NewActivityHandler(activityStore *store.ActivityStore) *ActivityHandler {
	return &ActivityHandler{
		ActivityStore: activityStore,
	}
}

// LogActivity logs a user activity
func (h *ActivityHandler) LogActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var activity model.UserActivity
	if err := json.NewDecoder(r.Body).Decode(&activity); err != nil {
		log.Printf("Error decoding activity JSON: %v", err)
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if activity.UserID == "" {
		log.Printf("Activity missing userId")
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	if activity.Type == "" {
		log.Printf("Activity missing type for userId: %s", activity.UserID)
		http.Error(w, "type is required", http.StatusBadRequest)
		return
	}

	log.Printf("Received activity log: type=%s, userId=%s, songId=%s", activity.Type, activity.UserID, activity.SongID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.ActivityStore.Create(ctx, &activity); err != nil {
		log.Printf("Error logging activity: %v", err)
		http.Error(w, "failed to log activity", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(activity)
}

// GetUserActivities retrieves activities for a user
func (h *ActivityHandler) GetUserActivities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "userId parameter is required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	activityType := r.URL.Query().Get("type")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var activities []*model.UserActivity
	var err error

	if activityType != "" {
		activities, err = h.ActivityStore.GetByUserIDAndType(ctx, userID, model.ActivityType(activityType), limit)
	} else {
		activities, err = h.ActivityStore.GetByUserID(ctx, userID, limit)
	}

	if err != nil {
		log.Printf("Error getting activities: %v", err)
		http.Error(w, "failed to get activities", http.StatusInternalServerError)
		return
	}

	log.Printf("Retrieved %d activities for user %s (type filter: %s)", len(activities), userID, activityType)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(activities); err != nil {
		log.Printf("Error encoding activities: %v", err)
	}
}
