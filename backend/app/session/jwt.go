package session

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const sessionTTL = 24 * time.Hour

type Claims struct {
	DiscordID string `json:"discordId"`
	jwt.RegisteredClaims
}

// Sign creates a signed JWT for the given Discord user ID.
func Sign(discordID, secretKey string) (string, error) {
	claims := Claims{
		DiscordID: discordID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(sessionTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// Verify parses and validates a JWT, returning the claims if valid.
func Verify(tokenStr, secretKey string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ─── Event-log one-time tokens ────────────────────────────────────────────────

const eventLogTTL = 8 * time.Hour

// EventLogClaims is the payload for a one-time event-log submission token.
// Subject is always "event_log" so it cannot be confused with session tokens.
type EventLogClaims struct {
	EventID       string `json:"eventId"`
	GuildID       string `json:"guildId"`
	HostDiscordID string `json:"hostDiscordId"`
	jwt.RegisteredClaims
}

// SignEventLog creates a signed JWT granting the bearer permission to submit
// an event log for the given event. The token expires after 48 hours.
func SignEventLog(eventID, guildID, hostDiscordID, secretKey string) (string, error) {
	claims := EventLogClaims{
		EventID:       eventID,
		GuildID:       guildID,
		HostDiscordID: hostDiscordID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "event_log",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(eventLogTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// VerifyEventLog parses and validates an event-log token.
// Returns an error if the token is expired, tampered with, or is not an event_log subject.
func VerifyEventLog(tokenStr, secretKey string) (*EventLogClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &EventLogClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*EventLogClaims)
	if !ok || !token.Valid || claims.Subject != "event_log" {
		return nil, errors.New("invalid event log token")
	}
	return claims, nil
}
