package model

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActivityType represents the type of user activity
type ActivityType string

const (
	ActivityTypeSongPlayed      ActivityType = "SONG_PLAYED"
	ActivityTypeRatingGiven     ActivityType = "RATING_GIVEN"
	ActivityTypeGenreSubscribed ActivityType = "GENRE_SUBSCRIBED"
	ActivityTypeGenreUnsubscribed ActivityType = "GENRE_UNSUBSCRIBED"
	ActivityTypeArtistSubscribed ActivityType = "ARTIST_SUBSCRIBED"
	ActivityTypeArtistUnsubscribed ActivityType = "ARTIST_UNSUBSCRIBED"
)

// UserActivity represents a user activity record
type UserActivity struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    string      `json:"userId" bson:"userId"`
	Type      ActivityType `json:"type" bson:"type"`
	Timestamp time.Time   `json:"timestamp" bson:"timestamp"`
	
	// Activity-specific data
	SongID    string `json:"songId,omitempty" bson:"songId,omitempty"`
	SongName  string `json:"songName,omitempty" bson:"songName,omitempty"`
	Rating    int    `json:"rating,omitempty" bson:"rating,omitempty"`
	Genre     string `json:"genre,omitempty" bson:"genre,omitempty"`
	ArtistID  string `json:"artistId,omitempty" bson:"artistId,omitempty"`
	ArtistName string `json:"artistName,omitempty" bson:"artistName,omitempty"`
}

// MarshalJSON customizes JSON marshaling to convert ObjectID to string
func (ua UserActivity) MarshalJSON() ([]byte, error) {
	type Alias UserActivity
	return json.Marshal(&struct {
		ID string `json:"id"`
		*Alias
	}{
		ID:    ua.ID.Hex(),
		Alias: (*Alias)(&ua),
	})
}
