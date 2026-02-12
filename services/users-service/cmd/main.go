package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"users-service/config"
	"users-service/internal/handler"
	"users-service/internal/logger"
	"users-service/internal/mail"
	"users-service/internal/middleware"
	"users-service/internal/model"
	"users-service/internal/store"
)

func initAdminUser(ctx context.Context, userRepo *store.UserRepository, cfg *config.Config) {
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
		PasswordExpiresAt: now.Add(time.Duration(cfg.PasswordExpirationDays) * 24 * time.Hour),
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

	// Initialize logger
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = filepath.Join(".", "logs")
	}
	appLogger, err := logger.NewLogger(logDir)
	if err != nil {
		log.Printf("Warning: Failed to initialize logger: %v, using stdout", err)
		appLogger = logger.NewStdoutLogger()
	}
	defer appLogger.Close()

	// Initialize email service
	mail.Init(cfg)

	// Initialize admin user
	ctx := context.Background()
	initAdminUser(ctx, userRepo, cfg)

	// inicijalizacija handler-a
	registerHandler := handler.NewRegisterHandler(userRepo, cfg, appLogger)
	loginHandler := handler.NewLoginHandler(userRepo, cfg, appLogger)
	passwordHandler := handler.NewPasswordHandler(userRepo, cfg)
	magicLinkHandler := handler.NewMagicLinkHandler(userRepo, cfg)
	verificationHandler := handler.NewVerificationHandler(userRepo)

	// router
	mux := http.NewServeMux()

	// health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("users-service is running"))
	})

	// Rate limiting: 10 requests per minute for sensitive endpoints
	rateLimit := middleware.RateLimit(10, 1*time.Minute)

	// register endpoint (rate limited)
	mux.HandleFunc("/register", rateLimit(registerHandler.Register))

	// login / OTP endpoints (rate limited)
	mux.HandleFunc("/login/request-otp", rateLimit(loginHandler.RequestOTP))
	mux.HandleFunc("/login/verify-otp", rateLimit(loginHandler.VerifyOTP))
	mux.HandleFunc("/logout", rateLimit(loginHandler.Logout))

	// password endpoints (rate limited)
	mux.HandleFunc("/password/change", rateLimit(passwordHandler.ChangePassword))
	mux.HandleFunc("/password/reset/request", rateLimit(passwordHandler.RequestPasswordReset))
	mux.HandleFunc("/password/reset", rateLimit(passwordHandler.ResetPassword))

	// email verification endpoint (rate limited)
	mux.HandleFunc("/verify-email", rateLimit(verificationHandler.VerifyEmail))

	// magic link endpoints (account recovery) (rate limited)
	mux.HandleFunc("/recover/request", rateLimit(magicLinkHandler.RequestMagicLink))
	mux.HandleFunc("/recover/verify", rateLimit(magicLinkHandler.VerifyMagicLink))

	log.Println("Users service running on port", cfg.Port)
	
	// Support HTTPS if certificates are provided
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	if certFile != "" && keyFile != "" {
		log.Println("Starting HTTPS server on port", cfg.Port)
		server := &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: mux,
		}
		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
			if appLogger != nil {
				appLogger.LogTLSFailure("users-service", err.Error(), "")
			}
			log.Fatal("HTTPS server failed:", err)
		}
	} else {
		log.Println("Starting HTTP server on port", cfg.Port)
		log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
	}
}
