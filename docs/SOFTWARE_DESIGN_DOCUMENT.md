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
- app/db: Mongo connection and utility helpers
- app/discord: Discord OAuth2 client (authorization URL, code exchange, user and guild fetching)
- app/middleware: JWT validation middleware
- app/repositories: data access interfaces and MongoDB implementations
- app/routes: HTTP route modules
- app/schemas: request/response payload structures
- app/session: JWT signing, verification, and claims types

### Data Access Layering

GuildLogger2 uses the repository pattern to separate business logic from data access:

```
Route Handlers (app/routes)
    ↓
Repository Interfaces (app/repositories)
    ↓
MongoDB Implementations (app/repositories)
    ↓
Database Utilities (app/db/mongo_utils.go)
    ↓
MongoDB Driver & Database
```

**Benefits:**
- Route handlers call business-focused methods (e.g., `userRepo.FindByDiscordID()`) instead of BSON queries
- Repositories encapsulate all MongoDB query logic in one place
- Database utilities provide low-level helpers (collection access, BSON serialization)
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

The `summary` field is compressed on write and decompressed on read inside the MongoDB repository. Callers always receive a plain string.

### Key Workflows

1. Discord OAuth login
2. Guild connect and bot installation
3. Member sync: fetch from Discord → filter by active role → resolve ranked role → upsert → deactivate departed members
4. Event create/register/unregister
5. Scheduled reminders and anniversary notifications
6. Manual event log CRUD: guild owner or active member submits a log → stored as `EventReport` without a linked `eventId` → accessible via guild event logs page

## Deployment Model

- Local: Docker Compose for backend + mongo, local frontend dev server
- Production target: image-based deployment with managed secrets and persistent database

## Migration Notes

- Flask-to-Go migration is complete. The Go backend is the sole active runtime.
- Documentation describes current Go behavior only.
- Legacy Flask code is no longer referenced.
