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
- app/routes: HTTP route modules
- app/schemas: request/response payload structures

### Runtime

- Health endpoint supports readiness checks.
- CORS policy is environment-driven.
- Request logging and panic recovery middleware are enabled.

## Frontend Design

- Route-driven pages for auth, account, and operations.
- API client layer centralizes backend requests.
- Auth state managed in a shared context/provider.

## Data Design

Primary identifiers:

- guild_id
- user_id
- event_id

Collections are separated by concern (identity, guild config, events, participation, notifications).

## Key Workflows

1. Discord OAuth login
2. Guild connect and bot installation
3. Member sync and eligibility checks
4. Event create/register/unregister
5. Scheduled reminders and anniversary notifications

## Deployment Model

- Local: Docker Compose for backend + mongo, local frontend dev server
- Production target: image-based deployment with managed secrets and persistent database

## Migration Notes

- Legacy Flask code is retained temporarily as a reference baseline.
- Route migration proceeds domain-by-domain to Go modules.
- Documentation should describe current Go behavior, not historical Flask behavior.
