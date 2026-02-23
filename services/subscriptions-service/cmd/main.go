package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"subscriptions-service/config"
	"subscriptions-service/internal/model"
	"subscriptions-service/internal/store"
)

// Circuit Breaker implementation for resilience
type CircuitBreaker struct {
	mu           sync.RWMutex
	maxFailures  int
	failures     int
	lastFailTime time.Time
	state        string // "closed", "open", "half-open"
	resetTimeout time.Duration
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        "closed",
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if circuit should reset
	if cb.state == "open" {
		timeSinceFail := time.Since(cb.lastFailTime)
		if timeSinceFail > cb.resetTimeout {
			cb.state = "half-open"
			cb.failures = 0
			log.Println("Circuit breaker transitioning to half-open")
		} else {
			return &CircuitBreakerError{"Circuit breaker is open"}
		}
	}

	err := fn()
	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()
		if cb.failures >= cb.maxFailures {
			cb.state = "open"
			log.Printf("Circuit breaker opened after %d failures", cb.failures)
		}
		return err
	}

	// Success - reset failures
	cb.failures = 0
	if cb.state == "half-open" {
		cb.state = "closed"
		log.Println("Circuit breaker closed again")
	}
	return nil
}

type CircuitBreakerError struct {
	Message string
}

func (e *CircuitBreakerError) Error() string {
	return e.Message
}

// Check if artist exists by ID with circuit breaker and fallback
func checkArtistExists(client *http.Client, contentURL, artistID string, cb *CircuitBreaker) bool {
	checkURL := contentURL + "/artists/" + url.QueryEscape(artistID)

	// Use circuit breaker for the call
	var exists bool
	var callErr error

	err := cb.Call(func() error {
		// Retry logic with timeout
		for i := 0; i < 2; i++ { // retry 2 times
			resp, err := client.Get(checkURL)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					exists = true
					return nil
				}
			}
			if i < 1 { // Don't sleep on last attempt
				log.Printf("Retrying call to content-service for artist %s... (attempt %d)", artistID, i+1)
				time.Sleep(100 * time.Millisecond)
			}
		}
		callErr = fmt.Errorf("failed to verify artist existence")
		return callErr
	})

	if err != nil {
		if _, ok := err.(*CircuitBreakerError); ok {
			log.Printf("Circuit breaker open for artist %s, using fallback", artistID)
		} else {
			log.Printf("Error checking artist %s: %v, using fallback", artistID, err)
		}
		// Fallback logic: return false (artist not found) when service is unavailable
		return false
	}

	return exists
}

func addCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// createNotification sends a notification creation request to notifications-service
func createNotification(notificationsServiceURL, userID, notifType, message, contentID string) {
	notificationData := map[string]interface{}{
		"userId":    userID,
		"type":      notifType,
		"message":   message,
		"contentId": contentID,
	}

	jsonData, err := json.Marshal(notificationData)
	if err != nil {
		log.Printf("Failed to marshal notification: %v", err)
		return
	}

	// Configure TLS transport for HTTPS support
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	client := &http.Client{
		Timeout:   5 * time.Second, // Increased timeout from 2 to 5 seconds
		Transport: tr,
	}

	// Retry mechanism: try up to 3 times
	maxRetries := 3
	retryDelay := 500 * time.Millisecond
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("POST", notificationsServiceURL+"/notifications", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Failed to create notification request: %v", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			if attempt < maxRetries {
				log.Printf("Failed to create notification (attempt %d/%d): %v, retrying...", attempt, maxRetries, err)
				time.Sleep(retryDelay)
				retryDelay *= 2 // Exponential backoff
				continue
			}
			log.Printf("Failed to create notification after %d attempts: %v", maxRetries, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			log.Printf("Notification created for user %s: %s", userID, message)
			return
		}

		if attempt < maxRetries {
			log.Printf("Notifications-service returned non-OK status: %d (attempt %d/%d), retrying...", resp.StatusCode, attempt, maxRetries)
			time.Sleep(retryDelay)
			retryDelay *= 2
			continue
		}

		log.Printf("Notifications-service returned non-OK status after %d attempts: %d", maxRetries, resp.StatusCode)
		return
	}
}

// handleNewArtistEvent processes new artist events
func handleNewArtistEvent(ctx context.Context, event map[string]interface{}, repo *store.SubscriptionRepository, cfg *config.Config) {
	artistID, _ := event["artistId"].(string)
	artistName, _ := event["name"].(string)
	genres, _ := event["genres"].([]interface{})

	if artistID == "" {
		log.Printf("Invalid new_artist event: missing artistId")
		return
	}

	// Get all genre subscriptions
	genreMap := make(map[string]bool)
	for _, g := range genres {
		if genreStr, ok := g.(string); ok {
			genreMap[genreStr] = true
		}
	}

	// For each genre, find subscribers and notify them
	for genre := range genreMap {
		subscriptions, err := repo.GetByGenre(ctx, genre)
		if err != nil {
			log.Printf("Error getting subscriptions for genre %s: %v", genre, err)
			continue
		}

		for _, sub := range subscriptions {
			message := fmt.Sprintf("New artist '%s' in genre %s has been added", artistName, genre)
			createNotification(cfg.NotificationsServiceURL, sub.UserID, "new_artist", message, artistID)
		}
	}

	log.Printf("Processed new_artist event for artist %s", artistID)
}

// handleNewAlbumEvent processes new album events
func handleNewAlbumEvent(ctx context.Context, event map[string]interface{}, repo *store.SubscriptionRepository, cfg *config.Config) {
	albumID, _ := event["albumId"].(string)
	albumName, _ := event["name"].(string)
	genre, _ := event["genre"].(string)
	artistIDs, _ := event["artistIds"].([]interface{})
	artistNamesInterface, _ := event["artistNames"].([]interface{})

	if albumID == "" {
		log.Printf("Invalid new_album event: missing albumId")
		return
	}

	// Convert artist names to string slice
	artistNames := make([]string, 0, len(artistNamesInterface))
	for _, nameInterface := range artistNamesInterface {
		if name, ok := nameInterface.(string); ok {
			artistNames = append(artistNames, name)
		}
	}

	// Format artist names for message
	artistNamesStr := "artist"
	if len(artistNames) > 0 {
		if len(artistNames) == 1 {
			artistNamesStr = artistNames[0]
		} else if len(artistNames) == 2 {
			artistNamesStr = artistNames[0] + " and " + artistNames[1]
		} else {
			artistNamesStr = artistNames[0] + " and others"
		}
	}

	// Track notified users to avoid duplicate notifications
	notifiedUsers := make(map[string]bool)

	// For each artist, find subscribers and notify them
	for _, artistIDInterface := range artistIDs {
		artistID, ok := artistIDInterface.(string)
		if !ok {
			continue
		}

		subscriptions, err := repo.GetByArtistID(ctx, artistID)
		if err != nil {
			log.Printf("Error getting subscriptions for artist %s: %v", artistID, err)
			continue
		}

		for _, sub := range subscriptions {
			if !notifiedUsers[sub.UserID] {
				message := fmt.Sprintf("New album '%s' by %s has been released", albumName, artistNamesStr)
				createNotification(cfg.NotificationsServiceURL, sub.UserID, "new_album", message, albumID)
				notifiedUsers[sub.UserID] = true
			}
		}
	}

	// Also notify users subscribed to the genre (if genre is provided)
	if genre != "" {
		log.Printf("[GENRE] Checking genre subscriptions for album genre: '%s'", genre)
		genreSubscriptions, err := repo.GetByGenre(ctx, genre)
		if err != nil {
			log.Printf("[GENRE] Error getting subscriptions for genre '%s': %v", genre, err)
		} else {
			log.Printf("[GENRE] Found %d subscriptions for album genre '%s'", len(genreSubscriptions), genre)
			for _, sub := range genreSubscriptions {
				// Only notify if user hasn't been notified already (to avoid duplicates)
				if !notifiedUsers[sub.UserID] {
					message := fmt.Sprintf("New album '%s' in genre %s has been released", albumName, genre)
					log.Printf("[GENRE] Creating genre notification for user %s: %s", sub.UserID, message)
					createNotification(cfg.NotificationsServiceURL, sub.UserID, "new_album", message, albumID)
					notifiedUsers[sub.UserID] = true
				} else {
					log.Printf("[GENRE] User %s already notified (skipping genre notification to avoid duplicate)", sub.UserID)
				}
			}
		}
	} else {
		log.Printf("[GENRE] No genre provided in album event, skipping genre notifications")
	}

	log.Printf("Processed new_album event for album %s", albumID)
}

// handleNewSongEvent processes new song events
func handleNewSongEvent(ctx context.Context, event map[string]interface{}, repo *store.SubscriptionRepository, cfg *config.Config) {
	// Log entire event for debugging
	eventJSON, _ := json.Marshal(event)
	log.Printf("[DEBUG] handleNewSongEvent called with event: %s", string(eventJSON))
	
	songID, _ := event["songId"].(string)
	songName, _ := event["name"].(string)
	genre, _ := event["genre"].(string)
	artistIDs, _ := event["artistIds"].([]interface{})
	artistNamesInterface, _ := event["artistNames"].([]interface{})

	log.Printf("[DEBUG] Processing new_song event - SongID: %s, Name: %s, Genre: '%s' (empty: %v)", songID, songName, genre, genre == "")

	if songID == "" {
		log.Printf("Invalid new_song event: missing songId")
		return
	}

	// Convert artist names to string slice
	artistNames := make([]string, 0, len(artistNamesInterface))
	for _, nameInterface := range artistNamesInterface {
		if name, ok := nameInterface.(string); ok {
			artistNames = append(artistNames, name)
		}
	}

	// Format artist names for message
	artistNamesStr := "artist"
	if len(artistNames) > 0 {
		if len(artistNames) == 1 {
			artistNamesStr = artistNames[0]
		} else if len(artistNames) == 2 {
			artistNamesStr = artistNames[0] + " and " + artistNames[1]
		} else {
			artistNamesStr = artistNames[0] + " and others"
		}
	}

	// Track notified users to avoid duplicate notifications
	notifiedUsers := make(map[string]bool)

	// For each artist, find subscribers and notify them
	for _, artistIDInterface := range artistIDs {
		artistID, ok := artistIDInterface.(string)
		if !ok {
			continue
		}

		subscriptions, err := repo.GetByArtistID(ctx, artistID)
		if err != nil {
			log.Printf("Error getting subscriptions for artist %s: %v", artistID, err)
			continue
		}

		for _, sub := range subscriptions {
			if !notifiedUsers[sub.UserID] {
				message := fmt.Sprintf("New song '%s' by %s has been added", songName, artistNamesStr)
				createNotification(cfg.NotificationsServiceURL, sub.UserID, "new_song", message, songID)
				notifiedUsers[sub.UserID] = true
			}
		}
	}

	// Also notify users subscribed to the genre (if genre is provided)
	if genre != "" {
		log.Printf("[GENRE] Checking genre subscriptions for genre: '%s' (length: %d)", genre, len(genre))
		genreSubscriptions, err := repo.GetByGenre(ctx, genre)
		if err != nil {
			log.Printf("[GENRE] Error getting subscriptions for genre '%s': %v", genre, err)
		} else {
			log.Printf("[GENRE] Found %d subscriptions for genre '%s'", len(genreSubscriptions), genre)
			if len(genreSubscriptions) == 0 {
				log.Printf("[GENRE] WARNING: No subscriptions found for genre '%s' - this might indicate a matching issue", genre)
			}
			for _, sub := range genreSubscriptions {
				// Only notify if user hasn't been notified already (to avoid duplicates)
				if !notifiedUsers[sub.UserID] {
					message := fmt.Sprintf("New song '%s' in genre %s has been added", songName, genre)
					log.Printf("[GENRE] Creating genre notification for user %s: %s", sub.UserID, message)
					createNotification(cfg.NotificationsServiceURL, sub.UserID, "new_song", message, songID)
					notifiedUsers[sub.UserID] = true
				} else {
					log.Printf("[GENRE] User %s already notified (skipping genre notification to avoid duplicate)", sub.UserID)
				}
			}
		}
	} else {
		log.Printf("[GENRE] No genre provided in event, skipping genre notifications")
	}

	log.Printf("Processed new_song event for song %s", songID)
}

func main() {
	cfg := config.Load()

	// Initialize MongoDB connection
	dbStore, err := store.NewMongoDBStore(cfg.MongoDBURI, cfg.MongoDBDatabase)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer dbStore.Close()
	log.Println("Connected to MongoDB")

	// Initialize repository
	subscriptionRepo := store.NewSubscriptionRepository(dbStore.Database)

	// HTTP client configuration with timeout (2.7.1, 2.7.2)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	client := &http.Client{
		Timeout:   2 * time.Second, // Request timeout (2.7.2)
		Transport: tr,
	}

	// Circuit breaker for content service calls (2.7.4)
	contentServiceCB := NewCircuitBreaker(3, 5*time.Second) // Open after 3 failures, reset after 5 seconds

	mux := http.NewServeMux()

	// emitSubscriptionEvent sends subscription event to recommendation-service asynchronously
	emitSubscriptionEvent := func(recommendationServiceURL, userID, genre, eventType string) {
		go func() {
			event := map[string]interface{}{
				"type":   eventType,
				"userId": userID,
				"genre":  genre,
			}

			eventJSON, err := json.Marshal(event)
			if err != nil {
				log.Printf("Failed to marshal subscription event: %v", err)
				return
			}

			url := recommendationServiceURL + "/events"
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(eventJSON))
			if err != nil {
				log.Printf("Failed to create subscription event request: %v", err)
				return
			}

			req.Header.Set("Content-Type", "application/json")

			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{
				Timeout:   5 * time.Second,
				Transport: tr,
			}

			log.Printf("Sending subscription event to %s: type=%s, userId=%s, genre=%s", url, eventType, userID, genre)
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Failed to emit subscription event to recommendation-service: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
				log.Printf("Recommendation-service returned non-OK status for subscription event: %d", resp.StatusCode)
				return
			}

			log.Printf("Subscription event emitted successfully: %s", eventType)
		}()
	}

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("subscriptions-service is running"))
	})

	// Get all subscriptions for a user
	mux.HandleFunc("/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		addCORSHeaders(w)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := r.URL.Query().Get("userId")
		if userID == "" {
			http.Error(w, "userId parameter is required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		subscriptions, err := subscriptionRepo.GetByUserID(ctx, userID)
		if err != nil {
			log.Printf("Error getting subscriptions: %v", err)
			http.Error(w, "failed to get subscriptions", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(subscriptions)
	})

	// Subscribe to artist endpoint with synchronous validation
	mux.HandleFunc("/subscribe-artist", func(w http.ResponseWriter, r *http.Request) {
		addCORSHeaders(w)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodPost {
			artistID := r.URL.Query().Get("artistId")
			if artistID == "" {
				http.Error(w, "artistId parameter is required", http.StatusBadRequest)
				return
			}

			userID := r.URL.Query().Get("userId")
			if userID == "" {
				http.Error(w, "userId parameter is required", http.StatusBadRequest)
				return
			}

			// Check if already subscribed
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			existing, err := subscriptionRepo.GetByUserAndArtist(ctx, userID, artistID)
			if err != nil {
				log.Printf("Error checking subscription: %v", err)
				http.Error(w, "failed to check subscription", http.StatusInternalServerError)
				return
			}
			if existing != nil {
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte("Already subscribed to this artist"))
				return
			}

			// Synchronous call to check if artist exists with circuit breaker (2.5, 2.7.4)
			exists := checkArtistExists(client, cfg.ContentServiceURL, artistID, contentServiceCB)
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Artist not found"))
				return
			}

			// Create subscription
			subscription := &model.Subscription{
				UserID:   userID,
				Type:     "artist",
				ArtistID: artistID,
			}

			err = subscriptionRepo.Create(ctx, subscription)
			if err != nil {
				if err.Error() == "subscription already exists" {
					w.WriteHeader(http.StatusConflict)
					w.Write([]byte("Already subscribed to this artist"))
					return
				}
				log.Printf("Error creating subscription: %v", err)
				http.Error(w, "failed to create subscription", http.StatusInternalServerError)
				return
			}

			log.Printf("User %s subscribed to artist %s", userID, artistID)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Subscribed to artist successfully"))
		} else if r.Method == http.MethodDelete {
			// Unsubscribe from artist
			artistID := r.URL.Query().Get("artistId")
			if artistID == "" {
				http.Error(w, "artistId parameter is required", http.StatusBadRequest)
				return
			}

			userID := r.URL.Query().Get("userId")
			if userID == "" {
				http.Error(w, "userId parameter is required", http.StatusBadRequest)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := subscriptionRepo.DeleteByUserAndArtist(ctx, userID, artistID)
			if err != nil {
				if err.Error() == "subscription not found" {
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte("Subscription not found"))
					return
				}
				log.Printf("Error deleting subscription: %v", err)
				http.Error(w, "failed to delete subscription", http.StatusInternalServerError)
				return
			}

			log.Printf("User %s unsubscribed from artist %s", userID, artistID)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Unsubscribed from artist successfully"))
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Subscribe to genre endpoint
	mux.HandleFunc("/subscribe-genre", func(w http.ResponseWriter, r *http.Request) {
		addCORSHeaders(w)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodPost {
			genre, err := url.QueryUnescape(r.URL.Query().Get("genre"))
			if err != nil {
				genre = r.URL.Query().Get("genre") // Fallback to raw value
			}
			if genre == "" {
				http.Error(w, "genre parameter is required", http.StatusBadRequest)
				return
			}

			userID := r.URL.Query().Get("userId")
			if userID == "" {
				http.Error(w, "userId parameter is required", http.StatusBadRequest)
				return
			}

			// Check if already subscribed
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			existing, err := subscriptionRepo.GetByUserAndGenre(ctx, userID, genre)
			if err != nil {
				log.Printf("Error checking subscription: %v", err)
				http.Error(w, "failed to check subscription", http.StatusInternalServerError)
				return
			}
			if existing != nil {
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte("Already subscribed to this genre"))
				return
			}

			// Create subscription
			subscription := &model.Subscription{
				UserID: userID,
				Type:   "genre",
				Genre:  genre,
			}

			err = subscriptionRepo.Create(ctx, subscription)
			if err != nil {
				if err.Error() == "subscription already exists" {
					w.WriteHeader(http.StatusConflict)
					w.Write([]byte("Already subscribed to this genre"))
					return
				}
				log.Printf("Error creating subscription: %v", err)
				http.Error(w, "failed to create subscription", http.StatusInternalServerError)
				return
			}

			log.Printf("User %s subscribed to genre %s", userID, genre)
			// Emit event to recommendation-service
			emitSubscriptionEvent(cfg.RecommendationServiceURL, userID, genre, "subscription_created")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Subscribed to genre successfully"))
		} else if r.Method == http.MethodDelete {
			// Unsubscribe from genre
			genre, err := url.QueryUnescape(r.URL.Query().Get("genre"))
			if err != nil {
				genre = r.URL.Query().Get("genre") // Fallback to raw value
			}
			if genre == "" {
				http.Error(w, "genre parameter is required", http.StatusBadRequest)
				return
			}

			userID := r.URL.Query().Get("userId")
			if userID == "" {
				http.Error(w, "userId parameter is required", http.StatusBadRequest)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = subscriptionRepo.DeleteByUserAndGenre(ctx, userID, genre)
			if err != nil {
				if err.Error() == "subscription not found" {
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte("Subscription not found"))
					return
				}
				log.Printf("Error deleting subscription: %v", err)
				http.Error(w, "failed to delete subscription", http.StatusInternalServerError)
				return
			}

			log.Printf("User %s unsubscribed from genre %s", userID, genre)
			// Emit event to recommendation-service
			emitSubscriptionEvent(cfg.RecommendationServiceURL, userID, genre, "subscription_deleted")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Unsubscribed from genre successfully"))
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Event handler endpoint - receives events from content-service
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var event map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			log.Printf("Failed to decode event: %v", err)
			http.Error(w, "invalid event payload", http.StatusBadRequest)
			return
		}

		// Debug: log the entire event
		eventJSON, _ := json.Marshal(event)
		log.Printf("Received event: %s", string(eventJSON))

		eventType, ok := event["type"].(string)
		if !ok {
			http.Error(w, "event type is required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Handle different event types
		switch eventType {
		case "new_artist":
			handleNewArtistEvent(ctx, event, subscriptionRepo, cfg)
		case "new_album":
			handleNewAlbumEvent(ctx, event, subscriptionRepo, cfg)
		case "new_song":
			handleNewSongEvent(ctx, event, subscriptionRepo, cfg)
		default:
			log.Printf("Unknown event type: %s", eventType)
			http.Error(w, "unknown event type", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Event processed"))
	})

	log.Println("Subscriptions service running on port", cfg.Port)
	
	// Support HTTPS if certificates are provided
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	if certFile != "" && keyFile != "" {
		log.Println("Starting HTTPS server on port", cfg.Port)
		log.Fatal(http.ListenAndServeTLS(":"+cfg.Port, certFile, keyFile, mux))
	} else {
		log.Println("Starting HTTP server on port", cfg.Port)
		log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
	}
}
