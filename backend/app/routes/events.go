package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/P-wig/GuildLogger2/backend/app/repositories"
	"github.com/P-wig/GuildLogger2/backend/app/session"
	"github.com/labstack/echo/v4"
)

type createEventPayload struct {
	GuildID   string    `json:"guildId"`
	EventDate time.Time `json:"eventDate"`
	Notes     string    `json:"notes"`
}

// RegisterEventsProtected registers all event routes on the JWT-guarded group.
func RegisterEventsProtected(g *echo.Group, eventRepo repositories.EventRepository) {
	g.POST("/events", createEventHandler(eventRepo))
	g.GET("/events", listEventsHandler(eventRepo))
	g.GET("/events/:eventId", getEventHandler(eventRepo))
	g.POST("/events/:eventId/register", registerForEventHandler(eventRepo))
	g.POST("/events/:eventId/unregister", unregisterFromEventHandler(eventRepo))
}

func createEventHandler(eventRepo repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		var in createEventPayload
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "invalid request body"})
		}

		in.GuildID = strings.TrimSpace(in.GuildID)
		if in.GuildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId is required"})
		}

		event := &repositories.Event{
			GuildID:       in.GuildID,
			HostDiscordID: claims.DiscordID,
			EventDate:     in.EventDate,
			Notes:         in.Notes,
			AttendingIDs:  []string{},
		}

		if err := eventRepo.Create(c.Request().Context(), event); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to create event"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "event": event})
	}
}

func listEventsHandler(eventRepo repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		guildID := strings.TrimSpace(c.QueryParam("guildId"))
		if guildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId query parameter is required"})
		}

		events, err := eventRepo.FindByGuildID(c.Request().Context(), guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "events": events})
	}
}

func getEventHandler(eventRepo repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		eventID := strings.TrimSpace(c.Param("eventId"))
		if eventID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "eventId is required"})
		}

		event, err := eventRepo.FindByEventID(c.Request().Context(), eventID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if event == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event not found"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "event": event})
	}
}

func registerForEventHandler(eventRepo repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		eventID := strings.TrimSpace(c.Param("eventId"))
		if eventID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "eventId is required"})
		}

		if err := eventRepo.AddAttendee(c.Request().Context(), eventID, claims.DiscordID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to register for event"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func unregisterFromEventHandler(eventRepo repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		eventID := strings.TrimSpace(c.Param("eventId"))
		if eventID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "eventId is required"})
		}

		if err := eventRepo.RemoveAttendee(c.Request().Context(), eventID, claims.DiscordID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to unregister from event"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}
