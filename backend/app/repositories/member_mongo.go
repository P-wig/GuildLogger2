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

	member.JoinedAt = now
	member.UpdatedAt = now
	if member.RankedRoleID == "" {
		member.RankedRoleID = ""
	}
	member.NotificationsOptOut = member.NotificationsOptOut // explicit no-op for clarity

	_, err := db.MembersCollection(r.database).InsertOne(ctx, member)
	return err
}

func (r *MongoMemberRepository) Upsert(ctx context.Context, member *Member) error {
	now := time.Now()
	if member.RoleIDs == nil {
		member.RoleIDs = []string{}
	}

	_, err := db.MembersCollection(r.database).UpdateOne(
		ctx,
		bson.M{
			"guildId":   member.GuildID,
			"discordId": member.DiscordID,
		},
		bson.M{
			"$set": bson.M{
				"roleIds":   member.RoleIDs,
				"updatedAt": now,
			},
			"$setOnInsert": bson.M{
				"guildId":   member.GuildID,
				"discordId": member.DiscordID,
				"joinedAt":  now,
			},
		},
		options.Update().SetUpsert(true),
	)

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

func (r *MongoMemberRepository) GetEventStats(ctx context.Context, guildID, discordID string) (*MemberEventStats, error) {
	hostedCount, err := db.EventsCollection(r.database).CountDocuments(
		ctx,
		bson.M{
			"guildId":       guildID,
			"hostDiscordId": discordID,
		},
	)
	if err != nil {
		return nil, err
	}

	participatedCount, err := db.EventsCollection(r.database).CountDocuments(
		ctx,
		bson.M{
			"guildId":      guildID,
			"attendingIds": discordID, // matches when discordID exists in array
		},
	)
	if err != nil {
		return nil, err
	}

	return &MemberEventStats{
		HostedCount:       hostedCount,
		ParticipatedCount: participatedCount,
	}, nil
}
