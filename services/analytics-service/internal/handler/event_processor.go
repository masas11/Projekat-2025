package handler

import (
	"context"
	"log"
	"time"

	"analytics-service/config"
	"analytics-service/internal/cqrs"
	"analytics-service/internal/model"
	"analytics-service/internal/store"
)

// NewEventHandler creates a new event handler for updating projections
func NewEventHandler(projectionStore *store.ProjectionStore, cfg *config.Config) *cqrs.EventHandler {
	return cqrs.NewEventHandler(projectionStore, cfg)
}

// StartEventProcessor starts a background worker that processes events and updates projections (2.15 CQRS)
func StartEventProcessor(ctx context.Context, eventStore *store.EventStore, eventHandler *cqrs.EventHandler) {
	ticker := time.NewTicker(5 * time.Second) // Process events every 5 seconds
	defer ticker.Stop()

	log.Println("Event processor started - processing events every 5 seconds")

	for {
		select {
		case <-ctx.Done():
			log.Println("Event processor stopped")
			return
		case <-ticker.C:
			processNewEvents(ctx, eventStore, eventHandler)
		}
	}
}

// processNewEvents processes new events from the event store and updates projections
// It tracks the last processed event version per user to avoid processing the same events multiple times
// NOTE: Events are already processed synchronously in LogActivity, so this is mainly a catch-up mechanism
func processNewEvents(ctx context.Context, eventStore *store.EventStore, eventHandler *cqrs.EventHandler) {
	// This function is called periodically to catch up on any missed events
	// The main event processing happens synchronously in LogActivity when events are created
	// We only process events from the last 1 minute to catch any that might have been missed
	
	log.Println("Event processor tick - checking for missed events (last 1 minute only)")
	
	// Only process events from the last 1 minute to avoid reprocessing old events
	cutoffTime := time.Now().Add(-1 * time.Minute)
	
	// Get recent events of each type
	eventTypes := []model.EventType{
		model.EventTypeSongPlayed,
		model.EventTypeRatingGiven,
		model.EventTypeArtistSubscribed,
		model.EventTypeArtistUnsubscribed,
	}

	processedCount := 0
	
	for _, eventType := range eventTypes {
		events, err := eventStore.GetEventsByType(ctx, eventType, 50) // Get last 50 events of each type
		if err != nil {
			log.Printf("Error getting events by type %s: %v", eventType, err)
			continue
		}

		// Process only very recent events (last 1 minute) that haven't been processed yet
		// The EventHandler will check LastProcessedEventVersion to skip already processed events
		for _, event := range events {
			if event.Timestamp.Before(cutoffTime) {
				continue // Skip old events
			}
			
			// EventHandler will check LastProcessedEventVersion internally and skip if already processed
			if err := eventHandler.HandleEvent(ctx, event); err != nil {
				log.Printf("Error handling event %s for user %s (version %d): %v", event.EventType, event.StreamID, event.Version, err)
			} else {
				processedCount++
			}
		}
	}

	if processedCount > 0 {
		log.Printf("Processed %d new events and updated projections", processedCount)
	}
}
