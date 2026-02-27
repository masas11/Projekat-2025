package store

import (
	"context"
	"log"
	"time"

	"analytics-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ActivityStore struct {
	collection *mongo.Collection
}

func NewActivityStore(db *mongo.Database) *ActivityStore {
	collection := db.Collection("user_activities")
	
	// Create indexes for better query performance
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Index on userId and timestamp for efficient queries
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "userId", Value: 1},
			{Key: "timestamp", Value: -1},
		},
	}
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("Warning: Failed to create index: %v", err)
	}
	
	return &ActivityStore{
		collection: collection,
	}
}

func (as *ActivityStore) Create(ctx context.Context, activity *model.UserActivity) error {
	// MongoDB will auto-generate _id if not provided
	if activity.ID.IsZero() {
		activity.ID = primitive.NewObjectID()
	}
	if activity.Timestamp.IsZero() {
		activity.Timestamp = time.Now()
	}

	_, err := as.collection.InsertOne(ctx, activity)
	if err != nil {
		log.Printf("Error creating activity: %v", err)
		return err
	}

	log.Printf("Activity logged successfully: type=%s, userId=%s", activity.Type, activity.UserID)
	return nil
}

func (as *ActivityStore) GetByUserID(ctx context.Context, userID string, limit int) ([]*model.UserActivity, error) {
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := as.collection.Find(ctx, bson.M{"userId": userID}, opts)
	if err != nil {
		log.Printf("Error getting activities: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []*model.UserActivity
	if err = cursor.All(ctx, &activities); err != nil {
		log.Printf("Error decoding activities: %v", err)
		return nil, err
	}

	return activities, nil
}

func (as *ActivityStore) GetByUserIDAndType(ctx context.Context, userID string, activityType model.ActivityType, limit int) ([]*model.UserActivity, error) {
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	filter := bson.M{
		"userId": userID,
		"type":   activityType,
	}

	cursor, err := as.collection.Find(ctx, filter, opts)
	if err != nil {
		log.Printf("Error getting activities by type: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []*model.UserActivity
	if err = cursor.All(ctx, &activities); err != nil {
		log.Printf("Error decoding activities: %v", err)
		return nil, err
	}

	return activities, nil
}
