package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"recommendation-service/config"
	"recommendation-service/internal/model"
	"recommendation-service/internal/store"
	"shared/tracing"
)

func main() {
	cfg := config.Load()

	// Initialize tracing (2.10)
	cleanup, err := tracing.InitTracing("recommendation-service")
	if err != nil {
		log.Printf("Warning: Failed to initialize tracing: %v", err)
	} else {
		defer cleanup()
		log.Println("Tracing initialized for recommendation-service")
	}

	// Initialize Neo4j store
	neo4jStore, err := store.NewNeo4jStore(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Neo4j store: %v", err)
	}
	defer neo4jStore.Close()

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("recommendation-service is running"))
	})

	// Sync endpoint - populates graph from existing MongoDB data
	mux.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		log.Println("Starting sync from MongoDB to Neo4j...")
		if err := neo4jStore.SyncFromMongoDB(ctx, cfg.ContentServiceURL, cfg.RatingsServiceURL, cfg.SubscriptionsServiceURL); err != nil {
			log.Printf("Sync failed: %v", err)
			http.Error(w, fmt.Sprintf("sync failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Sync completed successfully"))
	})

	// Sync subscriptions for a specific user endpoint
	mux.HandleFunc("/sync-user-subscriptions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := r.URL.Query().Get("userId")
		if userID == "" {
			http.Error(w, "userId parameter is required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Ensure user exists
		if err := neo4jStore.AddOrUpdateUser(ctx, userID); err != nil {
			log.Printf("Error adding user: %v", err)
		}

		// Get user subscriptions from subscriptions service
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(fmt.Sprintf("%s/subscriptions?userId=%s", cfg.SubscriptionsServiceURL, userID))
		if err != nil {
			log.Printf("Error fetching subscriptions for user %s: %v", userID, err)
			http.Error(w, fmt.Sprintf("failed to fetch subscriptions: %v", err), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var subscriptions []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&subscriptions); err != nil {
			log.Printf("Error decoding subscriptions: %v", err)
			http.Error(w, "failed to decode subscriptions", http.StatusInternalServerError)
			return
		}

		syncedCount := 0
		for _, sub := range subscriptions {
			subType, _ := sub["type"].(string)
			if subType == "genre" {
				genre, _ := sub["genre"].(string)
				if genre != "" {
					if err := neo4jStore.AddSubscription(ctx, userID, genre); err != nil {
						log.Printf("Error syncing subscription for user %s to genre %s: %v", userID, genre, err)
					} else {
						syncedCount++
						log.Printf("Synced subscription: user %s -> genre %s", userID, genre)
					}
				}
			}
		}

		log.Printf("Synced %d genre subscriptions for user %s", syncedCount, userID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": fmt.Sprintf("Synced %d subscriptions for user %s", syncedCount, userID),
			"count":   syncedCount,
		})
	})

	// Get recommendations endpoint
	mux.HandleFunc("/recommendations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := r.URL.Query().Get("userId")
		if userID == "" {
			http.Error(w, "userId parameter is required", http.StatusBadRequest)
			return
		}
		
		// Clean userId - remove any query parameters that might have been appended
		if idx := strings.Index(userID, "?"); idx != -1 {
			userID = userID[:idx]
		}
		if idx := strings.Index(userID, "&"); idx != -1 {
			userID = userID[:idx]
		}
		
		log.Printf("Cleaned userId: %s (original had query params: %v)", userID, strings.Contains(r.URL.Query().Get("userId"), "?"))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Ensure user exists in graph
		if err := neo4jStore.AddOrUpdateUser(ctx, userID); err != nil {
			log.Printf("Warning: Failed to ensure user exists in graph: %v", err)
		}

		// Auto-sync subscriptions and ratings for this user if not already synced
		// This ensures existing data in MongoDB is available in Neo4j
		log.Printf("Auto-syncing subscriptions and ratings for user: %s", userID)
		
		// Sync subscriptions
		client := &http.Client{Timeout: 10 * time.Second}
		subsResp, err := client.Get(fmt.Sprintf("%s/subscriptions?userId=%s", cfg.SubscriptionsServiceURL, userID))
		if err == nil {
			defer subsResp.Body.Close()
			if subsResp.StatusCode == http.StatusOK {
				var subscriptions []map[string]interface{}
				if err := json.NewDecoder(subsResp.Body).Decode(&subscriptions); err == nil {
					for _, sub := range subscriptions {
						subType, _ := sub["type"].(string)
						if subType == "genre" {
							genre, _ := sub["genre"].(string)
							if genre != "" {
								if err := neo4jStore.AddSubscription(ctx, userID, genre); err != nil {
									log.Printf("Error syncing subscription for user %s to genre %s: %v", userID, genre, err)
								} else {
									log.Printf("Auto-synced subscription: user %s -> genre %s", userID, genre)
								}
							}
						}
					}
				}
			}
		}

		// Sync ratings - get all ratings for this user from ratings service
		ratingsResp, err := client.Get(fmt.Sprintf("%s/ratings-by-user?userId=%s", cfg.RatingsServiceURL, userID))
		if err == nil {
			defer ratingsResp.Body.Close()
			if ratingsResp.StatusCode == http.StatusOK {
				var ratings []map[string]interface{}
				if err := json.NewDecoder(ratingsResp.Body).Decode(&ratings); err == nil {
					for _, rating := range ratings {
						songID, _ := rating["songId"].(string)
						ratingValue, _ := rating["rating"].(float64)
						if songID != "" && ratingValue > 0 {
							if err := neo4jStore.AddRating(ctx, userID, songID, int(ratingValue)); err != nil {
								log.Printf("Error syncing rating for user %s to song %s: %v", userID, songID, err)
							} else {
								log.Printf("Auto-synced rating: user %s -> song %s (rating: %d)", userID, songID, int(ratingValue))
							}
						}
					}
				}
			}
		}

		// Get songs from subscribed genres
		log.Printf("Getting recommendations for user: %s", userID)
		subscribedSongs, err := neo4jStore.GetSubscribedGenreSongs(ctx, userID)
		if err != nil {
			log.Printf("Error getting subscribed genre songs: %v", err)
			http.Error(w, "failed to get recommendations", http.StatusInternalServerError)
			return
		}
		log.Printf("Found %d subscribed genre songs for user %s", len(subscribedSongs), userID)
		if len(subscribedSongs) > 0 {
			log.Printf("First subscribed song for user %s: %s (genre: %s)", userID, subscribedSongs[0].Name, subscribedSongs[0].Genre)
		} else {
			log.Printf("WARNING: No subscribed genre songs found for user %s - user may not have subscriptions in Neo4j", userID)
		}

		// Get top rated song from unsubscribed genre
		topRatedSong, err := neo4jStore.GetTopRatedSongFromUnsubscribedGenre(ctx, userID)
		if err != nil {
			log.Printf("Error getting top rated song: %v", err)
			// Continue even if this fails
		} else if topRatedSong != nil {
			log.Printf("Found top rated song: %s", topRatedSong.Name)
		} else {
			log.Printf("No top rated song found for user %s", userID)
		}

		// Fetch additional song details from content service
		// For now, we'll return basic info and let the frontend fetch details
		response := &model.RecommendationResponse{
			SubscribedGenreSongs: subscribedSongs,
			TopRatedSong:         topRatedSong,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Event handler for updating graph
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

		log.Printf("Received event: type=%s", eventType)

		// Handle different event types asynchronously
		go func() {
			// Use longer timeout for event processing
			eventCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			switch eventType {
			case "rating_created", "rating_updated":
				handleRatingEvent(eventCtx, event, neo4jStore)
			case "rating_deleted":
				handleRatingDeleted(eventCtx, event, neo4jStore)
			case "subscription_created":
				handleSubscriptionCreated(eventCtx, event, neo4jStore)
			case "subscription_deleted":
				handleSubscriptionDeleted(eventCtx, event, neo4jStore)
			case "song_created":
				handleSongCreated(eventCtx, event, neo4jStore)
			case "song_deleted":
				handleSongDeleted(eventCtx, event, neo4jStore)
			case "artist_created":
				handleArtistCreated(eventCtx, event, neo4jStore)
			case "artist_deleted":
				handleArtistDeleted(eventCtx, event, neo4jStore)
			case "album_deleted":
				handleAlbumDeleted(eventCtx, event, neo4jStore)
			default:
				log.Printf("Unknown event type: %s", eventType)
			}
		}()

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Event accepted"))
	})

	log.Println("Recommendation service running on port", cfg.Port)

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

func handleRatingEvent(ctx context.Context, event map[string]interface{}, store *store.Neo4jStore) {
	userID, _ := event["userId"].(string)
	songID, _ := event["songId"].(string)
	rating, _ := event["rating"].(float64)

	if userID == "" || songID == "" {
		log.Printf("Invalid rating event: missing userId or songId")
		return
	}

	// Ensure user exists
	if err := store.AddOrUpdateUser(ctx, userID); err != nil {
		log.Printf("Error adding user: %v", err)
	}

	// Add rating relationship
	if err := store.AddRating(ctx, userID, songID, int(rating)); err != nil {
		log.Printf("Error adding rating: %v", err)
		return
	}

	log.Printf("Rating event processed: user %s rated song %s with %d", userID, songID, int(rating))
}

func handleSubscriptionCreated(ctx context.Context, event map[string]interface{}, store *store.Neo4jStore) {
	userID, _ := event["userId"].(string)
	genre, _ := event["genre"].(string)

	if userID == "" || genre == "" {
		log.Printf("Invalid subscription event: missing userId or genre")
		return
	}

	log.Printf("Processing subscription_created event: userId=%s, genre=%s", userID, genre)

	// Ensure user exists
	if err := store.AddOrUpdateUser(ctx, userID); err != nil {
		log.Printf("Error adding user: %v", err)
	}

	// Add subscription relationship
	if err := store.AddSubscription(ctx, userID, genre); err != nil {
		log.Printf("Error adding subscription: %v", err)
		return
	}

	log.Printf("Subscription created successfully: user %s subscribed to genre %s", userID, genre)
	
	// Verify subscription was created by checking if songs exist for this genre
	songs, err := store.GetSubscribedGenreSongs(ctx, userID)
	if err != nil {
		log.Printf("Error verifying subscription: %v", err)
	} else {
		log.Printf("After subscription: found %d songs for user %s in genre %s", len(songs), userID, genre)
		if len(songs) == 0 {
			log.Printf("WARNING: No songs found for genre %s - genre may not have songs in Neo4j or genre name mismatch", genre)
		}
	}
}

func handleSubscriptionDeleted(ctx context.Context, event map[string]interface{}, store *store.Neo4jStore) {
	userID, _ := event["userId"].(string)
	genre, _ := event["genre"].(string)

	if userID == "" || genre == "" {
		log.Printf("Invalid subscription deletion event: missing userId or genre")
		return
	}

	if err := store.RemoveSubscription(ctx, userID, genre); err != nil {
		log.Printf("Error removing subscription: %v", err)
		return
	}

	log.Printf("Subscription deleted: user %s unsubscribed from genre %s", userID, genre)
}

func handleSongCreated(ctx context.Context, event map[string]interface{}, store *store.Neo4jStore) {
	songID, _ := event["songId"].(string)
	songName, _ := event["name"].(string)
	genre, _ := event["genre"].(string)
	artistIDsInterface, _ := event["artistIds"].([]interface{})
	albumID, _ := event["albumId"].(string)
	durationInterface, _ := event["duration"]

	if songID == "" || songName == "" || genre == "" {
		log.Printf("Invalid song event: missing required fields")
		return
	}

	artistIDs := make([]string, 0, len(artistIDsInterface))
	for _, id := range artistIDsInterface {
		if idStr, ok := id.(string); ok {
			artistIDs = append(artistIDs, idStr)
		}
	}

	duration := 0
	if durationInterface != nil {
		if d, ok := durationInterface.(float64); ok {
			duration = int(d)
		} else if d, ok := durationInterface.(int); ok {
			duration = d
		}
	}

	if err := store.AddOrUpdateSong(ctx, songID, songName, genre, artistIDs, albumID, duration); err != nil {
		log.Printf("Error adding song: %v", err)
		return
	}

	log.Printf("Song created: %s in genre %s", songName, genre)
}

func handleArtistCreated(ctx context.Context, event map[string]interface{}, store *store.Neo4jStore) {
	artistID, _ := event["artistId"].(string)
	artistName, _ := event["name"].(string)
	genresInterface, _ := event["genres"].([]interface{})

	if artistID == "" || artistName == "" {
		log.Printf("Invalid artist event: missing required fields")
		return
	}

	genres := make([]string, 0, len(genresInterface))
	for _, g := range genresInterface {
		if genreStr, ok := g.(string); ok {
			genres = append(genres, genreStr)
		}
	}

	if err := store.AddOrUpdateArtist(ctx, artistID, artistName, genres); err != nil {
		log.Printf("Error adding artist: %v", err)
		return
	}

	log.Printf("Artist created: %s", artistName)
}

func handleSongDeleted(ctx context.Context, event map[string]interface{}, store *store.Neo4jStore) {
	songID, _ := event["songId"].(string)

	if songID == "" {
		log.Printf("Invalid song_deleted event: missing songId")
		return
	}

	if err := store.DeleteSong(ctx, songID); err != nil {
		log.Printf("Error deleting song: %v", err)
		return
	}

	log.Printf("Song deleted: %s", songID)
}

func handleArtistDeleted(ctx context.Context, event map[string]interface{}, store *store.Neo4jStore) {
	artistID, _ := event["artistId"].(string)

	if artistID == "" {
		log.Printf("Invalid artist_deleted event: missing artistId")
		return
	}

	if err := store.DeleteArtist(ctx, artistID); err != nil {
		log.Printf("Error deleting artist: %v", err)
		return
	}

	log.Printf("Artist deleted: %s", artistID)
}

func handleAlbumDeleted(ctx context.Context, event map[string]interface{}, store *store.Neo4jStore) {
	albumID, _ := event["albumId"].(string)

	if albumID == "" {
		log.Printf("Invalid album_deleted event: missing albumId")
		return
	}

	if err := store.DeleteAlbum(ctx, albumID); err != nil {
		log.Printf("Error deleting album: %v", err)
		return
	}

	log.Printf("Album deleted: %s", albumID)
}

func handleRatingDeleted(ctx context.Context, event map[string]interface{}, store *store.Neo4jStore) {
	userID, _ := event["userId"].(string)
	songID, _ := event["songId"].(string)

	if userID == "" || songID == "" {
		log.Printf("Invalid rating_deleted event: missing userId or songId")
		return
	}

	if err := store.DeleteRating(ctx, userID, songID); err != nil {
		log.Printf("Error deleting rating: %v", err)
		return
	}

	log.Printf("Rating deleted: user %s, song %s", userID, songID)
}
