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

Response (401): missing, invalid, or expired token.
Response (403): authenticated user does not own this guild.
Response (404): guild not found.

#### POST /api/guilds/:guildId/bot/verify

Verifies the bot is present in the guild via the Discord API and syncs guild roles into GuildLogger.
Only the guild owner can verify. On success, the guild is marked as bot-installed with roles populated.

Response (200):
```json
{ "ok": true }
```

Response (401): missing, invalid, or expired token.
Response (403): authenticated user does not own this guild.
Response (404): guild not found.
Response (422): bot is not currently installed in this guild.
Response (502): Discord API unreachable or bot token invalid.

#### POST /api/guilds/:guildId/members/sync

Synchronizes Discord guild members into GuildLogger. No request body — the guild ID is taken from the URL.
Only the guild owner can trigger a sync.

Response (200):
```json
{
  "ok": true,
  "synced": true,
  "memberCount": 42
}
```

Response (401): missing, invalid, or expired token.
Response (403): authenticated user does not own this guild.
Response (404): guild not found.
Response (502): failed to fetch members from Discord.

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

Response (403): user is not a member of the specified Discord guild.
Response (404): guild not found.

#### GET /api/guilds/:guildId/dashboard

Returns aggregated summary data from guild, member, and event collections.
Includes a queryable `members` array of projected member summaries.

Response (200):
```json
{
  "ok": true,
  "dashboard": {
    "guild": { "guildId": "...", "name": "My Server", "ownerDiscordId": "...", "botInstalled": true },
    "memberCount": 42,
    "members": [
      {
        "discordId": "1234567890",
        "rankedRoleId": "...",
        "status": "active",
        "discordJoinedAt": "2023-06-08T00:00:00Z",
        "roleIds": ["...", "..."]
      }
    ],
    "eventCount": 7
  }
}
```

Response (401): missing, invalid, or expired token.
Response (404): guild not found.

### Event Operations

All event endpoints require `Authorization: Bearer <token>`.

#### POST /api/events/:eventId/start

Transitions an event from `open` to `active`. Only the event host can start the event.

Response (200):
```json
{ "ok": true }
```

Response (400): missing or invalid `eventId`.
Response (401): missing, invalid, or expired token.
Response (403): authenticated user is not the event host.
Response (404): event not found.
Response (409): event is not currently in `open` status.

#### POST /api/events/:eventId/close

Transitions an event from `active` to `closed` and creates a permanent EventReport record.
Only the event host can close the event.

Request body:
```json
{
  "summary": "Event wrap-up notes",
  "participantIds": ["1234567890", "0987654321"],
  "eventDate": "2026-06-04T20:00:00Z"
}
```

Response (200):
```json
{
  "ok": true,
  "report": {
    "id": "...",
    "eventId": "...",
    "guildId": "...",
    "hostDiscordId": "...",
    "eventDate": "2026-06-04T20:00:00Z",
    "participantIds": ["1234567890", "0987654321"],
    "summary": "Event wrap-up notes",
    "submittedAt": "..."
  }
}
```

Response (400): invalid request body or missing `eventId`.
Response (401): missing, invalid, or expired token.
Response (403): authenticated user is not the event host.
Response (404): event not found.
Response (409): event is not currently in `active` status.

#### POST /api/events

Creates a new event for a guild. The authenticated user becomes the host.

Request body:
```json
{
  "guildId": "...",
  "title": "Raid Night",
  "description": "Optional description, max 3000 characters",
  "scheduledAt": "2026-06-10T20:00:00Z",
  "channelId": "...",
  "reminderEnabled": true,
  "capacity": 20,
  "cutoffAt": "2026-06-10T19:00:00Z"
}
```

Response (200):
```json
{ "ok": true, "event": { "id": "...", "guildId": "...", "title": "Raid Night", "status": "open", "attendingIds": [] } }
```

Response (400): `guildId` or `title` missing.
Response (401): missing, invalid, or expired token.

#### GET /api/events

Returns all events for a guild.

Query parameters:
- `guildId` (required)

Response (200):
```json
{ "ok": true, "events": [] }
```

Response (400): `guildId` query parameter missing.
Response (401): missing, invalid, or expired token.

#### GET /api/events/:eventId

Returns a single event by ID.

Response (200):
```json
{ "ok": true, "event": { "id": "...", "status": "open" } }
```

Response (401): missing, invalid, or expired token.
Response (404): event not found.

#### POST /api/events/:eventId/register

Registers the authenticated user for an event using their Discord ID from the session token.

Response (200):
```json
{ "ok": true }
```

Response (401): missing, invalid, or expired token.
Response (404): event not found.
Response (409): already registered, registration is closed, or event is at capacity.

#### POST /api/events/:eventId/unregister

Unregisters the authenticated user from an event.

Response (200):
```json
{ "ok": true }
```

Response (401): missing, invalid, or expired token.
Response (404): event not found.
Response (409): not registered for this event.

### Member Analytics

All member endpoints require `Authorization: Bearer <token>`.

#### GET /api/members/:discordId/stats

Returns aggregated stats for a member within a guild. Includes hosted and participation counts derived from the `event_reports` collection, plus tenure timestamps.

Query parameters:
- `guildId` (required)

Response (200):
```json
{
  "ok": true,
  "stats": {
    "hostedCount": 5,
    "participatedCount": 12,
    "discordJoinedAt": "2023-06-08T00:00:00Z",
    "firstSyncedAt": "2024-01-01T00:00:00Z",
    "deactivatedAt": null
  }
}
```

Response (400): `discordId` or `guildId` missing.
Response (401): missing, invalid, or expired token.
Response (404): member not found.

### Notifications

All notification endpoints require `Authorization: Bearer <token>`.

#### POST /api/notifications/members/anniversaries/run

Scans guild members for Discord join date anniversaries matching the guild's configured milestone years and posts a congratulatory message to the guild's configured notification channel for each match. Respects each member's `notificationsOptOut` preference.

Request body:
```json
{
  "guildId": "..."
}
```

Response (200):
```json
{
  "ok": true,
  "notified": 2,
  "members": ["1234567890", "0987654321"],
  "failed": []
}
```

Response (200, disabled): `{ "ok": true, "notified": 0, "skipped": "milestone notifications are disabled for this guild" }`
Response (200, none configured): `{ "ok": true, "notified": 0, "skipped": "no anniversary years configured" }`
Response (400): `guildId` missing or invalid.
Response (401): missing, invalid, or expired token.
Response (404): guild not found.
Response (422): no notification channel configured for this guild.
Response (502): one or more Discord channel messages failed to send (partial success reported in `failed` array).

## Planned Endpoints (Target Contract)

### Notifications

- POST /api/notifications/events/reminders/run

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
- 422: precondition not met (e.g. bot not yet installed in guild)
- 502: upstream Discord API error

## Notes

- Legacy Flask endpoint documentation is intentionally removed from this file.
- Add endpoint examples here as each Go route is implemented.
