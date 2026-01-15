package store

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"notifications-service/internal/model"
)

type NotificationRepository struct {
	collection *mongo.Collection
}

func NewNotificationRepository(db *mongo.Database) *NotificationRepository {
	return &NotificationRepository{
		collection: db.Collection("notifications"),
	}
}

func (r *NotificationRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Notification, error) {
	findOptions := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"userId": userID}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*model.Notification
	if err = cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	_, err := r.collection.InsertOne(ctx, notification)
	return err
}
