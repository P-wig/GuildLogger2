// backend/app/config/config.go
package config

import (
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
}

func Load() Config {
	_ = godotenv.Load() // reads backend/.env if present, no-op in Docker

	origins := strings.Split(os.Getenv("CORS_ORIGINS"), ",")
	if len(origins) == 0 || (len(origins) == 1 && strings.TrimSpace(origins[0]) == "") {
		origins = []string{"http://localhost:5173"}
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
	}
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
