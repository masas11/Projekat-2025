package main

import (
	"log"
	"net/http"

	"users-service/config"
	"users-service/internal/handler"
	"users-service/internal/store"
)

func main() {
	cfg := config.Load()

	// initialize in-memory store
	userStore := store.NewUserStore()

	// initialize handlers
	registerHandler := handler.NewRegisterHandler(userStore)

	// router
	mux := http.NewServeMux()

	// health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("users-service is running"))
	})

	// register endpoint
	mux.HandleFunc("/register", registerHandler.Register)

	log.Println("Users service running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
