package repositories

import (
	"context"
	"errors"
	"time"
)

var ErrGuildAlreadyExists = errors.New("guild already connected")

type GuildRoleType string

const (
	GuildRoleTypeDefault GuildRoleType = "default"
	GuildRoleTypeRanked  GuildRoleType = "ranked"
)

// GuildRole mirrors a Discord role and stores app-specific metadata.
// Position should mirror Discord role position (higher number = higher role).
type GuildRole struct {
	DiscordRoleID  string        `bson:"discordRoleId"`
	Position       int           `bson:"position"`
	Type           GuildRoleType `bson:"type"`
	AppPermissions []string      `bson:"appPermissions"`
	Managed        bool          `bson:"managed"`
	IsDefault      bool          `bson:"isDefault"`
}

// GuildStatusRoleConfig defines which role IDs represent active/inactive members.
type GuildStatusRoleConfig struct {
	ActiveRoleID   string `bson:"activeRoleId"`
	InactiveRoleID string `bson:"inactiveRoleId"`
}

// Milestones are now notification-only (no automatic Discord role assignment).
type GuildMilestoneNotificationConfig struct {
	Enabled             bool  `bson:"enabled"`
	AnniversaryYears    []int `bson:"anniversaryYears"`
	HostedEventCounts   []int `bson:"hostedEventCounts"`
	AttendedEventCounts []int `bson:"attendedEventCounts"`
}

// GuildNotificationConfig stores guild-level communication settings.
type GuildNotificationConfig struct {
	StatusRoles            GuildStatusRoleConfig            `bson:"statusRoles"`
	MilestoneNotifications GuildMilestoneNotificationConfig `bson:"milestoneNotifications"`
}

// Guild is the aggregate root for guild configuration and role hierarchy.
type Guild struct {
	ID                 string                  `bson:"_id,omitempty"`
	GuildID            string                  `bson:"guildId"`
	Name               string                  `bson:"name"`
	OwnerDiscordID     string                  `bson:"ownerDiscordId"`
	Roles              []GuildRole             `bson:"roles"`
	NotificationConfig GuildNotificationConfig `bson:"notificationConfig"`
	BotInstalled       bool                    `bson:"botInstalled"`
	BotInstalledAt     *time.Time              `bson:"botInstalledAt,omitempty"`
	CreatedAt          time.Time               `bson:"createdAt"`
	UpdatedAt          time.Time               `bson:"updatedAt"`
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

	// Role operations
	UpsertRole(ctx context.Context, guildID string, role GuildRole) error
	RemoveRole(ctx context.Context, guildID, discordRoleID string) error
	ReorderRoles(ctx context.Context, guildID string, orderedRoleIDs []string) error

	// Guild-level configuration
	UpdateStatusRoleConfig(ctx context.Context, guildID string, cfg GuildStatusRoleConfig) error
	UpdateMilestoneNotificationConfig(ctx context.Context, guildID string, cfg GuildMilestoneNotificationConfig) error

	EnsureIndexes(ctx context.Context) error
}
