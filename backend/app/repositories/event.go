package repositories

import (
	"context"
	"time"
)

// Event represents an event in a Discord guild where members can register to attend.
type Event struct {
	ID            string    `bson:"_id,omitempty"`
	EventID       string    `bson:"eventId"`       // User-friendly event identifier
	GuildID       string    `bson:"guildId"`       // Discord guild this event belongs to
	HostDiscordID string    `bson:"hostDiscordId"` // Discord ID of the member hosting
	AttendingIDs  []string  `bson:"attendingIds"`  // List of Discord IDs registered to attend
	EventDate     time.Time `bson:"eventDate"`     // When the event occurs
	Notes         string    `bson:"notes"`         // Event description/notes (max 3000 chars)
	CreatedAt     time.Time `bson:"createdAt"`
	UpdatedAt     time.Time `bson:"updatedAt"`
}

// EventRepository defines the contract for event data access operations.
type EventRepository interface {
	// Create inserts a new event.
	Create(ctx context.Context, event *Event) error

	// FindByEventID retrieves an event by its event ID.
	// Returns (nil, nil) if not found; only returns error on DB failure.
	FindByEventID(ctx context.Context, eventID string) (*Event, error)

	// FindByGuildID retrieves all events for a specific guild.
	FindByGuildID(ctx context.Context, guildID string) ([]Event, error)

	// Update modifies an existing event.
	Update(ctx context.Context, eventID string, event *Event) error

	// AddAttendee registers a member to attend the event.
	// No-op if already attending (idempotent).
	AddAttendee(ctx context.Context, eventID, discordID string) error

	// RemoveAttendee unregisters a member from attending.
	// No-op if not attending (idempotent).
	RemoveAttendee(ctx context.Context, eventID, discordID string) error

	// Delete removes an event entirely.
	Delete(ctx context.Context, eventID string) error

	// EnsureIndexes creates the necessary indexes for the event collection.
	EnsureIndexes(ctx context.Context) error
}
