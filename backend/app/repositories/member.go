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

type MemberSummary struct {
	DiscordID       string       `bson:"discordId"    json:"discordId"`
	RankedRoleID    string       `bson:"rankedRoleId" json:"rankedRoleId"`
	Status          MemberStatus `bson:"status"       json:"status"`
	DiscordJoinedAt time.Time    `bson:"discordJoinedAt" json:"discordJoinedAt"`
	RoleIDs         []string     `bson:"roleIds"      json:"roleIds"`
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

	// FindAnniversaryMembers returns opted-in members whose discordJoinedAt
	// falls on today's month/day and whose tenure in years matches one of anniversaryYears.
	FindAnniversaryMembers(ctx context.Context, guildID string, anniversaryYears []int) ([]Member, error)

	// FindSummariesByGuildID returns a projected member list for dashboard use.
	// Only display-relevant fields are fetched from the database.
	FindSummariesByGuildID(ctx context.Context, guildID string) ([]MemberSummary, error)
	// GetGuildMemberCounts returns total/active/inactive member counts for a guild.
	GetGuildMemberCounts(ctx context.Context, guildID string) (*GuildDashboardStats, error)

	// GetDashboardMemberRows returns member rows and inactive members using the configured inactivity window.
	GetDashboardMemberRows(
		ctx context.Context,
		guildID string,
		inactivityDays int,
	) ([]GuildDashboardMemberRow, []GuildDashboardInactiveMember, error)

	EnsureIndexes(ctx context.Context) error
}
