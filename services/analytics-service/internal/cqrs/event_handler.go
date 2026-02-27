package cqrs

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"analytics-service/config"
	"analytics-service/internal/model"
	"analytics-service/internal/store"
)

// EventHandler handles events and updates read model (CQRS - 2.15)
// This processes events from the event store and updates the projection
type EventHandler struct {
	projectionStore *store.ProjectionStore
	config          *config.Config
}

func NewEventHandler(projectionStore *store.ProjectionStore, cfg *config.Config) *EventHandler {
	return &EventHandler{
		projectionStore: projectionStore,
		config:          cfg,
	}
}

// HandleEvent processes an event and updates the read model (projection)
// It checks the last processed event version to avoid processing the same event multiple times
func (eh *EventHandler) HandleEvent(ctx context.Context, event *model.UserEvent) error {
	// CRITICAL: Check if this event has already been processed to prevent duplicate counting
	projection, err := eh.projectionStore.GetProjection(ctx, event.StreamID)
	if err == nil && projection != nil {
		// If event version is less than or equal to last processed version, skip it
		if event.Version > 0 && event.Version <= projection.LastProcessedEventVersion {
			// Event already processed, skip to prevent duplicate counting
			log.Printf("Skipping already processed event: streamId=%s, eventType=%s, version=%d (lastProcessed=%d)", 
				event.StreamID, event.EventType, event.Version, projection.LastProcessedEventVersion)
			return nil
		}
	}
	
	var handleErr error
	switch event.EventType {
	case model.EventTypeSongPlayed:
		handleErr = eh.handleSongPlayedEvent(ctx, event)
	case model.EventTypeRatingGiven:
		handleErr = eh.handleRatingGivenEvent(ctx, event)
	case model.EventTypeArtistSubscribed:
		handleErr = eh.handleArtistSubscribedEvent(ctx, event)
	case model.EventTypeArtistUnsubscribed:
		handleErr = eh.handleArtistUnsubscribedEvent(ctx, event)
	default:
		log.Printf("Unknown event type: %s", event.EventType)
		return nil
	}
	
	// Update last processed event version if processing was successful
	// This MUST be done to prevent reprocessing the same event
	if handleErr == nil && event.Version > 0 {
		if err := eh.projectionStore.UpdateLastProcessedEventVersion(ctx, event.StreamID, event.Version); err != nil {
			log.Printf("Warning: Failed to update last processed event version: %v", err)
		}
	}
	
	return handleErr
}

func (eh *EventHandler) handleSongPlayedEvent(ctx context.Context, event *model.UserEvent) error {
	genre := ""
	if g, ok := event.Payload["genre"].(string); ok {
		genre = g
	}
	
	artistIDs := []string{}
	if ids, ok := event.Payload["artistIds"].([]interface{}); ok {
		for _, id := range ids {
			if strID, ok := id.(string); ok {
				artistIDs = append(artistIDs, strID)
			}
		}
	}
	
	// Always fetch artist names from content service to ensure we have real names
	artistNames := make(map[string]string)
	if len(artistIDs) > 0 {
		// Fetch artist names from content service
		artistNames = eh.fetchArtistNames(ctx, artistIDs)
	}
	
	// If genre is not in payload, try to fetch from content service
	if genre == "" {
		if songID, ok := event.Payload["songId"].(string); ok && songID != "" {
			genre = eh.fetchSongGenre(ctx, songID)
		}
	}
	
	// If artistIDs are not in payload, try to fetch from content service
	if len(artistIDs) == 0 {
		if songID, ok := event.Payload["songId"].(string); ok && songID != "" {
			artistIDs = eh.fetchSongArtists(ctx, songID)
			// Fetch artist names for all artists from the song
			if len(artistIDs) > 0 {
				artistNames = eh.fetchArtistNames(ctx, artistIDs)
			}
		}
	} else {
		// If we have artistIDs but no names (or only one name), fetch all names
		if len(artistNames) < len(artistIDs) {
			fetchedNames := eh.fetchArtistNames(ctx, artistIDs)
			// Merge fetched names with existing ones
			for id, name := range fetchedNames {
				if name != "" {
					artistNames[id] = name
				}
			}
		}
	}
	
	return eh.projectionStore.IncrementSongPlayed(ctx, event.StreamID, genre, artistIDs, artistNames)
}

func (eh *EventHandler) handleRatingGivenEvent(ctx context.Context, event *model.UserEvent) error {
	rating := 0
	if r, ok := event.Payload["rating"].(float64); ok {
		rating = int(r)
	} else if r, ok := event.Payload["rating"].(int); ok {
		rating = r
	}
	
	if rating > 0 {
		return eh.projectionStore.AddRating(ctx, event.StreamID, rating)
	}
	
	return nil
}

func (eh *EventHandler) handleArtistSubscribedEvent(ctx context.Context, event *model.UserEvent) error {
	artistID := ""
	artistName := ""
	
	if id, ok := event.Payload["artistId"].(string); ok {
		artistID = id
	}
	if name, ok := event.Payload["artistName"].(string); ok {
		artistName = name
	}
	
	if artistID == "" {
		return nil
	}
	
	// If artist name is not in payload, try to fetch from content service
	if artistName == "" {
		artistName = eh.fetchArtistName(ctx, artistID)
	}
	
	return eh.projectionStore.SubscribeToArtist(ctx, event.StreamID, artistID, artistName)
}

func (eh *EventHandler) handleArtistUnsubscribedEvent(ctx context.Context, event *model.UserEvent) error {
	artistID := ""
	if id, ok := event.Payload["artistId"].(string); ok {
		artistID = id
	}
	
	if artistID == "" {
		return nil
	}
	
	return eh.projectionStore.UnsubscribeFromArtist(ctx, event.StreamID, artistID)
}

// Helper methods to fetch data from content service

func (eh *EventHandler) fetchArtistNames(ctx context.Context, artistIDs []string) map[string]string {
	if len(artistIDs) == 0 {
		return make(map[string]string)
	}
	
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   3 * time.Second,
		Transport: tr,
	}
	
	artistsURL := eh.config.ContentServiceURL + "/artists"
	artistsReq, err := http.NewRequestWithContext(ctx, "GET", artistsURL, nil)
	if err != nil {
		return make(map[string]string)
	}
	
	artistsResp, err := client.Do(artistsReq)
	if err != nil || artistsResp == nil || artistsResp.StatusCode != http.StatusOK {
		return make(map[string]string)
	}
	defer artistsResp.Body.Close()
	
	var allArtists []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if json.NewDecoder(artistsResp.Body).Decode(&allArtists) != nil {
		return make(map[string]string)
	}
	
	artistNames := make(map[string]string)
	for _, artist := range allArtists {
		for _, artistID := range artistIDs {
			if artist.ID == artistID {
				artistNames[artistID] = artist.Name
				break
			}
		}
	}
	
	return artistNames
}

func (eh *EventHandler) fetchArtistName(ctx context.Context, artistID string) string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   3 * time.Second,
		Transport: tr,
	}
	
	artistURL := eh.config.ContentServiceURL + "/artists/" + artistID
	artistReq, err := http.NewRequestWithContext(ctx, "GET", artistURL, nil)
	if err != nil {
		return ""
	}
	
	artistResp, err := client.Do(artistReq)
	if err != nil || artistResp == nil || artistResp.StatusCode != http.StatusOK {
		return ""
	}
	defer artistResp.Body.Close()
	
	var artist struct {
		Name string `json:"name"`
	}
	if json.NewDecoder(artistResp.Body).Decode(&artist) != nil {
		return ""
	}
	
	return artist.Name
}

func (eh *EventHandler) fetchSongGenre(ctx context.Context, songID string) string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   3 * time.Second,
		Transport: tr,
	}
	
	songURL := eh.config.ContentServiceURL + "/songs/" + songID
	songReq, err := http.NewRequestWithContext(ctx, "GET", songURL, nil)
	if err != nil {
		return ""
	}
	
	songResp, err := client.Do(songReq)
	if err != nil || songResp == nil || songResp.StatusCode != http.StatusOK {
		return ""
	}
	defer songResp.Body.Close()
	
	var song struct {
		Genre string `json:"genre"`
	}
	if json.NewDecoder(songResp.Body).Decode(&song) != nil {
		return ""
	}
	
	return song.Genre
}

func (eh *EventHandler) fetchSongArtists(ctx context.Context, songID string) []string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   3 * time.Second,
		Transport: tr,
	}
	
	songURL := eh.config.ContentServiceURL + "/songs/" + songID
	songReq, err := http.NewRequestWithContext(ctx, "GET", songURL, nil)
	if err != nil {
		return []string{}
	}
	
	songResp, err := client.Do(songReq)
	if err != nil || songResp == nil || songResp.StatusCode != http.StatusOK {
		return []string{}
	}
	defer songResp.Body.Close()
	
	var song struct {
		ArtistIDs []string `json:"artistIds"`
	}
	if json.NewDecoder(songResp.Body).Decode(&song) != nil {
		return []string{}
	}
	
	return song.ArtistIDs
}
