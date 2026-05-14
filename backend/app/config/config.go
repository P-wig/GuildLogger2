// backend/app/config/config.go
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI            string
	MongoDB             string
	CORSOrigins         []string
	Port                string
	DiscordBotToken     string
	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURI  string
	SecretKey           string
	DiscordAuthBaseURL  string
	DiscordAPIBaseURL   string
	DiscordOAuthScopes  []string
}

func Load() Config {
	_ = godotenv.Load()

	origins := strings.Split(os.Getenv("CORS_ORIGINS"), ",")
	if len(origins) == 0 || (len(origins) == 1 && strings.TrimSpace(origins[0]) == "") {
		origins = []string{"http://localhost:5173"}
	}

	scopesRaw := os.Getenv("DISCORD_OAUTH_SCOPES")
	if strings.TrimSpace(scopesRaw) == "" {
		scopesRaw = "identify guilds"
	}

	return Config{
		MongoURI:            getenv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:             getenv("MONGO_DB", "guildlogger"),
		CORSOrigins:         trimAll(origins),
		Port:                getenv("PORT", "5001"),
		DiscordBotToken:     getenv("DISCORD_BOT_TOKEN", ""),
		DiscordClientID:     getenv("DISCORD_CLIENT_ID", ""),
		DiscordClientSecret: getenv("DISCORD_CLIENT_SECRET", ""),
		DiscordRedirectURI:  getenv("DISCORD_REDIRECT_URI", ""),
		SecretKey:           getenv("SECRET_KEY", ""),
		DiscordAuthBaseURL:  getenv("DISCORD_AUTH_BASE_URL", "https://discord.com/api/oauth2"),
		DiscordAPIBaseURL:   getenv("DISCORD_API_BASE_URL", "https://discord.com/api/v10"),
		DiscordOAuthScopes:  trimAll(strings.Split(scopesRaw, " ")),
	}
}

// ValidateOAuthConfig returns an error if any required Discord OAuth value is missing.
// Call this at startup before registering auth routes.
func ValidateOAuthConfig(cfg Config) error {
	missing := []string{}
	if cfg.DiscordClientID == "" {
		missing = append(missing, "DISCORD_CLIENT_ID")
	}
	if cfg.DiscordClientSecret == "" {
		missing = append(missing, "DISCORD_CLIENT_SECRET")
	}
	if cfg.DiscordRedirectURI == "" {
		missing = append(missing, "DISCORD_REDIRECT_URI")
	}
	if cfg.SecretKey == "" {
		missing = append(missing, "SECRET_KEY")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required OAuth config: %s", strings.Join(missing, ", "))
	}
	return nil
}

func getenv(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}

func trimAll(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		t := strings.TrimSpace(s)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}
