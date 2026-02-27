package analytics

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// ActivityType represents the type of user activity
type ActivityType string

const (
	ActivityTypeSongPlayed        ActivityType = "SONG_PLAYED"
	ActivityTypeRatingGiven       ActivityType = "RATING_GIVEN"
	ActivityTypeGenreSubscribed   ActivityType = "GENRE_SUBSCRIBED"
	ActivityTypeGenreUnsubscribed ActivityType = "GENRE_UNSUBSCRIBED"
	ActivityTypeArtistSubscribed  ActivityType = "ARTIST_SUBSCRIBED"
	ActivityTypeArtistUnsubscribed ActivityType = "ARTIST_UNSUBSCRIBED"
)

// Activity represents a user activity to be logged
type Activity struct {
	UserID     string       `json:"userId"`
	Type       ActivityType `json:"type"`
	SongID     string       `json:"songId,omitempty"`
	SongName   string       `json:"songName,omitempty"`
	Rating     int          `json:"rating,omitempty"`
	Genre      string       `json:"genre,omitempty"`
	ArtistID   string       `json:"artistId,omitempty"`
	ArtistName string       `json:"artistName,omitempty"`
}

// LogActivity logs a user activity to analytics service asynchronously
func LogActivity(analyticsServiceURL string, activity Activity) {
	if analyticsServiceURL == "" {
		log.Printf("Analytics service URL not configured, skipping activity log")
		return // Analytics service not configured
	}

	go func() {
		log.Printf("Logging activity: type=%s, userId=%s, songId=%s", activity.Type, activity.UserID, activity.SongID)
		
		activityJSON, err := json.Marshal(activity)
		if err != nil {
			log.Printf("Failed to marshal activity: %v", err)
			return
		}

		url := analyticsServiceURL + "/activities"
		req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(activityJSON))
		if err != nil {
			log.Printf("Failed to create activity request: %v", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		// Configure TLS transport for HTTPS
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{
			Timeout:   2 * time.Second,
			Transport: tr,
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to log activity to %s: %v", url, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			log.Printf("Analytics service returned non-OK status: %d for activity type=%s, userId=%s", resp.StatusCode, activity.Type, activity.UserID)
		} else {
			log.Printf("Activity logged successfully: type=%s, userId=%s", activity.Type, activity.UserID)
		}
	}()
}
