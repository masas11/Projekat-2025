package handler

import (
	"context"
	"log"
	"time"

	"analytics-service/config"
	"analytics-service/internal/cqrs"
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
// DISABLED: Events are already processed synchronously in LogActivity when they are created.
// This background processor was causing duplicate event processing and incorrect analytics counts.
// If we need to catch up on missed events, we can enable this with proper version checking.
func processNewEvents(ctx context.Context, _ *store.EventStore, _ *cqrs.EventHandler) {
	// DISABLED: Events are already processed synchronously in LogActivity when they are created.
	// This background processor was causing duplicate event processing and incorrect analytics counts.
	// If we need to catch up on missed events, we can enable this with proper version checking.
	
	// Do nothing - events are processed synchronously in LogActivity
	return
}
