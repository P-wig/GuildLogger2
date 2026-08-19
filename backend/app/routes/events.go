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
	GuildID         string     `json:"guildId"`
	Title           string     `json:"title"`
	EventType       string     `json:"eventType"`
	Description     string     `json:"description"`
	ScheduledAt     time.Time  `json:"scheduledAt"`
	ChannelID       string     `json:"channelId"`
	ReminderEnabled bool       `json:"reminderEnabled"`
	Capacity        int        `json:"capacity"`
	CutoffAt        *time.Time `json:"cutoffAt"`
}

type closeEventPayload struct {
	Summary        string    `json:"summary"`
	ParticipantIDs []string  `json:"participantIds"`
	EventDate      time.Time `json:"eventDate"`
}

// RegisterEventsProtected registers all event routes on the JWT-guarded group.
func RegisterEventsProtected(
	g *echo.Group,
	eventRepo repositories.EventRepository,
	reportRepo repositories.EventReportRepository,
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
	botClient *discord.BotClient,
) {
	g.POST("/events", createEventHandler(eventRepo))
	g.GET("/events", listEventsHandler(eventRepo))
	g.GET("/events/:eventId", getEventHandler(eventRepo))
	g.POST("/events/:eventId/register", registerForEventHandler(eventRepo))
	g.POST("/events/:eventId/unregister", unregisterFromEventHandler(eventRepo))
	g.POST("/events/:eventId/maybe", maybeEventHandler(eventRepo))
	g.POST("/events/:eventId/unmaybe", unmaybeEventHandler(eventRepo))
	g.POST("/events/:eventId/decline", declineEventHandler(eventRepo))
	g.POST("/events/:eventId/undecline", undeclineEventHandler(eventRepo))
	g.POST("/events/:eventId/mod-mail", modMailEventHandler(eventRepo, guildRepo, memberRepo, botClient))
	g.POST("/events/:eventId/start", startEventHandler(eventRepo))
	g.POST("/events/:eventId/close", closeEventHandler(eventRepo, reportRepo))
	g.GET("/event-reports", listEventReportsHandler(reportRepo))
}

func createEventHandler(eventRepo repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		var in createEventPayload
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "invalid request body"})
		}

		in.GuildID = strings.TrimSpace(in.GuildID)
		in.Title = strings.TrimSpace(in.Title)
		if in.GuildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId is required"})
		}
		if in.Title == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "title is required"})
		}

		event := &repositories.Event{
			GuildID:         in.GuildID,
			HostDiscordID:   claims.DiscordID,
			Title:           in.Title,
			EventType:       in.EventType,
			Description:     in.Description,
			ScheduledAt:     in.ScheduledAt,
			ChannelID:       in.ChannelID,
			ReminderEnabled: in.ReminderEnabled,
			Capacity:        in.Capacity,
			CutoffAt:        in.CutoffAt,
		}

		if err := eventRepo.Create(c.Request().Context(), event); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to create event"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "event": event})
	}
}

func listEventsHandler(eventRepo repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		guildID := strings.TrimSpace(c.QueryParam("guildId"))
		if guildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId query parameter is required"})
		}

		events, err := eventRepo.FindByGuildID(c.Request().Context(), guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "events": events})
	}
}

func getEventHandler(eventRepo repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		eventID := strings.TrimSpace(c.Param("eventId"))
		if eventID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "eventId is required"})
		}

		event, err := eventRepo.FindByID(c.Request().Context(), eventID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if event == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event not found"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "event": event})
	}
}

func registerForEventHandler(eventRepo repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		eventID := strings.TrimSpace(c.Param("eventId"))
		if eventID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "eventId is required"})
		}

		if err := eventRepo.AddAttendee(c.Request().Context(), eventID, claims.DiscordID); err != nil {
			switch {
			case errors.Is(err, repositories.ErrEventNotFound):
				return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event not found"})
			case errors.Is(err, repositories.ErrAlreadyRegistered):
				return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "already registered for this event"})
			case errors.Is(err, repositories.ErrRegistrationClosed):
				return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "registration is closed for this event"})
			case errors.Is(err, repositories.ErrEventAtCapacity):
				return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "event is at capacity"})
			default:
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to register for event"})
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func unregisterFromEventHandler(eventRepo repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		eventID := strings.TrimSpace(c.Param("eventId"))
		if eventID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "eventId is required"})
		}

		if err := eventRepo.RemoveAttendee(c.Request().Context(), eventID, claims.DiscordID); err != nil {
			switch {
			case errors.Is(err, repositories.ErrEventNotFound):
				return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event not found"})
			case errors.Is(err, repositories.ErrNotRegistered):
				return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "not registered for this event"})
			default:
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to unregister from event"})
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func startEventHandler(eventRepo repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		eventID := strings.TrimSpace(c.Param("eventId"))
		if eventID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "eventId is required"})
		}

		ctx := c.Request().Context()

		event, err := eventRepo.FindByID(ctx, eventID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if event == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event not found"})
		}
		if event.HostDiscordID != claims.DiscordID {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "only the host can start this event"})
		}

		if err := eventRepo.Start(ctx, eventID); err != nil {
			if errors.Is(err, repositories.ErrInvalidStatusTransition) {
				return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "event is not in open status"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to start event"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func closeEventHandler(eventRepo repositories.EventRepository, reportRepo repositories.EventReportRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		eventID := strings.TrimSpace(c.Param("eventId"))
		if eventID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "eventId is required"})
		}

		var in closeEventPayload
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "invalid request body"})
		}

		ctx := c.Request().Context()

		event, err := eventRepo.FindByID(ctx, eventID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if event == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event not found"})
		}
		if event.HostDiscordID != claims.DiscordID {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "only the host can close this event"})
		}

		if err := eventRepo.Close(ctx, eventID); err != nil {
			if errors.Is(err, repositories.ErrInvalidStatusTransition) {
				return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "event is not in active status"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to close event"})
		}

		report := &repositories.EventReport{
			EventID:        eventID,
			GuildID:        event.GuildID,
			HostDiscordID:  event.HostDiscordID,
			EventDate:      in.EventDate,
			ParticipantIDs: in.ParticipantIDs,
			Summary:        in.Summary,
		}

		if err := reportRepo.Create(ctx, report); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "event closed but failed to save report"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "report": report})
	}
}

func listEventReportsHandler(reportRepo repositories.EventReportRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		guildID := strings.TrimSpace(c.QueryParam("guildId"))
		if guildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId query parameter is required"})
		}

		reports, err := reportRepo.FindByGuildID(c.Request().Context(), guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "reports": reports})
	}
}

// --- RSVP: maybe / unmaybe / decline / undecline ---

func maybeEventHandler(eventRepo repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}
		eventID := strings.TrimSpace(c.Param("eventId"))
		if err := eventRepo.AddMaybe(c.Request().Context(), eventID, claims.DiscordID); err != nil {
			switch {
			case errors.Is(err, repositories.ErrEventNotFound):
				return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event not found"})
			case errors.Is(err, repositories.ErrAlreadyMaybe):
				return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "already marked as maybe"})
			default:
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to update attendance"})
			}
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func unmaybeEventHandler(eventRepo repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}
		eventID := strings.TrimSpace(c.Param("eventId"))
		if err := eventRepo.RemoveMaybe(c.Request().Context(), eventID, claims.DiscordID); err != nil {
			switch {
			case errors.Is(err, repositories.ErrEventNotFound):
				return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event not found"})
			case errors.Is(err, repositories.ErrNotMaybe):
				return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "not in the maybe list"})
			default:
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to update attendance"})
			}
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func declineEventHandler(eventRepo repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}
		eventID := strings.TrimSpace(c.Param("eventId"))
		if err := eventRepo.AddDecline(c.Request().Context(), eventID, claims.DiscordID); err != nil {
			switch {
			case errors.Is(err, repositories.ErrEventNotFound):
				return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event not found"})
			case errors.Is(err, repositories.ErrAlreadyDeclined):
				return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "already declined"})
			default:
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to update attendance"})
			}
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func undeclineEventHandler(eventRepo repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}
		eventID := strings.TrimSpace(c.Param("eventId"))
		if err := eventRepo.RemoveDecline(c.Request().Context(), eventID, claims.DiscordID); err != nil {
			switch {
			case errors.Is(err, repositories.ErrEventNotFound):
				return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event not found"})
			case errors.Is(err, repositories.ErrNotDeclined):
				return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "not in the declined list"})
			default:
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to update attendance"})
			}
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

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
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		eventID := strings.TrimSpace(c.Param("eventId"))
		ctx := c.Request().Context()

		event, err := eventRepo.FindByID(ctx, eventID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if event == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event not found"})
		}
		if event.Status == repositories.EventStatusClosed {
			return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "event is closed"})
		}
		if !strings.EqualFold(event.EventType, "skirmish") {
			return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "mod mail is only available for Skirmish events"})
		}

		guild, err := guildRepo.FindByGuildID(ctx, event.GuildID)
		if err != nil || guild == nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "could not load guild"})
		}

		// Verify caller is a moderator.
		callerMember, err := memberRepo.FindByGuildAndDiscordID(ctx, event.GuildID, claims.DiscordID)
		if err != nil || callerMember == nil {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "forbidden"})
		}
		modRoles := make(map[string]bool)
		for _, id := range guild.NotificationConfig.StatusRoles.ModeratorRoleIDs {
			modRoles[id] = true
		}
		isMod := false
		for _, id := range callerMember.RoleIDs {
			if modRoles[id] {
				isMod = true
				break
			}
		}
		if !isMod {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "forbidden: moderator role required"})
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
