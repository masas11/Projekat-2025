package store

import (
	"context"
	"log"
	"time"

	"saga-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SagaStore struct {
	collection *mongo.Collection
}

func NewSagaStore(db *mongo.Database) *SagaStore {
	return &SagaStore{
		collection: db.Collection("saga_transactions"),
	}
}

func (s *SagaStore) CreateTransaction(ctx context.Context, saga *model.SagaTransaction) error {
	if saga.ID == "" {
		saga.ID = primitive.NewObjectID().Hex()
	}
	saga.CreatedAt = time.Now()
	saga.UpdatedAt = time.Now()

	_, err := s.collection.InsertOne(ctx, saga)
	if err != nil {
		log.Printf("Error creating saga transaction: %v", err)
		return err
	}
	return nil
}

func (s *SagaStore) GetTransaction(ctx context.Context, id string) (*model.SagaTransaction, error) {
	var saga model.SagaTransaction
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&saga)
	if err != nil {
		return nil, err
	}
	return &saga, nil
}

func (s *SagaStore) UpdateTransaction(ctx context.Context, saga *model.SagaTransaction) error {
	saga.UpdatedAt = time.Now()
	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{"_id": saga.ID},
		bson.M{"$set": saga},
	)
	if err != nil {
		log.Printf("Error updating saga transaction: %v", err)
		return err
	}
	return nil
}

func (s *SagaStore) UpdateStepStatus(ctx context.Context, sagaID string, stepName string, status model.StepStatus, errorMsg string) error {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"updatedAt": now,
			"steps.$[step].status": status,
		},
	}

	arrayFilters := []interface{}{
		bson.M{"step.name": stepName},
	}

	if status == model.StepStatusCompleted {
		update["$set"].(bson.M)["steps.$[step].executedAt"] = &now
	} else if status == model.StepStatusCompensated {
		update["$set"].(bson.M)["steps.$[step].compensatedAt"] = &now
	}

	if errorMsg != "" {
		update["$set"].(bson.M)["steps.$[step].error"] = errorMsg
	}

	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{"_id": sagaID},
		update,
		options.Update().SetArrayFilters(options.ArrayFilters{Filters: arrayFilters}),
	)
	return err
}
