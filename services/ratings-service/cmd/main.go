package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"ratings-service/config"
	"ratings-service/internal/model"
	"ratings-service/internal/store"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Simple Circuit Breaker implementation
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

	log.Printf("Circuit breaker state: %s, failures: %d", cb.state, cb.failures)

	// Check if circuit should reset
	if cb.state == "open" {
		timeSinceFail := time.Since(cb.lastFailTime)
		log.Printf("Circuit breaker open, time since last fail: %v, timeout: %v", timeSinceFail, cb.resetTimeout)
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
		log.Printf("Function failed, failures: %d/%d", cb.failures, cb.maxFailures)
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

// Synchronous call with retry + fallback - DEPRECATED, use checkSpecificSongExists
func checkSongExists(client *http.Client, contentURL string) bool {
	// This function is deprecated - always return true to avoid blocking
	// Use checkSpecificSongExists for actual validation
	log.Println("WARNING: checkSongExists is deprecated, use checkSpecificSongExists")
	return true
}

// Check if specific song exists by ID
func checkSpecificSongExists(client *http.Client, contentURL, songID string) bool {
	checkURL := contentURL + "/songs/exists?id=" + url.QueryEscape(songID)

	for i := 0; i < 2; i++ { // retry 2 times
		resp, err := client.Get(checkURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			var result map[string]bool
			if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
				return result["exists"]
			}
		}
		log.Printf("Retrying call to content-service for song %s... (attempt %d)", songID, i+1)
		time.Sleep(100 * time.Millisecond) // brief delay between retries
	}

	// fallback logic
	log.Printf("Content-service unavailable for song %s, fallback activated", songID)
	return false
}

func main() {
	cfg := config.Load()

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDBName := os.Getenv("MONGODB_DATABASE")
	if mongoDBName == "" {
		mongoDBName = "ratings_db"
	}

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	// Test the connection
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	db := mongoClient.Database(mongoDBName)
	ratingStore := store.NewRatingStore(db)

	// Connect to other MongoDB instances for recommendations
	// Content service MongoDB
	contentClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongodb-content:27017"))
	if err != nil {
		log.Fatalf("Failed to connect to content MongoDB: %v", err)
	}
	defer contentClient.Disconnect(ctx)

	// Subscriptions service MongoDB
	subscriptionsClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongodb-subscriptions:27017"))
	if err != nil {
		log.Fatalf("Failed to connect to subscriptions MongoDB: %v", err)
	}
	defer subscriptionsClient.Disconnect(ctx)

	contentDB := contentClient.Database("music_streaming")
	subscriptionsDB := subscriptionsClient.Database("subscriptions_db")
	recommendationStore := store.NewRecommendationStore(db, contentDB, subscriptionsDB)

	log.Printf("Connected to MongoDB at %s, database: %s", mongoURI, mongoDBName)

	// HTTP client with timeout (MANDATORY for 3.3)
	clientHTTP := &http.Client{
		Timeout: 2 * time.Second,
	}

	// Circuit breaker for content service calls
	cb := NewCircuitBreaker(3, 5*time.Second) // Open after 3 failures, reset after 5 seconds

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ratings-service is running"))
	})

	// Dummy rate endpoint using synchronous communication
	mux.HandleFunc("/rate", func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		exists := checkSongExists(clientHTTP, cfg.ContentServiceURL)

		if !exists {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Cannot rate song, content-service unavailable"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Song rated successfully"))
	})

	// Rate specific song endpoint
	mux.HandleFunc("/rate-song", func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		songID := r.URL.Query().Get("songId")
		if songID == "" {
			http.Error(w, "songId parameter is required", http.StatusBadRequest)
			return
		}

		userID := r.URL.Query().Get("userId")
		if userID == "" {
			http.Error(w, "userId parameter is required", http.StatusBadRequest)
			return
		}

		rating := r.URL.Query().Get("rating")
		if rating == "" {
			http.Error(w, "rating parameter is required", http.StatusBadRequest)
			return
		}

		// Validate rating is between 1 and 5
		ratingValue, err := strconv.Atoi(rating)
		if err != nil || ratingValue < 1 || ratingValue > 5 {
			http.Error(w, "rating must be between 1 and 5", http.StatusBadRequest)
			return
		}

		// Use circuit breaker for synchronous call to check if song exists
		err = cb.Call(func() error {
			if !checkSpecificSongExists(clientHTTP, cfg.ContentServiceURL, songID) {
				return &CircuitBreakerError{"Song not found"}
			}
			return nil
		})

		if err != nil {
			if cbErr, ok := err.(*CircuitBreakerError); ok {
				if cbErr.Message == "Song not found" {
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte("Song not found"))
					return
				}
				// This is a real circuit breaker error
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("Service temporarily unavailable - circuit breaker open"))
				return
			}
			// Some other error
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
			return
		}

		// Save rating to database
		ratingModel := &model.Rating{
			SongID: songID,
			UserID: userID,
			Rating: ratingValue,
		}

		// Check if user already rated this song
		ratingCtx, ratingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer ratingCancel()

		existingRating, err := ratingStore.GetBySongAndUser(ratingCtx, songID, userID)
		if err != nil {
			log.Printf("Error checking existing rating: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error checking existing rating"))
			return
		}

		if existingRating != nil {
			// Update existing rating
			ratingModel.ID = existingRating.ID
			err = ratingStore.Update(ratingCtx, ratingModel)
			if err != nil {
				log.Printf("Error updating rating: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error updating rating"))
				return
			}
			log.Printf("Updated rating for song %s by user %s", songID, userID)
		} else {
			// Create new rating
			err = ratingStore.Create(ratingCtx, ratingModel)
			if err != nil {
				log.Printf("Error creating rating: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error creating rating"))
				return
			}
			log.Printf("Created rating for song %s by user %s", songID, userID)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Song rated successfully"))
	})

	// Delete rating endpoint
	mux.HandleFunc("/delete-rating", func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		songID := r.URL.Query().Get("songId")
		if songID == "" {
			http.Error(w, "songId parameter is required", http.StatusBadRequest)
			return
		}

		userID := r.URL.Query().Get("userId")
		if userID == "" {
			http.Error(w, "userId parameter is required", http.StatusBadRequest)
			return
		}

		// Delete rating from database
		ratingCtx, ratingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer ratingCancel()

		err := ratingStore.DeleteBySongAndUser(ratingCtx, songID, userID)
		if err != nil {
			if err.Error() == "rating not found" {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Rating not found"))
				return
			}
			log.Printf("Error deleting rating: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error deleting rating"))
			return
		}

		log.Printf("Deleted rating for song %s by user %s", songID, userID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Rating deleted successfully"))
	})

	// Get user's rating for a song
	mux.HandleFunc("/get-rating", func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		songID := r.URL.Query().Get("songId")
		if songID == "" {
			http.Error(w, "songId parameter is required", http.StatusBadRequest)
			return
		}

		userID := r.URL.Query().Get("userId")
		if userID == "" {
			http.Error(w, "userId parameter is required", http.StatusBadRequest)
			return
		}

		// Get rating from database
		ratingCtx, ratingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer ratingCancel()

		rating, err := ratingStore.GetBySongAndUser(ratingCtx, songID, userID)
		if err != nil {
			log.Printf("Error getting rating: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error getting rating"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if rating == nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"rating": null}`))
		} else {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]int{"rating": rating.Rating})
		}
	})

	// Recommendations endpoint
	mux.HandleFunc("/recommendations", func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight request
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

		// Get recommendations
		recCtx, recCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer recCancel()

		// Get songs from subscribed genres
		subscribedSongs, err := recommendationStore.GetSongsByUserSubscribedGenres(recCtx, userID)
		if err != nil {
			log.Printf("Error getting subscribed genre songs: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error getting recommendations"))
			return
		}

		// Get top rated song from unsubscribed genre
		topRatedSong, err := recommendationStore.GetTopRatedSongFromUnsubscribedGenre(recCtx, userID)
		if err != nil {
			log.Printf("Error getting top rated song: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error getting recommendations"))
			return
		}

		response := model.RecommendationResponse{
			SubscribedGenreSongs: subscribedSongs,
			TopRatedSong:         topRatedSong,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	log.Println("Ratings service running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
