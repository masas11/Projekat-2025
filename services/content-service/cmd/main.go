package main

import (
	"log"
	"net/http"
	"content-service/config"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("content-service is running"))
	})

	// Dummy endpoint for 3.3 (song existence check)
	mux.HandleFunc("/songs/exists", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	})
	
	log.Println("Content service running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
