package handler

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"analytics-service/config"
	"analytics-service/internal/cqrs"
	"analytics-service/internal/model"
	"analytics-service/internal/store"
)

type ActivityHandler struct {
	ActivityStore  *store.ActivityStore  // For backward compatibility
	EventStore     *store.EventStore     // Event Sourcing (2.14)
	ProjectionStore *store.ProjectionStore // CQRS Read Model (2.15)
	CommandHandler *cqrs.CommandHandler  // CQRS Command Handler (2.15)
	QueryHandler   *cqrs.QueryHandler    // CQRS Query Handler (2.15)
	Config         *config.Config
}

func NewActivityHandler(activityStore *store.ActivityStore, eventStore *store.EventStore, projectionStore *store.ProjectionStore, cfg *config.Config) *ActivityHandler {
	commandHandler := cqrs.NewCommandHandler(eventStore)
	queryHandler := cqrs.NewQueryHandler(projectionStore, cfg)
	
	return &ActivityHandler{
		ActivityStore:  activityStore,
		EventStore:     eventStore,
		ProjectionStore: projectionStore,
		CommandHandler: commandHandler,
		QueryHandler:   queryHandler,
		Config:         cfg,
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
		} else {
			log.Printf("Activity stored successfully in ActivityStore: type=%s, userId=%s", activity.Type, activity.UserID)
		}
	}

	// Store as event in Event Store using CQRS Command Handler (2.15)
	if h.CommandHandler != nil {
		cmd := h.activityToCommand(&activity)
		if cmd != nil {
			result := h.CommandHandler.HandleCommand(ctx, cmd)
			if !result.Success {
				log.Printf("Error handling command: %v", result.Error)
				// Fallback to direct event store append
				if h.EventStore != nil {
					event := h.activityToEvent(&activity)
					if err := h.EventStore.AppendEvent(ctx, event); err != nil {
						log.Printf("Error appending event to EventStore: %v", err)
					}
				}
			} else {
				log.Printf("Command handled successfully via CQRS: streamId=%s, eventType=%s", result.Event.StreamID, result.Event.EventType)
				// Process event immediately to update projection (synchronous for now)
				if h.ProjectionStore != nil {
					eventHandler := cqrs.NewEventHandler(h.ProjectionStore, h.Config)
					if err := eventHandler.HandleEvent(ctx, result.Event); err != nil {
						log.Printf("Error processing event for projection: %v", err)
					}
				}
			}
		} else {
			// Fallback to direct event store append if command conversion fails
			if h.EventStore != nil {
				event := h.activityToEvent(&activity)
				if err := h.EventStore.AppendEvent(ctx, event); err != nil {
					log.Printf("Error appending event to EventStore: %v", err)
				}
			}
		}
	} else if h.EventStore != nil {
		// Fallback if command handler is not available
		event := h.activityToEvent(&activity)
		if err := h.EventStore.AppendEvent(ctx, event); err != nil {
			log.Printf("Error appending event to EventStore: %v", err)
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

// activityToCommand converts UserActivity to CQRS Command (2.15)
func (h *ActivityHandler) activityToCommand(activity *model.UserActivity) cqrs.Command {
	switch model.ActivityType(activity.Type) {
	case model.ActivityTypeSongPlayed:
		artistIDs := []string{}
		if activity.ArtistID != "" {
			artistIDs = append(artistIDs, activity.ArtistID)
		}
		return &cqrs.PlaySongCommand{
			UserID:    activity.UserID,
			SongID:    activity.SongID,
			SongName:  activity.SongName,
			Genre:     activity.Genre,
			ArtistIDs: artistIDs,
			Timestamp: activity.Timestamp,
		}
	case model.ActivityTypeRatingGiven:
		return &cqrs.RateSongCommand{
			UserID:    activity.UserID,
			SongID:    activity.SongID,
			Rating:    activity.Rating,
			Timestamp: activity.Timestamp,
		}
	case model.ActivityTypeArtistSubscribed:
		return &cqrs.SubscribeToArtistCommand{
			UserID:     activity.UserID,
			ArtistID:   activity.ArtistID,
			ArtistName: activity.ArtistName,
			Timestamp:  activity.Timestamp,
		}
	case model.ActivityTypeArtistUnsubscribed:
		return &cqrs.UnsubscribeFromArtistCommand{
			UserID:    activity.UserID,
			ArtistID:  activity.ArtistID,
			Timestamp: activity.Timestamp,
		}
	default:
		return nil
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

// GetUserAnalytics returns analytics for a user using CQRS Query Side (2.15)
func (h *ActivityHandler) GetUserAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "userId parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Use CQRS Query Handler to read from read model (projection)
	query := &cqrs.GetUserAnalyticsQuery{
		UserID: userID,
	}
	
	result := h.QueryHandler.HandleQuery(ctx, query)
	if result.Error != nil {
		log.Printf("Error getting analytics from query handler: %v", result.Error)
		// Fallback to old method if query handler fails
		h.GetUserAnalyticsLegacy(w, r)
		return
	}

	if result.Analytics == nil {
		// If projection is empty, try to rebuild from events or use legacy method
		log.Printf("Projection is empty for user %s, trying legacy method", userID)
		h.GetUserAnalyticsLegacy(w, r)
		return
	}

	// Check if projection has any data - if not, try to rebuild from events or use legacy method
	if result.Analytics.TotalSongsPlayed == 0 {
		log.Printf("Projection has no songs played for user %s, checking events and activities", userID)
		
		// Try to rebuild projection from events first
		if h.EventStore != nil {
			events, err := h.EventStore.GetEventStream(ctx, userID, 0, 0)
			if err == nil && len(events) > 0 {
				log.Printf("Found %d events for user %s, processing to update projection", len(events), userID)
				// Process events to update projection
				eventHandler := cqrs.NewEventHandler(h.ProjectionStore, h.Config)
				for _, event := range events {
					if err := eventHandler.HandleEvent(ctx, event); err != nil {
						log.Printf("Error processing event for projection: %v", err)
					}
				}
				// Try query again
				result = h.QueryHandler.HandleQuery(ctx, query)
				if result.Analytics != nil && result.Analytics.TotalSongsPlayed > 0 {
					log.Printf("Projection updated successfully, returning analytics")
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					if err := json.NewEncoder(w).Encode(result.Analytics); err != nil {
						log.Printf("Error encoding analytics: %v", err)
					}
					return
				}
			}
		}
		
		// If still no data, use legacy method which reads from ActivityStore
		log.Printf("No data in projection or events, using legacy method for user %s", userID)
		h.GetUserAnalyticsLegacy(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result.Analytics); err != nil {
		log.Printf("Error encoding analytics: %v", err)
	}
}

// GetUserAnalyticsLegacy is the old implementation (kept for backward compatibility)
func (h *ActivityHandler) GetUserAnalyticsLegacy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "userId parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Initialize analytics response
	analytics := &model.UserAnalytics{
		UserID:             userID,
		SongsPlayedByGenre: make(map[string]int),
		Top5Artists:        []model.ArtistPlayCount{},
	}

	// 1. Get all song play activities
	activities, err := h.ActivityStore.GetByUserIDAndType(ctx, userID, model.ActivityTypeSongPlayed, 0) // 0 = no limit
	if err != nil {
		log.Printf("Error getting activities: %v", err)
		http.Error(w, "failed to get activities", http.StatusInternalServerError)
		return
	}

	analytics.TotalSongsPlayed = len(activities)

	// Create HTTP client for service calls
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: tr,
	}

	// Maps to track data
	genreCount := make(map[string]int)
	artistCount := make(map[string]int) // artistID -> count
	artistNames := make(map[string]string) // artistID -> name
	songIDs := make(map[string]bool) // unique song IDs
	songMap := make(map[string]struct { // songID -> song data
		Genre     string
		ArtistIDs []string
	})

	// Collect unique song IDs from activities
	for _, activity := range activities {
		if activity.SongID != "" {
			songIDs[activity.SongID] = true
		}
	}

	// Load all songs from content service at once
	if len(songIDs) > 0 {
		songsURL := h.Config.ContentServiceURL + "/songs"
		songsReq, err := http.NewRequestWithContext(ctx, "GET", songsURL, nil)
		if err == nil {
			songsResp, err := client.Do(songsReq)
			if err == nil && songsResp != nil && songsResp.StatusCode == http.StatusOK {
				var allSongs []struct {
					ID        string   `json:"id"`
					Genre     string   `json:"genre"`
					ArtistIDs []string `json:"artistIds"`
				}
				if json.NewDecoder(songsResp.Body).Decode(&allSongs) == nil {
					// Build song map for quick lookup
					for _, song := range allSongs {
						if songIDs[song.ID] {
							songMap[song.ID] = struct {
								Genre     string
								ArtistIDs []string
							}{
								Genre:     song.Genre,
								ArtistIDs: song.ArtistIDs,
							}
						}
					}
				}
				songsResp.Body.Close()
			}
		}
	}

	// 2. Process activities to get genre and artist counts
	for _, activity := range activities {
		if activity.SongID == "" {
			continue
		}

		// Try to get song data from map
		songData, found := songMap[activity.SongID]
		if !found {
			// Fallback: try to get from activity if available
			if activity.Genre != "" {
				genreCount[activity.Genre]++
			}
			if activity.ArtistID != "" {
				artistCount[activity.ArtistID]++
				if activity.ArtistName != "" {
					artistNames[activity.ArtistID] = activity.ArtistName
				}
			}
			continue
		}

		// Count by genre
		if songData.Genre != "" {
			genreCount[songData.Genre]++
		}

		// Count by artist
		for _, artistID := range songData.ArtistIDs {
			if artistID != "" {
				artistCount[artistID]++
				// Try to get artist name from activity
				if activity.ArtistID == artistID && activity.ArtistName != "" {
					artistNames[artistID] = activity.ArtistName
				}
			}
		}
	}

	analytics.SongsPlayedByGenre = genreCount

	// 3. Get top 5 artists - fetch artist names from content service
	type artistData struct {
		ID       string
		Name     string
		PlayCount int
	}
	artists := make([]artistData, 0, len(artistCount))
	
	// Fetch all artists from content service to get names
	artistsURL := h.Config.ContentServiceURL + "/artists"
	artistsReq, err := http.NewRequestWithContext(ctx, "GET", artistsURL, nil)
	if err == nil {
		artistsResp, err := client.Do(artistsReq)
		if err == nil && artistsResp != nil && artistsResp.StatusCode == http.StatusOK {
			var allArtists []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			if json.NewDecoder(artistsResp.Body).Decode(&allArtists) == nil {
				// Build artist name map
				for _, artist := range allArtists {
					if artistNames[artist.ID] == "" {
						artistNames[artist.ID] = artist.Name
					}
				}
			}
			artistsResp.Body.Close()
		}
	}
	
	// Build artist list with names
	for artistID, count := range artistCount {
		name := artistNames[artistID]
		if name == "" {
			// Fallback: try to get from content service individually
			artistURL := h.Config.ContentServiceURL + "/artists/" + artistID
			artistReq, err := http.NewRequestWithContext(ctx, "GET", artistURL, nil)
			if err == nil {
				artistResp, err := client.Do(artistReq)
				if err == nil && artistResp != nil && artistResp.StatusCode == http.StatusOK {
					var artist struct {
						Name string `json:"name"`
					}
					if json.NewDecoder(artistResp.Body).Decode(&artist) == nil {
						name = artist.Name
						artistNames[artistID] = name
					}
					artistResp.Body.Close()
				}
			}
			// If still no name, use a formatted version of ID
			if name == "" {
				name = "Umetnik " + artistID[:8] + "..."
			}
		}
		artists = append(artists, artistData{
			ID:        artistID,
			Name:      name,
			PlayCount: count,
		})
	}

	// Sort by play count descending
	sort.Slice(artists, func(i, j int) bool {
		return artists[i].PlayCount > artists[j].PlayCount
	})

	// Take top 5
	top5Count := 5
	if len(artists) < top5Count {
		top5Count = len(artists)
	}
	analytics.Top5Artists = make([]model.ArtistPlayCount, top5Count)
	for i := 0; i < top5Count; i++ {
		analytics.Top5Artists[i] = model.ArtistPlayCount{
			ArtistID:   artists[i].ID,
			ArtistName: artists[i].Name,
			PlayCount:  artists[i].PlayCount,
		}
	}

	// 4. Get average rating from ratings service
	ratingsURL := h.Config.RatingsServiceURL + "/ratings-by-user?userId=" + userID
	ratingsReq, err := http.NewRequestWithContext(ctx, "GET", ratingsURL, nil)
	if err == nil {
		ratingsResp, err := client.Do(ratingsReq)
		if err == nil && ratingsResp != nil && ratingsResp.StatusCode == http.StatusOK {
			var ratings []struct {
				Rating int `json:"rating"`
			}
			if json.NewDecoder(ratingsResp.Body).Decode(&ratings) == nil {
				if len(ratings) > 0 {
					sum := 0
					for _, r := range ratings {
						sum += r.Rating
					}
					analytics.AverageRating = float64(sum) / float64(len(ratings))
				}
			}
			ratingsResp.Body.Close()
		}
	}

	// 5. Get subscribed artists count from subscriptions service
	subsURL := h.Config.SubscriptionsServiceURL + "/subscriptions?userId=" + userID
	subsReq, err := http.NewRequestWithContext(ctx, "GET", subsURL, nil)
	if err == nil {
		subsResp, err := client.Do(subsReq)
		if err == nil && subsResp != nil && subsResp.StatusCode == http.StatusOK {
			var subscriptions []struct {
				Type     string `json:"type"`
				ArtistID string `json:"artistId,omitempty"`
			}
			if json.NewDecoder(subsResp.Body).Decode(&subscriptions) == nil {
				count := 0
				uniqueArtists := make(map[string]bool)
				for _, sub := range subscriptions {
					if sub.Type == "artist" && sub.ArtistID != "" {
						// Count unique artists (in case of duplicates)
						if !uniqueArtists[sub.ArtistID] {
							uniqueArtists[sub.ArtistID] = true
							count++
						}
					}
				}
				analytics.SubscribedArtistsCount = count
			}
			subsResp.Body.Close()
		} else {
			if subsResp != nil {
				log.Printf("Error getting subscriptions: %v, status: %d", err, subsResp.StatusCode)
			} else {
				log.Printf("Error getting subscriptions: %v (no response)", err)
			}
		}
	} else {
		log.Printf("Error creating subscriptions request: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(analytics); err != nil {
		log.Printf("Error encoding analytics: %v", err)
	}
}
