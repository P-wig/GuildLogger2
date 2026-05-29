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
	RoleIDs             []string     `bson:"roleIds"`                 // full mirrored discord role IDs
	RankedRoleID        string       `bson:"rankedRoleId"`            // one app-selected ranked role
	Status              MemberStatus `bson:"status"`                  // active|inactive
	NotificationsOptOut bool         `bson:"notificationsOptOut"`     // true = do not send announcements
	DiscordJoinedAt     time.Time    `bson:"discordJoinedAt"`         // when the member joined the Discord guild
	FirstSyncedAt       time.Time    `bson:"firstSyncedAt"`           // when we first recorded this member
	LastSyncedAt        time.Time    `bson:"lastSyncedAt"`            // updated on every sync that confirms presence
	DeactivatedAt       *time.Time   `bson:"deactivatedAt,omitempty"` // set when status → inactive, nil when active
	UpdatedAt           time.Time    `bson:"updatedAt"`
}

type MemberStats struct {
	HostedCount       int64      `json:"hostedCount"`
	ParticipatedCount int64      `json:"participatedCount"`
	DiscordJoinedAt   time.Time  `json:"discordJoinedAt"`
	FirstSyncedAt     time.Time  `json:"firstSyncedAt"`
	DeactivatedAt     *time.Time `json:"deactivatedAt,omitempty"`
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

	// Derived from events collection combined with member record.
	GetStats(ctx context.Context, guildID, discordID string) (*MemberStats, error)

	EnsureIndexes(ctx context.Context) error
}
