package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"subscriptions-service/config"
	"subscriptions-service/internal/model"
	"subscriptions-service/internal/store"
)

// Check if artist exists by ID
func checkArtistExists(client *http.Client, contentURL, artistID string) bool {
	checkURL := contentURL + "/artists/" + url.QueryEscape(artistID)

	for i := 0; i < 2; i++ { // retry 2 times
		resp, err := client.Get(checkURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			return true
		}
		log.Printf("Retrying call to content-service for artist %s... (attempt %d)", artistID, i+1)
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Content-service unavailable for artist %s, fallback activated", artistID)
	return false
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

	req, err := http.NewRequest("POST", notificationsServiceURL+"/notifications", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create notification request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to create notification: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		log.Printf("Notifications-service returned non-OK status: %d", resp.StatusCode)
		return
	}

	log.Printf("Notification created for user %s: %s", userID, message)
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
	artistIDs, _ := event["artistIds"].([]interface{})

	if albumID == "" {
		log.Printf("Invalid new_album event: missing albumId")
		return
	}

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
			message := fmt.Sprintf("New album '%s' by artist has been released", albumName)
			createNotification(cfg.NotificationsServiceURL, sub.UserID, "new_album", message, albumID)
		}
	}

	log.Printf("Processed new_album event for album %s", albumID)
}

// handleNewSongEvent processes new song events
func handleNewSongEvent(ctx context.Context, event map[string]interface{}, repo *store.SubscriptionRepository, cfg *config.Config) {
	songID, _ := event["songId"].(string)
	songName, _ := event["name"].(string)
	artistIDs, _ := event["artistIds"].([]interface{})

	if songID == "" {
		log.Printf("Invalid new_song event: missing songId")
		return
	}

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
			message := fmt.Sprintf("New song '%s' by artist has been added", songName)
			createNotification(cfg.NotificationsServiceURL, sub.UserID, "new_song", message, songID)
		}
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

	// HTTP client with timeout
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	mux := http.NewServeMux()

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

			// Synchronous call to check if artist exists
			exists := checkArtistExists(client, cfg.ContentServiceURL, artistID)
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
			genre := r.URL.Query().Get("genre")
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
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Subscribed to genre successfully"))
		} else if r.Method == http.MethodDelete {
			// Unsubscribe from genre
			genre := r.URL.Query().Get("genre")
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

			err := subscriptionRepo.DeleteByUserAndGenre(ctx, userID, genre)
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
