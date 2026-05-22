# API Reference

GuildLogger2 backend API reference for the Go + Echo migration.

## Status Legend

- Implemented: available in current backend
- Planned: defined target contract, not fully implemented yet

## Current Base URL

- Local: http://localhost:5001
- Route prefix: /api

## Implemented Endpoints

### GET /

Service root endpoint.

Response (200):

```json
{
  "service": "backend",
  "status": "ok"
}
```

### GET /api/health

Health endpoint for local/dev orchestration checks.

Response (200):

```json
{
  "ok": true
}
```

### Auth and Identity

#### Discord OAuth2 Login Flow

```
Frontend → GET  /api/auth/discord/url       → returns Discord authorization URL
User authorizes app on Discord
Discord  → redirects browser to redirectUri with ?code=...
Frontend → POST /api/auth/discord/login     → backend exchanges code, returns { token, user }
Frontend stores token in localStorage
Frontend → GET  /api/auth/session           → sends Bearer token, returns current user
Frontend → POST /api/auth/logout            → client discards token from localStorage
```

#### GET /api/auth/discord/url

Query parameters:
- `redirectUri` (required): the URI Discord will redirect back to after authorization.
- `state` (optional): CSRF token forwarded to Discord and returned in the callback.

Response (200):
```json
{
  "ok": true,
  "url": "https://discord.com/api/oauth2/authorize?..."
}
```

#### POST /api/auth/discord/login

Request body:
```json
{
  "code": "discord_authorization_code",
  "redirectUri": "http://localhost:5173/auth/callback"
}
```

Response (200):
```json
{
  "ok": true,
  "token": "signed_jwt",
  "user": {
    "discordId": "...",
    "createdAt": "...",
    "updatedAt": "..."
  }
}
```

#### GET /api/auth/session

Headers:
- `Authorization: Bearer <token>`

Response (200):
```json
{
  "ok": true,
  "user": { "discordId": "..." }
}
```

Response (401): missing, invalid, or expired token.

#### POST /api/auth/logout

Stateless. Client is responsible for discarding the token.

Response (200):
```json
{ "ok": true }
```

#### GET /api/auth/users/:discordId

Response (200):
```json
{
  "ok": true,
  "user": { "discordId": "..." }
}
```

Response (404): user not found.

### Guild and Bot Integration

All guild endpoints require `Authorization: Bearer <token>`.

#### GET /api/guilds

Returns guilds the authenticated user has connected to GuildLogger.

Response (200):
```json
{
  "ok": true,
  "guilds": [
    {
      "guildId": "...",
      "name": "My Server",
      "ownerDiscordId": "...",
      "botInstalled": false,
      "createdAt": "...",
      "updatedAt": "..."
    }
  ]
}
```

#### GET /api/guilds/discord

Returns the authenticated user's Discord server list fetched from Discord using their stored access token.

Response (200):
```json
{
  "ok": true,
  "guilds": [
    { "id": "...", "name": "My Server", "icon": "..." }
  ]
}
```

Response (401): no Discord access token stored for the user.
Response (502): Discord API unreachable or access token expired.

#### POST /api/guilds/connect

Links a Discord server to the authenticated user's account. Verifies the user is a member of the guild on Discord before creating the record.

Request body:
```json
{
  "guildId": "discord_guild_id",
  "name": "My Server"
}
```

Response (200):
```json
{
  "ok": true,
  "guild": {
    "guildId": "...",
    "name": "My Server",
    "ownerDiscordId": "...",
    "botInstalled": false
  }
}
```

Response (403): user is not a member of the specified Discord guild.
Response (409): guild is already connected.

#### GET /api/guilds/:guildId/bot/invite-url

Returns the Discord OAuth2 URL to add the bot to the specified guild. The authenticated user must be the guild owner in GuildLogger.

Response (200):
```json
{
  "ok": true,
  "url": "https://discord.com/api/oauth2/authorize?..."
}
```

Response (403): authenticated user does not own this guild.
Response (404): guild not found.

#### POST /api/guilds/:guildId/bot/install

Marks the bot as installed for the specified guild. Call after the user completes the Discord bot add flow.

Response (200):
```json
{ "ok": true }
```

Response (404): guild not found.

#### GET /api/guilds/:guildId/members/sync-status

Returns member sync status for a guild.

Response (200):
```json
{
  "ok": true,
  "memberCount": 42,
  "synced": true
}
```

#### GET /api/guilds/:guildId/dashboard

Returns aggregated summary data from guild, member, and event collections.

Response (200):
```json
{
  "ok": true,
  "dashboard": {
    "guild": { "..." : "..." },
    "memberCount": 42,
    "eventCount": 7
  }
}
```

Response (404): guild not found.

## Planned Endpoints (Target Contract)

### Event Management

- POST /api/events
- GET /api/events
- GET /api/events/{eventId}
- POST /api/events/{eventId}/register
- POST /api/events/{eventId}/unregister

### Member Analytics

- GET /api/members/{memberId}/stats

### Notifications

- POST /api/notifications/events/reminders/run
- POST /api/notifications/members/anniversaries/run

## Response Format Guidance

During migration, response shapes may vary by endpoint. New and migrated endpoints should follow:

```json
{
  "ok": true,
  "data": {}
}
```

Error shape:

```json
{
  "ok": false,
  "error": "Human readable message"
}
```

## Error Codes

- 200: successful read/update operation
- 201: resource created
- 400: invalid request payload
- 401: unauthenticated
- 403: authenticated but unauthorized
- 404: resource not found
- 409: state conflict
- 500: internal server error

## Notes

- Legacy Flask endpoint documentation is intentionally removed from this file.
- Add endpoint examples here as each Go route is implemented.
