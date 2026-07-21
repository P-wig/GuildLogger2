package repositories

import (
	"context"
	"errors"
	"time"
)

var ErrGuildAlreadyExists = errors.New("guild already connected")
var ErrGuildNotFound = errors.New("guild not found")

type GuildRoleType string

const (
	GuildRoleTypeDefault GuildRoleType = "default"
	GuildRoleTypeRanked  GuildRoleType = "ranked"
)

// GuildRole mirrors a Discord role and stores app-specific metadata.
// Position should mirror Discord role position (higher number = higher role).
type GuildRole struct {
	DiscordRoleID  string        `bson:"discordRoleId"  json:"discordRoleId"`
	Name           string        `bson:"name"           json:"name"`
	Position       int           `bson:"position"       json:"position"`
	Type           GuildRoleType `bson:"type"           json:"type"`
	AppPermissions []string      `bson:"appPermissions" json:"appPermissions"`
	Managed        bool          `bson:"managed"        json:"managed"`
	IsDefault      bool          `bson:"isDefault"      json:"isDefault"`
}

// GuildStatusRoleConfig defines which role IDs represent active/inactive members.
type GuildStatusRoleConfig struct {
	ActiveRoleID   string `bson:"activeRoleId"   json:"activeRoleId"`
	InactiveRoleID string `bson:"inactiveRoleId" json:"inactiveRoleId"`
}

// Milestones are notification-only.
type GuildMilestoneNotificationConfig struct {
	Enabled               bool   `bson:"enabled"               json:"enabled"`
	NotificationChannelID string `bson:"notificationChannelId" json:"notificationChannelId"`
	AnniversaryYears      []int  `bson:"anniversaryYears"      json:"anniversaryYears"`
	HostedEventCounts     []int  `bson:"hostedEventCounts"     json:"hostedEventCounts"`
	AttendedEventCounts   []int  `bson:"attendedEventCounts"   json:"attendedEventCounts"`
}

// GuildNotificationConfig stores guild-level communication settings.
type GuildNotificationConfig struct {
	StatusRoles            GuildStatusRoleConfig            `bson:"statusRoles"            json:"statusRoles"`
	MilestoneNotifications GuildMilestoneNotificationConfig `bson:"milestoneNotifications" json:"milestoneNotifications"`
}

// Guild is the aggregate root for guild configuration and role hierarchy.
type Guild struct {
	ID                 string                  `bson:"_id,omitempty"            json:"-"`
	GuildID            string                  `bson:"guildId"                  json:"guildId"`
	Name               string                  `bson:"name"                     json:"name"`
	OwnerDiscordID     string                  `bson:"ownerDiscordId"           json:"ownerDiscordId"`
	Roles              []GuildRole             `bson:"roles"                    json:"roles"`
	NotificationConfig GuildNotificationConfig `bson:"notificationConfig"       json:"notificationConfig"`
	BotInstalled       bool                    `bson:"botInstalled"             json:"botInstalled"`
	BotInstalledAt     *time.Time              `bson:"botInstalledAt,omitempty" json:"botInstalledAt,omitempty"`
	CreatedAt          time.Time               `bson:"createdAt"                json:"createdAt"`
	UpdatedAt          time.Time               `bson:"updatedAt"                json:"updatedAt"`
}

type GuildRepository interface {
	// Base CRUD
	Create(ctx context.Context, guild *Guild) error
	FindByGuildID(ctx context.Context, guildID string) (*Guild, error)
	FindByOwnerDiscordID(ctx context.Context, ownerDiscordID string) ([]Guild, error)
	Update(ctx context.Context, guildID string, guild *Guild) error
	Delete(ctx context.Context, guildID string) error

	// Bot installation
	SetBotInstalled(ctx context.Context, guildID string) error
	SetBotInstalledWithRoles(ctx context.Context, guildID string, roles []GuildRole) error

	// Role operations
	UpsertRole(ctx context.Context, guildID string, role GuildRole) error
	RemoveRole(ctx context.Context, guildID, discordRoleID string) error
	ReorderRoles(ctx context.Context, guildID string, orderedRoleIDs []string) error

	// Guild-level configuration
	UpdateStatusRoleConfig(ctx context.Context, guildID string, cfg GuildStatusRoleConfig) error
	UpdateMilestoneNotificationConfig(ctx context.Context, guildID string, cfg GuildMilestoneNotificationConfig) error
	UpdateConfigAndRoles(ctx context.Context, guildID string, cfg GuildStatusRoleConfig, roles []GuildRole) error

	EnsureIndexes(ctx context.Context) error
}
