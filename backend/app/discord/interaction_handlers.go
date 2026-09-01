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

func (h *InteractionHandler) handleEventList(c echo.Context, i *Interaction) error {
	ctx := c.Request().Context()
	allEvents, err := h.eventRepo.FindByGuildID(ctx, i.GuildID)
	if err != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Failed to retrieve events."))
	}

	// Filter to open/active only, sort by ScheduledAt ascending.
	var upcoming []repositories.Event
	for _, e := range allEvents {
		if e.Status == repositories.EventStatusOpen || e.Status == repositories.EventStatusActive {
			upcoming = append(upcoming, e)
		}
	}
	for i := 1; i < len(upcoming); i++ {
		for j := i; j > 0 && upcoming[j].ScheduledAt.Before(upcoming[j-1].ScheduledAt); j-- {
			upcoming[j], upcoming[j-1] = upcoming[j-1], upcoming[j]
		}
	}

	if len(upcoming) == 0 {
		return c.JSON(http.StatusOK, ephemeralMsg("No upcoming events scheduled."))
	}

	const maxShown = 10
	shown := upcoming
	footerText := fmt.Sprintf("%d event(s) scheduled", len(upcoming))
	if len(upcoming) > maxShown {
		shown = upcoming[:maxShown]
		footerText = fmt.Sprintf("Showing %d of %d upcoming events", maxShown, len(upcoming))
	}

	fields := make([]EmbedField, 0, len(shown))
	for _, e := range shown {
		attendees := len(e.AttendingIDs)
		capStr := "∞"
		if e.Capacity > 0 {
			capStr = strconv.Itoa(e.Capacity)
		}
		statusEmoji := "🟢"
		if e.Status == repositories.EventStatusActive {
			statusEmoji = "🔵"
		}
		fields = append(fields, EmbedField{
			Name: statusEmoji + " " + e.Title,
			Value: fmt.Sprintf(
				"Host: <@%s>\nStarts: <t:%d:F>\nAttending: %d / %s",
				e.HostDiscordID,
				e.ScheduledAt.Unix(),
				attendees,
				capStr,
			),
		})
	}

	return c.JSON(http.StatusOK, interactionResponse{
		Type: InteractionResponseChannelMessage,
		Data: map[string]interface{}{
			"flags": MessageFlagEphemeral,
			"embeds": []Embed{{
				Title:  "📅 Upcoming Events",
				Color:  0x5865F2, // Discord blurple
				Fields: fields,
				Footer: &EmbedFooter{Text: footerText},
			}},
		},
	})
}

// --- APPLICATION_COMMAND: /log ---

func (h *InteractionHandler) handleLog(c echo.Context, i *Interaction) error {
	discordID := interactionUserID(i)
	ctx := c.Request().Context()
	member, err := h.memberRepo.FindByGuildAndDiscordID(ctx, i.GuildID, discordID)
	if err != nil || member == nil || member.Status != repositories.MemberStatusActive {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ You must be a registered active member of this guild to log events."))
	}

	return c.JSON(http.StatusOK, interactionResponse{
		Type: InteractionResponseModal,
		Data: map[string]interface{}{
			"custom_id": "log_event",
			"title":     "Log Completed Event",
			"components": []TextInputRow{
				{Type: 1, Components: []TextInput{{
					Type:        4,
					CustomID:    "date",
					Style:       1,
					Label:       "Event Date",
					Required:    true,
					Placeholder: "YYYY-MM-DD",
					MaxLength:   10,
				}}},
				{Type: 1, Components: []TextInput{{
					Type:        4,
					CustomID:    "summary",
					Style:       2,
					Label:       "Summary",
					Required:    true,
					Placeholder: "What happened at this event?",
					MaxLength:   1000,
				}}},
			},
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

	stats, err := h.memberRepo.GetStats(ctx, i.GuildID, targetID)
	if err != nil || stats == nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Could not find stats for that member. They may not be registered."))
	}

	fields := []EmbedField{
		{Name: "Events Hosted", Value: strconv.FormatInt(stats.HostedCount, 10), Inline: true},
		{Name: "Events Attended", Value: strconv.FormatInt(stats.ParticipatedCount, 10), Inline: true},
		{Name: "Discord Joined", Value: "<t:" + strconv.FormatInt(stats.DiscordJoinedAt.Unix(), 10) + ":D>", Inline: false},
	}

	return c.JSON(http.StatusOK, interactionResponse{
		Type: InteractionResponseChannelMessage,
		Data: map[string]interface{}{
			"flags": MessageFlagEphemeral,
			"embeds": []Embed{{
				Title:  "📊 Stats for <@" + targetID + ">",
				Color:  0x5865F2,
				Fields: fields,
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
		{Name: "/stats [@member]", Value: "View event participation stats (defaults to yourself)."},
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

// --- MODAL_SUBMIT: log_event ---

func (h *InteractionHandler) handleLogModalSubmit(c echo.Context, i *Interaction, eventID string) error {
	dateStr := strings.TrimSpace(getModalValue(i, "date"))
	summary := strings.TrimSpace(getModalValue(i, "summary"))
	if dateStr == "" || summary == "" {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Date and summary are required."))
	}

	eventDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Invalid date format. Use YYYY-MM-DD."))
	}

	ctx := c.Request().Context()
	discordID := interactionUserID(i)

	// Use the event's host if found; fall back to the submitter.
	hostID := discordID
	if event, err := h.eventRepo.FindByID(ctx, eventID); err == nil && event != nil {
		hostID = event.HostDiscordID
	}

	report := &repositories.EventReport{
		EventID:              eventID,
		GuildID:              i.GuildID,
		HostDiscordID:        hostID,
		EventDate:            eventDate.UTC(),
		ParticipantIDs:       []string{},
		Summary:              summary,
		SubmittedByDiscordID: discordID,
	}

	if err := h.eventReportRepo.Create(ctx, report); err != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Failed to save event log."))
	}

	return c.JSON(http.StatusOK, ephemeralMsg(fmt.Sprintf("✅ Event logged for %s.", eventDate.Format("January 2, 2006"))))
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

	if eventID, ok := strings.CutPrefix(i.Data.CustomID, "log_event|"); ok {
		return h.handleLogModalSubmit(c, i, eventID)
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
	capacity := 0

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
	capturedCapacity := capacity
	capturedDesc := description
	capturedStart := startTime
	capturedType := eventType

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

		// Resolve the channel configured for this event type.
		channelID := ""
		isQuick := false
		for _, et := range guild.EventConfig.EventTypes {
			if strings.EqualFold(et.Name, capturedType) {
				channelID = et.ChannelID
				isQuick = et.IsQuickEvent
				break
			}
		}
		if channelID == "" {
			editResp("⚠️ No channel configured for event type \"" + capturedType + "\". Set one in the GuildLogger dashboard under Guild Configuration.")
			return
		}

		// Large events use a fixed capacity of 99.
		if !isQuick {
			capturedCapacity = 99
		}

		// Build the announcement message content that appears above the embed.
		// This must be plain text (not inside the embed) so the @role mention
		// fires a push notification for both quick and non-quick event types.
		epoch := capturedStart.Unix()
		activeRoleID := guild.NotificationConfig.StatusRoles.ActiveRoleID
		typeName := strings.ToLower(capturedType)
		messageContent := ""
		if activeRoleID != "" {
			messageContent = fmt.Sprintf("<@%s> is hosting a %s at <t:%d:t> <t:%d:R>", hostID, typeName, epoch, epoch)
			// comment out the role mention to avoid unnecessary pings during testing ONLY, switch back later
			//messageContent = fmt.Sprintf("<@&%s> <@%s> is hosting a %s at <t:%d:t> <t:%d:R>", activeRoleID, hostID, typeName, epoch, epoch)
		} else {
			messageContent = fmt.Sprintf("<@%s> is hosting a %s at <t:%d:t> <t:%d:R>", hostID, typeName, epoch, epoch)
			//messageContent = fmt.Sprintf("<@&%s> <@%s> is hosting a %s at <t:%d:t> <t:%d:R>", activeRoleID, hostID, typeName, epoch, epoch)
		}
		if isQuick {
			capturedDesc = "" // Quick events carry no user description; the message content is the announcement
		}

		event := &repositories.Event{
			GuildID:         guildID,
			HostDiscordID:   hostID,
			Title:           capturedType,
			EventType:       capturedType,
			Description:     capturedDesc,
			ScheduledAt:     capturedStart,
			Status:          repositories.EventStatusOpen,
			AttendingIDs:    []string{},
			MaybeIDs:        []string{},
			NotAttendingIDs: []string{},
			ChannelID:       channelID,
			Capacity:        capturedCapacity,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		if err := h.eventRepo.Create(ctx, event); err != nil {
			editResp("⚠️ Failed to create event record.")
			return
		}

		embed := buildEventEmbed(capturedType, isQuick, capturedDesc, hostID, event.ID, "open", capturedStart.Unix(), []string{}, []string{}, []string{})
		msgID, err := h.botClient.PostEmbedMessage(ctx, channelID, messageContent, []Embed{embed}, buildRSVPButtons(event.ID, isQuick, "open"))
		if err != nil {
			// Event created but announcement failed — still usable via API.
			editResp(fmt.Sprintf("✅ Event created (ID: %s) but announcement post failed: %v", event.ID, err))
			return
		}

		// Persist the Discord message ID so the embed can be updated when attendees change.
		event.AnnouncementMessageID = msgID
		event.UpdatedAt = time.Now()
		_ = h.eventRepo.Update(ctx, event.ID, event)

		editResp(fmt.Sprintf("✅ **%s** event posted in <#%s>.", capturedType, channelID))
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

	var repoErr error
	switch action {
	case "join":
		repoErr = h.eventRepo.AddAttendee(ctx, eventID, discordID)
	case "leave":
		repoErr = h.eventRepo.RemoveAttendee(ctx, eventID, discordID)
	case "maybe":
		repoErr = h.eventRepo.AddMaybe(ctx, eventID, discordID)
	case "unmaybe":
		repoErr = h.eventRepo.RemoveMaybe(ctx, eventID, discordID)
	case "decline":
		repoErr = h.eventRepo.AddDecline(ctx, eventID, discordID)
	case "undecline":
		repoErr = h.eventRepo.RemoveDecline(ctx, eventID, discordID)
	}

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
		isQuick := h.lookupIsQuick(ctx, i.GuildID, event.EventType)
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

	event, err := h.eventRepo.FindByID(ctx, eventID)
	if err != nil || event == nil {
		return c.JSON(http.StatusOK, interactionResponse{Type: InteractionResponseDeferredUpdate})
	}

	isQuick := h.lookupIsQuick(ctx, i.GuildID, event.EventType)
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

// collectVoiceChannelMembers returns the user IDs currently in channelID.
// It first tries the guild-wide voice-states endpoint; if that returns an
// error or no members in the target channel it falls back to querying each
// candidate user individually (attendees + maybe-RSVPs).
func (h *InteractionHandler) collectVoiceChannelMembers(
	ctx context.Context,
	guildID, channelID string,
	candidates []string,
	logger echo.Logger,
) []string {
	// Attempt bulk lookup.
	states, err := h.botClient.GetGuildVoiceStates(ctx, guildID)
	if err == nil {
		var ids []string
		for _, vs := range states {
			if vs.ChannelID == channelID && vs.UserID != "" {
				ids = append(ids, vs.UserID)
			}
		}
		if len(ids) > 0 {
			return ids
		}
		logger.Infof("collectVoiceChannelMembers: bulk endpoint returned 0 members in %s — falling back to per-user lookup", channelID)
	} else {
		logger.Errorf("collectVoiceChannelMembers: bulk endpoint error: %v — falling back to per-user lookup", err)
	}

	// Per-user fallback: check every candidate (attendees + maybe RSVPs).
	var ids []string
	for _, uid := range candidates {
		vs, vsErr := h.botClient.GetUserVoiceState(ctx, guildID, uid)
		if vsErr != nil {
			logger.Errorf("collectVoiceChannelMembers: GetUserVoiceState %s: %v", uid, vsErr)
			continue
		}
		if vs != nil && vs.ChannelID == channelID {
			ids = append(ids, uid)
		}
	}
	return ids
}

func (h *InteractionHandler) handleCtrlStart(c echo.Context, i *Interaction, eventID string) error {
	discordID := interactionUserID(i)
	ctx := c.Request().Context()

	event, err := h.eventRepo.FindByID(ctx, eventID)
	if err != nil || event == nil || event.GuildID != i.GuildID {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Event not found."))
	}
	if event.HostDiscordID != discordID {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Only the event host can start this event."))
	}
	if event.Status != repositories.EventStatusOpen {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ This event has already been started or closed."))
	}

	// Transition event to active (fast DB write — must succeed before responding).
	if err := h.eventRepo.Start(ctx, eventID); err != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Failed to start event. Please try again."))
	}

	// Run all Discord API work in a background goroutine so we respond to Discord
	// within the 3-second interaction deadline.
	hostName := memberDisplayName(i.Member)
	capturedGuildID := i.GuildID
	capturedChannelID := event.ChannelID
	capturedMsgID := event.AnnouncementMessageID
	capturedEventID := event.ID
	capturedHostID := discordID
	capturedEventType := event.EventType
	startLogger := c.Logger()
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 14*time.Second)
		defer cancel()

		// Create the voice channel now that the event has started.
		guild, gErr := h.guildRepo.FindByGuildID(bgCtx, capturedGuildID)
		if gErr != nil || guild == nil {
			startLogger.Errorf("ctrl_start: guild lookup failed for %s: %v", capturedGuildID, gErr)
		} else if guild.EventConfig.VoiceCategoryID == "" {
			startLogger.Errorf("ctrl_start: guild %s has no VoiceCategoryID configured — skipping voice channel creation", capturedGuildID)
		} else {
			channelName := memberChannelName(hostName, capturedEventType)
			vcID, vcErr := h.botClient.CreateEventVoiceChannel(bgCtx, capturedGuildID, channelName, guild.EventConfig.VoiceCategoryID, capturedHostID)
			if vcErr != nil {
				startLogger.Errorf("ctrl_start: CreateEventVoiceChannel failed: %v", vcErr)
			} else {
				if ev, fErr := h.eventRepo.FindByID(bgCtx, capturedEventID); fErr == nil && ev != nil {
					ev.VoiceChannelID = vcID
					ev.UpdatedAt = time.Now()
					if err := h.eventRepo.Update(bgCtx, capturedEventID, ev); err != nil {
						startLogger.Errorf("ctrl_start: failed to save VoiceChannelID %s: %v", vcID, err)
					}
				}
				_ = h.botClient.MoveGuildMember(bgCtx, capturedGuildID, capturedHostID, vcID)
			}
		}

		if capturedMsgID != "" {
			updatedEvent, fErr := h.eventRepo.FindByID(bgCtx, capturedEventID)
			if fErr == nil && updatedEvent != nil {
				isQuick := h.lookupIsQuick(bgCtx, capturedGuildID, updatedEvent.EventType)
				embed := buildEventEmbed(updatedEvent.EventType, isQuick, updatedEvent.Description, updatedEvent.HostDiscordID, capturedEventID, string(updatedEvent.Status), updatedEvent.ScheduledAt.Unix(), updatedEvent.AttendingIDs, updatedEvent.MaybeIDs, updatedEvent.NotAttendingIDs)
				_ = h.botClient.EditMessage(bgCtx, capturedChannelID, capturedMsgID, []Embed{embed}, buildRSVPButtons(updatedEvent.ID, isQuick, string(updatedEvent.Status)))
			}
		}
	}()

	return c.JSON(http.StatusOK, ephemeralMsg("▶️ Event started! The voice channel is now open."))
}

// --- MESSAGE_COMPONENT: ctrl_end|{eventID} ---

func (h *InteractionHandler) handleCtrlEnd(c echo.Context, i *Interaction, eventID string) error {
	discordID := interactionUserID(i)
	ctx := c.Request().Context()

	event, err := h.eventRepo.FindByID(ctx, eventID)
	if err != nil || event == nil || event.GuildID != i.GuildID {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Event not found."))
	}
	if event.HostDiscordID != discordID {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Only the event host can end this event."))
	}
	if event.Status != repositories.EventStatusActive {
		if event.Status == repositories.EventStatusClosed {
			return c.JSON(http.StatusOK, ephemeralMsg("⚠️ This event is already closed."))
		}
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Start the event before ending it."))
	}

	// Block early if a report already exists.
	existing, _ := h.eventReportRepo.FindByEventID(ctx, eventID)
	if existing != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ This event has already been logged."))
	}

	// Generate the log token synchronously (no I/O — stays well within the 3s deadline).
	token, err := session.SignEventLog(eventID, i.GuildID, discordID, h.secretKey)
	if err != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Failed to generate log link. Please try again."))
	}

	// Snapshot who is in the voice channel now so the log form can pre-fill participants.
	if event.VoiceChannelID != "" {
		candidates := append(event.AttendingIDs, event.MaybeIDs...)
		ids := h.collectVoiceChannelMembers(ctx, i.GuildID, event.VoiceChannelID, candidates, c.Logger())
		if len(ids) > 0 {
			event.VoiceMemberIDs = ids
			event.UpdatedAt = time.Now()
			_ = h.eventRepo.Update(ctx, event.ID, event)
		}
	}

	// Transition to closed so the Close Channel button becomes enabled.
	if err := h.eventRepo.Close(ctx, eventID); err != nil {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Failed to close event. Please try again."))
	}

	capturedVoiceChannelID := event.VoiceChannelID
	capturedGuildID := i.GuildID
	capturedEventID := eventID
	capturedMsgID := event.AnnouncementMessageID
	capturedChannelID := event.ChannelID
	bgLogger := c.Logger()
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 14*time.Second)
		defer cancel()

		// ctrl_start goroutine may not have written VoiceChannelID yet — re-fetch and snapshot.
		if capturedVoiceChannelID == "" {
			ev, err := h.eventRepo.FindByID(bgCtx, capturedEventID)
			if err == nil && ev != nil && ev.VoiceChannelID != "" {
				candidates := append(ev.AttendingIDs, ev.MaybeIDs...)
				memberIDs := h.collectVoiceChannelMembers(bgCtx, capturedGuildID, ev.VoiceChannelID, candidates, bgLogger)
				if len(memberIDs) > 0 {
					ev.VoiceMemberIDs = memberIDs
					ev.UpdatedAt = time.Now()
					_ = h.eventRepo.Update(bgCtx, capturedEventID, ev)
				}
			}
		}

		if capturedMsgID == "" {
			bgLogger.Warnf("ctrl_end: no AnnouncementMessageID for event %s — Close Channel button will not be enabled", capturedEventID)
			return
		}
		ev, fErr := h.eventRepo.FindByID(bgCtx, capturedEventID)
		if fErr != nil || ev == nil {
			return
		}
		isQuick := h.lookupIsQuick(bgCtx, capturedGuildID, ev.EventType)
		embed := buildEventEmbed(ev.EventType, isQuick, ev.Description, ev.HostDiscordID, capturedEventID, string(ev.Status), ev.ScheduledAt.Unix(), ev.AttendingIDs, ev.MaybeIDs, ev.NotAttendingIDs)
		if err := h.botClient.EditMessage(bgCtx, capturedChannelID, capturedMsgID, []Embed{embed}, buildRSVPButtons(capturedEventID, isQuick, "closed")); err != nil {
			bgLogger.Errorf("ctrl_end: EditMessage %s/%s: %v", capturedChannelID, capturedMsgID, err)
		}
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

	event, err := h.eventRepo.FindByID(ctx, eventID)
	if err != nil || event == nil || event.GuildID != i.GuildID {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Event not found."))
	}
	if event.HostDiscordID != discordID {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ Only the event host can close the channel."))
	}
	if event.Status != repositories.EventStatusClosed {
		return c.JSON(http.StatusOK, ephemeralMsg("⚠️ End the event before closing the channel."))
	}

	isQuick := h.lookupIsQuick(ctx, i.GuildID, event.EventType)
	embed := buildEventEmbed(event.EventType, isQuick, event.Description, event.HostDiscordID, eventID, string(event.Status), event.ScheduledAt.Unix(), event.AttendingIDs, event.MaybeIDs, event.NotAttendingIDs)

	capturedEventID := eventID
	capturedGuildID := i.GuildID
	capturedVoiceChannelID := event.VoiceChannelID
	capturedVoiceMemberIDs := event.VoiceMemberIDs
	bgLogger := c.Logger()
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		defer func() { _ = h.eventRepo.Delete(bgCtx, capturedEventID) }()

		if capturedVoiceChannelID == "" {
			return
		}
		lobbyID := ""
		if guild, gErr := h.guildRepo.FindByGuildID(bgCtx, capturedGuildID); gErr == nil && guild != nil {
			lobbyID = guild.EventConfig.LobbyChannelID
		}
		if lobbyID != "" {
			for _, uid := range capturedVoiceMemberIDs {
				if err := h.botClient.MoveGuildMember(bgCtx, capturedGuildID, uid, lobbyID); err != nil {
					bgLogger.Errorf("ctrl_close_channel: MoveGuildMember %s: %v", uid, err)
				}
			}
		}
		if err := h.botClient.DeleteChannel(bgCtx, capturedVoiceChannelID); err != nil {
			bgLogger.Errorf("ctrl_close_channel: DeleteChannel %s: %v", capturedVoiceChannelID, err)
		}
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

// lookupIsQuick returns true if the given eventType is configured as a quick event in the guild.
func (h *InteractionHandler) lookupIsQuick(ctx context.Context, guildID, eventType string) bool {
	guild, err := h.guildRepo.FindByGuildID(ctx, guildID)
	if err != nil || guild == nil {
		return false
	}
	for _, et := range guild.EventConfig.EventTypes {
		if strings.EqualFold(et.Name, eventType) {
			return et.IsQuickEvent
		}
	}
	return false
}
