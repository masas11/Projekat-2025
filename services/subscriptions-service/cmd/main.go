package main

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"subscriptions-service/config"
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

func main() {
	cfg := config.Load()

	// HTTP client with timeout
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("subscriptions-service is running"))
	})

	// Subscribe to artist endpoint with synchronous validation
	mux.HandleFunc("/subscribe-artist", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

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

		// Synchronous call to check if artist exists
		exists := checkArtistExists(client, cfg.ContentServiceURL, artistID)
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Artist not found"))
			return
		}

		// Here you would normally save the subscription to database
		log.Printf("User %s subscribed to artist %s", userID, artistID)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Subscribed to artist successfully"))
	})

	log.Println("Subscriptions service running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
