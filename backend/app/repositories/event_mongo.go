package repositories

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/P-wig/GuildLogger2/backend/app/db"
)

// MongoEventRepository implements EventRepository using MongoDB.
type MongoEventRepository struct {
	database *mongo.Database
}

// NewMongoEventRepository creates a new MongoDB-backed event repository.
func NewMongoEventRepository(database *mongo.Database) *MongoEventRepository {
	return &MongoEventRepository{database: database}
}

// Create inserts a new event.
func (r *MongoEventRepository) Create(ctx context.Context, event *Event) error {
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()
	_, err := db.EventsCollection(r.database).InsertOne(ctx, event)
	return err
}

// FindByEventID retrieves an event by its event ID.
func (r *MongoEventRepository) FindByEventID(ctx context.Context, eventID string) (*Event, error) {
	var event Event
	err := db.EventsCollection(r.database).FindOne(ctx, bson.M{"eventId": eventID}).Decode(&event)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// FindByGuildID retrieves all events for a specific guild.
func (r *MongoEventRepository) FindByGuildID(ctx context.Context, guildID string) ([]Event, error) {
	cursor, err := db.EventsCollection(r.database).Find(ctx, bson.M{"guildId": guildID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []Event
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// Update modifies an existing event.
func (r *MongoEventRepository) Update(ctx context.Context, eventID string, event *Event) error {
	event.UpdatedAt = time.Now()
	result := db.EventsCollection(r.database).FindOneAndReplace(
		ctx,
		bson.M{"eventId": eventID},
		event,
	)
	return result.Err()
}

// AddAttendee registers a member to attend the event (idempotent).
func (r *MongoEventRepository) AddAttendee(ctx context.Context, eventID, discordID string) error {
	_, err := db.EventsCollection(r.database).UpdateOne(
		ctx,
		bson.M{"eventId": eventID},
		bson.M{
			"$addToSet": bson.M{"attendingIds": discordID}, // $addToSet prevents duplicates
			"$set":      bson.M{"updatedAt": time.Now()},
		},
	)
	return err
}

// RemoveAttendee unregisters a member from attending (idempotent).
func (r *MongoEventRepository) RemoveAttendee(ctx context.Context, eventID, discordID string) error {
	_, err := db.EventsCollection(r.database).UpdateOne(
		ctx,
		bson.M{"eventId": eventID},
		bson.M{
			"$pull": bson.M{"attendingIds": discordID}, // $pull removes the element
			"$set":  bson.M{"updatedAt": time.Now()},
		},
	)
	return err
}

// Delete removes an event entirely.
func (r *MongoEventRepository) Delete(ctx context.Context, eventID string) error {
	_, err := db.EventsCollection(r.database).DeleteOne(ctx, bson.M{"eventId": eventID})
	return err
}
