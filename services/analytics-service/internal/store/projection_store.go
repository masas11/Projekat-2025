package store

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ProjectionStore implements CQRS Read Model (2.15)
// This is an optimized read model (projection) for fast queries
type ProjectionStore struct {
	collection *mongo.Collection
}

func NewProjectionStore(db *mongo.Database) *ProjectionStore {
	collection := db.Collection("analytics_projections")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Create unique index on userId
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "userId", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("Warning: Failed to create index: %v", err)
	}
	
	return &ProjectionStore{
		collection: collection,
	}
}

// AnalyticsProjection represents the read model for user analytics
type AnalyticsProjection struct {
	ID                     primitive.ObjectID `bson:"_id,omitempty"`
	UserID                 string             `bson:"userId"`
	TotalSongsPlayed       int                `bson:"totalSongsPlayed"`
	TotalRatingsSum         float64            `bson:"totalRatingsSum"`
	TotalRatingsCount       int                `bson:"totalRatingsCount"`
	SongsPlayedByGenre      map[string]int     `bson:"songsPlayedByGenre"`
	ArtistPlayCounts        map[string]int     `bson:"artistPlayCounts"` // artistID -> count
	ArtistNames              map[string]string  `bson:"artistNames"`       // artistID -> name
	SubscribedArtists        map[string]bool    `bson:"subscribedArtists"` // artistID -> true
	LastUpdated              time.Time          `bson:"lastUpdated"`
	LastProcessedEventVersion int64            `bson:"lastProcessedEventVersion"` // Last event version processed
}

// GetProjection retrieves the analytics projection for a user
func (ps *ProjectionStore) GetProjection(ctx context.Context, userID string) (*AnalyticsProjection, error) {
	var projection AnalyticsProjection
	err := ps.collection.FindOne(ctx, bson.M{"userId": userID}).Decode(&projection)
	if err == mongo.ErrNoDocuments {
		// Return empty projection if not found
		return &AnalyticsProjection{
			UserID:                userID,
			SongsPlayedByGenre:    make(map[string]int),
			ArtistPlayCounts:       make(map[string]int),
			ArtistNames:           make(map[string]string),
			SubscribedArtists:     make(map[string]bool),
			LastUpdated:           time.Now(),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	
	// Initialize maps if nil
	if projection.SongsPlayedByGenre == nil {
		projection.SongsPlayedByGenre = make(map[string]int)
	}
	if projection.ArtistPlayCounts == nil {
		projection.ArtistPlayCounts = make(map[string]int)
	}
	if projection.ArtistNames == nil {
		projection.ArtistNames = make(map[string]string)
	}
	if projection.SubscribedArtists == nil {
		projection.SubscribedArtists = make(map[string]bool)
	}
	
	return &projection, nil
}

// UpsertProjection creates or updates the analytics projection
func (ps *ProjectionStore) UpsertProjection(ctx context.Context, projection *AnalyticsProjection) error {
	if projection.UserID == "" {
		return mongo.ErrNoDocuments
	}
	
	projection.LastUpdated = time.Now()
	
	// Ensure maps are initialized
	if projection.SongsPlayedByGenre == nil {
		projection.SongsPlayedByGenre = make(map[string]int)
	}
	if projection.ArtistPlayCounts == nil {
		projection.ArtistPlayCounts = make(map[string]int)
	}
	if projection.ArtistNames == nil {
		projection.ArtistNames = make(map[string]string)
	}
	if projection.SubscribedArtists == nil {
		projection.SubscribedArtists = make(map[string]bool)
	}
	
	filter := bson.M{"userId": projection.UserID}
	update := bson.M{
		"$set": bson.M{
			"totalSongsPlayed":          projection.TotalSongsPlayed,
			"totalRatingsSum":            projection.TotalRatingsSum,
			"totalRatingsCount":          projection.TotalRatingsCount,
			"songsPlayedByGenre":         projection.SongsPlayedByGenre,
			"artistPlayCounts":           projection.ArtistPlayCounts,
			"artistNames":                projection.ArtistNames,
			"subscribedArtists":          projection.SubscribedArtists,
			"lastUpdated":                projection.LastUpdated,
			"lastProcessedEventVersion":  projection.LastProcessedEventVersion,
		},
		"$setOnInsert": bson.M{
			"userId": projection.UserID,
		},
	}
	
	opts := options.Update().SetUpsert(true)
	_, err := ps.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Printf("Error upserting projection: %v", err)
		return err
	}
	
	return nil
}

// IncrementSongPlayed increments the song play count in the projection
func (ps *ProjectionStore) IncrementSongPlayed(ctx context.Context, userID string, genre string, artistIDs []string, artistNames map[string]string) error {
	// First, get the current projection to update it properly
	projection, err := ps.GetProjection(ctx, userID)
	if err != nil {
		return err
	}
	
	// Increment counters
	projection.TotalSongsPlayed++
	
	if genre != "" {
		if projection.SongsPlayedByGenre == nil {
			projection.SongsPlayedByGenre = make(map[string]int)
		}
		projection.SongsPlayedByGenre[genre]++
	}
	
	// Increment artist play counts
	if projection.ArtistPlayCounts == nil {
		projection.ArtistPlayCounts = make(map[string]int)
	}
	for _, artistID := range artistIDs {
		if artistID != "" {
			projection.ArtistPlayCounts[artistID]++
		}
	}
	
	// Update artist names
	if projection.ArtistNames == nil {
		projection.ArtistNames = make(map[string]string)
	}
	for artistID, name := range artistNames {
		if artistID != "" && name != "" {
			projection.ArtistNames[artistID] = name
		}
	}
	
	// Save the updated projection
	return ps.UpsertProjection(ctx, projection)
}

// AddRating adds a rating to the projection
func (ps *ProjectionStore) AddRating(ctx context.Context, userID string, rating int) error {
	// Get current projection
	projection, err := ps.GetProjection(ctx, userID)
	if err != nil {
		return err
	}
	
	// Increment rating counters
	projection.TotalRatingsSum += float64(rating)
	projection.TotalRatingsCount++
	
	// Save updated projection
	return ps.UpsertProjection(ctx, projection)
}

// SubscribeToArtist adds an artist subscription to the projection
func (ps *ProjectionStore) SubscribeToArtist(ctx context.Context, userID string, artistID string, artistName string) error {
	// Get current projection
	projection, err := ps.GetProjection(ctx, userID)
	if err != nil {
		return err
	}
	
	// Initialize maps if needed
	if projection.SubscribedArtists == nil {
		projection.SubscribedArtists = make(map[string]bool)
	}
	if projection.ArtistNames == nil {
		projection.ArtistNames = make(map[string]string)
	}
	
	// Add subscription
	projection.SubscribedArtists[artistID] = true
	
	// Update artist name if provided
	if artistName != "" {
		projection.ArtistNames[artistID] = artistName
	}
	
	// Save updated projection
	return ps.UpsertProjection(ctx, projection)
}

// UnsubscribeFromArtist removes an artist subscription from the projection
func (ps *ProjectionStore) UnsubscribeFromArtist(ctx context.Context, userID string, artistID string) error {
	// Get current projection
	projection, err := ps.GetProjection(ctx, userID)
	if err != nil {
		return err
	}
	
	// Initialize map if needed
	if projection.SubscribedArtists == nil {
		projection.SubscribedArtists = make(map[string]bool)
	}
	
	// Remove subscription
	delete(projection.SubscribedArtists, artistID)
	
	// Save updated projection
	return ps.UpsertProjection(ctx, projection)
}

// UpdateLastProcessedEventVersion updates the last processed event version for a user
func (ps *ProjectionStore) UpdateLastProcessedEventVersion(ctx context.Context, userID string, version int64) error {
	filter := bson.M{"userId": userID}
	update := bson.M{
		"$set": bson.M{
			"lastProcessedEventVersion": version,
			"lastUpdated":               time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := ps.collection.UpdateOne(ctx, filter, update, opts)
	return err
}
