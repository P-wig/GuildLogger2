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

		// Verify the event still exists and belongs to the correct guild.
		event, err := eventRepo.FindByID(ctx, claims.EventID)
		if err != nil || event == nil || event.GuildID != claims.GuildID {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"ok": false, "reason": "not_found",
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

		// Re-validate event state to prevent replay on deleted events.
		event, err := eventRepo.FindByID(ctx, claims.EventID)
		if err != nil || event == nil || event.GuildID != claims.GuildID {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"ok": false, "reason": "not_found",
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

		// Post embed to the configured logs channel in the background.
		capturedEvent := event
		capturedReport := report
		capturedLogger := c.Logger()
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			guild, gErr := guildRepo.FindByGuildID(bgCtx, capturedReport.GuildID)
			if gErr != nil || guild == nil {
				return
			}
			syncEventLogEmbed(bgCtx, capturedLogger, guild.EventConfig.LogsChannelID, eventReportRepo, botClient, capturedEvent, capturedReport)
		}()

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// buildEventLogEmbed constructs the Discord embed posted to the guild's logs channel.
// event is nil for logs created directly in the dashboard, which have no source event.
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

	title := "Event Log"
	fields := make([]discord.EmbedField, 0, 5)
	if event != nil {
		if event.Title != "" {
			title = event.Title
		} else if event.EventType != "" {
			title = event.EventType
		}
		fields = append(fields, discord.EmbedField{Name: "Event Type", Value: event.EventType, Inline: true})
	}
	fields = append(fields,
		discord.EmbedField{Name: "Date", Value: report.EventDate.UTC().Format("January 2, 2006"), Inline: true},
		discord.EmbedField{Name: "Host", Value: "<@" + report.HostDiscordID + ">", Inline: true},
		discord.EmbedField{Name: "Summary", Value: report.Summary},
		discord.EmbedField{Name: fmt.Sprintf("Participants (%d)", len(report.ParticipantIDs)), Value: participantValue},
	)

	return discord.Embed{
		Title:  title,
		Color:  0x5865F2, // Discord blurple
		Fields: fields,
		Footer: &discord.EmbedFooter{
			Text: "Logged " + report.SubmittedAt.UTC().Format("Jan 2, 2006 15:04 UTC"),
		},
	}
}

// syncEventLogEmbed upserts the report's embed in the logs channel: it edits the existing
// message when one is known and reposts otherwise, storing the new message reference.
func syncEventLogEmbed(
	ctx context.Context,
	logger echo.Logger,
	channelID string,
	reportRepo repositories.EventReportRepository,
	botClient *discord.BotClient,
	event *repositories.Event,
	report *repositories.EventReport,
) {
	if channelID == "" || report == nil {
		return
	}
	embed := buildEventLogEmbed(event, report)

	if report.LogsChannelID == channelID && report.LogsMessageID != "" {
		err := botClient.EditMessage(ctx, channelID, report.LogsMessageID, []discord.Embed{embed}, nil)
		if err == nil {
			return
		}
		logger.Warnf("event-log sync: edit failed for %s/%s, reposting: %v", channelID, report.LogsMessageID, err)
	}

	msgID, err := botClient.PostEmbedMessage(ctx, channelID, "", []discord.Embed{embed}, nil)
	if err != nil {
		logger.Errorf("event-log sync: post to %s failed: %v", channelID, err)
		return
	}
	if report.ID == "" {
		return
	}
	if err := reportRepo.SetLogMessageRef(ctx, report.ID, channelID, msgID); err != nil {
		logger.Errorf("event-log sync: failed to save message ref for %s: %v", report.ID, err)
		return
	}
	report.LogsChannelID = channelID
	report.LogsMessageID = msgID
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
