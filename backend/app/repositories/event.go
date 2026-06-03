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

type EventStatus string

const (
	EventStatusOpen   EventStatus = "open"
	EventStatusActive EventStatus = "active"
	EventStatusClosed EventStatus = "closed"
)

// EventReport is submitted by the host when closing an event.
type EventReport struct {
    Summary        string    `bson:"summary"`        // post-event summary, compressed in storage
    ParticipantIDs []string  `bson:"participantIds"` // confirmed attendees at close time
    SubmittedAt    time.Time `bson:"submittedAt"`
}

// Event is the aggregate root for a scheduled guild event.
type Event struct {
	ID              string       `bson:"_id,omitempty"`
	GuildID         string       `bson:"guildId"`
	HostDiscordID   string       `bson:"hostDiscordId"`
	Title           string       `bson:"title"`
	Description     string       `bson:"description"`     // pre-event description, max 3000 chars
	AttendingIDs    []string     `bson:"attendingIds"`    // Discord IDs registered to attend
	ScheduledAt     time.Time    `bson:"scheduledAt"`     // epoch set by the host
	Status          EventStatus  `bson:"status"`          // open|active|closed
	ChannelID       string       `bson:"channelId"`       // Discord channel the bot opens/closes
	ReminderEnabled bool         `bson:"reminderEnabled"` // send 1hr pre-event reminder to registrants
	Capacity        int          `bson:"capacity"`        // max registrations; 0 = unlimited
	CutoffAt        *time.Time   `bson:"cutoffAt,omitempty"`
	ReminderSentAt  *time.Time   `bson:"reminderSentAt,omitempty"`
	StartedAt       *time.Time   `bson:"startedAt,omitempty"`
	ClosedAt        *time.Time   `bson:"closedAt,omitempty"`
	Report          *EventReport `bson:"report,omitempty"`
	CreatedAt       time.Time    `bson:"createdAt"`
	UpdatedAt       time.Time    `bson:"updatedAt"`
}

// Participation records a single member's registration for an event.
// Stored in a separate collection keyed on (eventId, discordId).
type Participation struct {
	ID           string    `bson:"_id,omitempty"`
	EventID      string    `bson:"eventId"`
	GuildID      string    `bson:"guildId"`
	DiscordID    string    `bson:"discordId"`
	RegisteredAt time.Time `bson:"registeredAt"`
}

// EventRepository defines the contract for event data access operations.
type EventRepository interface {
	Create(ctx context.Context, event *Event) error

	// FindByID retrieves an event by its MongoDB ObjectID string.
	// Returns (nil, nil) if not found; only returns error on DB failure.
	FindByID(ctx context.Context, eventID string) (*Event, error)

	FindByGuildID(ctx context.Context, guildID string) ([]Event, error)

	// Update replaces the event document. Returns ErrEventNotFound if missing.
	Update(ctx context.Context, eventID string, event *Event) error

	Delete(ctx context.Context, eventID string) error

	EnsureIndexes(ctx context.Context) error
}

// ParticipationRepository defines the contract for event registration operations.
type ParticipationRepository interface {
	// Register inserts a new participation record. Returns ErrAlreadyRegistered if one exists.
	Register(ctx context.Context, p *Participation) error

	// Unregister removes a participation record. Returns ErrNotRegistered if none exists.
	Unregister(ctx context.Context, eventID, discordID string) error

	FindByEventID(ctx context.Context, eventID string) ([]Participation, error)
	CountByEventID(ctx context.Context, eventID string) (int64, error)

	// IsRegistered returns true if the member has an active registration.
	IsRegistered(ctx context.Context, eventID, discordID string) (bool, error)

	EnsureIndexes(ctx context.Context) error
}
