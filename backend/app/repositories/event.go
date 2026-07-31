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
}

// Event is the aggregate root for a scheduled guild event.
// Events may be deleted after closing; reports are the permanent record.
type Event struct {
	ID                    string      `bson:"_id,omitempty"`
	GuildID               string      `bson:"guildId"`
	HostDiscordID         string      `bson:"hostDiscordId"`
	Title                 string      `bson:"title"`
	Description           string      `bson:"description"`                     // pre-event description, max 3000 chars
	AttendingIDs          []string    `bson:"attendingIds"`                    // Discord IDs registered to attend
	ScheduledAt           time.Time   `bson:"scheduledAt"`                     // epoch set by the host
	Status                EventStatus `bson:"status"`                          // open|active|closed
	ChannelID             string      `bson:"channelId"`                       // Discord channel the bot opens/closes
	AnnouncementMessageID string      `bson:"announcementMessageId,omitempty"` // Discord message ID for the RSVP embed
	ReminderEnabled       bool        `bson:"reminderEnabled"`                 // send 1hr pre-event reminder to registrants
	Capacity              int         `bson:"capacity"`                        // max registrations; 0 = unlimited
	CutoffAt              *time.Time  `bson:"cutoffAt,omitempty"`
	ReminderSentAt        *time.Time  `bson:"reminderSentAt,omitempty"`
	StartedAt             *time.Time  `bson:"startedAt,omitempty"`
	ClosedAt              *time.Time  `bson:"closedAt,omitempty"`
	CreatedAt             time.Time   `bson:"createdAt"`
	UpdatedAt             time.Time   `bson:"updatedAt"`
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

	// Start transitions the event from open → active and records StartedAt.
	// Returns ErrInvalidStatusTransition if not currently open.
	Start(ctx context.Context, eventID string) error

	// Close transitions the event from active → closed and records ClosedAt.
	// Returns ErrInvalidStatusTransition if not currently active.
	// The report is created separately via EventReportRepository.
	Close(ctx context.Context, eventID string) error

	// GetLiveEventCounts returns count of open/active events from the events collection.
	GetLiveEventCounts(ctx context.Context, guildID string) (openOrActiveCount int64, err error)

	EnsureIndexes(ctx context.Context) error
}

// EventReportRepository defines the contract for permanent event report storage.
type EventReportRepository interface {
	// Create inserts a new report. SubmittedAt is set by the implementation.
	Create(ctx context.Context, report *EventReport) error

	// FindByEventID retrieves the report for a specific event.
	// Returns (nil, nil) if not found.
	FindByEventID(ctx context.Context, eventID string) (*EventReport, error)

	// FindByGuildID retrieves all reports for a guild, ordered by EventDate descending.
	FindByGuildID(ctx context.Context, guildID string) ([]EventReport, error)

	// GetGuildClosedEventCount returns closed-event count from event_reports (source of truth).
	GetGuildClosedEventCount(ctx context.Context, guildID string) (int64, error)

	// GetGuildMemberActivity returns hosted/attended aggregates per member for leaderboard/table enrichment.
	GetGuildMemberActivity(ctx context.Context, guildID string) ([]GuildDashboardLeaderboardEntry, error)

	// GetGuildParticipationStats returns participant slot totals and unique reported event count.
	GetGuildParticipationStats(ctx context.Context, guildID string) (participantSlots int64, uniqueReportedEvents int64, err error)

	// FindDashboardEvents retrieves and filters event reports for dashboard display.
	// Returns rows containing event summaries and participant info, suitable for search/filter UI.
	FindDashboardEvents(ctx context.Context, guildID string, filter GuildDashboardEventFilter) ([]GuildDashboardEventRow, error)

	// Update replaces mutable fields (host, eventDate, participantIds, summary) for an existing report.
	// Returns ErrReportNotFound if no document matches logID.
	Update(ctx context.Context, logID string, report *EventReport) error

	// Delete removes a report by its ID.
	// Returns ErrReportNotFound if no document matches logID.
	Delete(ctx context.Context, logID string) error

	EnsureIndexes(ctx context.Context) error
}
