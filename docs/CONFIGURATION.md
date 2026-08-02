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
| DISCORD_PUBLIC_KEY | — | No | Ed25519 public key from the Discord Developer Portal — used to verify incoming interaction request signatures. If empty, signature verification is skipped (dev only). Required for production. |
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
- MONGO_DB=GuildLoggerDB
- PORT=5001

## Frontend Environment Variables

| Variable | Default | Description |
|---|---|---|
| VITE_API_BASE_URL | http://localhost:5001 | Backend API base URL |

## Discord Slash Command Setup (Local Development)

Discord delivers slash command interactions by sending an HTTP POST to an **Interactions Endpoint URL** registered in the Discord Developer Portal. Discord's servers cannot reach `localhost`, so a public tunnel is required during local development.

The project uses **Cloudflare quicktunnel** (no account needed). The tunnel is installed at:
```
C:\Program Files (x86)\cloudflared\cloudflared.exe
```

> **Important:** The quicktunnel URL is randomly generated and **changes every time the tunnel is restarted**. You must update the Discord Developer Portal each session.

### Automated Tunnel (start_app.sh)

The `start_app.sh` script automatically starts cloudflared after the backend is healthy, waits for the tunnel URL, and prints it in the URL summary alongside the other service URLs:

```
Interactions: https://some-random-words.trycloudflare.com/api/interactions
  ↑ Paste into Discord Developer Portal → General Information → Interactions Endpoint URL
```

The tunnel is also automatically killed when the script exits (`Ctrl+C`). No manual tunnel management is needed when using `start_app.sh`.

The script looks for cloudflared at `C:\Program Files (x86)\cloudflared\cloudflared.exe` (the default winget install location). If cloudflared is not found, it prints a warning and skips the tunnel — the rest of the app still starts normally.

### Manual Tunnel (without start_app.sh)

1. Start MongoDB (or ensure Docker mongo container is running).
2. Start the backend:
   ```powershell
   cd backend
   go run .
   ```
3. In a **separate terminal**, start the Cloudflare tunnel:
   ```powershell
   & "C:\Program Files (x86)\cloudflared\cloudflared.exe" tunnel --url http://localhost:5001
   ```
4. Wait for the tunnel to print a URL like:
   ```
   https://some-random-words.trycloudflare.com
   ```
5. Go to the [Discord Developer Portal](https://discord.com/developers/applications) → your app → **General Information**.
6. Paste `https://some-random-words.trycloudflare.com/api/interactions` into **Interactions Endpoint URL** and click **Save Changes**.
   - Discord sends a PING to verify — the backend must be running for this to succeed.

### Session Shutdown Procedure

The cloudflared terminal is launched inside VS Code's Terminal panel (not a separate window).

**Option 1 — Ctrl+C (preferred):**
1. Find the cloudflared terminal tab in the VS Code Terminal panel. It will show log lines containing `INF Registered tunnel connection` and the `trycloudflare.com` URL.
2. Click that tab and press `Ctrl+C`. cloudflared will print a shutdown message and exit.
3. Press `Ctrl+C` in the backend terminal to stop the Go server.

**Option 2 — Stop-Process (if the terminal tab is lost or already closed):**
```powershell
Stop-Process -Name "cloudflared" -ErrorAction SilentlyContinue
```
Or by PID if you know it:
```powershell
Stop-Process -Id <pid>
```

**Verify the tunnel is stopped:**
```powershell
Get-Process -Name "cloudflared" -ErrorAction SilentlyContinue
```
No output means the process is gone.

The Discord Developer Portal does not need to be updated on shutdown — it will simply fail to deliver interactions until you provide a new URL next session.

### Persistent URL (Optional)

To avoid updating the portal on every restart, use a **named Cloudflare Tunnel** (requires a free Cloudflare account) or a **paid ngrok plan** with a reserved subdomain. See the [Cloudflare Tunnel docs](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps) for setup.

---

## Local Run Checklist

1. Ensure Mongo is running.
2. Set backend env variables (including `DISCORD_PUBLIC_KEY` for interaction signature verification).
3. Run backend from backend directory.
4. Start Cloudflare tunnel and update Discord Developer Portal Interactions Endpoint URL (see above).
5. Verify GET /api/health returns ok.

## Docker Run Checklist

1. Build and run mongo + backend via compose.
2. Verify backend command resolves to Go binary (/app/main).
3. Validate health endpoint on localhost:5001.

## Migration Notes

- Flask-specific variables such as FLASK_DEBUG are obsolete for the Go runtime.
- Keep this file synchronized with docker-compose.yml and backend/app/config/config.go.
