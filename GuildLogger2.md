# GuildLogger2 Instruction Set

## Mission
Build GuildLogger2 as a production-ready web application plus Discord bot system that helps gaming communities organize hosted events, automate participation management, and maintain long-term member engagement data.

## Product Vision
GuildLogger2 is a community operations assistant for Discord-based gaming groups.  
It should reduce manual admin work, improve event participation, and create clear accountability for hosts and members through reliable automation, clean workflows, and measurable activity history.

## Core Objectives
1. Support user authentication through Discord OAuth.
2. Let server owners add and configure the bot in their Discord servers.
3. Automatically register server members with appropriate role checks, including verified and member role handling.
4. Create and manage hosted event prompts that members can join.
5. Track member participation and hosting statistics over time.
6. Track membership tenure and trigger one-year anniversary notifications.
7. Provide reminder mail and notification features for upcoming events.
8. Expose all major workflows through a web app with clear admin and member UX.

## Required Tech Stack
1. Backend: Go with Echo framework.
2. Frontend: React with Vite.
3. Database: MongoDB.
4. Integrations: Discord OAuth2 and Discord bot APIs.

## Architecture Direction
1. Refactor legacy backend logic from Python to Go without carrying old coupling or unused patterns.
2. Keep backend domain-oriented around:
   - Auth and identity
   - Guild and server management
   - Member lifecycle
   - Events and participation
   - Notifications and reminders
   - Analytics and stats
3. Treat Discord bot actions as first-class backend workflows, not ad hoc scripts.
4. Keep API contracts stable and versioned once exposed to frontend clients.
5. Design for eventual background jobs and scheduled processing for reminders and anniversaries.

## Functional Requirements

### Authentication and Identity
1. Implement Discord OAuth login and account linking.
2. Store Discord user ID as primary identity key.
3. Create session or token flow with secure expiration and refresh behavior.
4. Enforce permission boundaries for server owner and admin operations.

### Bot and Server Onboarding
1. Allow users to invite the bot only to servers they have authority over.
2. Persist guild-level configuration after bot installation.
3. Verify bot permissions required for member and event workflows.
4. Provide clear status in the app for connected servers and integration health.

### Member Registration and Role Automation
1. Sync members from Discord guild into app records.
2. Apply registration logic for verified and member role holders.
3. Prevent duplicate member records and maintain idempotent sync behavior.
4. Record join timestamps and lifecycle milestones.

### Event Management
1. Allow authorized users to create hosted event prompts.
2. Support event metadata: host, title, schedule, capacity, notes, and registration state.
3. Allow members to register and unregister with validation and cutoff rules.
4. Keep an auditable participant list and host ownership history.

### Stats and Community Analytics
1. Track hosted event count per member.
2. Track participation count per member.
3. Track tenure duration for each member.
4. Support queryable summaries for guild dashboards and profiles.

### Anniversary and Reminder Notifications
1. Identify members reaching one-year milestones.
2. Generate and deliver anniversary notices.
3. Send reminders for upcoming events to registered participants.
4. Track delivery state, retries, and failures for notification jobs.

## Non-Functional Requirements
1. Reliability: workflows must be idempotent where retries are possible.
2. Security: protect OAuth tokens, secrets, and privileged operations.
3. Observability: structured logs, error traces, and operational metrics.
4. Performance: responsive UI and efficient API and database access for active guilds.
5. Maintainability: clear module boundaries, tests, and documented contracts.
6. Compliance: align with Discord platform policies and user data expectations.

## Data Design Principles
1. Model around stable identifiers: guild_id, user_id, event_id.
2. Use timestamps for all lifecycle actions and state transitions.
3. Store immutable activity records for analytics integrity.
4. Separate configuration documents from event and member transactional records.

## Build Phases

### Phase 1: Foundation
1. Go Echo service scaffold with health checks and environment config.
2. MongoDB connection layer and base repositories.
3. Discord OAuth login flow and identity persistence.
4. Basic React app shell with auth state handling.

### Phase 2: Guild and Bot Integration
1. Guild linking and bot install flow.
2. Member sync and verified and member role registration pipeline.
3. Admin dashboard for connected guild status.

### Phase 3: Event Operations
1. Event creation, publishing, and registration flows.
2. Host and participant tracking model.
3. Frontend event boards and profile summaries.

### Phase 4: Notifications and Milestones
1. Scheduled jobs for reminders and anniversaries.
2. Delivery tracking and retry policies.
3. Admin controls for notification behavior.

### Phase 5: Hardening
1. Integration and end-to-end tests for critical journeys.
2. Access control verification and security review.
3. Performance profiling and production readiness checklist.

## Agent Working Rules
1. Prioritize correctness and data integrity over rapid feature count.
2. Do not introduce hidden coupling between Discord API behavior and UI assumptions.
3. Keep migrations and schema changes explicit and reversible.
4. Write tests for high-risk logic first:
   - OAuth and authorization boundaries
   - Member sync idempotency
   - Event registration edge cases
   - Notification scheduling and retries
5. Prefer incremental delivery with clear acceptance criteria per feature.
6. Document every API endpoint and event state transition as features are added.

## Acceptance Criteria for MVP
1. A user can sign in with Discord.
2. A qualified user can connect at least one Discord server.
3. Bot-assisted member registration works for verified and member role users.
4. Authorized hosts can create events and members can register.
5. Host and participation counters update correctly.
6. One-year membership anniversaries can be detected and notified.
7. Upcoming event reminders can be dispatched and tracked.
8. Core workflows are covered by automated tests.

## Definition of Done
1. Feature behavior matches documented requirements.
2. Edge cases and failure paths are tested.
3. Security and permission checks are present.
4. Observability is adequate for troubleshooting.
5. Documentation is updated with API and workflow changes.