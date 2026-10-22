package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OAuthClient handles Discord OAuth2 token exchange and user profile fetching.
type OAuthClient struct {
	clientID     string
	clientSecret string
	authBaseURL  string
	apiBaseURL   string
	scopes       []string
	httpClient   *http.Client
}

func NewOAuthClient(clientID, clientSecret, authBaseURL, apiBaseURL string, scopes []string) *OAuthClient {
	return &OAuthClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		authBaseURL:  authBaseURL,
		apiBaseURL:   apiBaseURL,
		scopes:       scopes,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// AuthorizeURL builds the Discord OAuth2 authorization URL the user visits to log in.
func (c *OAuthClient) AuthorizeURL(redirectURI, state string) string {
	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(c.scopes, " "))
	if state != "" {
		params.Set("state", state)
	}
	return c.authBaseURL + "/authorize?" + params.Encode()
}

// ErrRefreshDenied means Discord rejected the refresh token itself — the user revoked the
// app, or the token was already rotated away. Retrying cannot help; only re-authorization can.
var ErrRefreshDenied = errors.New("discord refused the refresh token")

// TokenResponse holds the token data returned by Discord after code exchange.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// ExpiresAt converts the relative expires_in into an absolute UTC instant.
func (t *TokenResponse) ExpiresAt(now time.Time) time.Time {
	return now.UTC().Add(time.Duration(t.ExpiresIn) * time.Second)
}

// requestToken posts to Discord's /token endpoint with client credentials attached.
// Shared by the authorization-code exchange and the refresh-token grant.
func (c *OAuthClient) requestToken(ctx context.Context, body url.Values) (*TokenResponse, error) {
	body.Set("client_id", c.clientID)
	body.Set("client_secret", c.clientSecret)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.authBaseURL+"/token",
		strings.NewReader(body.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		// 4xx means the grant is bad; 5xx means Discord is having a bad day and a retry may work.
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, fmt.Errorf("%w: status %d body %s", ErrRefreshDenied, resp.StatusCode, string(raw))
		}
		return nil, fmt.Errorf("discord token request failed: status %d body %s", resp.StatusCode, string(raw))
	}

	var tokens TokenResponse
	if err := json.Unmarshal(raw, &tokens); err != nil {
		return nil, err
	}
	return &tokens, nil
}

// ExchangeCode exchanges the authorization code for Discord access + refresh tokens.
func (c *OAuthClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*TokenResponse, error) {
	body := url.Values{}
	body.Set("grant_type", "authorization_code")
	body.Set("code", code)
	body.Set("redirect_uri", redirectURI)
	return c.requestToken(ctx, body)
}

// RefreshToken trades a refresh token for a fresh access + refresh pair.
// Discord rotates the refresh token on every use, so the returned pair must be persisted
// in full — reusing the old refresh token afterwards fails.
// Returns ErrRefreshDenied when the grant is rejected and re-authorization is required.
func (c *OAuthClient) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	body := url.Values{}
	body.Set("grant_type", "refresh_token")
	body.Set("refresh_token", refreshToken)
	return c.requestToken(ctx, body)
}

// DiscordUser holds the identity fields returned by /users/@me.
type DiscordUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// GetCurrentUser fetches the authenticated Discord user's profile using an access token.
func (c *OAuthClient) GetCurrentUser(ctx context.Context, accessToken string) (*DiscordUser, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.apiBaseURL+"/users/@me",
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

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
		return nil, fmt.Errorf("discord user fetch failed: status %d body %s", resp.StatusCode, string(raw))
	}

	var user DiscordUser
	if err := json.Unmarshal(raw, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

type DiscordPartialGuild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

func (c *OAuthClient) GetUserGuilds(ctx context.Context, accessToken string) ([]DiscordPartialGuild, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.apiBaseURL+"/users/@me/guilds",
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

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
		return nil, fmt.Errorf("discord guilds fetch failed: status %d body %s", resp.StatusCode, string(raw))
	}

	var guilds []DiscordPartialGuild
	if err := json.Unmarshal(raw, &guilds); err != nil {
		return nil, err
	}
	return guilds, nil
}

// BotInviteURL builds the Discord OAuth2 URL to add the bot to a specific guild.
//
// When redirectURI is provided, Discord will redirect back to that URI after
// the user authorizes the bot, appending ?guild_id=...&code=... as query params.
// This enables the app to auto-verify installation instead of requiring a manual
// confirmation step. When redirectURI is empty, no redirect occurs.
func (c *OAuthClient) BotInviteURL(guildID, redirectURI string) string {
	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("scope", "bot applications.commands")
	params.Set("guild_id", guildID)
	if redirectURI != "" {
		params.Set("redirect_uri", redirectURI)
		params.Set("response_type", "code")
	}
	return c.authBaseURL + "/authorize?" + params.Encode()
}
