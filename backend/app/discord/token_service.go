package discord

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/P-wig/GuildLogger2/backend/app/repositories"
)

var (
	// ErrNoDiscordToken means the user has no stored Discord credentials at all.
	ErrNoDiscordToken = errors.New("no discord access token stored for user")
	// ErrDiscordReauthRequired means the stored credentials are unusable and cannot be
	// refreshed. The caller must invalidate the session and send the user back through OAuth.
	ErrDiscordReauthRequired = errors.New("discord authorization expired; re-authentication required")
)

// refreshSkew refreshes slightly before the real expiry so a token cannot die
// in flight between our check and Discord receiving the request.
const refreshSkew = 5 * time.Minute

// TokenService hands out a usable Discord access token for a user, refreshing it
// transparently when it has expired.
//
// It is a read model in the sense of the SDD — it owns no lifecycle — but unlike
// StatsService it does write, because refreshing is only correct if the rotated
// token pair is persisted. The write is incidental to the read, not a domain transition.
type TokenService struct {
	users repositories.UserRepository
	oauth *OAuthClient

	// mu serializes refreshes per user. Discord rotates the refresh token on every use,
	// so two concurrent refreshes would leave one goroutine holding a token that Discord
	// has already invalidated, logging the user out at random.
	mu       sync.Mutex
	inflight map[string]*sync.Mutex
}

func NewTokenService(users repositories.UserRepository, oauth *OAuthClient) *TokenService {
	return &TokenService{users: users, oauth: oauth, inflight: make(map[string]*sync.Mutex)}
}

func (s *TokenService) userLock(discordID string) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()
	lock, ok := s.inflight[discordID]
	if !ok {
		lock = &sync.Mutex{}
		s.inflight[discordID] = lock
	}
	return lock
}

// ValidAccessToken returns an access token that is safe to use against the Discord API
// right now, refreshing and persisting a new token pair first if the stored one has expired.
//
// Returns ErrNoDiscordToken when nothing is stored, and ErrDiscordReauthRequired when the
// stored credentials are beyond saving.
func (s *TokenService) ValidAccessToken(ctx context.Context, discordID string) (string, error) {
	user, err := s.users.FindByDiscordID(ctx, discordID)
	if err != nil {
		return "", err
	}
	if user == nil || user.AccessToken == "" {
		return "", ErrNoDiscordToken
	}
	if !s.needsRefresh(user) {
		return user.AccessToken, nil
	}

	lock := s.userLock(discordID)
	lock.Lock()
	defer lock.Unlock()

	// Another request may have refreshed while we waited for the lock.
	user, err = s.users.FindByDiscordID(ctx, discordID)
	if err != nil {
		return "", err
	}
	if user == nil || user.AccessToken == "" {
		return "", ErrNoDiscordToken
	}
	if !s.needsRefresh(user) {
		return user.AccessToken, nil
	}
	if user.RefreshToken == "" {
		return "", ErrDiscordReauthRequired
	}

	tokens, err := s.oauth.RefreshToken(ctx, user.RefreshToken)
	if errors.Is(err, ErrRefreshDenied) {
		// The grant is dead — drop the credentials so we stop retrying on every request.
		_ = s.users.ClearTokens(ctx, discordID)
		return "", ErrDiscordReauthRequired
	}
	if err != nil {
		return "", err
	}

	if err := s.users.UpdateTokens(ctx, discordID, tokens.AccessToken, tokens.RefreshToken, tokens.ExpiresAt(time.Now())); err != nil {
		return "", err
	}
	return tokens.AccessToken, nil
}

// needsRefresh treats a zero expiry as expired so users stored before expiry tracking
// existed get refreshed on their next Discord call rather than failing forever.
func (s *TokenService) needsRefresh(user *repositories.User) bool {
	return !time.Now().UTC().Before(user.TokenExpiresAt.Add(-refreshSkew))
}
