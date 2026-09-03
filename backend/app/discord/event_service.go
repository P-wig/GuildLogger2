package discord

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/P-wig/GuildLogger2/backend/app/repositories"
)

// Errors returned by EventService. Transport layers map these to their own
// error surface (ephemeral Discord replies or HTTP status codes).
var (
	ErrGuildNotConfigured  = errors.New("guild is not configured")
	ErrNoChannelConfigured = errors.New("no channel configured for this event type")
	ErrEventNotInGuild     = errors.New("event does not belong to this guild")
	ErrNotHost             = errors.New("only the event host can perform this action")
	ErrAlreadyLogged       = errors.New("event has already been logged")
	ErrAnnouncementFailed  = errors.New("event saved but the Discord announcement failed")
	ErrUnknownRSVPAction   = errors.New("unknown RSVP action")
)

// largeEventCapacity is the attendee cap applied to non-quick event types.
// Quick events are uncapped (capacity 0).
const largeEventCapacity = 99

// Logger is the minimal logging surface EventService needs.
// echo.Logger satisfies this, which keeps the service free of a transport dependency.
type Logger interface {
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Infof(format string, args ...interface{})
}

// EventService owns the event lifecycle and the Discord side effects that go with it.
// Both the Discord interaction handlers and the REST API call through here so the two
// interfaces cannot drift apart.
//
// Methods are split into two kinds:
//   - transitions (CreateEvent, SetRSVP, StartEvent, EndEvent, CloseEventChannel) validate
//     and write state. They are fast enough to run inside Discord's 3-second deadline.
//   - Apply*Effects perform the slow Discord API work. Discord handlers run these in a
//     goroutine after acknowledging; the REST API runs them inline.
type EventService struct {
	eventRepo  repositories.EventRepository
	reportRepo repositories.EventReportRepository
	guildRepo  repositories.GuildRepository
	bot        *BotClient
}

func NewEventService(
	eventRepo repositories.EventRepository,
	reportRepo repositories.EventReportRepository,
	guildRepo repositories.GuildRepository,
	bot *BotClient,
) *EventService {
	return &EventService{eventRepo: eventRepo, reportRepo: reportRepo, guildRepo: guildRepo, bot: bot}
}

// resolveEventType looks up the per-guild configuration for an event type name.
func resolveEventType(guild *repositories.Guild, eventType string) (channelID string, isQuick bool, found bool) {
	for _, et := range guild.EventConfig.EventTypes {
		if strings.EqualFold(et.Name, eventType) {
			return et.ChannelID, et.IsQuickEvent, true
		}
	}
	return "", false, false
}

// IsQuickEvent reports whether eventType is configured as a quick event in the guild.
func (s *EventService) IsQuickEvent(ctx context.Context, guildID, eventType string) bool {
	guild, err := s.guildRepo.FindByGuildID(ctx, guildID)
	if err != nil || guild == nil {
		return false
	}
	_, isQuick, _ := resolveEventType(guild, eventType)
	return isQuick
}

// announcementContent builds the plain-text line above the embed. It must sit outside the
// embed so an @role mention fires a push notification.
func announcementContent(event *repositories.Event) string {
	epoch := event.ScheduledAt.Unix()
	return fmt.Sprintf("<@%s> is hosting a %s at <t:%d:t> <t:%d:R>",
		event.HostDiscordID, strings.ToLower(event.EventType), epoch, epoch)
}

type CreateEventParams struct {
	GuildID     string
	HostID      string
	EventType   string
	Description string
	ScheduledAt time.Time
}

// CreateEvent persists an event and posts its RSVP announcement to the channel configured
// for the event type. On announcement failure the event is still returned along with
// ErrAnnouncementFailed, since the record itself is valid.
func (s *EventService) CreateEvent(ctx context.Context, p CreateEventParams) (*repositories.Event, error) {
	guild, err := s.guildRepo.FindByGuildID(ctx, p.GuildID)
	if err != nil {
		return nil, err
	}
	if guild == nil {
		return nil, ErrGuildNotConfigured
	}

	channelID, isQuick, found := resolveEventType(guild, p.EventType)
	if !found || channelID == "" {
		return nil, ErrNoChannelConfigured
	}

	description := strings.TrimSpace(p.Description)
	capacity := 0
	if isQuick {
		description = "" // quick events carry no description; the content line is the announcement
	} else {
		capacity = largeEventCapacity
	}

	now := time.Now()
	event := &repositories.Event{
		GuildID:         p.GuildID,
		HostDiscordID:   p.HostID,
		Title:           p.EventType,
		EventType:       p.EventType,
		Description:     description,
		ScheduledAt:     p.ScheduledAt,
		Status:          repositories.EventStatusOpen,
		AttendingIDs:    []string{},
		MaybeIDs:        []string{},
		NotAttendingIDs: []string{},
		ChannelID:       channelID,
		Capacity:        capacity,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, err
	}

	embed := buildEventEmbed(event.EventType, isQuick, event.Description, event.HostDiscordID,
		event.ID, string(event.Status), event.ScheduledAt.Unix(),
		event.AttendingIDs, event.MaybeIDs, event.NotAttendingIDs)

	msgID, err := s.bot.PostEmbedMessage(ctx, channelID, announcementContent(event),
		[]Embed{embed}, buildRSVPButtons(event.ID, isQuick, string(event.Status)))
	if err != nil {
		return event, fmt.Errorf("%w: %v", ErrAnnouncementFailed, err)
	}

	event.AnnouncementMessageID = msgID
	event.UpdatedAt = time.Now()
	_ = s.eventRepo.Update(ctx, event.ID, event)
	return event, nil
}

// RefreshAnnouncement re-renders the event's announcement embed and buttons in place.
// It is a no-op when the event has no announcement message (for example if the
// original post failed).
func (s *EventService) RefreshAnnouncement(ctx context.Context, event *repositories.Event) error {
	if event == nil || event.AnnouncementMessageID == "" || event.ChannelID == "" {
		return nil
	}
	isQuick := s.IsQuickEvent(ctx, event.GuildID, event.EventType)
	embed := buildEventEmbed(event.EventType, isQuick, event.Description, event.HostDiscordID,
		event.ID, string(event.Status), event.ScheduledAt.Unix(),
		event.AttendingIDs, event.MaybeIDs, event.NotAttendingIDs)
	return s.bot.EditMessage(ctx, event.ChannelID, event.AnnouncementMessageID,
		[]Embed{embed}, buildRSVPButtons(event.ID, isQuick, string(event.Status)))
}

// SetRSVP applies an RSVP action and returns the updated event.
// Valid actions: join, leave, maybe, unmaybe, decline, undecline.
// Repository sentinel errors (ErrAlreadyRegistered, ErrEventAtCapacity, ...) pass through.
func (s *EventService) SetRSVP(ctx context.Context, eventID, discordID, action string) (*repositories.Event, error) {
	var err error
	switch action {
	case "join":
		err = s.eventRepo.AddAttendee(ctx, eventID, discordID)
	case "leave":
		err = s.eventRepo.RemoveAttendee(ctx, eventID, discordID)
	case "maybe":
		err = s.eventRepo.AddMaybe(ctx, eventID, discordID)
	case "unmaybe":
		err = s.eventRepo.RemoveMaybe(ctx, eventID, discordID)
	case "decline":
		err = s.eventRepo.AddDecline(ctx, eventID, discordID)
	case "undecline":
		err = s.eventRepo.RemoveDecline(ctx, eventID, discordID)
	default:
		return nil, ErrUnknownRSVPAction
	}
	if err != nil {
		return nil, err
	}
	return s.eventRepo.FindByID(ctx, eventID)
}

// loadHostedEvent fetches an event and verifies guild scope and host ownership.
// guildID may be empty to skip the guild check.
func (s *EventService) loadHostedEvent(ctx context.Context, eventID, guildID, actorID string) (*repositories.Event, error) {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, repositories.ErrEventNotFound
	}
	if guildID != "" && event.GuildID != guildID {
		return nil, ErrEventNotInGuild
	}
	if event.HostDiscordID != actorID {
		return nil, ErrNotHost
	}
	return event, nil
}

// StartEvent transitions an open event to active. Call ApplyStartEffects afterwards to
// create the voice channel and refresh the announcement.
func (s *EventService) StartEvent(ctx context.Context, eventID, guildID, actorID string) (*repositories.Event, error) {
	event, err := s.loadHostedEvent(ctx, eventID, guildID, actorID)
	if err != nil {
		return nil, err
	}
	if event.Status != repositories.EventStatusOpen {
		return nil, repositories.ErrInvalidStatusTransition
	}
	if err := s.eventRepo.Start(ctx, eventID); err != nil {
		return nil, err
	}
	event.Status = repositories.EventStatusActive
	return event, nil
}

// ApplyStartEffects creates the event voice channel, moves the host into it, and
// refreshes the announcement. Failures are logged rather than returned: the event has
// already started and none of these steps should undo that.
func (s *EventService) ApplyStartEffects(ctx context.Context, event *repositories.Event, hostDisplayName string, logger Logger) {
	guild, err := s.guildRepo.FindByGuildID(ctx, event.GuildID)
	switch {
	case err != nil || guild == nil:
		logger.Errorf("start effects: guild lookup failed for %s: %v", event.GuildID, err)
	case guild.EventConfig.VoiceCategoryID == "":
		logger.Warnf("start effects: guild %s has no VoiceCategoryID configured — skipping voice channel creation", event.GuildID)
	default:
		name := memberChannelName(hostDisplayName, event.EventType)
		vcID, vcErr := s.bot.CreateEventVoiceChannel(ctx, event.GuildID, name, guild.EventConfig.VoiceCategoryID, event.HostDiscordID)
		if vcErr != nil {
			logger.Errorf("start effects: CreateEventVoiceChannel failed: %v", vcErr)
		} else {
			if fresh, fErr := s.eventRepo.FindByID(ctx, event.ID); fErr == nil && fresh != nil {
				fresh.VoiceChannelID = vcID
				fresh.UpdatedAt = time.Now()
				if uErr := s.eventRepo.Update(ctx, event.ID, fresh); uErr != nil {
					logger.Errorf("start effects: failed to save VoiceChannelID %s: %v", vcID, uErr)
				}
				event = fresh
			}
			_ = s.bot.MoveGuildMember(ctx, event.GuildID, event.HostDiscordID, vcID)
		}
	}

	if fresh, fErr := s.eventRepo.FindByID(ctx, event.ID); fErr == nil && fresh != nil {
		event = fresh
	}
	if rErr := s.RefreshAnnouncement(ctx, event); rErr != nil {
		logger.Errorf("start effects: RefreshAnnouncement %s: %v", event.ID, rErr)
	}
}

// EndEvent transitions an active event to closed, snapshotting the voice channel roster
// first so the log form can pre-fill participants. Call ApplyEndEffects afterwards to
// refresh the announcement.
func (s *EventService) EndEvent(ctx context.Context, eventID, guildID, actorID string, logger Logger) (*repositories.Event, error) {
	event, err := s.loadHostedEvent(ctx, eventID, guildID, actorID)
	if err != nil {
		return nil, err
	}
	if event.Status != repositories.EventStatusActive {
		return nil, repositories.ErrInvalidStatusTransition
	}
	if existing, _ := s.reportRepo.FindByEventID(ctx, eventID); existing != nil {
		return nil, ErrAlreadyLogged
	}

	if event.VoiceChannelID != "" {
		candidates := append(append([]string{}, event.AttendingIDs...), event.MaybeIDs...)
		if ids := s.CollectVoiceChannelMembers(ctx, event.GuildID, event.VoiceChannelID, candidates, logger); len(ids) > 0 {
			event.VoiceMemberIDs = ids
			event.UpdatedAt = time.Now()
			_ = s.eventRepo.Update(ctx, event.ID, event)
		}
	}

	if err := s.eventRepo.Close(ctx, eventID); err != nil {
		return nil, err
	}
	event.Status = repositories.EventStatusClosed
	return event, nil
}

// ApplyEndEffects snapshots the voice roster if the start goroutine had not yet written
// VoiceChannelID, then refreshes the announcement so Close Channel becomes enabled.
func (s *EventService) ApplyEndEffects(ctx context.Context, event *repositories.Event, logger Logger) {
	if event.VoiceChannelID == "" {
		if fresh, err := s.eventRepo.FindByID(ctx, event.ID); err == nil && fresh != nil && fresh.VoiceChannelID != "" {
			candidates := append(append([]string{}, fresh.AttendingIDs...), fresh.MaybeIDs...)
			if ids := s.CollectVoiceChannelMembers(ctx, fresh.GuildID, fresh.VoiceChannelID, candidates, logger); len(ids) > 0 {
				fresh.VoiceMemberIDs = ids
				fresh.UpdatedAt = time.Now()
				_ = s.eventRepo.Update(ctx, fresh.ID, fresh)
			}
			event = fresh
		}
	}

	if event.AnnouncementMessageID == "" {
		logger.Warnf("end effects: event %s has no AnnouncementMessageID — Close Channel button will not be enabled", event.ID)
		return
	}
	if fresh, err := s.eventRepo.FindByID(ctx, event.ID); err == nil && fresh != nil {
		event = fresh
	}
	if err := s.RefreshAnnouncement(ctx, event); err != nil {
		logger.Errorf("end effects: RefreshAnnouncement %s: %v", event.ID, err)
	}
}

// CloseEventChannel validates that a closed event may have its channel torn down.
// Call ApplyCloseChannelEffects to perform the teardown and delete the event.
func (s *EventService) CloseEventChannel(ctx context.Context, eventID, guildID, actorID string) (*repositories.Event, error) {
	event, err := s.loadHostedEvent(ctx, eventID, guildID, actorID)
	if err != nil {
		return nil, err
	}
	if event.Status != repositories.EventStatusClosed {
		return nil, repositories.ErrInvalidStatusTransition
	}
	return event, nil
}

// ApplyCloseChannelEffects returns members to the lobby, deletes the event voice channel,
// and then deletes the event record. The event report is the permanent trace.
func (s *EventService) ApplyCloseChannelEffects(ctx context.Context, event *repositories.Event, logger Logger) {
	defer func() {
		if err := s.eventRepo.Delete(ctx, event.ID); err != nil {
			logger.Errorf("close channel: delete event %s: %v", event.ID, err)
		}
	}()

	if event.VoiceChannelID == "" {
		return
	}
	lobbyID := ""
	if guild, err := s.guildRepo.FindByGuildID(ctx, event.GuildID); err == nil && guild != nil {
		lobbyID = guild.EventConfig.LobbyChannelID
	}
	if lobbyID != "" {
		for _, uid := range event.VoiceMemberIDs {
			if err := s.bot.MoveGuildMember(ctx, event.GuildID, uid, lobbyID); err != nil {
				logger.Errorf("close channel: MoveGuildMember %s: %v", uid, err)
			}
		}
	}
	if err := s.bot.DeleteChannel(ctx, event.VoiceChannelID); err != nil {
		logger.Errorf("close channel: DeleteChannel %s: %v", event.VoiceChannelID, err)
	}
}

// CollectVoiceChannelMembers returns the user IDs currently in channelID. It first tries
// the guild-wide voice-states endpoint and falls back to per-user lookups over candidates
// (attendees + maybe RSVPs) when that returns an error or no members.
func (s *EventService) CollectVoiceChannelMembers(
	ctx context.Context,
	guildID, channelID string,
	candidates []string,
	logger Logger,
) []string {
	states, err := s.bot.GetGuildVoiceStates(ctx, guildID)
	if err == nil {
		var ids []string
		for _, vs := range states {
			if vs.ChannelID == channelID && vs.UserID != "" {
				ids = append(ids, vs.UserID)
			}
		}
		if len(ids) > 0 {
			return ids
		}
		logger.Infof("collectVoiceChannelMembers: bulk endpoint returned 0 members in %s — falling back to per-user lookup", channelID)
	} else {
		logger.Errorf("collectVoiceChannelMembers: bulk endpoint error: %v — falling back to per-user lookup", err)
	}

	var ids []string
	for _, uid := range candidates {
		vs, vsErr := s.bot.GetUserVoiceState(ctx, guildID, uid)
		if vsErr != nil {
			logger.Errorf("collectVoiceChannelMembers: GetUserVoiceState %s: %v", uid, vsErr)
			continue
		}
		if vs != nil && vs.ChannelID == channelID {
			ids = append(ids, uid)
		}
	}
	return ids
}
