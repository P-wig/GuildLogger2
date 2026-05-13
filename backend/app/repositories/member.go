package repositories

import (
	"context"
	"time"
)

type MemberStatus string

const (
	MemberStatusActive   MemberStatus = "active"
	MemberStatusInactive MemberStatus = "inactive"
)

type Member struct {
	ID                  string       `bson:"_id,omitempty"`
	GuildID             string       `bson:"guildId"`
	DiscordID           string       `bson:"discordId"`
	RoleIDs             []string     `bson:"roleIds"`             // full mirrored discord role IDs
	RankedRoleID        string       `bson:"rankedRoleId"`        // one app-selected ranked role
	Status              MemberStatus `bson:"status"`              // active|inactive
	NotificationsOptOut bool         `bson:"notificationsOptOut"` // true = do not send announcements
	JoinedAt            time.Time    `bson:"joinedAt"`
	UpdatedAt           time.Time    `bson:"updatedAt"`
}

type MemberEventStats struct {
	HostedCount       int64 `json:"hostedCount"`
	ParticipatedCount int64 `json:"participatedCount"`
}

type MemberRepository interface {
	Create(ctx context.Context, member *Member) error
	Upsert(ctx context.Context, member *Member) error
	FindByGuildAndDiscordID(ctx context.Context, guildID, discordID string) (*Member, error)
	FindByGuildID(ctx context.Context, guildID string) ([]Member, error)

	UpdateRoles(ctx context.Context, guildID, discordID string, roleIDs []string) error
	UpdateStatusAndRank(ctx context.Context, guildID, discordID string, status MemberStatus, rankedRoleID string) error
	UpdateNotificationPreference(ctx context.Context, guildID, discordID string, optOut bool) error

	// Members eligible for announcements/newsletters.
	FindNotificationTargets(ctx context.Context, guildID string) ([]Member, error)

	Delete(ctx context.Context, guildID, discordID string) error

	// Derived from events collection (source of truth).
	GetEventStats(ctx context.Context, guildID, discordID string) (*MemberEventStats, error)

	EnsureIndexes(ctx context.Context) error
}
