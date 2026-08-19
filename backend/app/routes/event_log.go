package routes

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/P-wig/GuildLogger2/backend/app/discord"
	"github.com/P-wig/GuildLogger2/backend/app/repositories"
	"github.com/P-wig/GuildLogger2/backend/app/session"
	"github.com/labstack/echo/v4"
)

// RegisterEventLogRoutes mounts the public event-log token endpoints.
// These routes do NOT require a JWT session — they are protected by the
// one-time signed event-log token instead.
func RegisterEventLogRoutes(
	e *echo.Echo,
	eventRepo repositories.EventRepository,
	eventReportRepo repositories.EventReportRepository,
	memberRepo repositories.MemberRepository,
	guildRepo repositories.GuildRepository,
	botClient *discord.BotClient,
	secretKey string,
) {
	e.GET("/api/event-log/validate", validateEventLogHandler(eventRepo, eventReportRepo, memberRepo, secretKey))
	e.POST("/api/event-log/submit", submitEventLogHandler(eventRepo, eventReportRepo, guildRepo, botClient, secretKey))
}

// eventLogMemberRow is the minimal member data returned to the log-event form.
type eventLogMemberRow struct {
	DiscordID  string `json:"discordId"`
	Username   string `json:"username"`
	AvatarHash string `json:"avatarHash"`
}

// validateEventLogHandler validates a token and returns event/member data for the log form.
// GET /api/event-log/validate?token=<jwt>
//
// Success:  { ok: true, event, preSelectedIds, members }
// Blocked:  { ok: false, reason: "expired" | "not_found" | "already_submitted" }
func validateEventLogHandler(
	eventRepo repositories.EventRepository,
	eventReportRepo repositories.EventReportRepository,
	memberRepo repositories.MemberRepository,
	secretKey string,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenStr := strings.TrimSpace(c.QueryParam("token"))
		if tokenStr == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"ok": false, "reason": "missing_token",
			})
		}

		claims, err := session.VerifyEventLog(tokenStr, secretKey)
		if err != nil {
			c.Logger().Warnf("event-log validate: JWT error: %v", err)
			return c.JSON(http.StatusOK, map[string]interface{}{
				"ok": false, "reason": "expired",
			})
		}

		ctx := c.Request().Context()

		// Verify the event still exists, belongs to the correct guild, and is not closed.
		event, err := eventRepo.FindByID(ctx, claims.EventID)
		if err != nil || event == nil || event.GuildID != claims.GuildID {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"ok": false, "reason": "not_found",
			})
		}
		if event.Status == repositories.EventStatusClosed {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"ok": false, "reason": "already_submitted",
			})
		}

		// Block if a report already exists (idempotency guard).
		existing, _ := eventReportRepo.FindByEventID(ctx, claims.EventID)
		if existing != nil {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"ok": false, "reason": "already_submitted",
			})
		}

		// Fetch the guild member list so the form can display a participant picker.
		summaries, err := memberRepo.FindSummariesByGuildID(ctx, claims.GuildID)
		if err != nil {
			summaries = []repositories.MemberSummary{}
		}
		members := make([]eventLogMemberRow, 0, len(summaries))
		for _, m := range summaries {
			members = append(members, eventLogMemberRow{
				DiscordID:  m.DiscordID,
				Username:   m.Username,
				AvatarHash: m.AvatarHash,
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"ok":             true,
			"event":          event,
			"preSelectedIds": preSelectedIDs(event),
			"members":        members,
		})
	}
}

// submitEventLogHandler creates an EventReport and deletes the source event.
// POST /api/event-log/submit
// Body: { token, summary, participantIds }
//
// Success:  { ok: true }
// Blocked:  { ok: false, reason: "expired" | "not_found" | "already_submitted" }
func submitEventLogHandler(
	eventRepo repositories.EventRepository,
	eventReportRepo repositories.EventReportRepository,
	guildRepo repositories.GuildRepository,
	botClient *discord.BotClient,
	secretKey string,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		var in struct {
			Token          string   `json:"token"`
			Summary        string   `json:"summary"`
			ParticipantIDs []string `json:"participantIds"`
		}
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"ok": false, "error": "invalid request body",
			})
		}

		claims, err := session.VerifyEventLog(in.Token, secretKey)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"ok": false, "reason": "expired",
			})
		}

		ctx := c.Request().Context()

		// Re-validate event state to prevent replay on closed/deleted events.
		event, err := eventRepo.FindByID(ctx, claims.EventID)
		if err != nil || event == nil || event.GuildID != claims.GuildID {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"ok": false, "reason": "not_found",
			})
		}
		if event.Status == repositories.EventStatusClosed {
			return c.JSON(http.StatusConflict, map[string]interface{}{
				"ok": false, "reason": "already_submitted",
			})
		}

		// Idempotency: reject if a report was already created.
		existing, _ := eventReportRepo.FindByEventID(ctx, claims.EventID)
		if existing != nil {
			return c.JSON(http.StatusConflict, map[string]interface{}{
				"ok": false, "reason": "already_submitted",
			})
		}

		summary := strings.TrimSpace(in.Summary)
		if summary == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"ok": false, "error": "summary is required",
			})
		}

		if in.ParticipantIDs == nil {
			in.ParticipantIDs = []string{}
		}

		report := &repositories.EventReport{
			EventID:              claims.EventID,
			GuildID:              claims.GuildID,
			HostDiscordID:        claims.HostDiscordID,
			EventDate:            event.ScheduledAt,
			ParticipantIDs:       in.ParticipantIDs,
			Summary:              summary,
			SubmittedAt:          time.Now().UTC(),
			SubmittedByDiscordID: claims.HostDiscordID,
		}

		if err := eventReportRepo.Create(ctx, report); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"ok": false, "error": "failed to save event log",
			})
		}

		// Delete the live event — it is now fully logged.
		_ = eventRepo.Delete(ctx, claims.EventID)

		// Post embed to the configured logs channel in the background.
		capturedEvent := event
		capturedReport := report
		capturedLogger := c.Logger()
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			guild, gErr := guildRepo.FindByGuildID(bgCtx, capturedReport.GuildID)
			if gErr != nil || guild == nil || guild.EventConfig.LogsChannelID == "" {
				return
			}
			embed := buildEventLogEmbed(capturedEvent, capturedReport)
			if _, err := botClient.PostEmbedMessage(bgCtx, guild.EventConfig.LogsChannelID, "", []discord.Embed{embed}, nil); err != nil {
				capturedLogger.Errorf("event-log: post to logs channel %s: %v", guild.EventConfig.LogsChannelID, err)
			}
		}()

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// buildEventLogEmbed constructs the Discord embed posted to the guild's logs channel.
// Field order: header (Event Type / Date / Host inline) → Summary → Participants → Footer.
func buildEventLogEmbed(event *repositories.Event, report *repositories.EventReport) discord.Embed {
	mentions := make([]string, len(report.ParticipantIDs))
	for i, id := range report.ParticipantIDs {
		mentions[i] = "<@" + id + ">"
	}
	participantValue := strings.Join(mentions, " ")
	if participantValue == "" {
		participantValue = "_(none recorded)_"
	}
	if len(participantValue) > 1024 {
		participantValue = participantValue[:1021] + "…"
	}

	title := event.Title
	if title == "" {
		title = event.EventType
	}

	return discord.Embed{
		Title: title,
		Color: 0x5865F2, // Discord blurple
		Fields: []discord.EmbedField{
			{Name: "Event Type", Value: event.EventType, Inline: true},
			{Name: "Date", Value: report.EventDate.UTC().Format("January 2, 2006"), Inline: true},
			{Name: "Host", Value: "<@" + report.HostDiscordID + ">", Inline: true},
			{Name: "Summary", Value: report.Summary},
			{Name: fmt.Sprintf("Participants (%d)", len(report.ParticipantIDs)), Value: participantValue},
		},
		Footer: &discord.EmbedFooter{
			Text: "Logged " + report.SubmittedAt.UTC().Format("Jan 2, 2006 15:04 UTC"),
		},
	}
}

// preSelectedIDs returns the best available participant list for the event log form.
// Voice channel members captured at event-end are used when available; otherwise
// the RSVP attending list is returned as the default selection.
func preSelectedIDs(event *repositories.Event) []string {
	if len(event.VoiceMemberIDs) > 0 {
		return event.VoiceMemberIDs
	}
	if event.AttendingIDs != nil {
		return event.AttendingIDs
	}
	return []string{}
}
