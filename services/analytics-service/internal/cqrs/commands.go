package cqrs

import (
	"time"

	"analytics-service/internal/model"
)

// Command represents a command in CQRS pattern (2.15)
type Command interface {
	GetUserID() string
	GetTimestamp() time.Time
}

// PlaySongCommand represents a command to play a song
type PlaySongCommand struct {
	UserID    string
	SongID    string
	SongName  string
	Genre     string
	ArtistIDs []string
	Timestamp time.Time
}

func (c *PlaySongCommand) GetUserID() string {
	return c.UserID
}

func (c *PlaySongCommand) GetTimestamp() time.Time {
	if c.Timestamp.IsZero() {
		return time.Now()
	}
	return c.Timestamp
}

// RateSongCommand represents a command to rate a song
type RateSongCommand struct {
	UserID    string
	SongID    string
	Rating    int
	Timestamp time.Time
}

func (c *RateSongCommand) GetUserID() string {
	return c.UserID
}

func (c *RateSongCommand) GetTimestamp() time.Time {
	if c.Timestamp.IsZero() {
		return time.Now()
	}
	return c.Timestamp
}

// SubscribeToArtistCommand represents a command to subscribe to an artist
type SubscribeToArtistCommand struct {
	UserID    string
	ArtistID  string
	ArtistName string
	Timestamp time.Time
}

func (c *SubscribeToArtistCommand) GetUserID() string {
	return c.UserID
}

func (c *SubscribeToArtistCommand) GetTimestamp() time.Time {
	if c.Timestamp.IsZero() {
		return time.Now()
	}
	return c.Timestamp
}

// UnsubscribeFromArtistCommand represents a command to unsubscribe from an artist
type UnsubscribeFromArtistCommand struct {
	UserID    string
	ArtistID  string
	Timestamp time.Time
}

func (c *UnsubscribeFromArtistCommand) GetUserID() string {
	return c.UserID
}

func (c *UnsubscribeFromArtistCommand) GetTimestamp() time.Time {
	if c.Timestamp.IsZero() {
		return time.Now()
	}
	return c.Timestamp
}

// CommandResult represents the result of executing a command
type CommandResult struct {
	Event   *model.UserEvent
	Success bool
	Error   error
}
