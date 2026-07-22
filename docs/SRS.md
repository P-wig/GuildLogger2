# Software Requirements Specification (SRS)

## Project

GuildLogger2

## Purpose

GuildLogger2 is a web application plus Discord bot integration that automates event operations for gaming communities.

## Scope

The system supports Discord-authenticated users, guild onboarding, member sync, event registration workflows, and community analytics.

## Functional Requirements

### FR-01 Authentication

- Users can authenticate via Discord OAuth2.
- Session state is represented securely and expires appropriately.

### FR-02 Guild and Bot Onboarding

- Authorized users can connect guilds they manage.
- The bot can be installed and validated for required permissions.

### FR-03 Member Lifecycle

- Guild members are synchronized into application records.
- Verified/member role checks control registration eligibility.
- Sync process is idempotent and safe to retry.

### FR-04 Event Operations

- Authorized hosts can create events.
- Members can register and unregister.
- Capacity and cutoff rules are enforced.

### FR-05 Analytics

- Track hosted event counts per member.
- Track participation counts per member.
- Track membership tenure.

### FR-06 Notifications

- Trigger anniversary notifications at one-year milestones.
- Send reminders for upcoming events.
- Track delivery outcomes and retries.

## Non-Functional Requirements

- Reliability: retry-safe workflows and consistent state transitions.
- Security: protected secrets and role-based authorization.
- Observability: structured logs and actionable errors.
- Performance: responsive API and UI under expected guild load.
- Maintainability: modular design and testable components.

## Constraints

- Backend runtime is Go + Echo.
- Frontend runtime is React + Vite.
- Persistence is MongoDB.
- Discord APIs are external dependencies.

## Current Implementation Status

- Fully implemented: Go backend with all core domains — auth, guild management, member sync, event logs, analytics, and anniversary notifications.
- Flask-to-Go migration complete.
- Remaining planned work: event reminder notifications (POST /api/notifications/events/reminders/run).
