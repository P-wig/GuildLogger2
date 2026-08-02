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

#### DELETE /api/guilds/:guildId

Disconnects a guild from GuildLogger. Permanently deletes the guild record and all associated member records. Only the guild owner can disconnect.

Response (200):
```json
{ "ok": true }
```

Response (401): missing, invalid, or expired token.
Response (403): authenticated user does not own this guild.
Response (404): guild not found.

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
Only the guild owner can trigger a sync. Requires the member role to be configured first.

Response (200):
```json
{
  "ok": true,
  "synced": 42
}
```

Response (401): missing, invalid, or expired token.
Response (403): authenticated user does not own this guild.
Response (404): guild not found.
Response (422): member role not configured — call `PUT /api/guilds/:guildId/config` first.
Response (502): failed to fetch members from Discord.

#### PUT /api/guilds/:guildId/config

Sets the guild's member role, inactive role, and ranked roles in a single operation.
Only the guild owner can configure. All role IDs must exist in the guild's stored role list (populated by bot verify).

Request body:
```json
{
  "activeRoleId": "discord_role_id",
  "inactiveRoleId": "discord_role_id",
  "rankedRoleIds": ["role_id_1", "role_id_2"]
}
```

- `activeRoleId` (required): role that marks a member as active; members without this role are skipped during sync.
- `inactiveRoleId` (optional): role that marks inactive members.
- `rankedRoleIds` (optional): roles that represent member ranks used in leaderboard sorting. Roles not listed are reset to `default` type.

Response (200):
```json
{ "ok": true }
```

Response (400): `activeRoleId` missing or any provided role ID not found in guild roles.
Response (401): missing, invalid, or expired token.
Response (403): authenticated user does not own this guild.
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

Response (401): missing, invalid, or expired token.
Response (403): authenticated user does not own this guild.
Response (404): guild not found.

#### GET /api/guilds/:guildId/dashboard

Returns one aggregated dashboard payload from guild, members, events, and event_reports sources.

Query parameters:
- `leaderboardBy` (optional): `score` | `hosted` | `attended`. Default: `score`.
- `leaderboardLimit` (optional): positive integer, max 200. Default: 10.
- `inactiveDays` (optional): positive integer. Default: 30.
- `eventStart` (optional): RFC3339 timestamp.
- `eventEnd` (optional): RFC3339 timestamp, must be greater than or equal to `eventStart`.
- `attendeeId` (optional): Discord user ID; filters events to reports containing this attendee.

Stats definitions:
- `totalMembers`, `activeMembers`, `inactiveMembers`: from `members` collection.
- `liveEvents`: count of `events` where status is `open` or `active`.
- `closedEvents`: count of `event_reports` (source of truth for completed events).
- `totalEvents = liveEvents + closedEvents`.
- `participationRate` (percentage): `(distinct members with at least one attended closed event / activeMembers) * 100`.
- `participationRate` returns `0` when `activeMembers` is `0`.

Response (200):
```json
{
  "ok": true,
  "dashboard": {
    "guild": {
      "guildId": "...",
      "name": "My Server",
      "ownerDiscordId": "...",
      "botInstalled": true,
      "createdAt": "...",
      "updatedAt": "..."
    },
    "stats": {
      "totalMembers": 42,
      "activeMembers": 38,
      "inactiveMembers": 4,
      "totalEvents": 17,
      "closedEvents": 12,
      "participationRate": 62.5
    },
    "leaderboard": [
      {
        "discordId": "1234567890",
        "eventsHosted": 5,
        "eventsAttended": 11,
        "score": 21,
        "lastHostedDate": "2026-06-10T20:00:00Z",
        "lastAttendedDate": "2026-06-14T20:00:00Z",
        "rank": 1
      }
    ],
    "members": [
      {
        "discordId": "1234567890",
        "rankedRoleId": "...",
        "status": "active",
        "discordJoinedAt": "2023-06-08T00:00:00Z",
        "roleIds": ["...", "..."],
        "eventsHosted": 5,
        "eventsAttended": 11,
        "lastHostedDate": "2026-06-10T20:00:00Z",
        "lastAttendedDate": "2026-06-14T20:00:00Z",
        "isInactiveByCutoff": false
      }
    ],
    "inactiveMembers": [],
    "events": [
      {
        "eventId": "...",
        "hostDiscordId": "...",
        "eventDate": "2026-06-14T20:00:00Z",
        "participantIds": ["1234567890", "0987654321"],
        "summary": "Event wrap-up notes"
      }
    ]
  }
}
```

Response (400): invalid `leaderboardBy`, `leaderboardLimit`, `inactiveDays`, `eventStart`, or `eventEnd`.
Response (401): missing, invalid, or expired token.
Response (404): guild not found.
Response (500): failed to load one or more dashboard data sources.

### Guild Event Logs

All event log endpoints require `Authorization: Bearer <token>`.
Access is permitted for the guild owner or any active guild member.

#### GET /api/guilds/:guildId/event-logs

Returns all event log records for a guild.

Response (200):
```json
{
  "ok": true,
  "logs": [
    {
      "id": "...",
      "eventId": "",
      "guildId": "...",
      "hostDiscordId": "...",
      "eventDate": "2026-06-14T20:00:00Z",
      "participantIds": ["1234567890", "0987654321"],
      "summary": "Event wrap-up notes",
      "submittedAt": "..."
    }
  ]
}
```

Response (401): missing, invalid, or expired token.
Response (403): authenticated user is not the guild owner or an active guild member.
Response (404): guild not found.

#### POST /api/guilds/:guildId/event-logs

Creates a manual event log record. Does not require a linked event document — used for logging events tracked outside the automated flow.

Request body:
```json
{
  "summary": "Event wrap-up notes",
  "eventDate": "2026-06-14T20:00:00Z",
  "participantIds": ["1234567890", "0987654321"],
  "hostDiscordId": "1234567890"
}
```

- `summary` (required): wrap-up text for the event.
- `eventDate` (required): RFC3339 timestamp for when the event occurred.
- `participantIds` (optional): list of attendee Discord IDs.
- `hostDiscordId` (optional): Discord ID of the host. Defaults to the authenticated user.

Response (200):
```json
{
  "ok": true,
  "log": {
    "id": "...",
    "guildId": "...",
    "hostDiscordId": "...",
    "eventDate": "2026-06-14T20:00:00Z",
    "participantIds": ["1234567890", "0987654321"],
    "summary": "Event wrap-up notes",
    "submittedAt": "..."
  }
}
```

Response (400): `summary` or `eventDate` missing or invalid.
Response (401): missing, invalid, or expired token.
Response (403): authenticated user is not the guild owner or an active guild member.
Response (404): guild not found.

#### PUT /api/guilds/:guildId/event-logs/:logId

Updates an existing event log record.

Request body:
```json
{
  "summary": "Updated wrap-up notes",
  "eventDate": "2026-06-14T20:00:00Z",
  "participantIds": ["1234567890", "0987654321"],
  "hostDiscordId": "1234567890"
}
```

All fields follow the same rules as `POST /api/guilds/:guildId/event-logs`.

Response (200): `{ "ok": true }`

Response (400): `summary` or `eventDate` missing or invalid.
Response (401): missing, invalid, or expired token.
Response (403): authenticated user is not the guild owner or an active guild member.
Response (404): guild or event log not found.

#### DELETE /api/guilds/:guildId/event-logs/:logId

Permanently deletes an event log record.

Response (200): `{ "ok": true }`

Response (401): missing, invalid, or expired token.
Response (403): authenticated user is not the guild owner or an active guild member.
Response (404): guild or event log not found.

### Event Operations

All event endpoints require `Authorization: Bearer <token>`.

#### Event Type Behavior

The `eventType` field on an event controls the Discord embed button set and available RSVP options:

| Event Type | Buttons | Maybe Tracking | Reminder DMs | Mod Mail |
|------------|---------|---------------|--------------|----------|
| **Raid** | ✅ Attending \| ❌ Not Attending | No | No | No |
| **Skirmish** | ✅ Attending \| ❌ Not Attending \| ❓ Maybe | Yes (`maybeIds`) | Yes (1h before event, to `maybeIds`) | Yes (non-responders) |

**Non-responders** are guild members not present in `attendingIds`, `notAttendingIds`, or `maybeIds`. Moderators can trigger a mod-mail DM blast to this group for Skirmish events to drive headcount.

Event type values are configured per-guild in `eventConfig.eventTypes` and selected via the `/event create` slash command autocomplete.

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

`eventType` drives the Discord embed button set — use `"Raid"` for a 2-button embed or `"Skirmish"` for a 3-button embed with Maybe. Other configured event types default to Raid behavior.

Request body:
```json
{
  "guildId": "...",
  "title": "Raid Night",
  "eventType": "Raid",
  "description": "Optional description, max 3000 characters",
  "scheduledAt": "2026-06-10T20:00:00Z",
  "channelId": "...",
  "reminderEnabled": true,
  "capacity": 20,
  "cutoffAt": "2026-06-10T19:00:00Z"
}
```

- `eventType` (optional): `"Raid"` or `"Skirmish"`. Defaults to `"Raid"` behavior when omitted.

Response (200):
```json
{ "ok": true, "event": { "id": "...", "guildId": "...", "title": "Raid Night", "eventType": "Raid", "status": "open", "attendingIds": [], "maybeIds": [], "notAttendingIds": [] } }
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
{
  "ok": true,
  "event": {
    "id": "...",
    "guildId": "...",
    "hostDiscordId": "...",
    "title": "Skirmish Saturday",
    "eventType": "Skirmish",
    "description": "...",
    "scheduledAt": "2026-06-10T20:00:00Z",
    "status": "open",
    "channelId": "...",
    "attendingIds": ["1234567890"],
    "maybeIds": ["0987654321"],
    "notAttendingIds": [],
    "capacity": 0,
    "reminderEnabled": true,
    "reminderSentAt": null,
    "cutoffAt": null,
    "createdAt": "...",
    "updatedAt": "..."
  }
}
```

Field notes:
- `maybeIds`: Discord IDs of members who responded "Maybe" (Skirmish only; always empty for Raid).
- `notAttendingIds`: Discord IDs of members who explicitly declined.
- `reminderSentAt`: set when 1-hour reminder DMs have been dispatched to `maybeIds`.

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

#### POST /api/events/:eventId/maybe

Marks the authenticated user as "Maybe" attending. Only valid for Skirmish events. Automatically removes the user from `attendingIds` and `notAttendingIds` if present.

Response (200):
```json
{ "ok": true }
```

Response (401): missing, invalid, or expired token.
Response (404): event not found.
Response (409): already marked as maybe, event type does not support maybe (Raid), registration is closed, or event is at capacity.

#### POST /api/events/:eventId/unmaybe

Removes the authenticated user's "Maybe" response.

Response (200):
```json
{ "ok": true }
```

Response (401): missing, invalid, or expired token.
Response (404): event not found.
Response (409): user is not in the maybe list.

#### POST /api/events/:eventId/decline

Marks the authenticated user as "Not Attending". Automatically removes the user from `attendingIds` and `maybeIds` if present.

Response (200):
```json
{ "ok": true }
```

Response (401): missing, invalid, or expired token.
Response (404): event not found.
Response (409): already declined, or registration is closed.

#### POST /api/events/:eventId/undecline

Removes the authenticated user's "Not Attending" response.

Response (200):
```json
{ "ok": true }
```

Response (401): missing, invalid, or expired token.
Response (404): event not found.
Response (409): user has not declined this event.

#### POST /api/events/:eventId/mod-mail

Sends a Discord DM to every active guild member who has not yet responded to the event (not in `attendingIds`, `notAttendingIds`, or `maybeIds`). Only available for Skirmish events. Caller must hold a moderator role configured in the guild's `moderatorRoleIds`.

Request body (optional):
```json
{
  "message": "Custom message override (optional). Defaults to a generated rally message."
}
```

Response (200):
```json
{
  "ok": true,
  "sent": 12,
  "failed": []
}
```

Response (401): missing, invalid, or expired token.
Response (403): caller does not hold a moderator role for this guild.
Response (404): event not found.
Response (409): event type does not support mod mail (Raid), or event is closed.
Response (422): no non-responding members to contact.

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

### Event Notifications

#### POST /api/notifications/events/reminders/run

Scans all open/active Skirmish events whose `scheduledAt` falls within the next hour, and sends a Discord DM reminder to every member in the event's `maybeIds` list who has not already received one (`reminderSentAt` is null). Sets `reminderSentAt` on each event after dispatching.

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
  "processed": 2,
  "reminded": 7,
  "failed": []
}
```

- `processed`: number of events scanned.
- `reminded`: total DMs sent across all events.
- `failed`: Discord IDs for which DM delivery failed.

Response (200, none eligible): `{ "ok": true, "processed": 0, "reminded": 0, "skipped": "no Skirmish events starting within 1 hour" }`
Response (400): `guildId` missing or invalid.
Response (401): missing, invalid, or expired token.
Response (404): guild not found.

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
