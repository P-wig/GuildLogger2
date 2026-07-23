package repositories

import (
	"context"
	"time"
)

// User represents a user document in MongoDB, authenticated via Discord OAuth.
type User struct {
	ID           string    `bson:"_id,omitempty"   json:"-"`
	DiscordID    string    `bson:"discordId"       json:"discordId"` // Primary identity
	AccessToken  string    `bson:"accessToken"     json:"-"`         // For making Discord API calls
	RefreshToken string    `bson:"refreshToken"    json:"-"`         // Token refresh
	CreatedAt    time.Time `bson:"createdAt"       json:"createdAt"` // When they joined your app
	UpdatedAt    time.Time `bson:"updatedAt"       json:"updatedAt"` // Last activity
}

// UserRepository defines the contract for user data access operations.
type UserRepository interface {
	// CreateOrUpdateFromDiscord inserts a new user or updates an existing one
	// based on Discord ID. Called during OAuth login/registration.
	CreateOrUpdateFromDiscord(ctx context.Context, discordID, accessToken, refreshToken string) (*User, error)

	// FindByDiscordID retrieves a user by their Discord ID.
	// Returns (nil, nil) if not found; only returns error on DB failure.
	FindByDiscordID(ctx context.Context, discordID string) (*User, error)

	// Update modifies an existing user by Discord ID.
	Update(ctx context.Context, discordID string, user *User) error

	// Delete removes a user by Discord ID.
	Delete(ctx context.Context, discordID string) error
}
