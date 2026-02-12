package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	"notifications-service/config"
	"notifications-service/internal/handler"
	"notifications-service/internal/model"
	"notifications-service/internal/store"
)

func initSampleNotifications(ctx context.Context, repo *store.NotificationRepository) {
	// Real user IDs from users_db.users collection
	// Admin: '17b8a354-d7ff-402e-9059-1723f72f1098'
	// Ivana Markovic: 'f613665d-83bf-4c6c-bd3b-f712f9b04e84'
	// Ljubica: '55def55d-fed3-466a-9d6a-ed2b15100411'
	
	// Create sample notifications for real users
	sampleNotifications := []*model.Notification{
		// Notifications for Ivana Markovic (ivana_m)
		{
			ID:        uuid.NewString(),
			UserID:    "f613665d-83bf-4c6c-bd3b-f712f9b04e84",
			Type:      "new_album",
			Message:   "New album 'Thriller' by Michael Jackson has been released",
			ContentID: "album1",
			Read:      false,
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.NewString(),
			UserID:    "f613665d-83bf-4c6c-bd3b-f712f9b04e84",
			Type:      "new_song",
			Message:   "New song 'Billie Jean' by Michael Jackson has been added",
			ContentID: "song1",
			Read:      false,
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        uuid.NewString(),
			UserID:    "f613665d-83bf-4c6c-bd3b-f712f9b04e84",
			Type:      "new_artist",
			Message:   "New artist 'The Weeknd' in genre Pop has been added",
			ContentID: "artist1",
			Read:      true,
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		// Notifications for Ljubica
		{
			ID:        uuid.NewString(),
			UserID:    "55def55d-fed3-466a-9d6a-ed2b15100411",
			Type:      "new_album",
			Message:   "New album 'Abbey Road' by The Beatles has been released",
			ContentID: "album2",
			Read:      false,
			CreatedAt: time.Now(),
		},
	}

	for _, notif := range sampleNotifications {
		if err := repo.Create(ctx, notif); err != nil {
			log.Printf("Failed to create sample notification: %v", err)
		}
	}
	log.Println("Sample notifications initialized with real user IDs")
}

func main() {
	cfg := config.Load()

	// Retry mechanism to wait for Cassandra to be ready
	maxRetries := 30
	retryDelay := 2 * time.Second
	var err error

	log.Println("Waiting for Cassandra to be ready...")
	for i := 0; i < maxRetries; i++ {
		// Initialize keyspace and tables first
		if err = store.InitKeyspace(cfg.CassandraHosts, cfg.CassandraKeyspace); err == nil {
			log.Println("Cassandra keyspace and tables initialized successfully")
			break
		}
		log.Printf("Attempt %d/%d: Failed to initialize Cassandra keyspace: %v. Retrying in %v...", i+1, maxRetries, err, retryDelay)
		time.Sleep(retryDelay)
	}

	if err != nil {
		log.Fatal("Failed to initialize Cassandra keyspace after retries:", err)
	}

	// Initialize Cassandra connection with retry
	var dbStore *store.CassandraStore
	for i := 0; i < maxRetries; i++ {
		dbStore, err = store.NewCassandraStore(cfg.CassandraHosts, cfg.CassandraKeyspace)
		if err == nil {
			break
		}
		log.Printf("Attempt %d/%d: Failed to connect to Cassandra: %v. Retrying in %v...", i+1, maxRetries, err, retryDelay)
		time.Sleep(retryDelay)
	}

	if err != nil {
		log.Fatal("Failed to connect to Cassandra after retries:", err)
	}
	defer dbStore.Close()
	log.Println("Connected to Cassandra")

	// Initialize repository
	notificationRepo := store.NewNotificationRepository(dbStore.Session)

	// Initialize sample notifications
	ctx := context.Background()
	initSampleNotifications(ctx, notificationRepo)

	// Initialize handler
	notificationHandler := handler.NewNotificationHandler(notificationRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("notifications-service is running"))
	})

	// GET /notifications?userId={id} - get notifications for user
	// POST /notifications - create notification
	mux.HandleFunc("/notifications", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			notificationHandler.GetNotifications(w, r)
		case http.MethodPost:
			notificationHandler.CreateNotification(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Notifications service running on port", cfg.Port)
	
	// Support HTTPS if certificates are provided
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	if certFile != "" && keyFile != "" {
		log.Println("Starting HTTPS server on port", cfg.Port)
		log.Fatal(http.ListenAndServeTLS(":"+cfg.Port, certFile, keyFile, mux))
	} else {
		log.Println("Starting HTTP server on port", cfg.Port)
		log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
	}
}
