package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/P-wig/GuildLogger2/backend/app/repositories"
	"github.com/labstack/echo/v4"
)

// InteractionHandler processes Discord interaction webhook payloads.
// Inject dependencies via NewInteractionHandler and register Handle on the interactions route.
type InteractionHandler struct {
	guildRepo repositories.GuildRepository
	eventRepo repositories.EventRepository
	botClient *BotClient
	publicKey string // hex-encoded Ed25519 public key; empty = skip verification (dev)
}

// NewInteractionHandler creates an InteractionHandler with the given dependencies.
func NewInteractionHandler(
	guildRepo repositories.GuildRepository,
	eventRepo repositories.EventRepository,
	botClient *BotClient,
	publicKey string,
) *InteractionHandler {
	return &InteractionHandler{
		guildRepo: guildRepo,
		eventRepo: eventRepo,
		botClient: botClient,
		publicKey: publicKey,
	}
}

// Handle is the Echo handler for POST /api/interactions.
func (h *InteractionHandler) Handle(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "cannot read body"})
	}

	// Skip signature check when no public key is configured (local dev without bot token).
	if h.publicKey != "" {
		if err := VerifyRequest(c.Request(), body, h.publicKey); err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"error": "invalid signature"})
		}
	}

	var i Interaction
	if err := json.Unmarshal(body, &i); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "invalid payload"})
	}

	switch i.Type {
	case InteractionTypePing:
		return c.JSON(http.StatusOK, interactionResponse{Type: InteractionResponsePong})
	case InteractionTypeApplicationCommand:
		return h.handleCommand(c, &i)
	case InteractionTypeApplicationCommandAutocomplete:
		return h.handleAutocomplete(c, &i)
	case InteractionTypeModalSubmit:
		return h.handleModalSubmit(c, &i)
	case InteractionTypeMessageComponent:
		return h.handleComponent(c, &i)
	}
	return c.JSON(http.StatusOK, interactionResponse{Type: InteractionResponsePong})
}

// --- Shared helpers ---

func ephemeralMsg(content string) interactionResponse {
	return interactionResponse{
		Type: InteractionResponseChannelMessage,
		Data: map[string]interface{}{"content": content, "flags": MessageFlagEphemeral},
	}
}

func interactionUserID(i *Interaction) string {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}
	return ""
}

func getModalValue(i *Interaction, customID string) string {
	if i.Data == nil {
		return ""
	}
	for _, row := range i.Data.Components {
		for _, comp := range row.Components {
			if comp.CustomID == customID {
				return comp.Value
			}
		}
	}
	return ""
}

func parseStartTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if epoch, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(epoch, 0).UTC(), nil
	}
	for _, layout := range []string{"2006-01-02 15:04", "2006-01-02T15:04", "2006-01-02", time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse %q — use a Unix timestamp or YYYY-MM-DD HH:MM (24h UTC)", s)
}

// --- APPLICATION_COMMAND: /start event [eventtype] ---

func (h *InteractionHandler) handleCommand(c echo.Context, i *Interaction) error {
	if i.Data == nil || i.Data.Name != "start" {
		return c.JSON(http.StatusOK, ephemeralMsg("Unknown command."))
	}
	return h.handleStartCommand(c, i)
}

func (h *InteractionHandler) handleStartCommand(c echo.Context, i *Interaction) error {
	// Drill into the "event" subcommand to find the eventtype option value.
	var subOpts []InteractionDataOption
	for _, opt := range i.Data.Options {
		if opt.Name == "event" {
			subOpts = opt.Options
			break
		}
	}
	eventType := ""
	for _, opt := range subOpts {
		if opt.Name == "eventtype" {
			if v, ok := opt.Value.(string); ok {
				eventType = strings.TrimSpace(v)
			}
		}
	}
	if eventType == "" {
		return c.JSON(http.StatusOK, ephemeralMsg("Event type is required."))
	}

	// Respond with a modal — ephemeral window only the invoking user can see.
	return c.JSON(http.StatusOK, interactionResponse{
		Type: InteractionResponseModal,
		Data: map[string]interface{}{
			"custom_id": "start_event|" + eventType,
			"title":     "Start Event: " + eventType,
			"components": []TextInputRow{
				{Type: 1, Components: []TextInput{{
					Type:        4,
					CustomID:    "start_time",
					Style:       1,
					Label:       "Start Time",
					Required:    true,
					Placeholder: "Unix epoch or YYYY-MM-DD HH:MM (24h, UTC)",
					MaxLength:   30,
				}}},
				{Type: 1, Components: []TextInput{{
					Type:        4,
					CustomID:    "message",
					Style:       2,
					Label:       "Rally Message",
					Required:    false,
					Placeholder: "Motivate your team! (optional)",
					MaxLength:   500,
				}}},
			},
		},
	})
}

// --- APPLICATION_COMMAND_AUTOCOMPLETE: /start event eventtype ---

func (h *InteractionHandler) handleAutocomplete(c echo.Context, i *Interaction) error {
	focused := ""
	if i.Data != nil {
		for _, opt := range i.Data.Options {
			if opt.Name == "event" {
				for _, sub := range opt.Options {
					if sub.Name == "eventtype" && sub.Focused {
						if v, ok := sub.Value.(string); ok {
							focused = strings.ToLower(strings.TrimSpace(v))
						}
					}
				}
			}
		}
	}

	choices := make([]AutocompleteChoice, 0)
	if guild, err := h.guildRepo.FindByGuildID(c.Request().Context(), i.GuildID); err == nil && guild != nil {
		for _, t := range guild.EventConfig.EventTypes {
			if focused == "" || strings.Contains(strings.ToLower(t), focused) {
				choices = append(choices, AutocompleteChoice{Name: t, Value: t})
				if len(choices) >= 25 {
					break
				}
			}
		}
	}

	return c.JSON(http.StatusOK, interactionResponse{
		Type: InteractionResponseAutocomplete,
		Data: map[string]interface{}{"choices": choices},
	})
}

// --- MODAL_SUBMIT: start_event|{eventType} ---

func (h *InteractionHandler) handleModalSubmit(c echo.Context, i *Interaction) error {
	if i.Data == nil {
		return c.JSON(http.StatusOK, ephemeralMsg("Something went wrong."))
	}

	eventType, ok := strings.CutPrefix(i.Data.CustomID, "start_event|")
	if !ok || eventType == "" {
		return c.JSON(http.StatusOK, interactionResponse{Type: InteractionResponsePong})
	}

	startTimeStr := getModalValue(i, "start_time")
	description := strings.TrimSpace(getModalValue(i, "message"))

	// Validate start time before acknowledging — cheap, no I/O.
	startTime, err := parseStartTime(startTimeStr)
	if err != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Invalid start time: "+err.Error()))
	}

	// Acknowledge immediately with an ephemeral deferred response.
	// This releases the 3-second Discord deadline. The actual work (DB writes + Discord
	// API call) runs in a goroutine and edits the deferred message when done.
	appID := i.ApplicationID
	token := i.Token
	guildID := i.GuildID
	hostID := interactionUserID(i)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 14*time.Second)
		defer cancel()

		editResp := func(msg string) {
			_ = h.botClient.EditInteractionResponse(ctx, appID, token, msg)
		}

		guild, err := h.guildRepo.FindByGuildID(ctx, guildID)
		if err != nil || guild == nil {
			editResp("⚠️ Could not find guild configuration.")
			return
		}
		if guild.EventConfig.EventsChannelID == "" {
			editResp("⚠️ No events channel configured. Set one in the GuildLogger dashboard under Guild Configuration.")
			return
		}

		event := &repositories.Event{
			GuildID:       guildID,
			HostDiscordID: hostID,
			Title:         eventType,
			Description:   description,
			ScheduledAt:   startTime,
			Status:        repositories.EventStatusOpen,
			AttendingIDs:  []string{},
			ChannelID:     guild.EventConfig.EventsChannelID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := h.eventRepo.Create(ctx, event); err != nil {
			editResp("⚠️ Failed to create event record.")
			return
		}

		embed := buildEventEmbed(eventType, description, hostID, startTime.Unix(), []string{})
		msgID, err := h.botClient.PostEmbedMessage(ctx, guild.EventConfig.EventsChannelID, []Embed{embed}, buildRSVPButtons(event.ID))
		if err != nil {
			// Event created but announcement failed — still usable via API.
			editResp(fmt.Sprintf("✅ Event created (ID: %s) but announcement post failed: %v", event.ID, err))
			return
		}

		// Persist the Discord message ID so the embed can be updated when attendees change.
		event.AnnouncementMessageID = msgID
		event.UpdatedAt = time.Now()
		_ = h.eventRepo.Update(ctx, event.ID, event)

		editResp(fmt.Sprintf("✅ **%s** event posted in <#%s>.", eventType, guild.EventConfig.EventsChannelID))
	}()

	// Return the deferred ephemeral acknowledgement immediately.
	return c.JSON(http.StatusOK, interactionResponse{
		Type: InteractionResponseDeferredChannelMessage,
		Data: map[string]interface{}{"flags": MessageFlagEphemeral},
	})
}

// --- MESSAGE_COMPONENT: event_join|{eventID} / event_leave|{eventID} ---

func (h *InteractionHandler) handleComponent(c echo.Context, i *Interaction) error {
	if i.Data == nil {
		return c.JSON(http.StatusOK, ephemeralMsg("Something went wrong."))
	}

	discordID := interactionUserID(i)
	ctx := c.Request().Context()

	var action, eventID string
	if id, ok := strings.CutPrefix(i.Data.CustomID, "event_join|"); ok {
		action, eventID = "join", id
	} else if id, ok := strings.CutPrefix(i.Data.CustomID, "event_leave|"); ok {
		action, eventID = "leave", id
	} else {
		return c.JSON(http.StatusOK, interactionResponse{Type: InteractionResponsePong})
	}

	var repoErr error
	if action == "join" {
		repoErr = h.eventRepo.AddAttendee(ctx, eventID, discordID)
	} else {
		repoErr = h.eventRepo.RemoveAttendee(ctx, eventID, discordID)
	}

	switch repoErr {
	case repositories.ErrAlreadyRegistered:
		return c.JSON(http.StatusOK, ephemeralMsg("You're already attending this event!"))
	case repositories.ErrNotRegistered:
		return c.JSON(http.StatusOK, ephemeralMsg("You're not registered for this event."))
	case repositories.ErrEventNotFound:
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ This event no longer exists."))
	case nil:
		// fall through to rebuild the embed
	default:
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Failed to update attendance. Please try again."))
	}

	event, err := h.eventRepo.FindByID(ctx, eventID)
	if err != nil || event == nil {
		return c.JSON(http.StatusOK, interactionResponse{Type: InteractionResponseDeferredUpdate})
	}

	embed := buildEventEmbed(event.Title, event.Description, event.HostDiscordID, event.ScheduledAt.Unix(), event.AttendingIDs)

	// UPDATE_MESSAGE replaces the original RSVP embed in place, visible to all users in the channel.
	return c.JSON(http.StatusOK, interactionResponse{
		Type: InteractionResponseUpdateMessage,
		Data: map[string]interface{}{
			"embeds":     []Embed{embed},
			"components": buildRSVPButtons(event.ID),
		},
	})
}
