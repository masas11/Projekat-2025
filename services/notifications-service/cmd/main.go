package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"notifications-service/config"
	"notifications-service/internal/handler"
	"notifications-service/internal/model"
	"notifications-service/internal/store"
)

func initSampleNotifications(ctx context.Context, repo *store.NotificationRepository) {
	// Create sample notifications for testing
	sampleNotifications := []*model.Notification{
		{
			ID:        uuid.NewString(),
			UserID:    "user1",
			Type:      "new_album",
			Message:   "New album 'Thriller' by Michael Jackson has been released",
			ContentID: "album1",
			Read:      false,
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.NewString(),
			UserID:    "user1",
			Type:      "new_song",
			Message:   "New song 'Billie Jean' by Michael Jackson has been added",
			ContentID: "song1",
			Read:      false,
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        uuid.NewString(),
			UserID:    "user1",
			Type:      "new_artist",
			Message:   "New artist 'The Weeknd' in genre Pop has been added",
			ContentID: "artist1",
			Read:      true,
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        uuid.NewString(),
			UserID:    "user2",
			Type:      "new_album",
			Message:   "New album 'Abbey Road' by The Beatles has been released",
			ContentID: "album2",
			Read:      false,
			CreatedAt: time.Now(),
		},
	}

	for _, notif := range sampleNotifications {
		repo.Create(ctx, notif)
	}
	log.Println("Sample notifications initialized")
}

func main() {
	cfg := config.Load()

	// Initialize MongoDB connection
	dbStore, err := store.NewMongoDBStore(cfg.MongoDBURI, cfg.MongoDBDatabase)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer dbStore.Close()
	log.Println("Connected to MongoDB")

	// Initialize repository
	notificationRepo := store.NewNotificationRepository(dbStore.Database)

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
	mux.HandleFunc("/notifications", notificationHandler.GetNotifications)

	log.Println("Notifications service running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
