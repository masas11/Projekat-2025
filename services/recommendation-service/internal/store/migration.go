package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// SyncFromMongoDB syncs data from MongoDB databases to Neo4j graph
func (s *Neo4jStore) SyncFromMongoDB(ctx context.Context, contentServiceURL, ratingsServiceURL, subscriptionsServiceURL string) error {
	log.Println("Starting MongoDB to Neo4j sync...")

	// Sync users, artists, songs from content service
	if err := s.syncContentData(ctx, contentServiceURL); err != nil {
		log.Printf("Error syncing content data: %v", err)
		return err
	}

	// Sync ratings from ratings service
	if err := s.syncRatingsData(ctx, ratingsServiceURL); err != nil {
		log.Printf("Error syncing ratings data: %v", err)
		// Continue even if this fails
	}

	// Sync subscriptions from subscriptions service
	if err := s.syncSubscriptionsData(ctx, subscriptionsServiceURL); err != nil {
		log.Printf("Error syncing subscriptions data: %v", err)
		// Continue even if this fails
	}

	log.Println("MongoDB to Neo4j sync completed")
	return nil
}

func (s *Neo4jStore) syncContentData(ctx context.Context, contentServiceURL string) error {
	client := &http.Client{Timeout: 30 * time.Second}

	// Get all artists
	resp, err := client.Get(contentServiceURL + "/artists")
	if err != nil {
		return fmt.Errorf("failed to fetch artists: %w", err)
	}
	defer resp.Body.Close()

	var artists []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&artists); err != nil {
		return fmt.Errorf("failed to decode artists: %w", err)
	}

	for _, artist := range artists {
		artistID, _ := artist["id"].(string)
		artistName, _ := artist["name"].(string)
		genresInterface, _ := artist["genres"].([]interface{})

		genres := make([]string, 0, len(genresInterface))
		for _, g := range genresInterface {
			if genreStr, ok := g.(string); ok {
				genres = append(genres, genreStr)
			}
		}

		if artistID != "" && artistName != "" {
			if err := s.AddOrUpdateArtist(ctx, artistID, artistName, genres); err != nil {
				log.Printf("Error adding artist %s: %v", artistID, err)
			}
		}
	}

	// Get all songs
	resp, err = client.Get(contentServiceURL + "/songs")
	if err != nil {
		return fmt.Errorf("failed to fetch songs: %w", err)
	}
	defer resp.Body.Close()

	var songs []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&songs); err != nil {
		return fmt.Errorf("failed to decode songs: %w", err)
	}

	for _, song := range songs {
		songID, _ := song["id"].(string)
		songName, _ := song["name"].(string)
		genre, _ := song["genre"].(string)
		artistIDsInterface, _ := song["artistIds"].([]interface{})
		albumID, _ := song["albumId"].(string)
		durationInterface, _ := song["duration"]

		artistIDs := make([]string, 0, len(artistIDsInterface))
		for _, id := range artistIDsInterface {
			if idStr, ok := id.(string); ok {
				artistIDs = append(artistIDs, idStr)
			}
		}

		duration := 0
		if durationInterface != nil {
			if d, ok := durationInterface.(float64); ok {
				duration = int(d)
			} else if d, ok := durationInterface.(int); ok {
				duration = d
			}
		}

		if songID != "" && songName != "" && genre != "" {
			if err := s.AddOrUpdateSong(ctx, songID, songName, genre, artistIDs, albumID, duration); err != nil {
				log.Printf("Error adding song %s: %v", songID, err)
			}
		}
	}

	log.Printf("Synced %d artists and %d songs from content service", len(artists), len(songs))
	return nil
}

func (s *Neo4jStore) syncRatingsData(ctx context.Context, ratingsServiceURL string) error {
	// Note: ratings-service doesn't have a public endpoint to get all ratings
	// We'll need to add one or sync via events
	// For now, we'll skip this and rely on events
	log.Println("Ratings sync skipped - will be populated via events")
	return nil
}

func (s *Neo4jStore) syncSubscriptionsData(ctx context.Context, subscriptionsServiceURL string) error {
	// Note: subscriptions-service doesn't have an endpoint to get all subscriptions
	// We'll need to get subscriptions per user, but we don't have a list of all users
	// For now, we'll skip this and rely on events
	// Users should re-subscribe or we need to add an admin endpoint to sync all subscriptions
	log.Println("Subscriptions sync skipped - existing subscriptions should be re-synced via events")
	log.Println("To sync existing subscriptions, users should unsubscribe and re-subscribe, or trigger sync manually")
	return nil
}
