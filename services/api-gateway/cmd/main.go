package main

import (
	"io"
	"log"
	"net/http"

	"api-gateway/config"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()

	// USERS → HEALTH
	mux.HandleFunc("/api/users/health", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(cfg.UsersServiceURL + "/health")
		if err != nil {
			http.Error(w, "Users service unavailable", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		w.Write(body)
	})

	// CONTENT → HEALTH
	mux.HandleFunc("/api/content/health", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(cfg.ContentServiceURL + "/health")
		if err != nil {
			http.Error(w, "Content service unavailable", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		w.Write(body)
	})

	log.Println("API Gateway running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
