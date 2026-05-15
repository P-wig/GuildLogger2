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

// EnsureIndexes creates required indexes for user identity persistence.
func (r *MongoUserRepository) EnsureIndexes(ctx context.Context) error {
	_, err := db.UsersCollection(r.database).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "discordId", Value: 1},
		},
		Options: options.Index().
			SetUnique(true).
			SetName("uniq_discord_user"),
	})
	return err
}

// CreateOrUpdateFromDiscord inserts or updates a user based on Discord ID (upsert pattern).
func (r *MongoUserRepository) CreateOrUpdateFromDiscord(ctx context.Context, discordID, accessToken, refreshToken string) (*User, error) {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
			"updatedAt":    now,
		},
		"$setOnInsert": bson.M{
			"discordId": discordID,
			"createdAt": now,
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	result := db.UsersCollection(r.database).FindOneAndUpdate(
		ctx,
		bson.M{"discordId": discordID},
		update,
		opts,
	)
	if result.Err() != nil {
		return nil, result.Err()
	}
	var user User
	if err := result.Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
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
