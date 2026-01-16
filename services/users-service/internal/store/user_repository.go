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
	usersCollection    *mongo.Collection
	otpsCollection     *mongo.Collection
	magicLinksCollection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		usersCollection:      db.Collection("users"),
		otpsCollection:       db.Collection("otps"),
		magicLinksCollection: db.Collection("magic_links"),
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

// Magic Link methods
func (r *UserRepository) SetMagicLink(ctx context.Context, email, token string) error {
	entry := security.MagicLinkEntry{
		Token:     token,
		Email:     email,
		ExpiresAt: time.Now().Add(15 * time.Minute), // Magic link expires in 15 minutes
	}

	// Use upsert to replace existing magic link for email
	_, err := r.magicLinksCollection.ReplaceOne(
		ctx,
		bson.M{"email": email},
		bson.M{
			"email":     entry.Email,
			"token":     entry.Token,
			"expiresAt": entry.ExpiresAt,
		},
		options.Replace().SetUpsert(true),
	)
	return err
}

func (r *UserRepository) GetMagicLink(ctx context.Context, token string) (security.MagicLinkEntry, bool) {
	var result struct {
		Email     string    `bson:"email"`
		Token     string    `bson:"token"`
		ExpiresAt time.Time `bson:"expiresAt"`
	}

	err := r.magicLinksCollection.FindOne(ctx, bson.M{"token": token}).Decode(&result)
	if err != nil {
		return security.MagicLinkEntry{}, false
	}

	return security.MagicLinkEntry{
		Token:     result.Token,
		Email:     result.Email,
		ExpiresAt: result.ExpiresAt,
	}, true
}

func (r *UserRepository) DeleteMagicLink(ctx context.Context, token string) error {
	_, err := r.magicLinksCollection.DeleteOne(ctx, bson.M{"token": token})
	return err
}

// Verification token methods for email verification
func (r *UserRepository) SetVerificationToken(ctx context.Context, email, token string) error {
	entry := security.VerificationTokenEntry{
		Token:     token,
		Email:     email,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Verification token expires in 24 hours
	}

	_, err := r.magicLinksCollection.ReplaceOne(
		ctx,
		bson.M{"email": email, "type": "verification"},
		bson.M{
			"email":     entry.Email,
			"token":     entry.Token,
			"expiresAt": entry.ExpiresAt,
			"type":      "verification",
		},
		options.Replace().SetUpsert(true),
	)
	return err
}

func (r *UserRepository) GetVerificationToken(ctx context.Context, token string) (security.VerificationTokenEntry, bool) {
	var result struct {
		Email     string    `bson:"email"`
		Token     string    `bson:"token"`
		ExpiresAt time.Time `bson:"expiresAt"`
	}

	err := r.magicLinksCollection.FindOne(ctx, bson.M{"token": token, "type": "verification"}).Decode(&result)
	if err != nil {
		return security.VerificationTokenEntry{}, false
	}

	return security.VerificationTokenEntry{
		Token:     result.Token,
		Email:     result.Email,
		ExpiresAt: result.ExpiresAt,
	}, true
}

func (r *UserRepository) DeleteVerificationToken(ctx context.Context, token string) error {
	_, err := r.magicLinksCollection.DeleteOne(ctx, bson.M{"token": token, "type": "verification"})
	return err
}

// Password reset token methods
func (r *UserRepository) SetPasswordResetToken(ctx context.Context, email, token string) error {
	entry := security.PasswordResetTokenEntry{
		Token:     token,
		Email:     email,
		ExpiresAt: time.Now().Add(1 * time.Hour), // Password reset token expires in 1 hour
	}

	_, err := r.magicLinksCollection.ReplaceOne(
		ctx,
		bson.M{"email": email, "type": "password_reset"},
		bson.M{
			"email":     entry.Email,
			"token":     entry.Token,
			"expiresAt": entry.ExpiresAt,
			"type":      "password_reset",
		},
		options.Replace().SetUpsert(true),
	)
	return err
}

func (r *UserRepository) GetPasswordResetToken(ctx context.Context, token string) (security.PasswordResetTokenEntry, bool) {
	var result struct {
		Email     string    `bson:"email"`
		Token     string    `bson:"token"`
		ExpiresAt time.Time `bson:"expiresAt"`
	}

	err := r.magicLinksCollection.FindOne(ctx, bson.M{"token": token, "type": "password_reset"}).Decode(&result)
	if err != nil {
		return security.PasswordResetTokenEntry{}, false
	}

	return security.PasswordResetTokenEntry{
		Token:     result.Token,
		Email:     result.Email,
		ExpiresAt: result.ExpiresAt,
	}, true
}

func (r *UserRepository) DeletePasswordResetToken(ctx context.Context, token string) error {
	_, err := r.magicLinksCollection.DeleteOne(ctx, bson.M{"token": token, "type": "password_reset"})
	return err
}