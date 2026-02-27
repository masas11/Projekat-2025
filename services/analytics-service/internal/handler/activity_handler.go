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
	ActivityStore *store.ActivityStore // For backward compatibility
	EventStore    *store.EventStore    // Event Sourcing (2.14)
}

func NewActivityHandler(activityStore *store.ActivityStore, eventStore *store.EventStore) *ActivityHandler {
	return &ActivityHandler{
		ActivityStore: activityStore,
		EventStore:    eventStore,
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

	// Store in ActivityStore for backward compatibility
	if h.ActivityStore != nil {
		if err := h.ActivityStore.Create(ctx, &activity); err != nil {
			log.Printf("Error logging activity to ActivityStore: %v", err)
		}
	}

	// Store as event in Event Store (2.14 Event Sourcing)
	if h.EventStore != nil {
		event := h.activityToEvent(&activity)
		if err := h.EventStore.AppendEvent(ctx, event); err != nil {
			log.Printf("Error appending event to EventStore: %v", err)
			// Don't fail the request if event store fails, but log it
		} else {
			log.Printf("Event appended successfully: streamId=%s, eventType=%s, version=%d", event.StreamID, event.EventType, event.Version)
		}
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

// activityToEvent converts UserActivity to UserEvent for Event Sourcing
func (h *ActivityHandler) activityToEvent(activity *model.UserActivity) *model.UserEvent {
	payload := make(map[string]interface{})
	
	if activity.SongID != "" {
		payload["songId"] = activity.SongID
	}
	if activity.SongName != "" {
		payload["songName"] = activity.SongName
	}
	if activity.Rating > 0 {
		payload["rating"] = activity.Rating
	}
	if activity.Genre != "" {
		payload["genre"] = activity.Genre
	}
	if activity.ArtistID != "" {
		payload["artistId"] = activity.ArtistID
	}
	if activity.ArtistName != "" {
		payload["artistName"] = activity.ArtistName
	}
	
	return &model.UserEvent{
		EventType: model.EventType(activity.Type),
		StreamID:  activity.UserID,
		Timestamp: activity.Timestamp,
		Payload:   payload,
	}
}

// GetEventStream retrieves the event stream for a user (2.14 Event Sourcing)
func (h *ActivityHandler) GetEventStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "userId parameter is required", http.StatusBadRequest)
		return
	}

	fromVersionStr := r.URL.Query().Get("fromVersion")
	fromVersion := int64(0)
	if fromVersionStr != "" {
		if parsed, err := strconv.ParseInt(fromVersionStr, 10, 64); err == nil {
			fromVersion = parsed
		}
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 0 // 0 means all events
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	events, err := h.EventStore.GetEventStream(ctx, userID, fromVersion, limit)
	if err != nil {
		log.Printf("Error getting event stream: %v", err)
		http.Error(w, "failed to get event stream", http.StatusInternalServerError)
		return
	}

	log.Printf("Retrieved %d events for stream %s (fromVersion: %d)", len(events), userID, fromVersion)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(events); err != nil {
		log.Printf("Error encoding events: %v", err)
	}
}

// ReplayEvents reconstructs the state by replaying all events for a user (2.14 Event Sourcing)
func (h *ActivityHandler) ReplayEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "userId parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	state, err := h.EventStore.ReplayEvents(ctx, userID)
	if err != nil {
		log.Printf("Error replaying events: %v", err)
		http.Error(w, "failed to replay events", http.StatusInternalServerError)
		return
	}

	log.Printf("Replayed events for user %s: %d total songs played, %d ratings given", userID, state.TotalSongsPlayed, state.TotalRatingsGiven)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(state); err != nil {
		log.Printf("Error encoding state: %v", err)
	}
}
