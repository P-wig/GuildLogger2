package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ApplicationCommandType is the Discord command type discriminator.
type ApplicationCommandType int

const (
	CommandTypeChatInput ApplicationCommandType = 1 // Slash command ("/name")
)

// ApplicationCommandOptionType maps to Discord's option type enum.
type ApplicationCommandOptionType int

const (
	OptionTypeSubCommand      ApplicationCommandOptionType = 1
	OptionTypeSubCommandGroup ApplicationCommandOptionType = 2
	OptionTypeString          ApplicationCommandOptionType = 3
	OptionTypeInteger         ApplicationCommandOptionType = 4
	OptionTypeBoolean         ApplicationCommandOptionType = 5
	OptionTypeUser            ApplicationCommandOptionType = 6
	OptionTypeChannel         ApplicationCommandOptionType = 7
	OptionTypeRole            ApplicationCommandOptionType = 8
	OptionTypeMentionable     ApplicationCommandOptionType = 9
	OptionTypeNumber          ApplicationCommandOptionType = 10
	OptionTypeAttachment      ApplicationCommandOptionType = 11
)

// ApplicationCommandOption defines one argument within a slash command.
type ApplicationCommandOption struct {
	Type         ApplicationCommandOptionType `json:"type"`
	Name         string                       `json:"name"`
	Description  string                       `json:"description"`
	Required     bool                         `json:"required,omitempty"`
	Autocomplete bool                         `json:"autocomplete,omitempty"`
	Options      []ApplicationCommandOption   `json:"options,omitempty"`
}

// ApplicationCommand is the top-level schema sent to Discord for one slash command.
type ApplicationCommand struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Type        ApplicationCommandType     `json:"type"`
	Options     []ApplicationCommandOption `json:"options,omitempty"`
}

// BotCommands is the authoritative list of slash commands the GuildLogger bot exposes in
// Discord. This slice is bulk-overwritten on every startup via PUT
// /applications/{id}/commands, which is idempotent — adding, editing, or removing a
// command here and restarting the backend is the only step required.
var BotCommands = []ApplicationCommand{
	{
		Name:        "stats",
		Description: "View event participation stats for yourself or another member",
		Type:        CommandTypeChatInput,
		Options: []ApplicationCommandOption{
			{
				Type:        OptionTypeUser,
				Name:        "member",
				Description: "The member to view stats for (defaults to yourself)",
				Required:    false,
			},
		},
	},
	{
		Name:        "anniversary",
		Description: "Trigger anniversary milestone notifications for the guild",
		Type:        CommandTypeChatInput,
	},
	{
		Name:        "help",
		Description: "List all available GuildLogger commands",
		Type:        CommandTypeChatInput,
	},
	{
		Name:        "event",
		Description: "Manage guild events",
		Type:        CommandTypeChatInput,
		Options: []ApplicationCommandOption{
			{
				Type:        OptionTypeSubCommand,
				Name:        "create",
				Description: "Create a new event and post an RSVP announcement to the events channel",
				Options: []ApplicationCommandOption{
					{
						Type:         OptionTypeString,
						Name:         "eventtype",
						Description:  "The type of event (configured in the GuildLogger dashboard)",
						Required:     true,
						Autocomplete: true,
					},
				},
			},
			{
				Type:        OptionTypeSubCommand,
				Name:        "log",
				Description: "Log a completed event and create a permanent record",
				Options: []ApplicationCommandOption{
					{
						Type:        OptionTypeString,
						Name:        "eventid",
						Description: "The ID of the event to log",
						Required:    true,
					},
				},
			},
		},
	},
}

// RegisterGlobalCommands bulk-overwrites the bot's global slash commands with Discord
// using PUT /applications/{applicationID}/commands.
//
// Global commands are available in every guild the bot has joined. Discord propagates
// changes within ~1 hour; they appear immediately in the Discord client after that window.
//
// This call is safe to run on every restart (idempotent). If botToken or applicationID
// are empty (for example in environments without a bot configured) the call is skipped
// and nil is returned.
func (c *BotClient) RegisterGlobalCommands(ctx context.Context, applicationID string) error {
	if c.botToken == "" || applicationID == "" {
		return nil
	}

	body, err := json.Marshal(BotCommands)
	if err != nil {
		return fmt.Errorf("marshal commands: %w", err)
	}

	url := fmt.Sprintf("%s/applications/%s/commands", c.apiBaseURL, applicationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bot "+c.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("discord request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// 200 OK on bulk overwrite success.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discord register commands failed: status %d body %s", resp.StatusCode, string(raw))
	}

	return nil
}
