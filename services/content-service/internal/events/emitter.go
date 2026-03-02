package events

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"shared/tracing"
	"go.opentelemetry.io/otel/propagation"
)

// Event types
type EventType string

const (
	EventTypeNewArtist EventType = "new_artist"
	EventTypeNewAlbum  EventType = "new_album"
	EventTypeNewSong   EventType = "new_song"
	EventTypeDeletedArtist EventType = "artist_deleted"
	EventTypeDeletedAlbum  EventType = "album_deleted"
	EventTypeDeletedSong   EventType = "song_deleted"
)

// Event payloads
type NewArtistEvent struct {
	Type    EventType `json:"type"`
	ArtistID string   `json:"artistId"`
	Name     string   `json:"name"`
	Genres   []string `json:"genres"`
}

type NewAlbumEvent struct {
	Type      EventType `json:"type"`
	AlbumID   string    `json:"albumId"`
	Name      string    `json:"name"`
	Genre     string    `json:"genre"`
	ArtistIDs []string  `json:"artistIds"`
	ArtistNames []string `json:"artistNames"` // Added for better notification messages
}

type NewSongEvent struct {
	Type        EventType `json:"type"`
	SongID      string    `json:"songId"`
	Name        string    `json:"name"`
	Genre       string    `json:"genre"`
	ArtistIDs   []string  `json:"artistIds"`
	ArtistNames []string  `json:"artistNames"` // Added for better notification messages
	AlbumID     string    `json:"albumId"`
}

// Deletion event payloads
type DeletedSongEvent struct {
	Type   EventType `json:"type"`
	SongID string    `json:"songId"`
}

type DeletedAlbumEvent struct {
	Type    EventType `json:"type"`
	AlbumID string    `json:"albumId"`
}

type DeletedArtistEvent struct {
	Type     EventType `json:"type"`
	ArtistID string   `json:"artistId"`
}

// EmitEvent sends an event to subscriptions-service asynchronously
// Note: Logger parameter is optional - if nil, uses standard log
// Tracing (2.10): Added context parameter for async tracing
func EmitEvent(ctx context.Context, subscriptionsServiceURL string, event interface{}) {
	go func() {
		// Use background context for async operations to avoid cancellation
		// when the original request context is cancelled
		bgCtx := context.Background()
		
		// Start span for async event emission (2.10)
		eventCtx, span := tracing.StartSpan(bgCtx, "emit.event")
		defer span.End()

		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %v", err)
			return
		}

		url := subscriptionsServiceURL + "/events"
		req, err := http.NewRequestWithContext(eventCtx, "POST", url, bytes.NewBuffer(eventJSON))
		if err != nil {
			log.Printf("Failed to create event request: %v", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		// Propagate trace context to downstream service (2.10)
		propagator := tracing.GetPropagator()
		if propagator != nil {
			propagator.Inject(eventCtx, propagation.HeaderCarrier(req.Header))
		}

		// Configure TLS transport for HTTPS
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{
			Timeout:   10 * time.Second, // Increased timeout for event delivery
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
