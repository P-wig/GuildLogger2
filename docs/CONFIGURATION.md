# Configuration Reference

Configuration reference for GuildLogger2 (Go backend, React frontend, MongoDB, Discord integrations).

## Backend Environment Variables

| Variable | Default | Required | Description |
|---|---|---|---|
| PORT | 5001 | No | Backend listen port |
| MONGO_URI | mongodb://localhost:27017 | No | MongoDB connection URI |
| MONGO_DB | guildlogger | No | MongoDB database name |
| CORS_ORIGINS | http://localhost:5173 | No | Comma-separated allowed CORS origins |
| SECRET_KEY | — | **Yes** | JWT signing secret — use a strong random value |
| DISCORD_CLIENT_ID | — | **Yes** | Discord application client ID |
| DISCORD_CLIENT_SECRET | — | **Yes** | Discord application client secret |
| DISCORD_REDIRECT_URI | — | **Yes** | OAuth2 callback URL — must match frontend `/auth` route and Discord Developer Portal allowlist |
| DISCORD_BOT_TOKEN | — | No | Bot token for Discord bot API calls |
| DISCORD_AUTH_BASE_URL | https://discord.com/api/oauth2 | No | Discord OAuth2 base URL |
| DISCORD_API_BASE_URL | https://discord.com/api/v10 | No | Discord REST API base URL |
| DISCORD_OAUTH_SCOPES | identify guilds | No | Space-separated OAuth2 scopes |

The backend will refuse to start if `SECRET_KEY`, `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`, or `DISCORD_REDIRECT_URI` are missing.

### Local development values

```env
DISCORD_REDIRECT_URI=http://localhost:5173/auth
CORS_ORIGINS=http://localhost:5173
MONGO_URI=mongodb://localhost:27017
MONGO_DB=guildlogger
```

### Container Runtime Values

When running in Docker Compose, backend should use:

- MONGO_URI=mongodb://mongo:27017
- MONGO_DB=cloudnative
- PORT=5001

## Frontend Environment Variables

| Variable | Default | Description |
|---|---|---|
| VITE_API_BASE_URL | http://localhost:5001 | Backend API base URL |

## Discord Integration Variables (Planned)

| Variable | Purpose |
|---|---|
| DISCORD_CLIENT_ID | OAuth2 client id |
| DISCORD_CLIENT_SECRET | OAuth2 secret |
| DISCORD_REDIRECT_URI | OAuth callback URL |
| DISCORD_BOT_TOKEN | Bot API token |

## Local Run Checklist

1. Ensure Mongo is running.
2. Set backend env variables.
3. Run backend from backend directory.
4. Verify GET /api/health returns ok.

## Docker Run Checklist

1. Build and run mongo + backend via compose.
2. Verify backend command resolves to Go binary (/app/main).
3. Validate health endpoint on localhost:5001.

## Migration Notes

- Flask-specific variables such as FLASK_DEBUG are obsolete for the Go runtime.
- Keep this file synchronized with docker-compose.yml and backend/app/config/config.go.
