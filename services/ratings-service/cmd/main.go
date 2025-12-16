package main

import (
	"log"
	"net/http"
	"time"

	"ratings-service/config"
)

// Synchronous call with retry + fallback
func checkSongExists(client *http.Client, contentURL string) bool {
	url := contentURL + "/songs/exists"

	for i := 0; i < 2; i++ { // retry 2 times
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			return true
		}
		log.Println("Retrying call to content-service...")
	}

	// fallback logic
	log.Println("Content-service unavailable, fallback activated")
	return false
}

func main() {
	cfg := config.Load()

	// HTTP client with timeout (MANDATORY for 3.3)
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ratings-service is running"))
	})

	// Dummy rate endpoint using synchronous communication
	mux.HandleFunc("/rate", func(w http.ResponseWriter, r *http.Request) {
		exists := checkSongExists(client, cfg.ContentServiceURL)

		if !exists {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Cannot rate song, content-service unavailable"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Song rated successfully"))
	})

	log.Println("Ratings service running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
