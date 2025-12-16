package main

import (
	"log"
	"net/http"

	"notifications-service/config"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("notifications-service is running"))
	})

	log.Println("Notifications service running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
