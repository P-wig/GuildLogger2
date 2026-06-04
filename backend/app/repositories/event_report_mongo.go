package repositories

import (
	"context"
	"errors"
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
	ID                string    `bson:"_id,omitempty"`
	EventID           string    `bson:"eventId"`
	GuildID           string    `bson:"guildId"`
	HostDiscordID     string    `bson:"hostDiscordId"`
	EventDate         time.Time `bson:"eventDate"`
	ParticipantIDs    []string  `bson:"participantIds"`
	SummaryCompressed []byte    `bson:"summary"`
	SubmittedAt       time.Time `bson:"submittedAt"`
}

func toEventReportDoc(report *EventReport) (*eventReportDoc, error) {
	compressed, err := zlibCompress(report.Summary)
	if err != nil {
		return nil, err
	}
	return &eventReportDoc{
		ID:                report.ID,
		EventID:           report.EventID,
		GuildID:           report.GuildID,
		HostDiscordID:     report.HostDiscordID,
		EventDate:         report.EventDate,
		ParticipantIDs:    report.ParticipantIDs,
		SummaryCompressed: compressed,
		SubmittedAt:       report.SubmittedAt,
	}, nil
}

func fromEventReportDoc(doc *eventReportDoc) (*EventReport, error) {
	summary, err := zlibDecompress(doc.SummaryCompressed)
	if err != nil {
		return nil, err
	}
	return &EventReport{
		ID:             doc.ID,
		EventID:        doc.EventID,
		GuildID:        doc.GuildID,
		HostDiscordID:  doc.HostDiscordID,
		EventDate:      doc.EventDate,
		ParticipantIDs: doc.ParticipantIDs,
		Summary:        summary,
		SubmittedAt:    doc.SubmittedAt,
	}, nil
}

type MongoEventReportRepository struct {
	database *mongo.Database
}

func NewMongoEventReportRepository(database *mongo.Database) *MongoEventReportRepository {
	return &MongoEventReportRepository{database: database}
}

func (r *MongoEventReportRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "eventId", Value: 1}},
			Options: options.Index().SetName("idx_report_event").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "guildId", Value: 1}, {Key: "eventDate", Value: -1}},
			Options: options.Index().SetName("idx_report_guild_date"),
		},
	}
	_, err := db.EventReportsCollection(r.database).Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *MongoEventReportRepository) Create(ctx context.Context, report *EventReport) error {
	report.ID = primitive.NewObjectID().Hex()
	report.SubmittedAt = time.Now()
	doc, err := toEventReportDoc(report)
	if err != nil {
		return err
	}
	_, err = db.EventReportsCollection(r.database).InsertOne(ctx, doc)
	return err
}

func (r *MongoEventReportRepository) FindByEventID(ctx context.Context, eventID string) (*EventReport, error) {
	var doc eventReportDoc
	err := db.EventReportsCollection(r.database).FindOne(ctx, bson.M{"eventId": eventID}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromEventReportDoc(&doc)
}

func (r *MongoEventReportRepository) FindByGuildID(ctx context.Context, guildID string) ([]EventReport, error) {
	opts := options.Find().SetSort(bson.D{{Key: "eventDate", Value: -1}})
	cursor, err := db.EventReportsCollection(r.database).Find(ctx, bson.M{"guildId": guildID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []eventReportDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	reports := make([]EventReport, 0, len(docs))
	for _, doc := range docs {
		report, err := fromEventReportDoc(&doc)
		if err != nil {
			return nil, err
		}
		reports = append(reports, *report)
	}
	return reports, nil
}
