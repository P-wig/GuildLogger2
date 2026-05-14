package discord

import (
	"context"
	"encoding/json"
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

// TokenResponse holds the token data returned by Discord after code exchange.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// ExchangeCode exchanges the authorization code for Discord access + refresh tokens.
func (c *OAuthClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*TokenResponse, error) {
	body := url.Values{}
	body.Set("client_id", c.clientID)
	body.Set("client_secret", c.clientSecret)
	body.Set("grant_type", "authorization_code")
	body.Set("code", code)
	body.Set("redirect_uri", redirectURI)

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
		return nil, fmt.Errorf("discord token exchange failed: status %d body %s", resp.StatusCode, string(raw))
	}

	var tokens TokenResponse
	if err := json.Unmarshal(raw, &tokens); err != nil {
		return nil, err
	}
	return &tokens, nil
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
