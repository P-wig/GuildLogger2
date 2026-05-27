package discord

import (
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
		ID string `json:"id"`
	} `json:"user"`
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

		if len(page) < limit {
			break
		}

		after = page[len(page)-1].User.ID
	}

	return all, nil
}
