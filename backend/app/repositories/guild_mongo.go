package repositories

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/P-wig/GuildLogger2/backend/app/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoGuildRepository implements GuildRepository using MongoDB.
type MongoGuildRepository struct {
	database *mongo.Database
}

func NewMongoGuildRepository(database *mongo.Database) *MongoGuildRepository {
	return &MongoGuildRepository{database: database}
}

// EnsureIndexes creates required indexes for guild identity/config persistence.
func (r *MongoGuildRepository) EnsureIndexes(ctx context.Context) error {
	_, err := db.GuildsCollection(r.database).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "guildId", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("uniq_guild_id"),
	})
	return err
}

func (r *MongoGuildRepository) Create(ctx context.Context, guild *Guild) error {
	now := time.Now()
	guild.CreatedAt = now
	guild.UpdatedAt = now

	if guild.Roles == nil {
		guild.Roles = []GuildRole{}
	}
	if guild.NotificationConfig.MilestoneNotifications.AnniversaryYears == nil {
		guild.NotificationConfig.MilestoneNotifications.AnniversaryYears = []int{}
	}
	if guild.NotificationConfig.MilestoneNotifications.HostedEventCounts == nil {
		guild.NotificationConfig.MilestoneNotifications.HostedEventCounts = []int{}
	}
	if guild.NotificationConfig.MilestoneNotifications.AttendedEventCounts == nil {
		guild.NotificationConfig.MilestoneNotifications.AttendedEventCounts = []int{}
	}

	_, err := db.GuildsCollection(r.database).InsertOne(ctx, guild)
	if mongo.IsDuplicateKeyError(err) {
		return ErrGuildAlreadyExists
	}
	return err
}

func (r *MongoGuildRepository) FindByGuildID(ctx context.Context, guildID string) (*Guild, error) {
	var guild Guild
	err := db.GuildsCollection(r.database).FindOne(ctx, bson.M{"guildId": guildID}).Decode(&guild)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &guild, nil
}

func (r *MongoGuildRepository) FindByOwnerDiscordID(ctx context.Context, ownerDiscordID string) ([]Guild, error) {
	cursor, err := db.GuildsCollection(r.database).Find(ctx, bson.M{"ownerDiscordId": ownerDiscordID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var guilds []Guild
	if err := cursor.All(ctx, &guilds); err != nil {
		return nil, err
	}
	return guilds, nil
}

func (r *MongoGuildRepository) Update(ctx context.Context, guildID string, guild *Guild) error {
	guild.UpdatedAt = time.Now()
	result := db.GuildsCollection(r.database).FindOneAndReplace(
		ctx,
		bson.M{"guildId": guildID},
		guild,
	)
	return result.Err()
}

func (r *MongoGuildRepository) Delete(ctx context.Context, guildID string) error {
	_, err := db.GuildsCollection(r.database).DeleteOne(ctx, bson.M{"guildId": guildID})
	return err
}

func (r *MongoGuildRepository) UpsertRole(ctx context.Context, guildID string, role GuildRole) error {
	guild, err := r.FindByGuildID(ctx, guildID)
	if err != nil {
		return err
	}
	if guild == nil {
		return mongo.ErrNoDocuments
	}

	found := false
	for i := range guild.Roles {
		if guild.Roles[i].DiscordRoleID == role.DiscordRoleID {
			guild.Roles[i] = role
			found = true
			break
		}
	}
	if !found {
		guild.Roles = append(guild.Roles, role)
	}

	// Keep roles sorted by hierarchy (highest position first).
	sort.SliceStable(guild.Roles, func(i, j int) bool {
		return guild.Roles[i].Position > guild.Roles[j].Position
	})

	guild.UpdatedAt = time.Now()
	return r.Update(ctx, guildID, guild)
}

func (r *MongoGuildRepository) RemoveRole(ctx context.Context, guildID, discordRoleID string) error {
	guild, err := r.FindByGuildID(ctx, guildID)
	if err != nil {
		return err
	}
	if guild == nil {
		return mongo.ErrNoDocuments
	}

	newRoles := make([]GuildRole, 0, len(guild.Roles))
	for _, role := range guild.Roles {
		if role.DiscordRoleID != discordRoleID {
			newRoles = append(newRoles, role)
		}
	}
	guild.Roles = newRoles
	guild.UpdatedAt = time.Now()

	return r.Update(ctx, guildID, guild)
}

func (r *MongoGuildRepository) ReorderRoles(ctx context.Context, guildID string, orderedRoleIDs []string) error {
	guild, err := r.FindByGuildID(ctx, guildID)
	if err != nil {
		return err
	}
	if guild == nil {
		return mongo.ErrNoDocuments
	}

	// Assign positions based on order from top (index 0) to bottom.
	// Higher position means higher hierarchy.
	top := len(orderedRoleIDs)
	posByID := make(map[string]int, len(orderedRoleIDs))
	for i, roleID := range orderedRoleIDs {
		posByID[roleID] = top - i
	}

	for i := range guild.Roles {
		if p, ok := posByID[guild.Roles[i].DiscordRoleID]; ok {
			guild.Roles[i].Position = p
		}
	}

	sort.SliceStable(guild.Roles, func(i, j int) bool {
		return guild.Roles[i].Position > guild.Roles[j].Position
	})

	guild.UpdatedAt = time.Now()
	return r.Update(ctx, guildID, guild)
}

func (r *MongoGuildRepository) UpdateStatusRoleConfig(ctx context.Context, guildID string, cfg GuildStatusRoleConfig) error {
	guild, err := r.FindByGuildID(ctx, guildID)
	if err != nil {
		return err
	}
	if guild == nil {
		return mongo.ErrNoDocuments
	}

	guild.NotificationConfig.StatusRoles = cfg
	guild.UpdatedAt = time.Now()

	return r.Update(ctx, guildID, guild)
}

func (r *MongoGuildRepository) UpdateMilestoneNotificationConfig(ctx context.Context, guildID string, cfg GuildMilestoneNotificationConfig) error {
	guild, err := r.FindByGuildID(ctx, guildID)
	if err != nil {
		return err
	}
	if guild == nil {
		return mongo.ErrNoDocuments
	}

	if cfg.AnniversaryYears == nil {
		cfg.AnniversaryYears = []int{}
	}
	if cfg.HostedEventCounts == nil {
		cfg.HostedEventCounts = []int{}
	}
	if cfg.AttendedEventCounts == nil {
		cfg.AttendedEventCounts = []int{}
	}

	guild.NotificationConfig.MilestoneNotifications = cfg
	guild.UpdatedAt = time.Now()

	return r.Update(ctx, guildID, guild)
}

func (r *MongoGuildRepository) SetBotInstalled(ctx context.Context, guildID string) error {
	now := time.Now()
	_, err := db.GuildsCollection(r.database).UpdateOne(
		ctx,
		bson.M{"guildId": guildID},
		bson.M{"$set": bson.M{
			"botInstalled":   true,
			"botInstalledAt": now,
			"updatedAt":      now,
		}},
	)
	return err
}
