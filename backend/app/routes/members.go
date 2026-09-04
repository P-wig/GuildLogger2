package routes

import (
	"errors"
	"net/http"
	"strings"

	"github.com/P-wig/GuildLogger2/backend/app/discord"
	"github.com/P-wig/GuildLogger2/backend/app/repositories"
	"github.com/labstack/echo/v4"
)

// RegisterMembersProtected registers all member routes on the JWT-guarded group.
func RegisterMembersProtected(
	g *echo.Group,
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
	statsService *discord.StatsService,
) {
	g.GET("/members/:discordId/stats", getMemberStatsHandler(guildRepo, memberRepo, statsService))
}

// getMemberStatsHandler returns a member's activity profile. It serves the same data the
// /stats slash command renders, composed by the shared stats service.
func getMemberStatsHandler(
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
	statsService *discord.StatsService,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		discordID := strings.TrimSpace(c.Param("discordId"))
		guildID := strings.TrimSpace(c.QueryParam("guildId"))
		if discordID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "discordId path param is required"})
		}

		// Stats expose another member's activity, so the caller must belong to the guild.
		if _, _, resp := authorizeGuild(c, guildRepo, memberRepo, guildID, tierMember); resp != nil {
			return resp
		}

		profile, err := statsService.MemberProfile(c.Request().Context(), guildID, discordID)
		if errors.Is(err, discord.ErrMemberNotFound) {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "member not found in this guild"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "stats": profile})
	}
}
