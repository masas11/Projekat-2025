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
	"strconv"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"ratings-service/config"
	"ratings-service/internal/model"
	"ratings-service/internal/store"
	"shared/tracing"
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

// RetryConfig holds configuration for retry mechanism (2.7.5)
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffMultiplier float64
}

// DefaultRetryConfig returns default retry configuration (2.7.5)
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:       3,
		InitialDelay:    100 * time.Millisecond,
		MaxDelay:         2 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// RetryWithExponentialBackoff executes a function with retry and exponential backoff (2.7.5)
func RetryWithExponentialBackoff(ctx context.Context, config RetryConfig, fn func() error) error {
	delay := config.InitialDelay
	
	for attempt := 0; attempt < config.MaxRetries; attempt++ {
		// Check if context is cancelled (2.7.7)
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled during retry attempt %d: %v", attempt+1, ctx.Err())
			return ctx.Err()
		default:
		}
		
		err := fn()
		if err == nil {
			if attempt > 0 {
				log.Printf("Retry succeeded on attempt %d", attempt+1)
			}
			return nil
		}
		
		// Don't retry on last attempt
		if attempt < config.MaxRetries-1 {
			log.Printf("Retry attempt %d/%d failed: %v, retrying in %v...", attempt+1, config.MaxRetries, err, delay)
			select {
			case <-ctx.Done():
				log.Printf("Context cancelled during retry delay: %v", ctx.Err())
				return ctx.Err()
			case <-time.After(delay):
				// Exponential backoff
				delay = time.Duration(float64(delay) * config.BackoffMultiplier)
				if delay > config.MaxDelay {
					delay = config.MaxDelay
				}
			}
		}
	}
	
	return fmt.Errorf("all %d retry attempts failed", config.MaxRetries)
}

// Check if specific song exists by ID with retry and fallback (2.7.2, 2.7.3, 2.7.5)
func checkSpecificSongExists(client *http.Client, contentURL, songID string) bool {
	checkURL := contentURL + "/songs/exists?id=" + url.QueryEscape(songID)
	retryConfig := DefaultRetryConfig()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	var lastErr error
	
	err := RetryWithExponentialBackoff(ctx, retryConfig, func() error {
		// Check if context is cancelled (2.7.7)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		reqCtx, reqCancel := context.WithTimeout(ctx, 2*time.Second)
		defer reqCancel()
		
		req, err := http.NewRequestWithContext(reqCtx, "GET", checkURL, nil)
		if err != nil {
			lastErr = err
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var result map[string]bool
			if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
				exists = result["exists"]
				return nil // Success
			}
			lastErr = fmt.Errorf("failed to decode response: %v", err)
			return lastErr
		}
		
		lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		return lastErr
	})

	if err != nil {
		// Fallback logic (2.7.3): return false when service is unavailable
		log.Printf("Content-service unavailable for song %s after retries, fallback activated - assuming song does not exist. Last error: %v", songID, lastErr)
		return false
	}

	return exists
}

func main() {
	cfg := config.Load()

	// Initialize tracing (2.10)
	cleanup, err := tracing.InitTracing("ratings-service")
	if err != nil {
		log.Printf("Warning: Failed to initialize tracing: %v", err)
	} else {
		defer cleanup()
		log.Println("Tracing initialized for ratings-service")
	}

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

	// HTTP client configuration with timeout (2.7.1, 2.7.2)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	clientHTTP := &http.Client{
		Timeout:   2 * time.Second, // Request timeout (2.7.2)
		Transport: tr,
	}

	// Circuit breaker for content service calls
	cb := NewCircuitBreaker(3, 5*time.Second) // Open after 3 failures, reset after 5 seconds

	mux := http.NewServeMux()

	// emitRatingEvent sends rating event to recommendation-service asynchronously
	emitRatingEvent := func(recommendationServiceURL, userID, songID string, rating int, eventType string) {
		go func() {
			event := map[string]interface{}{
				"type":    eventType,
				"userId":  userID,
				"songId":  songID,
				"rating":  rating,
			}

			eventJSON, err := json.Marshal(event)
			if err != nil {
				log.Printf("Failed to marshal rating event: %v", err)
				return
			}

			url := recommendationServiceURL + "/events"
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(eventJSON))
			if err != nil {
				log.Printf("Failed to create rating event request: %v", err)
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

			log.Printf("Sending rating event to %s: type=%s, userId=%s, songId=%s, rating=%d", url, eventType, userID, songID, rating)
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Failed to emit rating event to recommendation-service: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
				log.Printf("Recommendation-service returned non-OK status for rating event: %d", resp.StatusCode)
				return
			}

			log.Printf("Rating event emitted successfully: %s", eventType)
		}()
	}

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

		// Create context with timeout from request context (2.7.6, 2.7.7)
		// Use request context so it can be cancelled by API Gateway timeout
		ratingCtx, ratingCancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer ratingCancel()

		// Check if context is already cancelled (2.7.7)
		select {
		case <-ratingCtx.Done():
			log.Printf("Request context cancelled before processing: %v", ratingCtx.Err())
			w.WriteHeader(http.StatusRequestTimeout)
			w.Write([]byte("Request timeout - processing abandoned"))
			return
		default:
		}

		// Use circuit breaker for synchronous call to check if song exists
		err = cb.Call(func() error {
			if !checkSpecificSongExists(clientHTTP, cfg.ContentServiceURL, songID) {
				return &CircuitBreakerError{"Song not found"}
			}
			return nil
		})

		if err != nil {
			// Check if context was cancelled during circuit breaker call (2.7.7)
			select {
			case <-ratingCtx.Done():
				log.Printf("Request context cancelled during circuit breaker call: %v", ratingCtx.Err())
				w.WriteHeader(http.StatusRequestTimeout)
				w.Write([]byte("Request timeout - processing abandoned"))
				return
			default:
			}

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

		// Check if context is cancelled before database operations (2.7.7)
		select {
		case <-ratingCtx.Done():
			log.Printf("Request context cancelled before database operations: %v", ratingCtx.Err())
			w.WriteHeader(http.StatusRequestTimeout)
			w.Write([]byte("Request timeout - processing abandoned"))
			return
		default:
		}

		// Save rating to database
		ratingModel := &model.Rating{
			SongID: songID,
			UserID: userID,
			Rating: ratingValue,
		}

		// Check if user already rated this song

		existingRating, err := ratingStore.GetBySongAndUser(ratingCtx, songID, userID)
		if err != nil {
			// Check if context was cancelled (2.7.7)
			select {
			case <-ratingCtx.Done():
				log.Printf("Request context cancelled during database read: %v", ratingCtx.Err())
				w.WriteHeader(http.StatusRequestTimeout)
				w.Write([]byte("Request timeout - processing abandoned"))
				return
			default:
			}
			log.Printf("Error checking existing rating: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error checking existing rating"))
			return
		}

		// Check if context is cancelled before update/create (2.7.7)
		select {
		case <-ratingCtx.Done():
			log.Printf("Request context cancelled before rating save: %v", ratingCtx.Err())
			w.WriteHeader(http.StatusRequestTimeout)
			w.Write([]byte("Request timeout - processing abandoned"))
			return
		default:
		}

		if existingRating != nil {
			// Update existing rating
			ratingModel.ID = existingRating.ID
			err = ratingStore.Update(ratingCtx, ratingModel)
			if err != nil {
				// Check if context was cancelled (2.7.7)
				select {
				case <-ratingCtx.Done():
					log.Printf("Request context cancelled during rating update: %v", ratingCtx.Err())
					w.WriteHeader(http.StatusRequestTimeout)
					w.Write([]byte("Request timeout - processing abandoned"))
					return
				default:
				}
				log.Printf("Error updating rating: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error updating rating"))
				return
			}
			log.Printf("Updated rating for song %s by user %s", songID, userID)
			// Emit event to recommendation-service
			emitRatingEvent(cfg.RecommendationServiceURL, userID, songID, ratingValue, "rating_updated")
		} else {
			// Create new rating
			err = ratingStore.Create(ratingCtx, ratingModel)
			if err != nil {
				// Check if context was cancelled (2.7.7)
				select {
				case <-ratingCtx.Done():
					log.Printf("Request context cancelled during rating create: %v", ratingCtx.Err())
					w.WriteHeader(http.StatusRequestTimeout)
					w.Write([]byte("Request timeout - processing abandoned"))
					return
				default:
				}
				log.Printf("Error creating rating: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error creating rating"))
				return
			}
			log.Printf("Created rating for song %s by user %s", songID, userID)
			// Emit event to recommendation-service
			emitRatingEvent(cfg.RecommendationServiceURL, userID, songID, ratingValue, "rating_created")
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
		// Emit event to recommendation-service for rating deletion
		emitRatingEvent(cfg.RecommendationServiceURL, userID, songID, 0, "rating_deleted")
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
	// Get average rating and count for a song
	mux.HandleFunc("/average-rating", func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
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

		songID := r.URL.Query().Get("songId")
		if songID == "" {
			http.Error(w, "songId parameter is required", http.StatusBadRequest)
			return
		}

		// Get average rating and count from database
		ratingCtx, ratingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer ratingCancel()

		avgRating, count, err := ratingStore.GetAverageRating(ratingCtx, songID)
		if err != nil {
			log.Printf("Error getting average rating: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error getting average rating"))
			return
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"songId":      songID,
			"averageRating": avgRating,
			"ratingCount":   count,
		})
	})

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
	
	// Support HTTPS if certificates are provided
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	// Wrap mux with tracing middleware
	handler := tracing.HTTPMiddleware(mux)
	if certFile != "" && keyFile != "" {
		log.Println("Starting HTTPS server on port", cfg.Port)
		log.Fatal(http.ListenAndServeTLS(":"+cfg.Port, certFile, keyFile, handler))
	} else {
		log.Println("Starting HTTP server on port", cfg.Port)
		log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
	}
}
