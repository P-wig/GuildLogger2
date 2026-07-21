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
	EventID           string    `bson:"eventId,omitempty"`
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
			Options: options.Index().SetName("idx_report_event_sparse").SetUnique(true).SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "guildId", Value: 1}, {Key: "eventDate", Value: -1}},
			Options: options.Index().SetName("idx_report_guild_date"),
		},
	}
	_, err := db.EventReportsCollection(r.database).Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *MongoEventReportRepository) Update(ctx context.Context, logID string, report *EventReport) error {
	compressed, err := zlibCompress(report.Summary)
	if err != nil {
		return err
	}
	result, err := db.EventReportsCollection(r.database).UpdateOne(ctx,
		bson.M{"_id": logID},
		bson.M{"$set": bson.M{
			"hostDiscordId":  report.HostDiscordID,
			"eventDate":      report.EventDate,
			"participantIds": report.ParticipantIDs,
			"summary":        compressed,
		}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrReportNotFound
	}
	return nil
}

func (r *MongoEventReportRepository) Delete(ctx context.Context, logID string) error {
	result, err := db.EventReportsCollection(r.database).DeleteOne(ctx, bson.M{"_id": logID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrReportNotFound
	}
	return nil
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

func (r *MongoEventReportRepository) GetGuildClosedEventCount(ctx context.Context, guildID string) (int64, error) {
	count, err := db.EventReportsCollection(r.database).CountDocuments(ctx, bson.M{"guildId": guildID})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *MongoEventReportRepository) GetGuildMemberActivity(ctx context.Context, guildID string) ([]GuildDashboardLeaderboardEntry, error) {
	reports, err := r.FindByGuildID(ctx, guildID)
	if err != nil {
		return nil, err
	}

	byMember := make(map[string]*GuildDashboardLeaderboardEntry)
	for _, report := range reports {
		hostID := report.HostDiscordID
		if hostID != "" {
			entry, ok := byMember[hostID]
			if !ok {
				entry = &GuildDashboardLeaderboardEntry{DiscordID: hostID}
				byMember[hostID] = entry
			}
			entry.HostedCount++
			if entry.LastHostedDate == nil || report.EventDate.After(*entry.LastHostedDate) {
				t := report.EventDate
				entry.LastHostedDate = &t
			}
		}

		for _, attendeeID := range report.ParticipantIDs {
			if attendeeID == "" {
				continue
			}
			entry, ok := byMember[attendeeID]
			if !ok {
				entry = &GuildDashboardLeaderboardEntry{DiscordID: attendeeID}
				byMember[attendeeID] = entry
			}
			entry.AttendedCount++
			if entry.LastAttendedDate == nil || report.EventDate.After(*entry.LastAttendedDate) {
				t := report.EventDate
				entry.LastAttendedDate = &t
			}
		}
	}

	out := make([]GuildDashboardLeaderboardEntry, 0, len(byMember))
	for _, entry := range byMember {
		entry.Score = (entry.HostedCount * 2) + entry.AttendedCount
		out = append(out, *entry)
	}
	return out, nil
}

func (r *MongoEventReportRepository) GetGuildParticipationStats(ctx context.Context, guildID string) (int64, int64, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"guildId": guildID}},
		bson.M{"$group": bson.M{
			"_id": bson.M{"$const": 1},
			"participantSlots": bson.M{"$sum": bson.M{
				"$size": bson.M{"$ifNull": bson.A{"$participantIds", bson.A{}}},
			}},
			"uniqueReportedEvents": bson.M{"$sum": 1},
		}},
	}

	cursor, err := db.EventReportsCollection(r.database).Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		ParticipantSlots     int64 `bson:"participantSlots"`
		UniqueReportedEvents int64 `bson:"uniqueReportedEvents"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, 0, err
		}
	}

	return result.ParticipantSlots, result.UniqueReportedEvents, nil
}

func (r *MongoEventReportRepository) FindDashboardEvents(ctx context.Context, guildID string, filter GuildDashboardEventFilter) ([]GuildDashboardEventRow, error) {
	mongoFilter := bson.M{"guildId": guildID}

	if filter.AttendeeID != "" {
		mongoFilter["participantIds"] = filter.AttendeeID
	}

	if filter.StartDate != nil || filter.EndDate != nil {
		dateFilter := bson.M{}
		if filter.StartDate != nil {
			dateFilter["$gte"] = *filter.StartDate
		}
		if filter.EndDate != nil {
			dateFilter["$lte"] = *filter.EndDate
		}
		mongoFilter["eventDate"] = dateFilter
	}

	limit := int64(filter.Limit)
	if limit <= 0 {
		limit = 100
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "eventDate", Value: -1}}).
		SetLimit(limit).
		SetProjection(bson.M{
			"eventId":        1,
			"hostDiscordId":  1,
			"eventDate":      1,
			"participantIds": 1,
			"summary":        1,
		})

	cursor, err := db.EventReportsCollection(r.database).Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	rows := make([]GuildDashboardEventRow, 0)
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	return rows, nil
}
