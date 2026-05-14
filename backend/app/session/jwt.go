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
