package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/P-wig/GuildLogger2/backend/app/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func normalizeMemberDefaults(member *Member) {
	if member.RoleIDs == nil {
		member.RoleIDs = []string{}
	}
	if member.Status == "" {
		member.Status = MemberStatusActive
	}
}

type MongoMemberRepository struct {
	database *mongo.Database
}

func NewMongoMemberRepository(database *mongo.Database) *MongoMemberRepository {
	return &MongoMemberRepository{database: database}
}

func (r *MongoMemberRepository) EnsureIndexes(ctx context.Context) error {
	_, err := db.MembersCollection(r.database).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "guildId", Value: 1},
			{Key: "discordId", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("uniq_guild_member"),
	})
	return err
}

func (r *MongoMemberRepository) Create(ctx context.Context, member *Member) error {
	now := time.Now()
	normalizeMemberDefaults(member)

	member.FirstSyncedAt = now
	member.LastSyncedAt = now
	member.UpdatedAt = now

	_, err := db.MembersCollection(r.database).InsertOne(ctx, member)
	return err
}

func (r *MongoMemberRepository) Upsert(ctx context.Context, member *Member) error {
	now := time.Now()
	normalizeMemberDefaults(member)

	_, err := db.MembersCollection(r.database).UpdateOne(
		ctx,
		bson.M{
			"guildId":   member.GuildID,
			"discordId": member.DiscordID,
		},
		bson.M{
			"$set": bson.M{
				"roleIds":      member.RoleIDs,
				"status":       member.Status,
				"rankedRoleId": member.RankedRoleID,
				"username":     member.Username,
				"avatarHash":   member.AvatarHash,
				"lastSyncedAt": now,
				"updatedAt":    now,
			},
			"$setOnInsert": bson.M{
				"guildId":         member.GuildID,
				"discordId":       member.DiscordID,
				"discordJoinedAt": member.DiscordJoinedAt,
				"firstSyncedAt":   now,
			},
		},
		options.Update().SetUpsert(true),
	)

	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}

func (r *MongoMemberRepository) FindByGuildAndDiscordID(ctx context.Context, guildID, discordID string) (*Member, error) {
	var member Member
	err := db.MembersCollection(r.database).FindOne(
		ctx,
		bson.M{"guildId": guildID, "discordId": discordID},
	).Decode(&member)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *MongoMemberRepository) FindByGuildID(ctx context.Context, guildID string) ([]Member, error) {
	cursor, err := db.MembersCollection(r.database).Find(ctx, bson.M{"guildId": guildID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []Member
	if err := cursor.All(ctx, &members); err != nil {
		return nil, err
	}
	return members, nil
}

func (r *MongoMemberRepository) UpdateRoles(ctx context.Context, guildID, discordID string, roleIDs []string) error {
	if roleIDs == nil {
		roleIDs = []string{}
	}

	_, err := db.MembersCollection(r.database).UpdateOne(
		ctx,
		bson.M{"guildId": guildID, "discordId": discordID},
		bson.M{
			"$set": bson.M{
				"roleIds":   roleIDs,
				"updatedAt": time.Now(),
			},
		},
	)
	return err
}

func (r *MongoMemberRepository) Delete(ctx context.Context, guildID, discordID string) error {
	_, err := db.MembersCollection(r.database).DeleteOne(
		ctx,
		bson.M{"guildId": guildID, "discordId": discordID},
	)
	return err
}

func (r *MongoMemberRepository) DeleteAllByGuildID(ctx context.Context, guildID string) error {
	_, err := db.MembersCollection(r.database).DeleteMany(
		ctx,
		bson.M{"guildId": guildID},
	)
	return err
}

func (r *MongoMemberRepository) GetStats(ctx context.Context, guildID, discordID string) (*MemberStats, error) {
	member, err := r.FindByGuildAndDiscordID(ctx, guildID, discordID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, nil
	}

	pipeline := bson.A{
		bson.M{"$match": bson.M{
			"guildId": guildID,
		}},
		bson.M{"$group": bson.M{
			"_id": nil,
			"hostedCount": bson.M{"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$hostDiscordId", discordID}},
					1, 0,
				},
			}},
			"participatedCount": bson.M{"$sum": bson.M{
				"$cond": bson.A{
					bson.M{"$in": bson.A{discordID, "$participantIds"}},
					1, 0,
				},
			}},
		}},
	}

	cursor, err := db.EventReportsCollection(r.database).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result struct {
		HostedCount       int64 `bson:"hostedCount"`
		ParticipatedCount int64 `bson:"participatedCount"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
	}

	return &MemberStats{
		HostedCount:       result.HostedCount,
		ParticipatedCount: result.ParticipatedCount,
		DiscordJoinedAt:   member.DiscordJoinedAt,
		FirstSyncedAt:     member.FirstSyncedAt,
		DeactivatedAt:     member.DeactivatedAt,
	}, nil
}

func (r *MongoMemberRepository) UpdateStatusAndRank(ctx context.Context, guildID, discordID string, status MemberStatus, rankedRoleID string) error {
	now := time.Now()

	fields := bson.M{
		"status":       status,
		"rankedRoleId": rankedRoleID,
		"updatedAt":    now,
	}

	if status == MemberStatusInactive {
		fields["deactivatedAt"] = now
	} else {
		fields["deactivatedAt"] = nil
	}

	_, err := db.MembersCollection(r.database).UpdateOne(
		ctx,
		bson.M{"guildId": guildID, "discordId": discordID},
		bson.M{"$set": fields},
	)
	return err
}

func (r *MongoMemberRepository) UpdateNotificationPreference(ctx context.Context, guildID, discordID string, optOut bool) error {
	_, err := db.MembersCollection(r.database).UpdateOne(
		ctx,
		bson.M{"guildId": guildID, "discordId": discordID},
		bson.M{"$set": bson.M{
			"notificationsOptOut": optOut,
			"updatedAt":           time.Now(),
		}},
	)
	return err
}

func (r *MongoMemberRepository) FindNotificationTargets(ctx context.Context, guildID string) ([]Member, error) {
	cursor, err := db.MembersCollection(r.database).Find(ctx, bson.M{
		"guildId":             guildID,
		"notificationsOptOut": false,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []Member
	if err := cursor.All(ctx, &members); err != nil {
		return nil, err
	}
	return members, nil
}

func (r *MongoMemberRepository) FindAnniversaryMembers(ctx context.Context, guildID string, anniversaryYears []int) ([]Member, error) {
	if len(anniversaryYears) == 0 {
		return []Member{}, nil
	}

	today := time.Now().UTC()

	// For each configured year, compute the exact join date window that represents
	// that anniversary today. AddDate handles Feb 29 → Feb 28 automatically.
	dateRanges := make([]bson.M, 0, len(anniversaryYears))
	for _, years := range anniversaryYears {
		target := today.AddDate(-years, 0, 0)
		dayStart := time.Date(target.Year(), target.Month(), target.Day(), 0, 0, 0, 0, time.UTC)
		dayEnd := dayStart.Add(24 * time.Hour)
		dateRanges = append(dateRanges, bson.M{
			"discordJoinedAt": bson.M{"$gte": dayStart, "$lt": dayEnd},
		})
	}

	cursor, err := db.MembersCollection(r.database).Find(ctx, bson.M{
		"guildId":             guildID,
		"notificationsOptOut": false,
		"$or":                 dateRanges,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []Member
	if err := cursor.All(ctx, &members); err != nil {
		return nil, err
	}
	return members, nil
}

func (r *MongoMemberRepository) FindSummariesByGuildID(ctx context.Context, guildID string) ([]MemberSummary, error) {
	opts := options.Find().SetProjection(bson.M{
		"discordId":       1,
		"username":        1,
		"avatarHash":      1,
		"rankedRoleId":    1,
		"status":          1,
		"discordJoinedAt": 1,
		"roleIds":         1,
	})

	cursor, err := db.MembersCollection(r.database).Find(ctx, bson.M{"guildId": guildID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var summaries []MemberSummary
	if err := cursor.All(ctx, &summaries); err != nil {
		return nil, err
	}
	return summaries, nil
}

func (r *MongoMemberRepository) GetGuildMemberCounts(ctx context.Context, guildID string) (*GuildDashboardStats, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"guildId": guildID}},
		bson.M{"$group": bson.M{
			"_id":          nil,
			"totalMembers": bson.M{"$sum": 1},
			"activeMembers": bson.M{"$sum": bson.M{
				"$cond": bson.A{bson.M{"$eq": bson.A{"$status", MemberStatusActive}}, 1, 0},
			}},
			"inactiveMembers": bson.M{"$sum": bson.M{
				"$cond": bson.A{bson.M{"$eq": bson.A{"$status", MemberStatusInactive}}, 1, 0},
			}},
		}},
	}

	cursor, err := db.MembersCollection(r.database).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	stats := &GuildDashboardStats{}
	if cursor.Next(ctx) {
		if err := cursor.Decode(stats); err != nil {
			return nil, err
		}
	}
	return stats, nil
}

func (r *MongoMemberRepository) GetDashboardMemberRows(
	ctx context.Context,
	guildID string,
	inactivityDays int,
) ([]GuildDashboardMemberRow, []GuildDashboardInactiveMember, error) {
	summaries, err := r.FindSummariesByGuildID(ctx, guildID)
	if err != nil {
		return nil, nil, err
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -inactivityDays)
	rows := make([]GuildDashboardMemberRow, 0, len(summaries))
	inactive := make([]GuildDashboardInactiveMember, 0)

	for _, m := range summaries {
		row := GuildDashboardMemberRow{
			DiscordID:       m.DiscordID,
			Username:        m.Username,
			AvatarHash:      m.AvatarHash,
			RankedRoleID:    m.RankedRoleID,
			Status:          m.Status,
			DiscordJoinedAt: m.DiscordJoinedAt,
			RoleIDs:         m.RoleIDs,
		}

		rows = append(rows, row)

		if m.Status == MemberStatusInactive {
			lastActivity := m.DiscordJoinedAt
			if lastActivity.Before(cutoff) {
				days := int64(time.Since(lastActivity).Hours() / 24)
				activityCopy := lastActivity
				inactive = append(inactive, GuildDashboardInactiveMember{
					DiscordID:         m.DiscordID,
					RankedRoleID:      m.RankedRoleID,
					Status:            m.Status,
					DiscordJoinedAt:   m.DiscordJoinedAt,
					LastActivityDate:  &activityCopy,
					DaysSinceActivity: &days,
				})
			}
		}
	}

	return rows, inactive, nil
}
