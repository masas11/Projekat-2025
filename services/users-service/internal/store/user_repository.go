package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"users-service/internal/model"
	"users-service/internal/security"
)

var ErrUserExists = errors.New("user already exists")

type UserRepository struct {
	usersCollection *mongo.Collection
	otpsCollection  *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		usersCollection: db.Collection("users"),
		otpsCollection:  db.Collection("otps"),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	// Check if user with same username or email already exists
	var existingUser model.User
	err := r.usersCollection.FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"username": user.Username},
			{"email": user.Email},
		},
	}).Decode(&existingUser)

	if err == nil {
		return ErrUserExists
	}
	if err != mongo.ErrNoDocuments {
		return err
	}

	// Set ID if not set
	if user.ID == "" {
		user.ID = uuid.NewString()
	}

	_, err = r.usersCollection.InsertOne(ctx, user)
	return err
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.usersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.usersCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.usersCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	update := bson.M{
		"$set": bson.M{
			"passwordHash":        user.PasswordHash,
			"passwordChangedAt":  user.PasswordChangedAt,
			"passwordExpiresAt":  user.PasswordExpiresAt,
			"failedLoginAttempts": user.FailedLoginAttempts,
			"lockedUntil":        user.LockedUntil,
			"verified":           user.Verified,
		},
	}

	result, err := r.usersCollection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepository) SetOTP(ctx context.Context, username, code string) error {
	entry := security.OTPEntry{
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	// Use upsert to replace existing OTP for username
	_, err := r.otpsCollection.ReplaceOne(
		ctx,
		bson.M{"username": username},
		bson.M{
			"username":  username,
			"code":      entry.Code,
			"expiresAt": entry.ExpiresAt,
		},
		options.Replace().SetUpsert(true),
	)
	return err
}

func (r *UserRepository) GetOTP(ctx context.Context, username string) (security.OTPEntry, bool) {
	var result struct {
		Username  string    `bson:"username"`
		Code      string    `bson:"code"`
		ExpiresAt time.Time `bson:"expiresAt"`
	}

	err := r.otpsCollection.FindOne(ctx, bson.M{"username": username}).Decode(&result)
	if err != nil {
		return security.OTPEntry{}, false
	}

	return security.OTPEntry{
		Code:      result.Code,
		ExpiresAt: result.ExpiresAt,
	}, true
}

func (r *UserRepository) DeleteOTP(ctx context.Context, username string) error {
	_, err := r.otpsCollection.DeleteOne(ctx, bson.M{"username": username})
	return err
}
