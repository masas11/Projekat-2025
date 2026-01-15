package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"users-service/config"
	"users-service/internal/handler"
	"users-service/internal/model"
	"users-service/internal/store"
)

func initAdminUser(ctx context.Context, userRepo *store.UserRepository) {
	// Check if admin already exists
	admin, err := userRepo.GetByUsername(ctx, "admin")
	if err == nil && admin != nil {
		log.Println("Admin user already exists")
		return
	}

	// Create admin user
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	now := time.Now()

	adminUser := &model.User{
		ID:                uuid.NewString(),
		FirstName:         "Admin",
		LastName:          "User",
		Email:             "admin@musicstreaming.com",
		Username:          "admin",
		PasswordHash:      string(hash),
		Role:              "ADMIN",
		Verified:          true,
		PasswordChangedAt: now,
		PasswordExpiresAt: now.Add(60 * 24 * time.Hour),
		CreatedAt:         now,
	}

	if err := userRepo.Create(ctx, adminUser); err != nil {
		log.Printf("Failed to create admin user: %v", err)
	} else {
		log.Println("Admin user created successfully (username: admin, password: admin123)")
	}
}

func main() {
	// učitavanje konfiguracije
	cfg := config.Load()

	// Initialize MongoDB connection
	dbStore, err := store.NewMongoDBStore(cfg.MongoDBURI, cfg.MongoDBDatabase)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer dbStore.Close()
	log.Println("Connected to MongoDB")

	// Initialize repository
	userRepo := store.NewUserRepository(dbStore.Database)

	// Initialize admin user
	ctx := context.Background()
	initAdminUser(ctx, userRepo)

	// inicijalizacija handler-a
	registerHandler := handler.NewRegisterHandler(userRepo)
	loginHandler := handler.NewLoginHandler(userRepo, cfg)
	passwordHandler := handler.NewPasswordHandler(userRepo)

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
