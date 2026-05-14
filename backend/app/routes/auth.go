package routes

import (
	"net/http"
	"strings"

	"github.com/P-wig/GuildLogger2/backend/app/discord"
	"github.com/P-wig/GuildLogger2/backend/app/repositories"
	"github.com/P-wig/GuildLogger2/backend/app/session"
	"github.com/labstack/echo/v4"
)

type discordLoginPayload struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirectUri"`
}

func RegisterAuth(e *echo.Echo, userRepo repositories.UserRepository, oauthClient *discord.OAuthClient, secretKey string) {
	e.GET("/api/auth/discord/url", getDiscordAuthURLHandler(oauthClient))
	e.POST("/api/auth/discord/login", discordLoginHandler(userRepo, oauthClient, secretKey))
	e.GET("/api/auth/session", getSessionHandler(userRepo, secretKey))
	e.POST("/api/auth/logout", logoutHandler())
	e.GET("/api/auth/users/:discordId", getUserByDiscordIDHandler(userRepo))
}

// getDiscordAuthURLHandler returns the Discord authorization URL the frontend redirects the user to.
func getDiscordAuthURLHandler(oauthClient *discord.OAuthClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		redirectURI := strings.TrimSpace(c.QueryParam("redirectUri"))
		if redirectURI == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"ok":    false,
				"error": "redirectUri query parameter is required",
			})
		}

		authURL := oauthClient.AuthorizeURL(redirectURI, "")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"ok":  true,
			"url": authURL,
		})
	}
}

// discordLoginHandler exchanges a Discord authorization code for tokens,
// fetches the user's Discord identity, and persists the user record.
func discordLoginHandler(userRepo repositories.UserRepository, oauthClient *discord.OAuthClient, secretKey string) echo.HandlerFunc {
	return func(c echo.Context) error {
		var in discordLoginPayload
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"ok":    false,
				"error": "invalid request body",
			})
		}

		in.Code = strings.TrimSpace(in.Code)
		in.RedirectURI = strings.TrimSpace(in.RedirectURI)

		if in.Code == "" || in.RedirectURI == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"ok":    false,
				"error": "code and redirectUri are required",
			})
		}

		ctx := c.Request().Context()

		tokens, err := oauthClient.ExchangeCode(ctx, in.Code, in.RedirectURI)
		if err != nil {
			return c.JSON(http.StatusBadGateway, map[string]interface{}{
				"ok":    false,
				"error": "failed to exchange code with Discord",
			})
		}

		discordUser, err := oauthClient.GetCurrentUser(ctx, tokens.AccessToken)
		if err != nil {
			return c.JSON(http.StatusBadGateway, map[string]interface{}{
				"ok":    false,
				"error": "failed to fetch Discord user profile",
			})
		}

		user, err := userRepo.CreateOrUpdateFromDiscord(
			ctx,
			discordUser.ID,
			tokens.AccessToken,
			tokens.RefreshToken,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"ok":    false,
				"error": "failed to persist user",
			})
		}

		token, err := session.Sign(user.DiscordID, secretKey)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"ok":    false,
				"error": "failed to create session",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"ok":    true,
			"token": token,
			"user":  user,
		})
	}
}

// getSessionHandler validates the JWT from the Authorization header and returns the current user.
func getSessionHandler(userRepo repositories.UserRepository, secretKey string) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == "" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"ok":    false,
				"error": "missing token",
			})
		}

		claims, err := session.Verify(tokenStr, secretKey)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"ok":    false,
				"error": "invalid or expired token",
			})
		}

		user, err := userRepo.FindByDiscordID(c.Request().Context(), claims.DiscordID)
		if err != nil || user == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
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

// logoutHandler is stateless — the client discards the token.
// No server-side invalidation is needed until token blocklisting is required.
func logoutHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"ok": true,
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
