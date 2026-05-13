package repositories

import (
	"bytes"
	"compress/zlib"
	"context"
	"errors"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/P-wig/GuildLogger2/backend/app/db"
)

// eventDoc is the internal MongoDB storage struct.
// Notes are stored as compressed bytes rather than a plain string.
type eventDoc struct {
	ID              string    `bson:"_id,omitempty"`
	EventID         string    `bson:"eventId"`
	GuildID         string    `bson:"guildId"`
	HostDiscordID   string    `bson:"hostDiscordId"`
	AttendingIDs    []string  `bson:"attendingIds"`
	EventDate       time.Time `bson:"eventDate"`
	NotesCompressed []byte    `bson:"notes"` // zlib-compressed notes
	CreatedAt       time.Time `bson:"createdAt"`
	UpdatedAt       time.Time `bson:"updatedAt"`
}

// compressNotes compresses a notes string using zlib.
func compressNotes(notes string) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write([]byte(notes)); err != nil {
		w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// decompressNotes decompresses zlib bytes back to a string.
func decompressNotes(data []byte) (string, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// toDoc converts an Event to the internal storage struct, compressing notes.
func toDoc(event *Event) (*eventDoc, error) {
	compressed, err := compressNotes(event.Notes)
	if err != nil {
		return nil, err
	}
	return &eventDoc{
		ID:              event.ID,
		EventID:         event.EventID,
		GuildID:         event.GuildID,
		HostDiscordID:   event.HostDiscordID,
		AttendingIDs:    event.AttendingIDs,
		EventDate:       event.EventDate,
		NotesCompressed: compressed,
		CreatedAt:       event.CreatedAt,
		UpdatedAt:       event.UpdatedAt,
	}, nil
}

// fromDoc converts the internal storage struct back to an Event, decompressing notes.
func fromDoc(doc *eventDoc) (*Event, error) {
	notes, err := decompressNotes(doc.NotesCompressed)
	if err != nil {
		return nil, err
	}
	return &Event{
		ID:            doc.ID,
		EventID:       doc.EventID,
		GuildID:       doc.GuildID,
		HostDiscordID: doc.HostDiscordID,
		AttendingIDs:  doc.AttendingIDs,
		EventDate:     doc.EventDate,
		Notes:         notes,
		CreatedAt:     doc.CreatedAt,
		UpdatedAt:     doc.UpdatedAt,
	}, nil
}

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
	doc, err := toDoc(event)
	if err != nil {
		return err
	}
	_, err = db.EventsCollection(r.database).InsertOne(ctx, doc)
	return err
}

// FindByEventID retrieves an event by its event ID.
func (r *MongoEventRepository) FindByEventID(ctx context.Context, eventID string) (*Event, error) {
	var doc eventDoc
	err := db.EventsCollection(r.database).FindOne(ctx, bson.M{"eventId": eventID}).Decode(&doc)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromDoc(&doc)
}

// FindByGuildID retrieves all events for a specific guild.
func (r *MongoEventRepository) FindByGuildID(ctx context.Context, guildID string) ([]Event, error) {
	cursor, err := db.EventsCollection(r.database).Find(ctx, bson.M{"guildId": guildID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []eventDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	events := make([]Event, 0, len(docs))
	for _, doc := range docs {
		event, err := fromDoc(&doc)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
	}
	return events, nil
}

// Update modifies an existing event.
func (r *MongoEventRepository) Update(ctx context.Context, eventID string, event *Event) error {
	event.UpdatedAt = time.Now()
	doc, err := toDoc(event)
	if err != nil {
		return err
	}
	result := db.EventsCollection(r.database).FindOneAndReplace(
		ctx,
		bson.M{"eventId": eventID},
		doc,
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
