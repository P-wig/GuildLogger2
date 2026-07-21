package repositories

import "time"

// GuildDashboardStats contains aggregated analytics for a guild dashboard.
//
// Stats sourcing:
// - TotalMembers, ActiveMembers, InactiveMembers: from members collection
// - TotalEvents: count of live (open/active) events + closed events (from event_reports)
// - ClosedEvents: count of event_reports (source of truth for completed events)
// - ParticipationRate: percentage of active members who have attended at least one closed event
//   Formula: (distinct members with AttendedCount > 0) / activeMembers * 100
//   Clamped to [0, 100]. Returns 0 when activeMembers is 0.
type GuildDashboardStats struct {
	TotalMembers      int64   `json:"totalMembers"`
	ActiveMembers     int64   `json:"activeMembers"`
	InactiveMembers   int64   `json:"inactiveMembers"`
	LiveEvents        int64   `json:"liveEvents"`
	TotalEvents       int64   `json:"totalEvents"`
	ClosedEvents      int64   `json:"closedEvents"`
	ParticipationRate float64 `json:"participationRate"`
}

type GuildDashboardLeaderboardEntry struct {
	DiscordID        string     `json:"discordId" bson:"discordId"`
	HostedCount      int64      `json:"eventsHosted" bson:"eventsHosted"`
	AttendedCount    int64      `json:"eventsAttended" bson:"eventsAttended"`
	Score            int64      `json:"score" bson:"score"`
	LastHostedDate   *time.Time `json:"lastHostedDate,omitempty" bson:"lastHostedDate,omitempty"`
	LastAttendedDate *time.Time `json:"lastAttendedDate,omitempty" bson:"lastAttendedDate,omitempty"`
	Rank             int64      `json:"rank" bson:"rank"`
}

type GuildDashboardMemberRow struct {
	DiscordID          string       `json:"discordId" bson:"discordId"`
	Username           string       `json:"username" bson:"username"`
	AvatarHash         string       `json:"avatarHash" bson:"avatarHash"`
	RankedRoleID       string       `json:"rankedRoleId" bson:"rankedRoleId"`
	Status             MemberStatus `json:"status" bson:"status"`
	DiscordJoinedAt    time.Time    `json:"discordJoinedAt" bson:"discordJoinedAt"`
	RoleIDs            []string     `json:"roleIds" bson:"roleIds"`
	EventsHosted       int64        `json:"eventsHosted" bson:"eventsHosted"`
	EventsAttended     int64        `json:"eventsAttended" bson:"eventsAttended"`
	LastHostedDate     *time.Time   `json:"lastHostedDate,omitempty" bson:"lastHostedDate,omitempty"`
	LastAttendedDate   *time.Time   `json:"lastAttendedDate,omitempty" bson:"lastAttendedDate,omitempty"`
	IsInactiveByCutoff bool         `json:"isInactiveByCutoff" bson:"isInactiveByCutoff"`
}

type GuildDashboardInactiveMember struct {
	DiscordID         string       `json:"discordId" bson:"discordId"`
	RankedRoleID      string       `json:"rankedRoleId" bson:"rankedRoleId"`
	Status            MemberStatus `json:"status" bson:"status"`
	DiscordJoinedAt   time.Time    `json:"discordJoinedAt" bson:"discordJoinedAt"`
	LastHostedDate    *time.Time   `json:"lastHostedDate,omitempty" bson:"lastHostedDate,omitempty"`
	LastAttendedDate  *time.Time   `json:"lastAttendedDate,omitempty" bson:"lastAttendedDate,omitempty"`
	LastActivityDate  *time.Time   `json:"lastActivityDate,omitempty" bson:"lastActivityDate,omitempty"`
	DaysSinceActivity *int64       `json:"daysSinceActivity,omitempty" bson:"daysSinceActivity,omitempty"`
}

type GuildDashboardEventFilter struct {
	StartDate  *time.Time // optional: filter events >= StartDate
	EndDate    *time.Time // optional: filter events <= EndDate
	AttendeeID string     // optional: filter to events where this member attended
	Limit      int        // max results; 0 = no limit
}

type GuildDashboardEventRow struct {
	EventID        string    `json:"eventId" bson:"eventId"`
	HostDiscordID  string    `json:"hostDiscordId" bson:"hostDiscordId"`
	EventDate      time.Time `json:"eventDate" bson:"eventDate"`
	ParticipantIDs []string  `json:"participantIds" bson:"participantIds"`
	Summary        string    `json:"summary" bson:"summary"`
}

type GuildDashboardData struct {
	Guild           *Guild                           `json:"guild"`
	Stats           GuildDashboardStats              `json:"stats"`
	Leaderboard     []GuildDashboardLeaderboardEntry `json:"leaderboard"`
	Members         []GuildDashboardMemberRow        `json:"members"`
	InactiveMembers []GuildDashboardInactiveMember   `json:"inactiveMembers"`
	Events          []GuildDashboardEventRow         `json:"events"`
}

type GuildDashboardQuery struct {
	LeaderboardBy    string     // score|hosted|attended
	LeaderboardLimit int        // default 10
	InactiveDays     int        // default 30
	EventStart       *time.Time // optional
	EventEnd         *time.Time // optional
	AttendeeID       string     // optional
}
