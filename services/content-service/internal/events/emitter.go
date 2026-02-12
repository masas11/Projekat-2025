package events

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"strings"
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
// Note: Logger parameter is optional - if nil, uses standard log
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
			errorMsg := err.Error()
			// Check if it's a TLS error
			if strings.Contains(errorMsg, "tls") || strings.Contains(errorMsg, "TLS") ||
				strings.Contains(errorMsg, "certificate") || strings.Contains(errorMsg, "handshake") {
				log.Printf("[TLS_FAILURE] Failed to emit event to subscriptions-service: %v", errorMsg)
			} else {
				log.Printf("Failed to emit event to subscriptions-service: %v", err)
			}
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
