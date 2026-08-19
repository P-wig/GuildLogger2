package routes

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/P-wig/GuildLogger2/backend/app/discord"
	"github.com/P-wig/GuildLogger2/backend/app/repositories"
	"github.com/labstack/echo/v4"
)

type runAnniversaryPayload struct {
	GuildID string `json:"guildId"`
}

// RegisterNotificationsProtected registers all notification job routes on the JWT-guarded group.
func RegisterNotificationsProtected(g *echo.Group, guildRepo repositories.GuildRepository, memberRepo repositories.MemberRepository, botClient *discord.BotClient, eventRepo repositories.EventRepository) {
	g.POST("/notifications/members/anniversaries/run", runAnniversaryNotificationsHandler(guildRepo, memberRepo, botClient))
	g.POST("/notifications/events/reminders/run", runEventRemindersHandler(eventRepo, memberRepo, botClient))
}

func runAnniversaryNotificationsHandler(guildRepo repositories.GuildRepository, memberRepo repositories.MemberRepository, botClient *discord.BotClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		var in runAnniversaryPayload
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "invalid request body"})
		}
		in.GuildID = strings.TrimSpace(in.GuildID)
		if in.GuildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId is required"})
		}

		ctx := c.Request().Context()

		guild, err := guildRepo.FindByGuildID(ctx, in.GuildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if guild == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "guild not found"})
		}

		cfg := guild.NotificationConfig.MilestoneNotifications
		if !cfg.Enabled {
			return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "notified": 0, "skipped": "milestone notifications are disabled for this guild"})
		}
		if cfg.NotificationChannelID == "" {
			return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "error": "no notification channel configured for this guild"})
		}
		if len(cfg.AnniversaryYears) == 0 {
			return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "notified": 0, "skipped": "no anniversary years configured"})
		}

		members, err := memberRepo.FindAnniversaryMembers(ctx, in.GuildID, cfg.AnniversaryYears)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to query anniversary members"})
		}

		now := time.Now().UTC()
		notified := make([]string, 0, len(members))
		failed := make([]string, 0)

		for _, member := range members {
			years := now.Year() - member.DiscordJoinedAt.UTC().Year()
			msg := fmt.Sprintf("Happy %d-year anniversary <@%s>! Thank you for being a member of the server.", years, member.DiscordID)
			if err := botClient.SendChannelMessage(ctx, cfg.NotificationChannelID, msg); err != nil {
				failed = append(failed, member.DiscordID)
				continue
			}
			notified = append(notified, member.DiscordID)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"ok":       true,
			"notified": len(notified),
			"members":  notified,
			"failed":   failed,
		})
	}
}

// StartReminderScheduler launches a background goroutine that fires reminder DMs
// once per hour, aligned to the top of the clock hour. If the server starts at
// 12:50, the first fire is at 13:00; subsequent fires are at 14:00, 15:00, …
// This ensures an 8 pm event is always caught at the 7 pm tick, not at 7:50.
// ctx cancellation stops the goroutine cleanly on server shutdown.
func StartReminderScheduler(ctx context.Context, eventRepo repositories.EventRepository, botClient *discord.BotClient, logger echo.Logger) {
	go func() {
		// Sleep until the top of the next clock hour.
		now := time.Now()
		nextHour := now.Truncate(time.Hour).Add(time.Hour)
		initial := time.NewTimer(nextHour.Sub(now))
		defer initial.Stop()
		logger.Infof("reminder scheduler: first fire at %s", nextHour.Format("15:04:05 UTC"))

		select {
		case <-ctx.Done():
			return
		case <-initial.C:
		}

		// Fire immediately at the top of the first hour, then every hour.
		fireRemindersOnce(ctx, eventRepo, botClient, logger)
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fireRemindersOnce(ctx, eventRepo, botClient, logger)
			}
		}
	}()
}

// fireRemindersOnce runs one reminder pass: queries for Skirmish events in the
// next hour and sends DMs to their MaybeIDs. Called by both the scheduler and
// the manual HTTP trigger.
func fireRemindersOnce(ctx context.Context, eventRepo repositories.EventRepository, botClient *discord.BotClient, logger echo.Logger) {
	bgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	now := time.Now().UTC()
	events, err := eventRepo.FindUpcomingSkirmishForReminders(bgCtx, now, now.Add(time.Hour))
	if err != nil {
		logger.Errorf("reminder scheduler: query failed: %v", err)
		return
	}
	for _, event := range events {
		for _, discordID := range event.MaybeIDs {
			msg := fmt.Sprintf(
				"⏰ Reminder: The **%s** event starts <t:%d:R>! Don't forget to update your RSVP if your plans have changed.",
				event.Title, event.ScheduledAt.Unix(),
			)
			if err := botClient.SendDMToUser(bgCtx, discordID, msg); err != nil {
				logger.Errorf("reminder scheduler: DM to %s failed: %v", discordID, err)
			}
		}
		if err := eventRepo.MarkReminderSent(bgCtx, event.ID, now); err != nil {
			logger.Errorf("reminder scheduler: MarkReminderSent %s: %v", event.ID, err)
		}
	}
	if len(events) > 0 {
		logger.Infof("reminder scheduler: sent reminders for %d event(s)", len(events))
	}
}

// runEventRemindersHandler sends DM reminders to "maybe" attendees for Skirmish events
// that start within the next hour and have not yet received a reminder.
func runEventRemindersHandler(eventRepo repositories.EventRepository, memberRepo repositories.MemberRepository, botClient *discord.BotClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		now := time.Now().UTC()
		cutoff := now.Add(time.Hour)

		events, err := eventRepo.FindUpcomingSkirmishForReminders(ctx, now, cutoff)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"ok":    false,
				"error": "failed to query events",
			})
		}

		totalSent := 0
		var failures []string

		for _, event := range events {
			for _, discordID := range event.MaybeIDs {
				msg := fmt.Sprintf(
					"⏰ Reminder: The **%s** event starts <t:%d:R>! Don't forget to update your RSVP if your plans have changed.",
					event.Title, event.ScheduledAt.Unix(),
				)
				if err := botClient.SendDMToUser(ctx, discordID, msg); err == nil {
					totalSent++
				} else {
					failures = append(failures, discordID)
				}
			}
			if err := eventRepo.MarkReminderSent(ctx, event.ID, now); err != nil {
				failures = append(failures, "mark:"+event.ID)
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"ok":     true,
			"sent":   totalSent,
			"failed": failures,
			"events": len(events),
		})
	}
}
