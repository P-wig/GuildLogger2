package db

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// Collection accessors — single source of truth for collection names.

func UsersCollection(database *mongo.Database) *mongo.Collection {
	return database.Collection("users")
}

func GuildsCollection(database *mongo.Database) *mongo.Collection {
	return database.Collection("guilds")
}

func EventsCollection(database *mongo.Database) *mongo.Collection {
	return database.Collection("events")
}

func MembersCollection(database *mongo.Database) *mongo.Collection {
	return database.Collection("members")
}

func EventReportsCollection(database *mongo.Database) *mongo.Collection {
	return database.Collection("event_reports")
}
