package events

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Event types
type EventType string

const (
	EventTypeNewArtist EventType = "new_artist"
	EventTypeNewAlbum  EventType = "new_album"
	EventTypeNewSong   EventType = "new_song"
)

// Event payloads
type NewArtistEvent struct {
	Type    EventType `json:"type"`
	ArtistID string   `json:"artistId"`
	Name     string   `json:"name"`
	Genres   []string `json:"genres"`
}

type NewAlbumEvent struct {
	Type     EventType `json:"type"`
	AlbumID  string    `json:"albumId"`
	Name     string    `json:"name"`
	Genre    string    `json:"genre"`
	ArtistIDs []string `json:"artistIds"`
}

type NewSongEvent struct {
	Type      EventType `json:"type"`
	SongID    string    `json:"songId"`
	Name      string    `json:"name"`
	Genre     string    `json:"genre"`
	ArtistIDs []string  `json:"artistIds"`
	AlbumID   string    `json:"albumId"`
}

// EmitEvent sends an event to subscriptions-service asynchronously
func EmitEvent(subscriptionsServiceURL string, event interface{}) {
	go func() {
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %v", err)
			return
		}

		url := subscriptionsServiceURL + "/events"
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(eventJSON))
		if err != nil {
			log.Printf("Failed to create event request: %v", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{
			Timeout: 2 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to emit event to subscriptions-service: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
			log.Printf("Subscriptions-service returned non-OK status: %d", resp.StatusCode)
			return
		}

		log.Printf("Event emitted successfully: %v", event)
	}()
}
