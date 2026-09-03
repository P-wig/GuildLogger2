package repositories

import (
	"context"
	"errors"
	"time"
)

var ErrEventNotFound = errors.New("event not found")
var ErrAlreadyRegistered = errors.New("already registered for this event")
var ErrNotRegistered = errors.New("not registered for this event")
var ErrRegistrationClosed = errors.New("registration is closed for this event")
var ErrEventAtCapacity = errors.New("event is at capacity")
var ErrInvalidStatusTransition = errors.New("invalid event status transition")
var ErrReportNotFound = errors.New("event report not found")
var ErrAlreadyMaybe = errors.New("already marked as maybe for this event")
var ErrNotMaybe = errors.New("not in maybe list for this event")
var ErrAlreadyDeclined = errors.New("already declined this event")
var ErrNotDeclined = errors.New("not declined for this event")

type EventStatus string

const (
	EventStatusOpen   EventStatus = "open"
	EventStatusActive EventStatus = "active"
	EventStatusClosed EventStatus = "closed"
)

// EventReport is a permanent record submitted by the host when closing an event.
// Stored in its own collection and outlives the source event document.
type EventReport struct {
	ID                   string    `bson:"_id,omitempty" json:"id"`
	EventID              string    `bson:"eventId,omitempty" json:"eventId,omitempty"`
	GuildID              string    `bson:"guildId" json:"guildId"`
	HostDiscordID        string    `bson:"hostDiscordId" json:"hostDiscordId"`
	EventDate            time.Time `bson:"eventDate" json:"eventDate"`
	ParticipantIDs       []string  `bson:"participantIds" json:"participantIds"`
	Summary              string    `bson:"summary" json:"summary"`
	SubmittedAt          time.Time `bson:"submittedAt" json:"submittedAt"`
	SubmittedByDiscordID string    `bson:"submittedByDiscordId" json:"submittedByDiscordId"`
	LogsChannelID        string    `bson:"logsChannelId,omitempty" json:"logsChannelId,omitempty"`
	LogsMessageID        string    `bson:"logsMessageId,omitempty" json:"logsMessageId,omitempty"`
}

// Event is the aggregate root for a scheduled guild event.
// Events may be deleted after closing; reports are the permanent record.
type Event struct {
	ID                    string      `bson:"_id,omitempty"                    json:"id"`
	GuildID               string      `bson:"guildId"                          json:"guildId"`
	HostDiscordID         string      `bson:"hostDiscordId"                    json:"hostDiscordId"`
	Title                 string      `bson:"title"                            json:"title"`
	EventType             string      `bson:"eventType"                        json:"eventType"`
	Description           string      `bson:"description"                      json:"description"`
	AttendingIDs          []string    `bson:"attendingIds"                     json:"attendingIds"`
	MaybeIDs              []string    `bson:"maybeIds"                         json:"maybeIds"`
	NotAttendingIDs       []string    `bson:"notAttendingIds"                  json:"notAttendingIds"`
	ScheduledAt           time.Time   `bson:"scheduledAt"                      json:"scheduledAt"`
	Status                EventStatus `bson:"status"                           json:"status"`
	ChannelID             string      `bson:"channelId"                        json:"channelId"`
	AnnouncementMessageID string      `bson:"announcementMessageId,omitempty"  json:"announcementMessageId,omitempty"`
	ReminderEnabled       bool        `bson:"reminderEnabled"                  json:"reminderEnabled"`
	Capacity              int         `bson:"capacity"                         json:"capacity"`
	VoiceChannelID        string      `bson:"voiceChannelId,omitempty"          json:"voiceChannelId,omitempty"`
	VoiceMemberIDs        []string    `bson:"voiceMemberIds,omitempty"          json:"voiceMemberIds,omitempty"`
	CutoffAt              *time.Time  `bson:"cutoffAt,omitempty"               json:"cutoffAt,omitempty"`
	ReminderSentAt        *time.Time  `bson:"reminderSentAt,omitempty"         json:"reminderSentAt,omitempty"`
	ModMailSentAt         *time.Time  `bson:"modMailSentAt,omitempty"          json:"modMailSentAt,omitempty"`
	StartedAt             *time.Time  `bson:"startedAt,omitempty"              json:"startedAt,omitempty"`
	ClosedAt              *time.Time  `bson:"closedAt,omitempty"               json:"closedAt,omitempty"`
	CreatedAt             time.Time   `bson:"createdAt"                        json:"createdAt"`
	UpdatedAt             time.Time   `bson:"updatedAt"                        json:"updatedAt"`
}

// EventRepository defines the contract for event data access operations.
type EventRepository interface {
	Create(ctx context.Context, event *Event) error

	// FindByID retrieves an event by its ID string.
	// Returns (nil, nil) if not found; only returns error on DB failure.
	FindByID(ctx context.Context, eventID string) (*Event, error)

	FindByGuildID(ctx context.Context, guildID string) ([]Event, error)

	// Update replaces the event document. Returns ErrEventNotFound if missing.
	Update(ctx context.Context, eventID string, event *Event) error

	Delete(ctx context.Context, eventID string) error

	// AddAttendee registers a member for the event.
	// Returns ErrAlreadyRegistered if already attending, ErrEventNotFound if event missing.
	AddAttendee(ctx context.Context, eventID, discordID string) error

	// RemoveAttendee unregisters a member from the event.
	// Returns ErrNotRegistered if not attending, ErrEventNotFound if event missing.
	RemoveAttendee(ctx context.Context, eventID, discordID string) error

	// AddMaybe marks a member as "maybe" attending (Skirmish events).
	// Automatically removes the member from attendingIds and notAttendingIds.
	// Returns ErrAlreadyMaybe if already maybe, ErrEventNotFound if event missing.
	AddMaybe(ctx context.Context, eventID, discordID string) error

	// RemoveMaybe removes a member's "maybe" response.
	// Returns ErrNotMaybe if not in the maybe list.
	RemoveMaybe(ctx context.Context, eventID, discordID string) error

	// AddDecline marks a member as "not attending".
	// Automatically removes the member from attendingIds and maybeIds.
	// Returns ErrAlreadyDeclined if already declined.
	AddDecline(ctx context.Context, eventID, discordID string) error

	// RemoveDecline removes a member's "not attending" response.
	// Returns ErrNotDeclined if not in the declined list.
	RemoveDecline(ctx context.Context, eventID, discordID string) error

	// Start transitions the event from open → active and records StartedAt.
	// Returns ErrInvalidStatusTransition if not currently open.
	Start(ctx context.Context, eventID string) error

	// Close transitions the event from active → closed and records ClosedAt.
	// Returns ErrInvalidStatusTransition if not currently active.
	// The report is created separately via EventReportRepository.
	Close(ctx context.Context, eventID string) error

	// GetLiveEventCounts returns count of open/active events from the events collection.
	GetLiveEventCounts(ctx context.Context, guildID string) (openOrActiveCount int64, err error)

	// FindUpcomingForReminders returns events with scheduledAt in (now, cutoff) that have at
	// least one maybe RSVP and have not yet had a reminder sent (reminderSentAt == nil).
	FindUpcomingForReminders(ctx context.Context, now, cutoff time.Time) ([]Event, error)

	// MarkReminderSent sets reminderSentAt on an event to record that reminders were dispatched.
	MarkReminderSent(ctx context.Context, eventID string, sentAt time.Time) error

	// MarkModMailSent sets modMailSentAt to prevent the host from triggering a second wave.
	MarkModMailSent(ctx context.Context, eventID string, sentAt time.Time) error

	EnsureIndexes(ctx context.Context) error
}

// EventReportRepository defines the contract for permanent event report storage.
type EventReportRepository interface {
	// Create inserts a new report. SubmittedAt is set by the implementation.
	Create(ctx context.Context, report *EventReport) error

	// FindByID retrieves a report by its log ID.
	// Returns (nil, nil) if not found.
	FindByID(ctx context.Context, logID string) (*EventReport, error)

	// FindByEventID retrieves the report for a specific event.
	// Returns (nil, nil) if not found.
	FindByEventID(ctx context.Context, eventID string) (*EventReport, error)

	// FindByGuildID retrieves all reports for a guild, ordered by EventDate descending.
	FindByGuildID(ctx context.Context, guildID string) ([]EventReport, error)

	// GetGuildClosedEventCount returns closed-event count from event_reports (source of truth).
	GetGuildClosedEventCount(ctx context.Context, guildID string) (int64, error)

	// GetGuildMemberActivity returns hosted/attended aggregates per member for leaderboard/table enrichment.
	GetGuildMemberActivity(ctx context.Context, guildID string) ([]GuildDashboardLeaderboardEntry, error)

	// FindDashboardEvents retrieves and filters event reports for dashboard display.
	// Returns rows containing event summaries and participant info, suitable for search/filter UI.
	FindDashboardEvents(ctx context.Context, guildID string, filter GuildDashboardEventFilter) ([]GuildDashboardEventRow, error)

	// Update replaces mutable fields (host, eventDate, participantIds, summary) for an existing report.
	// Returns ErrReportNotFound if no document matches logID.
	Update(ctx context.Context, logID string, report *EventReport) error

	// Delete removes a report by its ID.
	// Returns ErrReportNotFound if no document matches logID.
	Delete(ctx context.Context, logID string) error

	// SetLogMessageRef stores the Discord log channel/message IDs used for embed synchronization.
	SetLogMessageRef(ctx context.Context, logID, channelID, messageID string) error

	EnsureIndexes(ctx context.Context) error
}
