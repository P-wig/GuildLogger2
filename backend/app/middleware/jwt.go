package middleware

import (
	"net/http"
	"strings"

	"github.com/P-wig/GuildLogger2/backend/app/config"
	"github.com/P-wig/GuildLogger2/backend/app/session"
	"github.com/labstack/echo/v4"
)

// JWTMiddleware validates JWT from the Authorization header and attaches user info to context.
func JWTMiddleware(cfg config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid Authorization header")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := session.Verify(tokenString, cfg.SecretKey)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			// Attach claims (or user info) to context for handlers to use
			c.Set("user", claims)
			return next(c)
		}
	}
}
