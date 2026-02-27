package model

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventType represents the type of event (same as ActivityType for compatibility)
type EventType string

const (
	EventTypeSongPlayed        EventType = "SONG_PLAYED"
	EventTypeRatingGiven       EventType = "RATING_GIVEN"
	EventTypeGenreSubscribed   EventType = "GENRE_SUBSCRIBED"
	EventTypeGenreUnsubscribed EventType = "GENRE_UNSUBSCRIBED"
	EventTypeArtistSubscribed  EventType = "ARTIST_SUBSCRIBED"
	EventTypeArtistUnsubscribed EventType = "ARTIST_UNSUBSCRIBED"
)

// UserEvent represents an immutable event in the event store (2.14 Event Sourcing)
type UserEvent struct {
	// Event metadata
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	EventID   string             `json:"eventId" bson:"eventId"` // Unique event identifier
	EventType EventType          `json:"eventType" bson:"eventType"`
	StreamID  string             `json:"streamId" bson:"streamId"` // User ID (stream identifier)
	Version   int64              `json:"version" bson:"version"`   // Event version in stream (sequence number)
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
	
	// Event payload (immutable data)
	Payload map[string]interface{} `json:"payload" bson:"payload"`
	
	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// ToUserActivity converts an event to UserActivity for backward compatibility
func (e *UserEvent) ToUserActivity() *UserActivity {
	activity := &UserActivity{
		ID:        e.ID,
		UserID:    e.StreamID,
		Type:      ActivityType(e.EventType),
		Timestamp: e.Timestamp,
	}
	
	// Extract fields from payload
	if songID, ok := e.Payload["songId"].(string); ok {
		activity.SongID = songID
	}
	if songName, ok := e.Payload["songName"].(string); ok {
		activity.SongName = songName
	}
	if rating, ok := e.Payload["rating"].(float64); ok {
		activity.Rating = int(rating)
	}
	if genre, ok := e.Payload["genre"].(string); ok {
		activity.Genre = genre
	}
	if artistID, ok := e.Payload["artistId"].(string); ok {
		activity.ArtistID = artistID
	}
	if artistName, ok := e.Payload["artistName"].(string); ok {
		activity.ArtistName = artistName
	}
	
	return activity
}

// MarshalJSON customizes JSON marshaling
func (e *UserEvent) MarshalJSON() ([]byte, error) {
	type Alias UserEvent
	return json.Marshal(&struct {
		ID string `json:"id"`
		*Alias
	}{
		ID:    e.ID.Hex(),
		Alias: (*Alias)(e),
	})
}

// UserActivityState represents the reconstructed state from events
type UserActivityState struct {
	UserID              string                 `json:"userId"`
	TotalSongsPlayed    int                    `json:"totalSongsPlayed"`
	TotalRatingsGiven   int                    `json:"totalRatingsGiven"`
	SubscribedGenres    []string               `json:"subscribedGenres"`
	SubscribedArtists   []string               `json:"subscribedArtists"`
	LastActivityTime    *time.Time             `json:"lastActivityTime,omitempty"`
	ActivityBreakdown   map[string]int         `json:"activityBreakdown"` // Count by activity type
	RecentActivities    []*UserActivity         `json:"recentActivities,omitempty"`
}
