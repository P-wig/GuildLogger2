package routes

import (
	"github.com/P-wig/GuildLogger2/backend/app/discord"
	"github.com/labstack/echo/v4"
)

// RegisterInteractions wires the Discord interaction webhook at POST /api/interactions.
// This endpoint is intentionally unprotected by JWT — authentication is performed inside
// the handler via Discord's Ed25519 signature verification.
func RegisterInteractions(e *echo.Echo, handler *discord.InteractionHandler) {
	e.POST("/api/interactions", handler.Handle)
}
