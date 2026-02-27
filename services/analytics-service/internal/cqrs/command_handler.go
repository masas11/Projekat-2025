package cqrs

import (
	"context"
	"fmt"
	"log"

	"analytics-service/internal/model"
	"analytics-service/internal/store"
)

// CommandHandler handles commands and creates events (CQRS Command Side - 2.15)
type CommandHandler struct {
	eventStore *store.EventStore
}

func NewCommandHandler(eventStore *store.EventStore) *CommandHandler {
	return &CommandHandler{
		eventStore: eventStore,
	}
}

// HandleCommand processes a command and creates an event
func (ch *CommandHandler) HandleCommand(ctx context.Context, cmd Command) *CommandResult {
	var event *model.UserEvent
	
	switch c := cmd.(type) {
	case *PlaySongCommand:
		event = ch.handlePlaySongCommand(ctx, c)
	case *RateSongCommand:
		event = ch.handleRateSongCommand(ctx, c)
	case *SubscribeToArtistCommand:
		event = ch.handleSubscribeToArtistCommand(ctx, c)
	case *UnsubscribeFromArtistCommand:
		event = ch.handleUnsubscribeFromArtistCommand(ctx, c)
	default:
		return &CommandResult{
			Success: false,
			Error:   fmt.Errorf("unknown command type"),
		}
	}
	
	if event == nil {
		return &CommandResult{
			Success: false,
			Error:   fmt.Errorf("failed to create event"),
		}
	}
	
	// Append event to event store
	if err := ch.eventStore.AppendEvent(ctx, event); err != nil {
		log.Printf("Error appending event: %v", err)
		return &CommandResult{
			Event:   event,
			Success: false,
			Error:   err,
		}
	}
	
	log.Printf("Command handled successfully: userID=%s, eventType=%s", cmd.GetUserID(), event.EventType)
	
	return &CommandResult{
		Event:   event,
		Success: true,
	}
}

func (ch *CommandHandler) handlePlaySongCommand(ctx context.Context, cmd *PlaySongCommand) *model.UserEvent {
	payload := make(map[string]interface{})
	payload["songId"] = cmd.SongID
	if cmd.SongName != "" {
		payload["songName"] = cmd.SongName
	}
	if cmd.Genre != "" {
		payload["genre"] = cmd.Genre
	}
	if len(cmd.ArtistIDs) > 0 {
		payload["artistIds"] = cmd.ArtistIDs
	}
	
	return &model.UserEvent{
		EventType: model.EventTypeSongPlayed,
		StreamID:  cmd.UserID,
		Timestamp: cmd.GetTimestamp(),
		Payload:   payload,
	}
}

func (ch *CommandHandler) handleRateSongCommand(ctx context.Context, cmd *RateSongCommand) *model.UserEvent {
	payload := make(map[string]interface{})
	payload["songId"] = cmd.SongID
	payload["rating"] = cmd.Rating
	
	return &model.UserEvent{
		EventType: model.EventTypeRatingGiven,
		StreamID:  cmd.UserID,
		Timestamp: cmd.GetTimestamp(),
		Payload:   payload,
	}
}

func (ch *CommandHandler) handleSubscribeToArtistCommand(ctx context.Context, cmd *SubscribeToArtistCommand) *model.UserEvent {
	payload := make(map[string]interface{})
	payload["artistId"] = cmd.ArtistID
	if cmd.ArtistName != "" {
		payload["artistName"] = cmd.ArtistName
	}
	
	return &model.UserEvent{
		EventType: model.EventTypeArtistSubscribed,
		StreamID:  cmd.UserID,
		Timestamp: cmd.GetTimestamp(),
		Payload:   payload,
	}
}

func (ch *CommandHandler) handleUnsubscribeFromArtistCommand(ctx context.Context, cmd *UnsubscribeFromArtistCommand) *model.UserEvent {
	payload := make(map[string]interface{})
	payload["artistId"] = cmd.ArtistID
	
	return &model.UserEvent{
		EventType: model.EventTypeArtistUnsubscribed,
		StreamID:  cmd.UserID,
		Timestamp: cmd.GetTimestamp(),
		Payload:   payload,
	}
}
