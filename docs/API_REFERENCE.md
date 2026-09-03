# API Reference

GuildLogger2 backend API reference.

Every endpoint listed here is implemented in the current backend.

## Authorization

All `/api` routes except the service root, health check, auth endpoints, the Discord
interaction webhook, and the token-based event-log endpoints require
`Authorization: Bearer <token>`.

A valid token establishes identity only. Guild-scoped and event-scoped endpoints
additionally check the caller's membership tier in the target guild — see
[Event Operations](#event-operations) for the tier table.

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

#### PUT /api/guilds/:guildId/config/event

Sets the guild's event-type list and the channels used by the event lifecycle.
Only the guild owner can configure.

Request body:
```json
{
  "eventTypes": [
    { "name": "Raids", "channelId": "discord_text_channel_id", "isQuickEvent": true },
    { "name": "Skirms", "channelId": "discord_text_channel_id", "isQuickEvent": false }
  ],
  "voiceCategoryId": "discord_category_id",
  "lobbyChannelId": "discord_voice_channel_id",
  "logsChannelId": "discord_text_channel_id",
  "commandChannelId": "discord_text_channel_id"
}
```

- `eventTypes` (required): the guild's event types. Entries with a blank `name` are dropped.
  - `name`: displayed in `/event create` autocomplete and used as the event title.
  - `channelId`: where this type's announcement is posted. Also gates `/event create <name>` —
    the command is rejected outside this channel.
  - `isQuickEvent`: `true` gives a 2-button uncapped embed with no description; `false` gives a
    3-button embed (including Maybe) capped at 99 attendees.
- `voiceCategoryId` (optional): category under which event voice channels are created on start.
  Voice channel creation is skipped when unset.
- `lobbyChannelId` (optional): voice channel members are returned to when an event's channel is closed.
- `logsChannelId` (optional): text channel where submitted event logs are posted as embeds.
  Log embeds are edited in place when a log is updated and deleted when a log is deleted.
- `commandChannelId` (optional): when set, **all** slash commands are rejected outside this
  channel. Leave blank to allow commands anywhere. This is a coarser gate than the per-event-type
  `channelId` above and applies in addition to it.

Response (200):
```json
{ "ok": true }
```

Response (400): invalid request body.
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

**Authorization model.** A valid session is never sufficient on its own. Every endpoint in
this section additionally resolves the caller's membership tier in the event's guild:

| Tier | Who | Event access |
|------|-----|--------------|
| `member` | any synced, active guild member | read events, RSVP, run the lifecycle on events they host |
| `moderator` | holds a role in `moderatorRoleIds` | additionally may send mod mail |
| `owner` | guild owner | full access |

Callers who are not members of the target guild receive `403`. Host-only actions (start,
end, close-channel) additionally require the caller to be the event host.

#### Event Type Behavior

Event types are **configured per guild** in `eventConfig.eventTypes`. Each entry has a
`name`, a `channelId`, and an `isQuickEvent` flag. There are no built-in type names.

`isQuickEvent` controls the embed button set and the event's shape:

| `isQuickEvent` | RSVP buttons | Capacity | Description field |
|----------------|--------------|----------|-------------------|
| `true` | ✅ Attending \| ❌ Not Attending | uncapped | not used — the announcement line is the message |
| `false` | ✅ Attending \| ❌ Not Attending \| ❓ Maybe | 99 | host-supplied rally message |

The `channelId` on each event type is authoritative: `/event create <type>` is only accepted
in that channel, and both the slash command and `POST /api/events` post the announcement
there. Callers do not choose the channel or capacity.

**Non-responders** are guild members not present in `attendingIds`, `notAttendingIds`, or
`maybeIds`. Moderators can trigger a mod-mail DM blast to this group to drive headcount.

#### Shared lifecycle

The Discord bot and this REST API are two transports over one implementation
(`discord.EventService`), so both produce identical side effects:

```
open ──start──▶ active ──end──▶ closed ──close-channel──▶ (event deleted)
```

- **create** — writes the event and posts the RSVP announcement, storing its message ID
- **RSVP** — mutates the roster and re-renders the announcement embed in place
- **start** — creates the event voice channel, moves the host in, refreshes the embed
- **end** — snapshots the voice roster, marks the event `closed`, refreshes the embed
- **close-channel** — returns members to the lobby, deletes the voice channel, deletes the event

The event log is submitted separately via `POST /api/event-log/submit`. The **EventReport**
is the permanent record and outlives the event document.

#### POST /api/events

Creates an event and posts its Discord announcement. Equivalent to `/event create` in Discord.

The announcement channel and capacity are resolved from the guild's configuration for
`eventType` — they are not accepted from the client.

Request body:
```json
{
  "guildId": "...",
  "eventType": "Skirms",
  "description": "Optional rally message; ignored for quick event types",
  "scheduledAt": "2026-06-10T20:00:00Z"
}
```

Response (200):
```json
{ "ok": true, "event": { "id": "...", "guildId": "...", "eventType": "Skirms", "status": "open", "channelId": "...", "announcementMessageId": "...", "attendingIds": [], "maybeIds": [], "notAttendingIds": [] } }
```

Response (200, announcement failed): `{ "ok": true, "event": {...}, "warning": "event created but the Discord announcement failed" }`

Response (400): `guildId`, `eventType`, or `scheduledAt` missing, or no channel configured for the event type.
Response (401): missing, invalid, or expired token.
Response (403): caller is not a member of the guild.

#### POST /api/events/:eventId/start

Transitions an event from `open` to `active`, creates the event voice channel, moves the
host into it, and refreshes the announcement embed. Only the event host can start the event.

Response (200): `{ "ok": true }`

Response (401): missing, invalid, or expired token.
Response (403): caller is not a guild member, or is not the event host.
Response (404): event not found.
Response (409): event is not currently in `open` status.

#### POST /api/events/:eventId/end

Transitions an event from `active` to `closed`, snapshots the voice-channel roster into
`voiceMemberIds` (used to pre-fill the log form), and refreshes the announcement embed so
the Close Channel control becomes available. Only the event host can end the event.

Submit the log afterwards via `POST /api/event-log/submit`.

Response (200): `{ "ok": true }`

Response (401): missing, invalid, or expired token.
Response (403): caller is not a guild member, or is not the event host.
Response (404): event not found.
Response (409): event is not in `active` status, or has already been logged.

#### POST /api/events/:eventId/close-channel

Final lifecycle step. Returns members to the configured lobby channel, deletes the event
voice channel, and deletes the event document. The EventReport remains as the permanent
record. Only the event host can close the channel.

Response (200): `{ "ok": true }`

Response (401): missing, invalid, or expired token.
Response (403): caller is not a guild member, or is not the event host.
Response (404): event not found.
Response (409): event is not in `closed` status.

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
Response (403): caller is not a member of the guild.

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
    "title": "Skirms",
    "eventType": "Skirms",
    "description": "...",
    "scheduledAt": "2026-06-10T20:00:00Z",
    "status": "open",
    "channelId": "...",
    "attendingIds": ["1234567890"],
    "maybeIds": ["0987654321"],
    "notAttendingIds": [],
    "capacity": 0,
    "announcementMessageId": "...",
    "voiceChannelId": "...",
    "voiceMemberIds": [],
    "reminderEnabled": true,
    "reminderSentAt": null,
    "cutoffAt": null,
    "createdAt": "...",
    "updatedAt": "..."
  }
}
```

Field notes:
- `maybeIds`: Discord IDs of members who responded "Maybe" (non-quick event types only; always empty when `isQuickEvent` is true).
- `notAttendingIds`: Discord IDs of members who explicitly declined.
- `announcementMessageId`: Discord message ID of the RSVP embed; used to re-render it in place.
- `voiceChannelId` / `voiceMemberIds`: set when the event is started and snapshotted when it ends.
- `reminderSentAt`: set when 1-hour reminder DMs have been dispatched to `maybeIds`.

Response (401): missing, invalid, or expired token.
Response (403): caller is not a member of the event's guild.
Response (404): event not found.

#### RSVP endpoints

All six RSVP actions share the same behavior: they mutate the caller's own attendance using
the Discord ID from the session token, then re-render the Discord announcement embed so the
roster shown in Discord matches the web UI immediately.

| Endpoint | Effect |
|----------|--------|
| `POST /api/events/:eventId/register` | adds to `attendingIds` |
| `POST /api/events/:eventId/unregister` | removes from `attendingIds` |
| `POST /api/events/:eventId/maybe` | adds to `maybeIds`, clears the other two lists |
| `POST /api/events/:eventId/unmaybe` | removes from `maybeIds` |
| `POST /api/events/:eventId/decline` | adds to `notAttendingIds`, clears the other two lists |
| `POST /api/events/:eventId/undecline` | removes from `notAttendingIds` |

Response (200):
```json
{ "ok": true, "event": { "id": "...", "attendingIds": [], "maybeIds": [], "notAttendingIds": [] } }
```

Response (401): missing, invalid, or expired token.
Response (403): caller is not a member of the event's guild.
Response (404): event not found.
Response (409): already in the requested state, not in the list being removed from, registration is closed, or event is at capacity.

#### POST /api/events/:eventId/mod-mail

Sends a Discord DM to every active guild member who has not yet responded to the event (not in `attendingIds`, `notAttendingIds`, or `maybeIds`). Available for any event type. Caller must hold a moderator role configured in the guild's `moderatorRoleIds`.

Members who have opted out via `notificationsOptOut` are excluded.

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
Response (409): event is closed.
Response (422): no non-responding members to contact.

#### GET /api/event-reports

Returns all permanent event reports for a guild, newest first. Reports outlive the events
they came from — an event document is deleted once its channel is closed, but its report
remains.

Query parameters:
- `guildId` (required)

Response (200):
```json
{ "ok": true, "reports": [] }
```

Response (400): `guildId` query parameter missing.
Response (401): missing, invalid, or expired token.
Response (403): caller is not a member of the guild.

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

### Event Notifications

#### POST /api/notifications/events/reminders/run

Scans all open/active events whose `scheduledAt` falls within the next hour and that have at
least one "Maybe" RSVP, then sends a Discord DM reminder to every member in the event's
`maybeIds` list who has not already received one (`reminderSentAt` is null). Sets
`reminderSentAt` on each event after dispatching.

Because only non-quick event types expose a Maybe button, quick events are never eligible.
There is no event-type name filter.

This also runs automatically on an hourly scheduler; the endpoint triggers a pass on demand.

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
