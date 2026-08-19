package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BotClient makes Discord API calls authenticated as the bot application,
// using a Bot token rather than a user OAuth access token.
//
// This is separate from OAuthClient because the two use different auth schemes:
//   - OAuthClient: "Authorization: Bearer <user_access_token>"
//   - BotClient:   "Authorization: Bot <bot_token>"
//
// Use BotClient for any action the bot itself performs — verifying guild
// membership, reading guild data, checking permissions — rather than actions
// performed on behalf of a user.
type BotClient struct {
	botToken   string
	apiBaseURL string
	httpClient *http.Client
}

// NewBotClient creates a BotClient using the Discord bot token and API base URL
// from config. Both values come from the backend environment (DISCORD_BOT_TOKEN
// and DISCORD_API_BASE_URL).
func NewBotClient(botToken, apiBaseURL string) *BotClient {
	return &BotClient{
		botToken:   botToken,
		apiBaseURL: apiBaseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// VerifyBotInGuild checks whether the bot is currently a member of the
// specified Discord guild.
//
// It calls GET /guilds/{guildID} using the Bot token. Discord returns:
//   - 200: the bot is in the guild and can see it
//   - 403: the bot is not in the guild (or lacks access)
//   - 404: the guild does not exist
//
// Any non-200 status is treated as "not verified". Only a real HTTP or
// network failure returns a non-nil error.
func (c *BotClient) VerifyBotInGuild(ctx context.Context, guildID string) (bool, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.apiBaseURL+"/guilds/"+guildID,
		nil,
	)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bot "+c.botToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Read the body so the connection can be reused, even if we discard the content.
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	// 403 and 404 are expected negative cases — not errors in the Go sense.
	// Any other unexpected status is surfaced as an error for visibility.
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, fmt.Errorf("discord guild verification failed: status %d body %s", resp.StatusCode, string(raw))
}

type DiscordRole struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
	Managed  bool   `json:"managed"`
}

func (c *BotClient) GetGuildRoles(ctx context.Context, guildID string) ([]DiscordRole, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBaseURL+"/guilds/"+guildID+"/roles", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+c.botToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord get guild roles failed: status %d body %s", resp.StatusCode, string(raw))
	}

	var roles []DiscordRole
	if err := json.Unmarshal(raw, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

type DiscordMember struct {
	User *struct {
		ID         string `json:"id"`
		Username   string `json:"username"`
		GlobalName string `json:"global_name"`
		Avatar     string `json:"avatar"`
	} `json:"user"`
	Nick     string    `json:"nick"`
	Roles    []string  `json:"roles"`
	JoinedAt time.Time `json:"joined_at"`
}

func (c *BotClient) GetGuildMembers(ctx context.Context, guildID string) ([]DiscordMember, error) {
	const limit = 1000
	var all []DiscordMember
	after := ""

	for {
		rawURL := c.apiBaseURL + "/guilds/" + guildID + "/members?limit=1000"
		if after != "" {
			rawURL += "&after=" + after
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bot "+c.botToken)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		raw, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("discord get guild members failed: status %d body %s", resp.StatusCode, string(raw))
		}

		var page []DiscordMember
		if err := json.Unmarshal(raw, &page); err != nil {
			return nil, err
		}

		all = append(all, page...)

		if len(page) < limit { // handles 0, handles partial page
			break
		}

		last := page[len(page)-1] // safe: len >= 1000 here
		if last.User == nil {
			break
		}
		after = last.User.ID
	}

	return all, nil
}

// SendChannelMessage posts a message to a Discord channel as the bot.
// Calls POST /channels/{channelID}/messages.
func (c *BotClient) SendChannelMessage(ctx context.Context, channelID, content string) error {
	body, err := json.Marshal(map[string]string{"content": content})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.apiBaseURL+"/channels/"+channelID+"/messages",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bot "+c.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discord send message failed: status %d body %s", resp.StatusCode, string(raw))
	}

	return nil
}

// SendDMToUser sends a Discord direct message to a user identified by their Discord user ID.
// Opens (or reuses) a DM channel via POST /users/@me/channels, then sends the message.
func (c *BotClient) SendDMToUser(ctx context.Context, userDiscordID, content string) error {
	dmPayload, err := json.Marshal(map[string]string{"recipient_id": userDiscordID})
	if err != nil {
		return err
	}

	dmReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.apiBaseURL+"/users/@me/channels",
		bytes.NewReader(dmPayload),
	)
	if err != nil {
		return err
	}
	dmReq.Header.Set("Authorization", "Bot "+c.botToken)
	dmReq.Header.Set("Content-Type", "application/json")

	dmResp, err := c.httpClient.Do(dmReq)
	if err != nil {
		return err
	}
	dmRaw, _ := io.ReadAll(dmResp.Body)
	dmResp.Body.Close()

	if dmResp.StatusCode != http.StatusOK && dmResp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to open DM channel for user %s: status %d", userDiscordID, dmResp.StatusCode)
	}

	var dmChannel struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(dmRaw, &dmChannel); err != nil {
		return fmt.Errorf("failed to parse DM channel response: %w", err)
	}
	if dmChannel.ID == "" {
		return fmt.Errorf("empty DM channel ID for user %s", userDiscordID)
	}

	return c.SendChannelMessage(ctx, dmChannel.ID, content)
}

// PostEmbedMessage posts a message with rich embeds and interactive components to a channel.
// content is optional plain-text that appears above the embed (use for role/user pings on Raid events).
// Returns the Discord message ID of the posted message, used to update the embed later.
func (c *BotClient) PostEmbedMessage(ctx context.Context, channelID, content string, embeds []Embed, components []ActionRow) (string, error) {
	type payload struct {
		Content    string      `json:"content,omitempty"`
		Embeds     []Embed     `json:"embeds"`
		Components []ActionRow `json:"components,omitempty"`
	}
	body, err := json.Marshal(payload{Content: content, Embeds: embeds, Components: components})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.apiBaseURL+"/channels/"+channelID+"/messages",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bot "+c.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("discord post embed message failed: status %d body %s", resp.StatusCode, string(raw))
	}

	var msg struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return "", fmt.Errorf("parse message response: %w", err)
	}
	return msg.ID, nil
}

// EditInteractionResponse edits the deferred response for a slash-command interaction.
// Call this after returning InteractionResponseDeferredChannelMessage (type 5) to replace
// the "thinking..." state with the actual result.
// Uses the interaction webhook URL — no bot token auth needed, token is in the URL.
func (c *BotClient) EditInteractionResponse(ctx context.Context, applicationID, token, content string) error {
	body, err := json.Marshal(map[string]string{"content": content})
	if err != nil {
		return err
	}
	url := c.apiBaseURL + "/webhooks/" + applicationID + "/" + token + "/messages/@original"
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	// Interaction webhook edits do not require an Authorization header.
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discord edit interaction response failed: status %d body %s", resp.StatusCode, string(raw))
	}
	return nil
}

// EditMessage updates an existing channel message's embeds and components in place.
// Used to refresh the RSVP attendee list after a Join or Leave button interaction.
func (c *BotClient) EditMessage(ctx context.Context, channelID, messageID string, embeds []Embed, components []ActionRow) error {
	type payload struct {
		Embeds     []Embed     `json:"embeds"`
		Components []ActionRow `json:"components,omitempty"`
	}
	body, err := json.Marshal(payload{Embeds: embeds, Components: components})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch,
		c.apiBaseURL+"/channels/"+channelID+"/messages/"+messageID,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bot "+c.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discord edit message failed: status %d body %s", resp.StatusCode, string(raw))
	}
	return nil
}

// ─── Voice channel management ─────────────────────────────────────────────────

// Discord permission bitfield constants used for voice-channel overwrites.
const (
	permViewChannel = int64(1 << 10) // 1024
	permConnect     = int64(1 << 20) // 1048576
	permSpeak       = int64(1 << 21) // 2097152
	permMoveMembers = int64(1 << 24) // 16777216
)

// channelOverwrite is a Discord permission overwrite object.
// id is a role ID (type 0) or member ID (type 1).
type channelOverwrite struct {
	ID    string `json:"id"`
	Type  int    `json:"type"` // 0 = role, 1 = member
	Allow string `json:"allow"`
	Deny  string `json:"deny"`
}

// CreateEventVoiceChannel creates a private voice channel inside the given category.
// The host is granted VIEW_CHANNEL + CONNECT + SPEAK + MOVE_MEMBERS.
// @everyone is denied CONNECT (so they can see the channel but cannot join), which
// also lets the bot query voice states in the channel via the REST API.
// Returns the new channel ID.
func (c *BotClient) CreateEventVoiceChannel(ctx context.Context, guildID, name, categoryID, hostDiscordID string) (string, error) {
	hostPerms := permViewChannel | permConnect | permSpeak | permMoveMembers
	overwrites := []channelOverwrite{
		{
			ID:    guildID, // @everyone role has the same ID as the guild
			Type:  0,
			Allow: "0",
			Deny:  fmt.Sprintf("%d", permConnect), // deny join only; view kept for bot voice-state visibility
		},
		{
			ID:    hostDiscordID,
			Type:  1,
			Allow: fmt.Sprintf("%d", hostPerms),
			Deny:  "0",
		},
	}
	type payload struct {
		Name                 string             `json:"name"`
		Type                 int                `json:"type"` // 2 = GUILD_VOICE
		ParentID             string             `json:"parent_id,omitempty"`
		PermissionOverwrites []channelOverwrite `json:"permission_overwrites"`
	}
	body, err := json.Marshal(payload{
		Name:                 name,
		Type:                 2,
		ParentID:             categoryID,
		PermissionOverwrites: overwrites,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.apiBaseURL+"/guilds/"+guildID+"/channels",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bot "+c.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("create voice channel failed: status %d body %s", resp.StatusCode, string(raw))
	}
	var ch struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &ch); err != nil {
		return "", fmt.Errorf("parse create channel response: %w", err)
	}
	return ch.ID, nil
}

// MoveGuildMember moves a guild member to a voice channel.
// The member must already be connected to a voice channel; Discord will reject the move otherwise.
func (c *BotClient) MoveGuildMember(ctx context.Context, guildID, userID, channelID string) error {
	body, err := json.Marshal(map[string]string{"channel_id": channelID})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch,
		c.apiBaseURL+"/guilds/"+guildID+"/members/"+userID,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bot "+c.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// 200 = moved, 204 = already there or no-op
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("move guild member failed: status %d body %s", resp.StatusCode, string(raw))
	}
	return nil
}

// SetVoiceChannelLock updates the @everyone permission overwrite on a voice channel.
// locked=true: deny VIEW_CHANNEL + CONNECT (private).
// locked=false: allow VIEW_CHANNEL + CONNECT + SPEAK (open to all).
func (c *BotClient) SetVoiceChannelLock(ctx context.Context, channelID, guildID string, locked bool) error {
	var allow, deny string
	if locked {
		allow = "0"
		deny = fmt.Sprintf("%d", permViewChannel|permConnect)
	} else {
		allow = fmt.Sprintf("%d", permViewChannel|permConnect|permSpeak)
		deny = "0"
	}
	body, err := json.Marshal(map[string]interface{}{
		"allow": allow,
		"deny":  deny,
		"type":  0,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		c.apiBaseURL+"/channels/"+channelID+"/permissions/"+guildID,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bot "+c.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("set channel permissions failed: status %d body %s", resp.StatusCode, string(raw))
	}
	return nil
}

// VoiceState represents a single Discord guild voice-state entry.
type VoiceState struct {
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
}

// GetGuildVoiceStates returns all voice states for a guild.
// Filters out entries where ChannelID is empty (members who just disconnected).
func (c *BotClient) GetGuildVoiceStates(ctx context.Context, guildID string) ([]VoiceState, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.apiBaseURL+"/guilds/"+guildID+"/voice-states",
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+c.botToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get voice states failed: status %d body %s", resp.StatusCode, string(raw))
	}
	var states []VoiceState
	if err := json.Unmarshal(raw, &states); err != nil {
		return nil, fmt.Errorf("parse voice states: %w", err)
	}
	return states, nil
}

// DeleteChannel permanently deletes a Discord channel.
func (c *BotClient) DeleteChannel(ctx context.Context, channelID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		c.apiBaseURL+"/channels/"+channelID,
		nil,
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bot "+c.botToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete channel failed: status %d body %s", resp.StatusCode, string(raw))
	}
	return nil
}
