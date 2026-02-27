package store

import (
	"context"
	"fmt"
	"log"
	"time"

	"analytics-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EventStore implements Event Sourcing pattern (2.14)
type EventStore struct {
	collection *mongo.Collection
}

func NewEventStore(db *mongo.Database) *EventStore {
	collection := db.Collection("event_store")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Create indexes for efficient queries
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "streamId", Value: 1},
				{Key: "version", Value: 1},
			},
			Options: options.Index().SetUnique(true), // Ensure version uniqueness per stream
		},
		{
			Keys: bson.D{
				{Key: "streamId", Value: 1},
				{Key: "timestamp", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "eventType", Value: 1},
				{Key: "timestamp", Value: -1},
			},
		},
	}
	
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}
	
	return &EventStore{
		collection: collection,
	}
}

// AppendEvent appends a new event to the event store (append-only)
func (es *EventStore) AppendEvent(ctx context.Context, event *model.UserEvent) error {
	// Generate event ID if not provided
	if event.EventID == "" {
		event.EventID = primitive.NewObjectID().Hex()
	}
	
	// Generate ID if not provided
	if event.ID.IsZero() {
		event.ID = primitive.NewObjectID()
	}
	
	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	
	// Get next version for this stream
	if event.Version == 0 {
		nextVersion, err := es.getNextVersion(ctx, event.StreamID)
		if err != nil {
			return fmt.Errorf("failed to get next version: %w", err)
		}
		event.Version = nextVersion
	}
	
	// Insert event (append-only, immutable)
	_, err := es.collection.InsertOne(ctx, event)
	if err != nil {
		// Check if it's a duplicate version error
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("event version conflict: version %d already exists for stream %s", event.Version, event.StreamID)
		}
		log.Printf("Error appending event: %v", err)
		return fmt.Errorf("failed to append event: %w", err)
	}
	
	log.Printf("Event appended successfully: streamId=%s, eventType=%s, version=%d", event.StreamID, event.EventType, event.Version)
	return nil
}

// getNextVersion gets the next version number for a stream
func (es *EventStore) getNextVersion(ctx context.Context, streamID string) (int64, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})
	
	var lastEvent model.UserEvent
	err := es.collection.FindOne(ctx, bson.M{"streamId": streamID}, opts).Decode(&lastEvent)
	if err == mongo.ErrNoDocuments {
		return 1, nil // First event in stream
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get last version: %w", err)
	}
	
	return lastEvent.Version + 1, nil
}

// GetEventStream retrieves all events for a stream (user) in order
func (es *EventStore) GetEventStream(ctx context.Context, streamID string, fromVersion int64, limit int) ([]*model.UserEvent, error) {
	filter := bson.M{"streamId": streamID}
	if fromVersion > 0 {
		filter["version"] = bson.M{"$gte": fromVersion}
	}
	
	opts := options.Find().SetSort(bson.D{{Key: "version", Value: 1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	
	cursor, err := es.collection.Find(ctx, filter, opts)
	if err != nil {
		log.Printf("Error getting event stream: %v", err)
		return nil, fmt.Errorf("failed to get event stream: %w", err)
	}
	defer cursor.Close(ctx)
	
	var events []*model.UserEvent
	if err = cursor.All(ctx, &events); err != nil {
		log.Printf("Error decoding events: %v", err)
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}
	
	return events, nil
}

// GetEventsByType retrieves events of a specific type
func (es *EventStore) GetEventsByType(ctx context.Context, eventType model.EventType, limit int) ([]*model.UserEvent, error) {
	filter := bson.M{"eventType": eventType}
	
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	
	cursor, err := es.collection.Find(ctx, filter, opts)
	if err != nil {
		log.Printf("Error getting events by type: %v", err)
		return nil, fmt.Errorf("failed to get events by type: %w", err)
	}
	defer cursor.Close(ctx)
	
	var events []*model.UserEvent
	if err = cursor.All(ctx, &events); err != nil {
		log.Printf("Error decoding events: %v", err)
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}
	
	return events, nil
}

// ReplayEvents reconstructs the state by replaying all events for a stream
func (es *EventStore) ReplayEvents(ctx context.Context, streamID string) (*model.UserActivityState, error) {
	events, err := es.GetEventStream(ctx, streamID, 0, 0) // Get all events
	if err != nil {
		return nil, fmt.Errorf("failed to get events for replay: %w", err)
	}
	
	state := &model.UserActivityState{
		UserID:            streamID,
		SubscribedGenres:  make([]string, 0),
		SubscribedArtists: make([]string, 0),
		ActivityBreakdown: make(map[string]int),
		RecentActivities:  make([]*model.UserActivity, 0),
	}
	
	// Replay all events to reconstruct state
	for _, event := range events {
		es.applyEvent(state, event)
	}
	
	return state, nil
}

// applyEvent applies a single event to the state
func (es *EventStore) applyEvent(state *model.UserActivityState, event *model.UserEvent) {
	// Update activity breakdown
	state.ActivityBreakdown[string(event.EventType)]++
	
	// Update last activity time
	if state.LastActivityTime == nil || event.Timestamp.After(*state.LastActivityTime) {
		state.LastActivityTime = &event.Timestamp
	}
	
	// Apply event-specific logic
	switch event.EventType {
	case model.EventTypeSongPlayed:
		state.TotalSongsPlayed++
		activity := event.ToUserActivity()
		state.RecentActivities = append(state.RecentActivities, activity)
		
	case model.EventTypeRatingGiven:
		state.TotalRatingsGiven++
		activity := event.ToUserActivity()
		state.RecentActivities = append(state.RecentActivities, activity)
		
	case model.EventTypeGenreSubscribed:
		if genre, ok := event.Payload["genre"].(string); ok {
			// Add if not already subscribed
			found := false
			for _, g := range state.SubscribedGenres {
				if g == genre {
					found = true
					break
				}
			}
			if !found {
				state.SubscribedGenres = append(state.SubscribedGenres, genre)
			}
		}
		activity := event.ToUserActivity()
		state.RecentActivities = append(state.RecentActivities, activity)
		
	case model.EventTypeGenreUnsubscribed:
		if genre, ok := event.Payload["genre"].(string); ok {
			// Remove from subscribed genres
			for i, g := range state.SubscribedGenres {
				if g == genre {
					state.SubscribedGenres = append(state.SubscribedGenres[:i], state.SubscribedGenres[i+1:]...)
					break
				}
			}
		}
		activity := event.ToUserActivity()
		state.RecentActivities = append(state.RecentActivities, activity)
		
	case model.EventTypeArtistSubscribed:
		if artistID, ok := event.Payload["artistId"].(string); ok {
			// Add if not already subscribed
			found := false
			for _, a := range state.SubscribedArtists {
				if a == artistID {
					found = true
					break
				}
			}
			if !found {
				state.SubscribedArtists = append(state.SubscribedArtists, artistID)
			}
		}
		activity := event.ToUserActivity()
		state.RecentActivities = append(state.RecentActivities, activity)
		
	case model.EventTypeArtistUnsubscribed:
		if artistID, ok := event.Payload["artistId"].(string); ok {
			// Remove from subscribed artists
			for i, a := range state.SubscribedArtists {
				if a == artistID {
					state.SubscribedArtists = append(state.SubscribedArtists[:i], state.SubscribedArtists[i+1:]...)
					break
				}
			}
		}
		activity := event.ToUserActivity()
		state.RecentActivities = append(state.RecentActivities, activity)
	}
	
	// Keep only last 50 recent activities
	if len(state.RecentActivities) > 50 {
		state.RecentActivities = state.RecentActivities[len(state.RecentActivities)-50:]
	}
}
