# Requirements Compliance Snapshot

## Purpose

Track compliance against GuildLogger2 target requirements during the Flask-to-Go migration.

## Summary

| Area | Status | Notes |
|---|---|---|
| Runtime migration to Go backend | Implemented | Go service, health endpoint, full auth + guild routes active |
| Discord OAuth auth flow | Implemented | Full flow: URL generation, code exchange, JWT issuance, session validation |
| Guild and bot onboarding | Implemented | Guild linking, Discord membership validation, bot invite URL, install confirmation |
| Member sync and role eligibility | Not Started | Planned domain module |
| Event operations | Not Started | Planned domain module |
| Analytics counters and tenure | Not Started | Planned domain module |
| Notifications (anniversary/reminders) | Not Started | Planned background workflow |
| Frontend integration with new API domains | In Progress | Auth and guild domains complete; events, members, analytics pending |

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
