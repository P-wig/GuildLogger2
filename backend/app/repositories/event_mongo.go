package repositories

import (
	"bytes"
	"compress/zlib"
	"context"
	"errors"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/P-wig/GuildLogger2/backend/app/db"
)

// eventReportDoc is the storage representation of EventReport.
// Summary is stored as zlib-compressed bytes.
type eventReportDoc struct {
	SummaryCompressed []byte    `bson:"summary"`
	ParticipantIDs    []string  `bson:"participantIds"`
	SubmittedAt       time.Time `bson:"submittedAt"`
}

// eventDoc is the internal MongoDB storage struct.
// Description and report summary are stored as zlib-compressed bytes.
type eventDoc struct {
	ID                    string          `bson:"_id,omitempty"`
	GuildID               string          `bson:"guildId"`
	HostDiscordID         string          `bson:"hostDiscordId"`
	Title                 string          `bson:"title"`
	DescriptionCompressed []byte          `bson:"description"`
	AttendingIDs          []string        `bson:"attendingIds"`
	ScheduledAt           time.Time       `bson:"scheduledAt"`
	Status                EventStatus     `bson:"status"`
	ChannelID             string          `bson:"channelId"`
	ReminderEnabled       bool            `bson:"reminderEnabled"`
	Capacity              int             `bson:"capacity"`
	CutoffAt              *time.Time      `bson:"cutoffAt,omitempty"`
	ReminderSentAt        *time.Time      `bson:"reminderSentAt,omitempty"`
	StartedAt             *time.Time      `bson:"startedAt,omitempty"`
	ClosedAt              *time.Time      `bson:"closedAt,omitempty"`
	Report                *eventReportDoc `bson:"report,omitempty"`
	CreatedAt             time.Time       `bson:"createdAt"`
	UpdatedAt             time.Time       `bson:"updatedAt"`
}

func zlibCompress(s string) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write([]byte(s)); err != nil {
		w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func zlibDecompress(data []byte) (string, error) {
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

func toReportDoc(report *EventReport) (*eventReportDoc, error) {
	if report == nil {
		return nil, nil
	}
	compressed, err := zlibCompress(report.Summary)
	if err != nil {
		return nil, err
	}
	return &eventReportDoc{
		SummaryCompressed: compressed,
		ParticipantIDs:    report.ParticipantIDs,
		SubmittedAt:       report.SubmittedAt,
	}, nil
}

func fromReportDoc(doc *eventReportDoc) (*EventReport, error) {
	if doc == nil {
		return nil, nil
	}
	summary, err := zlibDecompress(doc.SummaryCompressed)
	if err != nil {
		return nil, err
	}
	return &EventReport{
		Summary:        summary,
		ParticipantIDs: doc.ParticipantIDs,
		SubmittedAt:    doc.SubmittedAt,
	}, nil
}

func toEventDoc(event *Event) (*eventDoc, error) {
	description, err := zlibCompress(event.Description)
	if err != nil {
		return nil, err
	}
	reportDoc, err := toReportDoc(event.Report)
	if err != nil {
		return nil, err
	}
	return &eventDoc{
		ID:                    event.ID,
		GuildID:               event.GuildID,
		HostDiscordID:         event.HostDiscordID,
		Title:                 event.Title,
		DescriptionCompressed: description,
		AttendingIDs:          event.AttendingIDs,
		ScheduledAt:           event.ScheduledAt,
		Status:                event.Status,
		ChannelID:             event.ChannelID,
		ReminderEnabled:       event.ReminderEnabled,
		Capacity:              event.Capacity,
		CutoffAt:              event.CutoffAt,
		ReminderSentAt:        event.ReminderSentAt,
		StartedAt:             event.StartedAt,
		ClosedAt:              event.ClosedAt,
		Report:                reportDoc,
		CreatedAt:             event.CreatedAt,
		UpdatedAt:             event.UpdatedAt,
	}, nil
}

func fromEventDoc(doc *eventDoc) (*Event, error) {
	description, err := zlibDecompress(doc.DescriptionCompressed)
	if err != nil {
		return nil, err
	}
	report, err := fromReportDoc(doc.Report)
	if err != nil {
		return nil, err
	}
	return &Event{
		ID:              doc.ID,
		GuildID:         doc.GuildID,
		HostDiscordID:   doc.HostDiscordID,
		Title:           doc.Title,
		Description:     description,
		AttendingIDs:    doc.AttendingIDs,
		ScheduledAt:     doc.ScheduledAt,
		Status:          doc.Status,
		ChannelID:       doc.ChannelID,
		ReminderEnabled: doc.ReminderEnabled,
		Capacity:        doc.Capacity,
		CutoffAt:        doc.CutoffAt,
		ReminderSentAt:  doc.ReminderSentAt,
		StartedAt:       doc.StartedAt,
		ClosedAt:        doc.ClosedAt,
		Report:          report,
		CreatedAt:       doc.CreatedAt,
		UpdatedAt:       doc.UpdatedAt,
	}, nil
}

// MongoEventRepository implements EventRepository using MongoDB.
type MongoEventRepository struct {
	database *mongo.Database
}

func NewMongoEventRepository(database *mongo.Database) *MongoEventRepository {
	return &MongoEventRepository{database: database}
}

func (r *MongoEventRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "guildId", Value: 1}},
			Options: options.Index().SetName("idx_event_guild"),
		},
		{
			Keys:    bson.D{{Key: "guildId", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("idx_event_guild_status"),
		},
		{
			Keys:    bson.D{{Key: "scheduledAt", Value: 1}},
			Options: options.Index().SetName("idx_event_scheduled_at"),
		},
	}

	_, err := db.EventsCollection(r.database).Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *MongoEventRepository) Create(ctx context.Context, event *Event) error {
	now := time.Now()
	event.ID = primitive.NewObjectID().Hex()
	event.CreatedAt = now
	event.UpdatedAt = now
	if event.Status == "" {
		event.Status = EventStatusOpen
	}
	if event.AttendingIDs == nil {
		event.AttendingIDs = []string{}
	}

	doc, err := toEventDoc(event)
	if err != nil {
		return err
	}
	_, err = db.EventsCollection(r.database).InsertOne(ctx, doc)
	return err
}

func (r *MongoEventRepository) FindByID(ctx context.Context, eventID string) (*Event, error) {
	var doc eventDoc
	err := db.EventsCollection(r.database).FindOne(ctx, bson.M{"_id": eventID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromEventDoc(&doc)
}

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
		event, err := fromEventDoc(&doc)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
	}
	return events, nil
}

func (r *MongoEventRepository) Update(ctx context.Context, eventID string, event *Event) error {
	event.UpdatedAt = time.Now()
	doc, err := toEventDoc(event)
	if err != nil {
		return err
	}
	result := db.EventsCollection(r.database).FindOneAndReplace(ctx, bson.M{"_id": eventID}, doc)
	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return ErrEventNotFound
	}
	return result.Err()
}

func (r *MongoEventRepository) Delete(ctx context.Context, eventID string) error {
	_, err := db.EventsCollection(r.database).DeleteOne(ctx, bson.M{"_id": eventID})
	return err
}

func (r *MongoEventRepository) AddAttendee(ctx context.Context, eventID, discordID string) error {
	result, err := db.EventsCollection(r.database).UpdateOne(
		ctx,
		bson.M{"_id": eventID},
		bson.M{
			"$addToSet": bson.M{"attendingIds": discordID},
			"$set":      bson.M{"updatedAt": time.Now()},
		},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrEventNotFound
	}
	if result.ModifiedCount == 0 {
		return ErrAlreadyRegistered
	}
	return nil
}

func (r *MongoEventRepository) RemoveAttendee(ctx context.Context, eventID, discordID string) error {
	result, err := db.EventsCollection(r.database).UpdateOne(
		ctx,
		bson.M{"_id": eventID},
		bson.M{
			"$pull": bson.M{"attendingIds": discordID},
			"$set":  bson.M{"updatedAt": time.Now()},
		},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrEventNotFound
	}
	if result.ModifiedCount == 0 {
		return ErrNotRegistered
	}
	return nil
}
