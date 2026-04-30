# Configuration Reference

Configuration reference for GuildLogger2 (Go backend, React frontend, MongoDB, Discord integrations).

## Backend Environment Variables

| Variable | Default | Description |
|---|---|---|
| PORT | 5001 | Backend listen port |
| MONGO_URI | mongodb://localhost:27017 | Mongo connection URI for local runtime |
| MONGO_DB | test | Mongo database name |
| CORS_ORIGINS | http://localhost:5173 | Comma-separated CORS origins |

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
