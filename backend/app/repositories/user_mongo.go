package repositories

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/P-wig/GuildLogger2/backend/app/db"
)

// MongoUserRepository implements UserRepository using MongoDB.
type MongoUserRepository struct {
	database *mongo.Database
}

// NewMongoUserRepository creates a new MongoDB-backed user repository.
func NewMongoUserRepository(database *mongo.Database) *MongoUserRepository {
	return &MongoUserRepository{database: database}
}

// CreateOrUpdateFromDiscord inserts or updates a user based on Discord ID (upsert pattern).
func (r *MongoUserRepository) CreateOrUpdateFromDiscord(ctx context.Context, discordID, accessToken, refreshToken string) (*User, error) {
	now := time.Now()
	user := &User{
		DiscordID:    discordID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	result := db.UsersCollection(r.database).FindOneAndUpdate(
		ctx,
		bson.M{"discordId": discordID},
		bson.M{"$set": user},
		opts,
	)

	if result.Err() != nil {
		return nil, result.Err()
	}

	if err := result.Decode(user); err != nil {
		return nil, err
	}

	return user, nil
}

// FindByDiscordID retrieves a user by their Discord ID.
func (r *MongoUserRepository) FindByDiscordID(ctx context.Context, discordID string) (*User, error) {
	var user User
	err := db.UsersCollection(r.database).FindOne(ctx, bson.M{"discordId": discordID}).Decode(&user)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil // not found, not an error
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update modifies an existing user by Discord ID.
func (r *MongoUserRepository) Update(ctx context.Context, discordID string, user *User) error {
	user.UpdatedAt = time.Now()
	result := db.UsersCollection(r.database).FindOneAndReplace(
		ctx,
		bson.M{"discordId": discordID},
		user,
	)
	return result.Err()
}

// Delete removes a user by Discord ID.
func (r *MongoUserRepository) Delete(ctx context.Context, discordID string) error {
	_, err := db.UsersCollection(r.database).DeleteOne(ctx, bson.M{"discordId": discordID})
	return err
}
