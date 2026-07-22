# Requirements Compliance Snapshot

## Purpose

Track compliance against GuildLogger2 target requirements during the Flask-to-Go migration.

## Summary

| Area | Status | Notes |
|---|---|---|
| Runtime migration to Go backend | Implemented | Flask-to-Go migration complete; Go is the sole active backend runtime |
| Discord OAuth auth flow | Implemented | Full flow: URL generation, code exchange, JWT issuance, session validation |
| Guild and bot onboarding | Implemented | Guild linking, Discord membership validation, bot invite URL, install confirmation |
| Member sync and role eligibility | Implemented | Sync pipeline, verified role filter, idempotency, departed member deactivation |
| Frontend integration with new API domains | Implemented | Auth, guild, dashboard, event logs, and member stats complete |
| Manual event log CRUD (guild owner / member) | Implemented | GET/POST/PUT/DELETE /api/guilds/:guildId/event-logs; GuildEvents frontend page |

## Migration Compliance Risks

1. Legacy documentation drift if route changes are not documented.
2. Partial endpoint migration creating mixed response contracts.
3. Stale container/image behavior during runtime transition.

## Mitigations

1. Keep API_REFERENCE.md updated in each backend route PR.
2. Standardize response envelopes for newly migrated endpoints.
3. Use compose rebuild flags in migration scripts when required.

## Definition of Done (Current Phase)

- Go backend builds and runs locally and in Docker.
- Health endpoint passes through local and compose startup paths.
- Legacy Flask runtime is no longer the active container path.
- Documentation reflects current architecture and migration status.
