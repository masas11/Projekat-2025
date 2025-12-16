package main

import (
	"log"
	"net/http"

	"subscriptions-service/config"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("subscriptions-service is running"))
	})

	log.Println("Subscriptions service running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
