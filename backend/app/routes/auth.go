package routes

import (
	"net/http"
	"strings"

	"github.com/P-wig/GuildLogger2/backend/app/repositories"
	"github.com/labstack/echo/v4"
)

type discordAuthPayload struct {
	DiscordID    string `json:"discordId"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func RegisterAuth(e *echo.Echo, userRepo repositories.UserRepository) {
	e.POST("/api/auth/discord", discordAuthHandler(userRepo))
	e.GET("/api/auth/users/:discordId", getUserByDiscordIDHandler(userRepo))
}

func discordAuthHandler(userRepo repositories.UserRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		var in discordAuthPayload
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"ok":    false,
				"error": "invalid request body",
			})
		}

		in.DiscordID = strings.TrimSpace(in.DiscordID)
		in.AccessToken = strings.TrimSpace(in.AccessToken)
		in.RefreshToken = strings.TrimSpace(in.RefreshToken)

		if in.DiscordID == "" || in.AccessToken == "" || in.RefreshToken == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"ok":    false,
				"error": "discordId, accessToken, and refreshToken are required",
			})
		}

		user, err := userRepo.CreateOrUpdateFromDiscord(
			c.Request().Context(),
			in.DiscordID,
			in.AccessToken,
			in.RefreshToken,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"ok":    false,
				"error": "failed to persist user",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"ok":   true,
			"user": user,
		})
	}
}

func getUserByDiscordIDHandler(userRepo repositories.UserRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		discordID := strings.TrimSpace(c.Param("discordId"))
		if discordID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"ok":    false,
				"error": "discordId is required",
			})
		}

		user, err := userRepo.FindByDiscordID(c.Request().Context(), discordID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"ok":    false,
				"error": "database error",
			})
		}
		if user == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"ok":    false,
				"error": "user not found",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"ok":   true,
			"user": user,
		})
	}
}
