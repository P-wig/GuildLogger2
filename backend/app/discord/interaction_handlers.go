package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/P-wig/GuildLogger2/backend/app/repositories"
	"github.com/P-wig/GuildLogger2/backend/app/session"
	"github.com/labstack/echo/v4"
)

// InteractionHandler processes Discord interaction webhook payloads.
// Inject dependencies via NewInteractionHandler and register Handle on the interactions route.
type InteractionHandler struct {
	guildRepo       repositories.GuildRepository
	memberRepo      repositories.MemberRepository
	eventRepo       repositories.EventRepository
	eventReportRepo repositories.EventReportRepository
	botClient       *BotClient
	events          *EventService
	stats           *StatsService
	publicKey       string // hex-encoded Ed25519 public key; empty = skip verification (dev)
	secretKey       string // JWT signing key used to generate event-log tokens
	appURL          string // frontend base URL, e.g. "https://app.example.com"
}

// NewInteractionHandler creates an InteractionHandler with the given dependencies.
func NewInteractionHandler(
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
	eventRepo repositories.EventRepository,
	eventReportRepo repositories.EventReportRepository,
	botClient *BotClient,
	events *EventService,
	statsService *StatsService,
	publicKey string,
	secretKey string,
	appURL string,
) *InteractionHandler {
	return &InteractionHandler{
		guildRepo:       guildRepo,
		memberRepo:      memberRepo,
		eventRepo:       eventRepo,
		eventReportRepo: eventReportRepo,
		botClient:       botClient,
		events:          events,
		stats:           statsService,
		publicKey:       publicKey,
		secretKey:       secretKey,
		appURL:          appURL,
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
			c.Logger().Warnf("interaction sig-verify failed: %v", err)
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"error": "invalid signature"})
		}
	}

	var i Interaction
	if err := json.Unmarshal(body, &i); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "invalid payload"})
	}

	c.Logger().Infof("interaction type=%d customID=%q guildID=%q", i.Type, func() string {
		if i.Data != nil {
			return i.Data.CustomID
		}
		return ""
	}(), i.GuildID)

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
	c.Logger().Warnf("unhandled interaction type=%d", i.Type)
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
		t := time.Unix(epoch, 0).UTC()
		if y := t.Year(); y < 2000 || y >= 10000 {
			return time.Time{}, fmt.Errorf("timestamp %q is out of range — use seconds, not milliseconds (e.g. 1753000000)", s)
		}
		return t, nil
	}
	for _, layout := range []string{"2006-01-02 15:04", "2006-01-02T15:04", "2006-01-02", time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse %q — use a Unix timestamp or YYYY-MM-DD HH:MM (24h UTC)", s)
}

// --- APPLICATION_COMMAND router ---

func (h *InteractionHandler) handleCommand(c echo.Context, i *Interaction) error {
	if i.Data == nil {
		return c.JSON(http.StatusOK, ephemeralMsg("Unknown command."))
	}
	// Enforce designated command channel if one is configured for this guild.
	if i.GuildID != "" {
		ctx := c.Request().Context()
		if guild, err := h.guildRepo.FindByGuildID(ctx, i.GuildID); err == nil && guild != nil {
			if ch := guild.EventConfig.CommandChannelID; ch != "" && i.ChannelID != ch {
				return c.JSON(http.StatusOK, ephemeralMsg(fmt.Sprintf("⚠️ This command can only be used in <#%s>.", ch)))
			}
		}
	}
	switch i.Data.Name {
	case "event":
		var subName string
		for _, opt := range i.Data.Options {
			if opt.Type == int(OptionTypeSubCommand) {
				subName = opt.Name
				break
			}
		}
		switch subName {
		case "create":
			return h.handleEventCreate(c, i)
		case "log":
			return h.handleEventLog(c, i)
		default:
			return c.JSON(http.StatusOK, ephemeralMsg("Unknown subcommand."))
		}
	case "stats":
		return h.handleStats(c, i)
	case "anniversary":
		return h.handleAnniversary(c, i)
	case "help":
		return h.handleHelp(c, i)
	default:
		return c.JSON(http.StatusOK, ephemeralMsg("Unknown command."))
	}
}

func (h *InteractionHandler) handleEventCreate(c echo.Context, i *Interaction) error {
	// Extract the eventtype option from the "create" subcommand.
	var subOpts []InteractionDataOption
	for _, opt := range i.Data.Options {
		if opt.Name == "create" {
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

	// Gate: invoking user must be an active registered member of this guild.
	discordID := interactionUserID(i)
	ctx := c.Request().Context()
	member, err := h.memberRepo.FindByGuildAndDiscordID(ctx, i.GuildID, discordID)
	if err != nil || member == nil || member.Status != repositories.MemberStatusActive {
		return c.JSON(http.StatusOK, ephemeralMsg(
			"⚠️ You must be a registered active member of this guild to create events.",
		))
	}

	// Look up whether this event type is configured as a quick event (minimal modal, auto-message).
	isQuick := false
	if guild, gErr := h.guildRepo.FindByGuildID(ctx, i.GuildID); gErr == nil && guild != nil {
		for _, et := range guild.EventConfig.EventTypes {
			if strings.EqualFold(et.Name, eventType) {
				isQuick = et.IsQuickEvent
				if et.ChannelID != "" && i.ChannelID != et.ChannelID {
					return c.JSON(http.StatusOK, ephemeralMsg(
						fmt.Sprintf("⚠️ /event create %s can only be used in <#%s>.", eventType, et.ChannelID),
					))
				}
				break
			}
		}
	}
	modalFields := []TextInputRow{
		{Type: 1, Components: []TextInput{{
			Type:        4,
			CustomID:    "start_time",
			Style:       1,
			Label:       "Start Time",
			Required:    true,
			Placeholder: "Unix epoch or YYYY-MM-DD HH:MM (24h, UTC)",
			MaxLength:   30,
		}}},
	}
	if !isQuick {
		modalFields = append(modalFields,
			TextInputRow{Type: 1, Components: []TextInput{{
				Type:        4,
				CustomID:    "message",
				Style:       2,
				Label:       "Rally Message",
				Required:    false,
				Placeholder: "Motivate your team! (optional)",
				MaxLength:   500,
			}}},
		)
	}
	// Respond with a modal — only the invoking user sees it.
	return c.JSON(http.StatusOK, interactionResponse{
		Type: InteractionResponseModal,
		Data: map[string]interface{}{
			"custom_id":  "start_event|" + eventType,
			"title":      "Create Event: " + eventType,
			"components": modalFields,
		},
	})
}

func (h *InteractionHandler) handleEventLog(c echo.Context, i *Interaction) error {
	// Extract the eventid option from the "log" subcommand.
	var subOpts []InteractionDataOption
	for _, opt := range i.Data.Options {
		if opt.Name == "log" {
			subOpts = opt.Options
			break
		}
	}
	eventID := ""
	for _, opt := range subOpts {
		if opt.Name == "eventid" {
			if v, ok := opt.Value.(string); ok {
				eventID = strings.TrimSpace(v)
			}
		}
	}
	if eventID == "" {
		return c.JSON(http.StatusOK, ephemeralMsg("Event ID is required."))
	}

	// Gate: invoking user must be an active registered member.
	discordID := interactionUserID(i)
	ctx := c.Request().Context()
	member, err := h.memberRepo.FindByGuildAndDiscordID(ctx, i.GuildID, discordID)
	if err != nil || member == nil || member.Status != repositories.MemberStatusActive {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ You must be a registered active member of this guild to log events."))
	}

	// Verify the event exists, belongs to this guild, and the caller is the host.
	event, err := h.eventRepo.FindByID(ctx, eventID)
	if err != nil || event == nil || event.GuildID != i.GuildID {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Event not found."))
	}
	if event.HostDiscordID != discordID {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Only the event host can log this event."))
	}
	if event.Status == repositories.EventStatusClosed {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ This event is already closed."))
	}

	// Ensure no report has already been submitted for this event.
	existing, _ := h.eventReportRepo.FindByEventID(ctx, eventID)
	if existing != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ This event has already been logged."))
	}

	// Generate a time-limited signed token the host can use in the webapp.
	token, err := session.SignEventLog(eventID, i.GuildID, discordID, h.secretKey)
	if err != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Failed to generate log link. Please try again."))
	}

	logURL := h.appURL + "/log-event?token=" + token
	return c.JSON(http.StatusOK, interactionResponse{
		Type: InteractionResponseChannelMessage,
		Data: map[string]interface{}{
			"flags":   MessageFlagEphemeral,
			"content": "✅ Use the link below to submit your event log (expires in 8 hours):\n" + logURL,
		},
	})
}

// --- APPLICATION_COMMAND: /stats ---

func (h *InteractionHandler) handleStats(c echo.Context, i *Interaction) error {
	ctx := c.Request().Context()

	targetID := interactionUserID(i)
	if i.Data != nil {
		for _, opt := range i.Data.Options {
			if opt.Name == "member" {
				if v, ok := opt.Value.(string); ok && v != "" {
					targetID = v
				}
			}
		}
	}
	if targetID == "" {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Could not determine the target member."))
	}

	profile, err := h.stats.MemberProfile(ctx, i.GuildID, targetID)
	if errors.Is(err, ErrMemberNotFound) {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ <@"+targetID+"> is not a registered member of this guild."))
	}
	if err != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Could not load stats. Please try again."))
	}

	rank := profile.RankName
	if rank == "" {
		rank = "_Unranked_"
	}
	status := "✅ Active"
	if profile.Status != repositories.MemberStatusActive {
		status = "💤 Inactive"
	}

	fields := []EmbedField{
		{Name: "Rank", Value: rank, Inline: true},
		{Name: "Status", Value: status, Inline: true},
		{Name: "Events Hosted", Value: strconv.FormatInt(profile.HostedCount, 10), Inline: true},
		{Name: "Events Attended", Value: strconv.FormatInt(profile.ParticipatedCount, 10), Inline: true},
	}
	if !profile.DiscordJoinedAt.IsZero() {
		fields = append(fields, EmbedField{
			Name:   "Joined Server",
			Value:  fmt.Sprintf("<t:%d:D> (%d days)", profile.DiscordJoinedAt.Unix(), profile.TenureDays()),
			Inline: false,
		})
	}
	if !profile.FirstSyncedAt.IsZero() {
		fields = append(fields, EmbedField{
			Name:   "Tracked Since",
			Value:  fmt.Sprintf("<t:%d:D>", profile.FirstSyncedAt.Unix()),
			Inline: false,
		})
	}

	return c.JSON(http.StatusOK, interactionResponse{
		Type: InteractionResponseChannelMessage,
		Data: map[string]interface{}{
			"flags": MessageFlagEphemeral,
			"embeds": []Embed{{
				Title: "📊 Member Stats",
				// Embed titles render mentions as raw text, so the mention goes in the description.
				Description: "<@" + profile.DiscordID + ">",
				Color:       0x5865F2,
				Fields:      fields,
			}},
		},
	})
}

// --- APPLICATION_COMMAND: /anniversary ---

func (h *InteractionHandler) handleAnniversary(c echo.Context, i *Interaction) error {
	ctx := c.Request().Context()
	discordID := interactionUserID(i)

	member, err := h.memberRepo.FindByGuildAndDiscordID(ctx, i.GuildID, discordID)
	if err != nil || member == nil || member.Status != repositories.MemberStatusActive {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ You must be a registered active member of this guild to run this command."))
	}

	guild, err := h.guildRepo.FindByGuildID(ctx, i.GuildID)
	if err != nil || guild == nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Could not find guild configuration."))
	}

	cfg := guild.NotificationConfig.MilestoneNotifications
	if !cfg.Enabled {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Anniversary notifications are not enabled for this guild."))
	}
	if cfg.NotificationChannelID == "" {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ No notification channel configured for this guild."))
	}
	if len(cfg.AnniversaryYears) == 0 {
		return c.JSON(http.StatusOK, ephemeralMsg("No anniversary years configured for this guild."))
	}

	members, err := h.memberRepo.FindAnniversaryMembers(ctx, i.GuildID, cfg.AnniversaryYears)
	if err != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Failed to query anniversary members."))
	}

	if len(members) == 0 {
		return c.JSON(http.StatusOK, ephemeralMsg("No anniversary milestones today."))
	}

	now := time.Now().UTC()
	notified := 0
	for _, m := range members {
		years := now.Year() - m.DiscordJoinedAt.UTC().Year()
		msg := fmt.Sprintf("Happy %d-year anniversary <@%s>! Thank you for being a member of the server.", years, m.DiscordID)
		if err := h.botClient.SendChannelMessage(ctx, cfg.NotificationChannelID, msg); err == nil {
			notified++
		}
	}

	return c.JSON(http.StatusOK, ephemeralMsg(fmt.Sprintf("✅ Sent anniversary messages to %d member(s).", notified)))
}

// --- APPLICATION_COMMAND: /help ---

func (h *InteractionHandler) handleHelp(c echo.Context, i *Interaction) error {
	fields := []EmbedField{
		{Name: "/event create [eventtype]", Value: "Create a new event and post an RSVP announcement."},
		{Name: "/event log [eventid]", Value: "Log a completed event and create a permanent record."},
		{Name: "/stats [@member]", Value: "View rank, tenure, and event participation (defaults to yourself)."},
		{Name: "/anniversary", Value: "Trigger anniversary milestone notifications."},
		{Name: "/help", Value: "Show this help message."},
	}

	return c.JSON(http.StatusOK, interactionResponse{
		Type: InteractionResponseChannelMessage,
		Data: map[string]interface{}{
			"flags": MessageFlagEphemeral,
			"embeds": []Embed{{
				Title:  "🤖 GuildLogger Commands",
				Color:  0x5865F2,
				Fields: fields,
			}},
		},
	})
}

// --- APPLICATION_COMMAND_AUTOCOMPLETE: /event create eventtype ---

func (h *InteractionHandler) handleAutocomplete(c echo.Context, i *Interaction) error {
	focused := ""
	if i.Data != nil {
		for _, opt := range i.Data.Options {
			if opt.Name == "create" {
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
			if focused == "" || strings.Contains(strings.ToLower(t.Name), focused) {
				choices = append(choices, AutocompleteChoice{Name: t.Name, Value: t.Name})
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

	if eventID, ok := strings.CutPrefix(i.Data.CustomID, "mail_event|"); ok {
		return h.handleMailModalSubmit(c, i, eventID)
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
	// Capture parsed values for the goroutine (closures over loop-free vars are safe).
	capturedDesc := description
	capturedStart := startTime
	capturedType := eventType

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 14*time.Second)
		defer cancel()

		editResp := func(msg string) {
			_ = h.botClient.EditInteractionResponse(ctx, appID, token, msg)
		}

		event, err := h.events.CreateEvent(ctx, CreateEventParams{
			GuildID:     guildID,
			HostID:      hostID,
			EventType:   capturedType,
			Description: capturedDesc,
			ScheduledAt: capturedStart,
		})
		switch {
		case errors.Is(err, ErrNoChannelConfigured):
			editResp("⚠️ No channel configured for event type \"" + capturedType + "\". Set one in the GuildLogger dashboard under Guild Configuration.")
		case errors.Is(err, ErrGuildNotConfigured):
			editResp("⚠️ Could not find guild configuration.")
		case errors.Is(err, ErrAnnouncementFailed):
			editResp(fmt.Sprintf("✅ Event created (ID: %s) but announcement post failed: %v", event.ID, err))
		case err != nil:
			editResp("⚠️ Failed to create event record.")
		default:
			editResp(fmt.Sprintf("✅ **%s** event posted in <#%s>.", capturedType, event.ChannelID))
		}
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
	} else if id, ok := strings.CutPrefix(i.Data.CustomID, "event_maybe|"); ok {
		action, eventID = "maybe", id
	} else if id, ok := strings.CutPrefix(i.Data.CustomID, "event_unmaybe|"); ok {
		action, eventID = "unmaybe", id
	} else if id, ok := strings.CutPrefix(i.Data.CustomID, "event_decline|"); ok {
		action, eventID = "decline", id
	} else if id, ok := strings.CutPrefix(i.Data.CustomID, "event_undecline|"); ok {
		action, eventID = "undecline", id
	} else if id, ok := strings.CutPrefix(i.Data.CustomID, "ctrl_start|"); ok {
		return h.handleCtrlStart(c, i, id)
	} else if id, ok := strings.CutPrefix(i.Data.CustomID, "ctrl_end|"); ok {
		return h.handleCtrlEnd(c, i, id)
	} else if id, ok := strings.CutPrefix(i.Data.CustomID, "ctrl_modmail|"); ok {
		return h.handleCtrlModMail(c, i, id)
	} else if id, ok := strings.CutPrefix(i.Data.CustomID, "ctrl_close_channel|"); ok {
		return h.handleCtrlCloseChannel(c, i, id)
	} else {
		return c.JSON(http.StatusOK, interactionResponse{Type: InteractionResponsePong})
	}

	updated, repoErr := h.events.SetRSVP(ctx, eventID, discordID, action)

	switch {
	case repoErr == nil:
		// fall through to rebuild the embed
	case errors.Is(repoErr, repositories.ErrEventNotFound):
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ This event no longer exists."))
	case errors.Is(repoErr, repositories.ErrEventAtCapacity):
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ This event is at capacity."))
	case errors.Is(repoErr, repositories.ErrAlreadyRegistered) ||
		errors.Is(repoErr, repositories.ErrAlreadyMaybe) ||
		errors.Is(repoErr, repositories.ErrAlreadyDeclined) ||
		errors.Is(repoErr, repositories.ErrNotRegistered) ||
		errors.Is(repoErr, repositories.ErrNotMaybe) ||
		errors.Is(repoErr, repositories.ErrNotDeclined):
		// User is already in the requested state — return the current embed unchanged
		// so Discord clears the loading indicator without showing an error.
		event, err := h.eventRepo.FindByID(ctx, eventID)
		if err != nil || event == nil {
			return c.JSON(http.StatusOK, interactionResponse{Type: InteractionResponseDeferredUpdate})
		}
		isQuick := h.events.IsQuickEvent(ctx, i.GuildID, event.EventType)
		noop := buildEventEmbed(event.EventType, isQuick, event.Description, event.HostDiscordID, eventID, string(event.Status), event.ScheduledAt.Unix(), event.AttendingIDs, event.MaybeIDs, event.NotAttendingIDs)
		return c.JSON(http.StatusOK, interactionResponse{
			Type: InteractionResponseUpdateMessage,
			Data: map[string]interface{}{
				"embeds":     []Embed{noop},
				"components": buildRSVPButtons(event.ID, isQuick, string(event.Status)),
			},
		})
	default:
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Failed to update attendance. Please try again."))
	}

	event := updated
	if event == nil {
		return c.JSON(http.StatusOK, interactionResponse{Type: InteractionResponseDeferredUpdate})
	}

	isQuick := h.events.IsQuickEvent(ctx, i.GuildID, event.EventType)
	embed := buildEventEmbed(event.EventType, isQuick, event.Description, event.HostDiscordID, eventID, string(event.Status), event.ScheduledAt.Unix(), event.AttendingIDs, event.MaybeIDs, event.NotAttendingIDs)

	// UPDATE_MESSAGE replaces the original RSVP embed in place, visible to all users in the channel.
	return c.JSON(http.StatusOK, interactionResponse{
		Type: InteractionResponseUpdateMessage,
		Data: map[string]interface{}{
			"embeds":     []Embed{embed},
			"components": buildRSVPButtons(event.ID, isQuick, string(event.Status)),
		},
	})
}

// --- MESSAGE_COMPONENT: ctrl_start|{eventID} ---

func (h *InteractionHandler) handleCtrlStart(c echo.Context, i *Interaction, eventID string) error {
	discordID := interactionUserID(i)
	ctx := c.Request().Context()

	event, err := h.events.StartEvent(ctx, eventID, i.GuildID, discordID)
	switch {
	case errors.Is(err, repositories.ErrEventNotFound) || errors.Is(err, ErrEventNotInGuild):
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Event not found."))
	case errors.Is(err, ErrNotHost):
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Only the event host can start this event."))
	case errors.Is(err, repositories.ErrInvalidStatusTransition):
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ This event has already been started or closed."))
	case err != nil:
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Failed to start event. Please try again."))
	}

	// Discord API work runs in the background so we stay inside the 3-second deadline.
	hostName := memberDisplayName(i.Member)
	startLogger := c.Logger()
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 14*time.Second)
		defer cancel()
		h.events.ApplyStartEffects(bgCtx, event, hostName, startLogger)
	}()

	return c.JSON(http.StatusOK, ephemeralMsg("▶️ Event started! The voice channel is now open."))
}

// --- MESSAGE_COMPONENT: ctrl_end|{eventID} ---

func (h *InteractionHandler) handleCtrlEnd(c echo.Context, i *Interaction, eventID string) error {
	discordID := interactionUserID(i)
	ctx := c.Request().Context()

	// Generate the log token first so a signing failure never leaves a closed event
	// without a way to submit its log.
	token, err := session.SignEventLog(eventID, i.GuildID, discordID, h.secretKey)
	if err != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Failed to generate log link. Please try again."))
	}

	event, err := h.events.EndEvent(ctx, eventID, i.GuildID, discordID, c.Logger())
	switch {
	case errors.Is(err, repositories.ErrEventNotFound) || errors.Is(err, ErrEventNotInGuild):
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Event not found."))
	case errors.Is(err, ErrNotHost):
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Only the event host can end this event."))
	case errors.Is(err, ErrAlreadyLogged):
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ This event has already been logged."))
	case errors.Is(err, repositories.ErrInvalidStatusTransition):
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Start the event before ending it."))
	case err != nil:
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Failed to close event. Please try again."))
	}

	bgLogger := c.Logger()
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 14*time.Second)
		defer cancel()
		h.events.ApplyEndEffects(bgCtx, event, bgLogger)
	}()

	logURL := h.appURL + "/log-event?token=" + token
	return c.JSON(http.StatusOK, interactionResponse{
		Type: InteractionResponseChannelMessage,
		Data: map[string]interface{}{
			"flags":   MessageFlagEphemeral,
			"content": "⏹️ Event ended. Use the link below to submit your event log (expires in 8 hours):\n" + logURL,
		},
	})
}

// memberDisplayName returns the best available display name for a guild member.
// Priority: global username > Discord username > server nickname.
func memberDisplayName(m *InteractionMember) string {
	if m == nil {
		return "event"
	}
	if m.User != nil {
		if m.User.GlobalName != "" {
			return m.User.GlobalName
		}
		if m.User.Username != "" {
			return m.User.Username
		}
	}
	if m.Nick != "" {
		return m.Nick
	}
	return "event"
}

// --- MESSAGE_COMPONENT: ctrl_modmail|{eventID} ---

func (h *InteractionHandler) handleCtrlModMail(c echo.Context, i *Interaction, eventID string) error {
	discordID := interactionUserID(i)
	ctx := c.Request().Context()

	event, err := h.eventRepo.FindByID(ctx, eventID)
	if err != nil || event == nil || event.GuildID != i.GuildID {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Event not found."))
	}
	if event.HostDiscordID != discordID {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Only the event host can send mail."))
	}
	if event.ModMailSentAt != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Mail has already been sent for this event."))
	}

	return c.JSON(http.StatusOK, interactionResponse{
		Type: InteractionResponseModal,
		Data: map[string]interface{}{
			"custom_id": "mail_event|" + eventID,
			"title":     "Send Mail",
			"components": []TextInputRow{
				{
					Type: 1,
					Components: []TextInput{
						{
							Type:        4,
							CustomID:    "message",
							Style:       2, // paragraph
							Label:       "Message",
							Required:    true,
							Placeholder: "Enter your message to non-responding members...",
							MaxLength:   1000,
						},
					},
				},
			},
		},
	})
}

// --- MODAL_SUBMIT: mail_event|{eventID} ---

func (h *InteractionHandler) handleMailModalSubmit(c echo.Context, i *Interaction, eventID string) error {
	discordID := interactionUserID(i)
	ctx := c.Request().Context()

	event, err := h.eventRepo.FindByID(ctx, eventID)
	if err != nil || event == nil || event.GuildID != i.GuildID {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Event not found."))
	}
	if event.HostDiscordID != discordID {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Only the event host can send mail."))
	}
	if event.ModMailSentAt != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Mail has already been sent for this event."))
	}

	message := strings.TrimSpace(getModalValue(i, "message"))
	if message == "" {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Message cannot be empty."))
	}

	// Mark sent before dispatch to prevent double-send on rapid re-submit.
	now := time.Now()
	if err := h.eventRepo.MarkModMailSent(ctx, eventID, now); err != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Could not lock mail. Please try again."))
	}

	capturedEvent := event
	capturedMessage := message
	capturedGuildID := i.GuildID
	capturedHostName := memberDisplayName(i.Member)
	bgLogger := c.Logger()
	go func() {
		// Long timeout: 500 ms gap × potentially many members.
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		allMembers, err := h.memberRepo.FindByGuildID(bgCtx, capturedGuildID)
		if err != nil {
			bgLogger.Errorf("mail: load members %s: %v", capturedGuildID, err)
			return
		}

		responded := make(map[string]bool)
		for _, id := range capturedEvent.AttendingIDs {
			responded[id] = true
		}
		for _, id := range capturedEvent.MaybeIDs {
			responded[id] = true
		}
		for _, id := range capturedEvent.NotAttendingIDs {
			responded[id] = true
		}

		epoch := capturedEvent.ScheduledAt.Unix()
		fields := []EmbedField{
			{Name: "📅 Scheduled", Value: fmt.Sprintf("<t:%d:F>", epoch), Inline: true},
		}
		if capturedEvent.AnnouncementMessageID != "" && capturedEvent.ChannelID != "" {
			link := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", capturedGuildID, capturedEvent.ChannelID, capturedEvent.AnnouncementMessageID)
			fields = append(fields, EmbedField{Name: "📌 Announcement", Value: "[Jump to event](" + link + ")", Inline: true})
		}
		embed := Embed{
			Title:       "📢 " + capturedEvent.Title + " — RSVP Reminder",
			Description: capturedMessage,
			Color:       0x5865F2,
			Fields:      fields,
			Footer:      &EmbedFooter{Text: "Hosted by " + capturedHostName},
		}

		sent, blocked, failed := 0, 0, 0
		for _, m := range allMembers {
			if m.Status != repositories.MemberStatusActive || responded[m.DiscordID] || m.NotificationsOptOut {
				continue
			}
			if err := h.botClient.SendEmbedDMToUser(bgCtx, m.DiscordID, embed); err != nil {
				e := err.Error()
				switch {
				case strings.Contains(e, "50278"):
					bgLogger.Warnf("mail: skipped %s — not in a mutual guild (likely left server): %v", m.DiscordID, err)
					blocked++
				case strings.Contains(e, "50007") || strings.Contains(e, "status 403"):
					bgLogger.Warnf("mail: skipped %s — DMs disabled (privacy settings): %v", m.DiscordID, err)
					blocked++
				default:
					bgLogger.Errorf("mail: DM to %s failed: %v", m.DiscordID, err)
					failed++
				}
			} else {
				sent++
			}
			time.Sleep(500 * time.Millisecond)
		}
		bgLogger.Infof("mail: event=%s sent=%d blocked=%d failed=%d", capturedEvent.ID, sent, blocked, failed)
	}()

	return c.JSON(http.StatusOK, ephemeralMsg("📧 Mail is being sent to non-responding active members."))
}

// --- MESSAGE_COMPONENT: ctrl_close_channel|{eventID} ---

func (h *InteractionHandler) handleCtrlCloseChannel(c echo.Context, i *Interaction, eventID string) error {
	discordID := interactionUserID(i)
	ctx := c.Request().Context()

	event, err := h.events.CloseEventChannel(ctx, eventID, i.GuildID, discordID)
	switch {
	case errors.Is(err, repositories.ErrEventNotFound) || errors.Is(err, ErrEventNotInGuild):
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Event not found."))
	case errors.Is(err, ErrNotHost):
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Only the event host can close the channel."))
	case errors.Is(err, repositories.ErrInvalidStatusTransition):
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ End the event before closing the channel."))
	case err != nil:
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Failed to close the channel. Please try again."))
	}

	isQuick := h.events.IsQuickEvent(ctx, i.GuildID, event.EventType)
	embed := buildEventEmbed(event.EventType, isQuick, event.Description, event.HostDiscordID, eventID, string(event.Status), event.ScheduledAt.Unix(), event.AttendingIDs, event.MaybeIDs, event.NotAttendingIDs)

	bgLogger := c.Logger()
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		h.events.ApplyCloseChannelEffects(bgCtx, event, bgLogger)
	}()

	return c.JSON(http.StatusOK, interactionResponse{
		Type: InteractionResponseUpdateMessage,
		Data: map[string]interface{}{
			"embeds":     []Embed{embed},
			"components": buildRSVPButtons(eventID, isQuick, "done"),
		},
	})
}

// memberChannelName builds a Discord voice channel name from a host display name
// and event type, e.g. "picklewig's skirms".
func memberChannelName(hostName, eventType string) string {
	name := strings.ToLower(hostName) + "'s " + strings.ToLower(eventType)
	if len(name) > 100 {
		name = name[:100]
	}
	return name
}
