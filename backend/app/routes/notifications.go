package routes

import (
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
func RegisterNotificationsProtected(g *echo.Group, guildRepo repositories.GuildRepository, memberRepo repositories.MemberRepository, botClient *discord.BotClient) {
	g.POST("/notifications/members/anniversaries/run", runAnniversaryNotificationsHandler(guildRepo, memberRepo, botClient))
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
