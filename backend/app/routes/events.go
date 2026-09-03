package routes

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/P-wig/GuildLogger2/backend/app/discord"
	"github.com/P-wig/GuildLogger2/backend/app/repositories"
	"github.com/P-wig/GuildLogger2/backend/app/session"
	"github.com/labstack/echo/v4"
)

type createEventPayload struct {
	GuildID     string    `json:"guildId"`
	EventType   string    `json:"eventType"`
	Description string    `json:"description"`
	ScheduledAt time.Time `json:"scheduledAt"`
}

// RegisterEventsProtected registers all event routes on the JWT-guarded group.
//
// Every handler authorizes against guild membership via getGuildMemberTier in addition to
// the JWT check performed by the group middleware. A valid session alone is never
// sufficient — the caller must be a synced member of the event's guild.
func RegisterEventsProtected(
	g *echo.Group,
	eventRepo repositories.EventRepository,
	reportRepo repositories.EventReportRepository,
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
	botClient *discord.BotClient,
	events *discord.EventService,
) {
	g.POST("/events", createEventHandler(guildRepo, memberRepo, events))
	g.GET("/events", listEventsHandler(eventRepo, guildRepo, memberRepo))
	g.GET("/events/:eventId", getEventHandler(eventRepo, guildRepo, memberRepo))
	g.POST("/events/:eventId/register", rsvpEventHandler(eventRepo, guildRepo, memberRepo, events, "join"))
	g.POST("/events/:eventId/unregister", rsvpEventHandler(eventRepo, guildRepo, memberRepo, events, "leave"))
	g.POST("/events/:eventId/maybe", rsvpEventHandler(eventRepo, guildRepo, memberRepo, events, "maybe"))
	g.POST("/events/:eventId/unmaybe", rsvpEventHandler(eventRepo, guildRepo, memberRepo, events, "unmaybe"))
	g.POST("/events/:eventId/decline", rsvpEventHandler(eventRepo, guildRepo, memberRepo, events, "decline"))
	g.POST("/events/:eventId/undecline", rsvpEventHandler(eventRepo, guildRepo, memberRepo, events, "undecline"))
	g.POST("/events/:eventId/mod-mail", modMailEventHandler(eventRepo, guildRepo, memberRepo, botClient))
	g.POST("/events/:eventId/start", startEventHandler(eventRepo, guildRepo, memberRepo, events))
	g.POST("/events/:eventId/end", endEventHandler(eventRepo, guildRepo, memberRepo, events))
	g.POST("/events/:eventId/close-channel", closeEventChannelHandler(eventRepo, guildRepo, memberRepo, events))
	g.GET("/event-reports", listEventReportsHandler(reportRepo, guildRepo, memberRepo))
}

// authorizeGuild resolves the caller's session and verifies they hold at least minTier in
// guildID. On failure it returns a ready-to-return error response as the third value.
func authorizeGuild(
	c echo.Context,
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
	guildID string,
	minTier memberTier,
) (*repositories.Guild, *session.Claims, error) {
	claims, ok := c.Get("user").(*session.Claims)
	if !ok || claims == nil {
		return nil, nil, c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
	}
	if guildID == "" {
		return nil, nil, c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId is required"})
	}
	ctx := c.Request().Context()
	guild, err := guildRepo.FindByGuildID(ctx, guildID)
	if err != nil {
		return nil, nil, c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
	}
	if guild == nil {
		return nil, nil, c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "guild not found"})
	}
	if getGuildMemberTier(ctx, memberRepo, guildID, guild, claims.DiscordID) < minTier {
		return nil, nil, c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "access denied"})
	}
	return guild, claims, nil
}

// loadAuthorizedEvent fetches an event and verifies the caller holds minTier in its guild.
func loadAuthorizedEvent(
	c echo.Context,
	eventRepo repositories.EventRepository,
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
	minTier memberTier,
) (*repositories.Event, *session.Claims, error) {
	eventID := strings.TrimSpace(c.Param("eventId"))
	if eventID == "" {
		return nil, nil, c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "eventId is required"})
	}
	event, err := eventRepo.FindByID(c.Request().Context(), eventID)
	if err != nil {
		return nil, nil, c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
	}
	if event == nil {
		return nil, nil, c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event not found"})
	}
	_, claims, resp := authorizeGuild(c, guildRepo, memberRepo, event.GuildID, minTier)
	if resp != nil {
		return nil, nil, resp
	}
	return event, claims, nil
}

// eventLifecycleError maps EventService and repository errors onto HTTP responses.
func eventLifecycleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, repositories.ErrEventNotFound):
		return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event not found"})
	case errors.Is(err, discord.ErrNotHost):
		return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "only the event host can perform this action"})
	case errors.Is(err, discord.ErrAlreadyLogged):
		return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "event has already been logged"})
	case errors.Is(err, repositories.ErrInvalidStatusTransition):
		return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "event is not in the required status for this action"})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to update event"})
	}
}

// createEventHandler creates an event and posts its Discord announcement. Channel and
// capacity come from the guild's event-type configuration, so the result is identical
// to running /event create in Discord.
func createEventHandler(
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
	events *discord.EventService,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		var in createEventPayload
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "invalid request body"})
		}
		in.GuildID = strings.TrimSpace(in.GuildID)
		in.EventType = strings.TrimSpace(in.EventType)

		_, claims, resp := authorizeGuild(c, guildRepo, memberRepo, in.GuildID, tierMember)
		if resp != nil {
			return resp
		}
		if in.EventType == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "eventType is required"})
		}
		if in.ScheduledAt.IsZero() {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "scheduledAt is required"})
		}

		event, err := events.CreateEvent(c.Request().Context(), discord.CreateEventParams{
			GuildID:     in.GuildID,
			HostID:      claims.DiscordID,
			EventType:   in.EventType,
			Description: in.Description,
			ScheduledAt: in.ScheduledAt,
		})
		switch {
		case errors.Is(err, discord.ErrNoChannelConfigured):
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"ok": false, "error": "no channel configured for event type \"" + in.EventType + "\"",
			})
		case errors.Is(err, discord.ErrAnnouncementFailed):
			// The event record is valid; only the Discord post failed.
			return c.JSON(http.StatusOK, map[string]interface{}{
				"ok": true, "event": event, "warning": "event created but the Discord announcement failed",
			})
		case err != nil:
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to create event"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "event": event})
	}
}

func listEventsHandler(
	eventRepo repositories.EventRepository,
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		guildID := strings.TrimSpace(c.QueryParam("guildId"))
		if _, _, resp := authorizeGuild(c, guildRepo, memberRepo, guildID, tierMember); resp != nil {
			return resp
		}
		events, err := eventRepo.FindByGuildID(c.Request().Context(), guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "events": events})
	}
}

func getEventHandler(
	eventRepo repositories.EventRepository,
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		event, _, resp := loadAuthorizedEvent(c, eventRepo, guildRepo, memberRepo, tierMember)
		if resp != nil {
			return resp
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "event": event})
	}
}

// rsvpEventHandler applies an RSVP action and re-renders the Discord announcement so the
// embed roster stays in sync with changes made from the web UI.
func rsvpEventHandler(
	eventRepo repositories.EventRepository,
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
	events *discord.EventService,
	action string,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		event, claims, resp := loadAuthorizedEvent(c, eventRepo, guildRepo, memberRepo, tierMember)
		if resp != nil {
			return resp
		}

		ctx := c.Request().Context()
		updated, err := events.SetRSVP(ctx, event.ID, claims.DiscordID, action)
		switch {
		case err == nil:
		case errors.Is(err, repositories.ErrEventNotFound):
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event not found"})
		case errors.Is(err, repositories.ErrEventAtCapacity):
			return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "event is at capacity"})
		case errors.Is(err, repositories.ErrRegistrationClosed):
			return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "registration is closed for this event"})
		case errors.Is(err, repositories.ErrAlreadyRegistered),
			errors.Is(err, repositories.ErrAlreadyMaybe),
			errors.Is(err, repositories.ErrAlreadyDeclined),
			errors.Is(err, repositories.ErrNotRegistered),
			errors.Is(err, repositories.ErrNotMaybe),
			errors.Is(err, repositories.ErrNotDeclined):
			return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to update attendance"})
		}

		if rErr := events.RefreshAnnouncement(ctx, updated); rErr != nil {
			c.Logger().Errorf("rsvp %s: RefreshAnnouncement %s: %v", action, updated.ID, rErr)
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "event": updated})
	}
}

// startEventHandler transitions open → active, creating the voice channel and refreshing
// the announcement exactly as the Discord Start button does.
func startEventHandler(
	eventRepo repositories.EventRepository,
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
	events *discord.EventService,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		event, claims, resp := loadAuthorizedEvent(c, eventRepo, guildRepo, memberRepo, tierMember)
		if resp != nil {
			return resp
		}
		ctx := c.Request().Context()
		started, err := events.StartEvent(ctx, event.ID, event.GuildID, claims.DiscordID)
		if err != nil {
			return eventLifecycleError(c, err)
		}
		hostName := hostDisplayName(c, memberRepo, event.GuildID, claims.DiscordID)
		events.ApplyStartEffects(ctx, started, hostName, c.Logger())
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// endEventHandler transitions active → closed. The event log is submitted separately
// through /api/event-log/submit, which is also what the Discord End button links to.
func endEventHandler(
	eventRepo repositories.EventRepository,
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
	events *discord.EventService,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		event, claims, resp := loadAuthorizedEvent(c, eventRepo, guildRepo, memberRepo, tierMember)
		if resp != nil {
			return resp
		}
		ctx := c.Request().Context()
		ended, err := events.EndEvent(ctx, event.ID, event.GuildID, claims.DiscordID, c.Logger())
		if err != nil {
			return eventLifecycleError(c, err)
		}
		events.ApplyEndEffects(ctx, ended, c.Logger())
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// closeEventChannelHandler tears down the event voice channel and deletes the event.
// The event report is the permanent record, so this is the final lifecycle step.
func closeEventChannelHandler(
	eventRepo repositories.EventRepository,
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
	events *discord.EventService,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		event, claims, resp := loadAuthorizedEvent(c, eventRepo, guildRepo, memberRepo, tierMember)
		if resp != nil {
			return resp
		}
		ctx := c.Request().Context()
		closed, err := events.CloseEventChannel(ctx, event.ID, event.GuildID, claims.DiscordID)
		if err != nil {
			return eventLifecycleError(c, err)
		}
		events.ApplyCloseChannelEffects(ctx, closed, c.Logger())
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func listEventReportsHandler(
	reportRepo repositories.EventReportRepository,
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		guildID := strings.TrimSpace(c.QueryParam("guildId"))
		if _, _, resp := authorizeGuild(c, guildRepo, memberRepo, guildID, tierMember); resp != nil {
			return resp
		}
		reports, err := reportRepo.FindByGuildID(c.Request().Context(), guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "reports": reports})
	}
}

// hostDisplayName resolves a member's username for voice-channel naming,
// falling back to "event" when the member record is unavailable.
func hostDisplayName(c echo.Context, memberRepo repositories.MemberRepository, guildID, discordID string) string {
	member, err := memberRepo.FindByGuildAndDiscordID(c.Request().Context(), guildID, discordID)
	if err != nil || member == nil || member.Username == "" {
		return "event"
	}
	return member.Username
}

// --- RSVP: maybe / unmaybe / decline / undecline ---

// --- Mod Mail ---

type modMailPayload struct {
	Message string `json:"message"`
}

func modMailEventHandler(
	eventRepo repositories.EventRepository,
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
	botClient *discord.BotClient,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		event, _, resp := loadAuthorizedEvent(c, eventRepo, guildRepo, memberRepo, tierModerator)
		if resp != nil {
			return resp
		}
		ctx := c.Request().Context()

		if event.Status == repositories.EventStatusClosed {
			return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "event is closed"})
		}

		guild, err := guildRepo.FindByGuildID(ctx, event.GuildID)
		if err != nil || guild == nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "could not load guild"})
		}

		// Determine non-responders.
		responded := make(map[string]bool)
		for _, id := range event.AttendingIDs {
			responded[id] = true
		}
		for _, id := range event.MaybeIDs {
			responded[id] = true
		}
		for _, id := range event.NotAttendingIDs {
			responded[id] = true
		}

		allMembers, err := memberRepo.FindByGuildID(ctx, event.GuildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to load members"})
		}
		var targets []string
		for _, m := range allMembers {
			if m.Status == repositories.MemberStatusActive && !responded[m.DiscordID] && !m.NotificationsOptOut {
				targets = append(targets, m.DiscordID)
			}
		}
		if len(targets) == 0 {
			return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "error": "no non-responding members to contact"})
		}

		var in modMailPayload
		_ = c.Bind(&in)
		msg := strings.TrimSpace(in.Message)
		if msg == "" {
			msg = fmt.Sprintf("📢 Hey! The **%s** event needs your RSVP. Head to the event channel and let us know if you're Attending, Not Attending, or Maybe.", event.Title)
		}

		sent := 0
		var failed []string
		for _, discordID := range targets {
			if err := botClient.SendDMToUser(ctx, discordID, msg); err == nil {
				sent++
			} else {
				failed = append(failed, discordID)
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "sent": sent, "failed": failed})
	}
}
