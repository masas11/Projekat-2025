package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"subscriptions-service/internal/model"
)

type SubscriptionRepository struct {
	collection *mongo.Collection
}

func NewSubscriptionRepository(db *mongo.Database) *SubscriptionRepository {
	return &SubscriptionRepository{
		collection: db.Collection("subscriptions"),
	}
}

func (r *SubscriptionRepository) Create(ctx context.Context, subscription *model.Subscription) error {
	// Check if subscription already exists
	var existing model.Subscription
	filter := bson.M{
		"userId": subscription.UserID,
		"type":   subscription.Type,
	}
	
	if subscription.Type == "artist" {
		filter["artistId"] = subscription.ArtistID
	} else if subscription.Type == "genre" {
		filter["genre"] = subscription.Genre
	}

	err := r.collection.FindOne(ctx, filter).Decode(&existing)
	if err == nil {
		return errors.New("subscription already exists")
	}
	if err != mongo.ErrNoDocuments {
		return err
	}

	// Set ID and timestamps
	if subscription.ID == "" {
		subscription.ID = uuid.NewString()
	}
	subscription.CreatedAt = time.Now()

	_, err = r.collection.InsertOne(ctx, subscription)
	return err
}

func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Subscription, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subscriptions []*model.Subscription
	if err = cursor.All(ctx, &subscriptions); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) GetByUserAndArtist(ctx context.Context, userID, artistID string) (*model.Subscription, error) {
	var subscription model.Subscription
	err := r.collection.FindOne(ctx, bson.M{
		"userId":  userID,
		"type":    "artist",
		"artistId": artistID,
	}).Decode(&subscription)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &subscription, nil
}

func (r *SubscriptionRepository) GetByUserAndGenre(ctx context.Context, userID, genre string) (*model.Subscription, error) {
	var subscription model.Subscription
	err := r.collection.FindOne(ctx, bson.M{
		"userId": userID,
		"type":   "genre",
		"genre":  genre,
	}).Decode(&subscription)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &subscription, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, subscriptionID string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": subscriptionID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("subscription not found")
	}
	return nil
}

func (r *SubscriptionRepository) DeleteByUserAndArtist(ctx context.Context, userID, artistID string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{
		"userId":  userID,
		"type":    "artist",
		"artistId": artistID,
	})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("subscription not found")
	}
	return nil
}

func (r *SubscriptionRepository) DeleteByUserAndGenre(ctx context.Context, userID, genre string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{
		"userId": userID,
		"type":   "genre",
		"genre":  genre,
	})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("subscription not found")
	}
	return nil
}

// GetByArtistID returns all subscriptions for a specific artist
func (r *SubscriptionRepository) GetByArtistID(ctx context.Context, artistID string) ([]*model.Subscription, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"type":    "artist",
		"artistId": artistID,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subscriptions []*model.Subscription
	if err = cursor.All(ctx, &subscriptions); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

// GetByGenre returns all subscriptions for a specific genre
func (r *SubscriptionRepository) GetByGenre(ctx context.Context, genre string) ([]*model.Subscription, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"type":  "genre",
		"genre": genre,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subscriptions []*model.Subscription
	if err = cursor.All(ctx, &subscriptions); err != nil {
		return nil, err
	}

	return subscriptions, nil
}
