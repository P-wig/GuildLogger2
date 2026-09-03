# Software Design Document (SDD)

## System Overview

GuildLogger2 uses a web frontend and API backend with MongoDB persistence and Discord integrations.

## High-Level Architecture

- Frontend: React + Vite
- Backend: Go + Echo
- Database: MongoDB
- Integrations: Discord OAuth2 and bot APIs

## Backend Design

### Application Wiring

- main entrypoint initializes app creation and graceful shutdown.
- app package configures middleware, CORS, DB lifecycle, and route registration.

### Package Responsibilities

- app/config: environment configuration
- app/db: Mongo connection and collection accessors
- app/discord: Discord OAuth2 client, bot API client, interaction webhook handlers, and the shared event lifecycle service
- app/middleware: JWT validation middleware
- app/repositories: data access interfaces and MongoDB implementations
- app/routes: HTTP route modules
- app/schemas: request/response payload structures
- app/session: JWT signing, verification, and claims types

### Dual-Transport Event Lifecycle

GuildLogger exposes the same event domain through two interfaces: the Discord bot and the
REST API. Both are **transports over a single implementation**, `discord.EventService`.
Neither owns lifecycle logic.

```
Discord interactions ──┐
                       ├──▶ discord.EventService ──▶ repositories ──▶ MongoDB
REST /api/events/* ────┘             │
                                     └──▶ Discord REST (embeds, voice channels)
```

This matters because event state and Discord state must stay consistent. Creating an event
also posts an announcement message whose ID is stored on the event; RSVPs re-render that
embed; starting an event creates a voice channel; closing the channel deletes both the
channel and the event document. A transport that wrote the database directly would leave
Discord stale, so neither is permitted to.

`EventService` splits each operation into two phases:

| Phase | Methods | Cost | Used by |
|---|---|---|---|
| transition | `CreateEvent`, `SetRSVP`, `StartEvent`, `EndEvent`, `CloseEventChannel` | fast, validated, DB-only | both; fits Discord's 3-second interaction deadline |
| effects | `ApplyStartEffects`, `ApplyEndEffects`, `ApplyCloseChannelEffects`, `RefreshAnnouncement` | slow Discord API calls | Discord runs these in a goroutine after acknowledging; REST runs them inline |

The service lives in `app/discord` rather than a separate package so it can use the
unexported embed and button builders without exporting them, and because `app/routes`
already imports `app/discord` (no import cycle).

### Event Lifecycle

```
open ──start──▶ active ──end──▶ closed ──close-channel──▶ (event document deleted)
```

| Step | State change | Discord side effects |
|---|---|---|
| create | event written, status `open` | announcement embed posted; `announcementMessageId` stored |
| RSVP | roster lists updated | announcement embed re-rendered in place |
| start | `open` → `active` | voice channel created under the configured category; host moved in; embed refreshed |
| end | `active` → `closed` | voice roster snapshotted to `voiceMemberIds`; embed refreshed to enable Close Channel |
| close-channel | event deleted | members returned to lobby; voice channel deleted |

The event log is submitted separately through a signed one-time token
(`/api/event-log/submit`). The resulting `EventReport` is the permanent record and outlives
the event document, which is why closing the channel can safely delete the event.

### Authorization Model

JWT validation only proves *who* the caller is. Every guild-scoped and event-scoped endpoint
additionally resolves a membership tier via `getGuildMemberTier`:

| Tier | Source | Grants |
|---|---|---|
| `none` | not a synced member | no access |
| `member` | synced active member | read guild data, RSVP, run the lifecycle on events they host |
| `moderator` | holds a role in `moderatorRoleIds` | write event logs, send mod mail |
| `owner` | guild owner | full access including configuration |

Host-only actions (start, end, close-channel) check event ownership on top of the tier check.
A valid session alone is never sufficient to reach another guild's data.

### Data Access Layering

GuildLogger2 uses the repository pattern to separate business logic from data access:

```
Route Handlers (app/routes) / Interaction Handlers (app/discord)
    ↓
Event Lifecycle Service (app/discord/event_service.go)
    ↓
Repository Interfaces (app/repositories)
    ↓
MongoDB Implementations (app/repositories)
    ↓
Collection Accessors (app/db/mongo_utils.go)
    ↓
MongoDB Driver & Database
```

Handlers that perform no Discord side effects (guild config, member sync, dashboards) call
repositories directly and skip the service layer.

**Benefits:**
- Route handlers call business-focused methods (e.g., `userRepo.FindByDiscordID()`) instead of BSON queries
- Repositories encapsulate all MongoDB query logic in one place
- Database utilities provide collection accessors as a single source of truth for collection names
- Testability: repository interfaces can be mocked for unit tests without touching the database

### Runtime

- Health endpoint supports readiness checks.
- CORS policy is environment-driven.
- Request logging and panic recovery middleware are enabled.

## Frontend Design

- Route-driven pages for auth, account, and operations.
- API client layer centralizes backend requests.
- Auth state managed in a shared context/provider.

### Frontend Routes

| Path | Component | Access |
|---|---|---|
| `/` | Welcome | Public |
| `/auth` | Auth | Unauthenticated only |
| `/app` | Home | Protected |
| `/app/account` | Account | Protected |
| `/app/guilds` | Guilds | Protected |
| `/app/guilds/:guildId` | GuildDashboard | Protected |
| `/app/guilds/:guildId/events` | GuildEvents | Protected |

## Data Design

Primary identifiers:
- `guildId` — Discord guild snowflake ID
- `discordId` — Discord user snowflake ID
- `eventId` — MongoDB ObjectID

### Member Lifecycle Fields

The `Member` document tracks the full lifecycle of a guild member:

| Field | Type | Description |
|---|---|---|
| `discordJoinedAt` | time | When the member joined the Discord guild (from Discord API) |
| `firstSyncedAt` | time | When GuildLogger first recorded this member |
| `lastSyncedAt` | time | Updated on every sync pass that confirms the member is present |
| `deactivatedAt` | *time (nullable) | Set when status transitions to `inactive`; cleared on reactivation |
| `status` | string | `active` or `inactive` |
| `rankedRoleId` | string | The highest-position ranked role the member holds |

### Error Sentinels

Repositories expose typed error sentinels for expected domain failures:

- `ErrGuildAlreadyExists` — returned by `Create` when a guild with the same `guildId` already exists
- `ErrGuildNotFound` — returned by `Update` and all read-modify-write methods when the guild document is missing
- `ErrReportNotFound` — returned by `Update` and `Delete` when the event report document is missing

### EventReport Document

| Field | Type | Description |
|---|---|---|
| `id` | string (ObjectID hex) | Unique report identifier, generated on create |
| `eventId` | string (optional) | Links to a bot-managed `Event` document; absent for manual logs |
| `guildId` | string | Discord guild snowflake ID |
| `hostDiscordId` | string | Discord ID of the event host |
| `eventDate` | time | When the event occurred |
| `participantIds` | []string | Discord IDs of attendees |
| `summary` | []byte | Event wrap-up text, stored zlib-compressed in MongoDB |
| `submittedAt` | time | When the log was created |
| `submittedByDiscordId` | string | Discord ID of whoever submitted the log |
| `logsChannelId` | string (optional) | Channel holding this report's embed in the logs channel |
| `logsMessageId` | string (optional) | Discord message ID of that embed |

The `summary` field is compressed on write and decompressed on read inside the MongoDB repository. Callers always receive a plain string.

`logsChannelId` and `logsMessageId` let the dashboard keep Discord in sync: creating a log
posts an embed and stores its ID, editing a log edits that message in place (reposting if it
has gone missing), and deleting a log deletes the message.

### Key Workflows

1. Discord OAuth login
2. Guild connect and bot installation
3. Member sync: fetch from Discord → filter by active role → resolve ranked role → upsert → deactivate departed members
4. Event lifecycle: create → RSVP → start → end → submit log → close channel, driven identically from Discord buttons or the REST API
5. Scheduled reminders and anniversary notifications
6. Manual event log CRUD: moderator submits a log from the dashboard → stored as `EventReport` without a linked `eventId` → mirrored to the guild's logs channel as an embed that is kept in sync on edit and delete

## Deployment Model

- Local: Docker Compose for backend + mongo, local frontend dev server
- Production target: image-based deployment with managed secrets and persistent database

## Migration Notes

- Flask-to-Go migration is complete. The Go backend is the sole active runtime.
- Documentation describes current Go behavior only.
- Legacy Flask code is no longer referenced.
