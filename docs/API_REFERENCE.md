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
    "accessToken": "...",
    "refreshToken": "..."
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

- GET /api/guilds
- POST /api/guilds/connect
- POST /api/guilds/{guildId}/bot/install
- GET /api/guilds/{guildId}/members/sync-status

### Event Management

- POST /api/events
- GET /api/events
- GET /api/events/{eventId}
- POST /api/events/{eventId}/register
- POST /api/events/{eventId}/unregister

### Member Analytics

- GET /api/members/{memberId}/stats
- GET /api/guilds/{guildId}/dashboard

### Notifications

- POST /api/notifications/events/reminders/run
- POST /api/notifications/members/anniversaries/run

## Planned Endpoints (Target Contract)

### Guild and Bot Integration

- GET /api/guilds
- POST /api/guilds/connect
- POST /api/guilds/{guildId}/bot/install
- GET /api/guilds/{guildId}/members/sync-status

### Event Management

- POST /api/events
- GET /api/events
- GET /api/events/{eventId}
- POST /api/events/{eventId}/register
- POST /api/events/{eventId}/unregister

### Member Analytics

- GET /api/members/{memberId}/stats
- GET /api/guilds/{guildId}/dashboard

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
