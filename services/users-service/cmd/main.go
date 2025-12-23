package main

import (
	"log"
	"net/http"

	"users-service/config"
	"users-service/internal/handler"
	"users-service/internal/store"
)

func main() {
	// učitavanje konfiguracije
	cfg := config.Load()

	// inicijalizacija in-memory store-a
	userStore := store.NewUserStore()

	// inicijalizacija handler-a
	registerHandler := handler.NewRegisterHandler(userStore)
	loginHandler := handler.NewLoginHandler(userStore)
	passwordHandler := handler.NewPasswordHandler(userStore)

	// router
	mux := http.NewServeMux()

	// health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("users-service is running"))
	})

	// register endpoint
	mux.HandleFunc("/register", registerHandler.Register)

	// login / OTP endpoints
	mux.HandleFunc("/login/request-otp", loginHandler.RequestOTP)
	mux.HandleFunc("/login/verify-otp", loginHandler.VerifyOTP)

	// password endpoints
	mux.HandleFunc("/password/change", passwordHandler.ChangePassword)
	mux.HandleFunc("/password/reset", passwordHandler.ResetPassword)

	log.Println("Users service running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
