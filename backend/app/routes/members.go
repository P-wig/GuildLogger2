package routes

import (
	"net/http"
	"strings"

	"github.com/P-wig/GuildLogger2/backend/app/repositories"
	"github.com/labstack/echo/v4"
)

// RegisterMembersProtected registers all member routes on the JWT-guarded group.
func RegisterMembersProtected(g *echo.Group, memberRepo repositories.MemberRepository) {
	g.GET("/members/:discordId/stats", getMemberStatsHandler(memberRepo))
}

func getMemberStatsHandler(memberRepo repositories.MemberRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		discordID := strings.TrimSpace(c.Param("discordId"))
		guildID := strings.TrimSpace(c.QueryParam("guildId"))

		if discordID == "" || guildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "discordId path param and guildId query param are required"})
		}

		stats, err := memberRepo.GetStats(c.Request().Context(), guildID, discordID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "stats": stats})
	}
}
