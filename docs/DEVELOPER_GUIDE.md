# Developer Guide

Development standards and workflow for GuildLogger2.

## Tech Stack

- Backend: Go + Echo
- Frontend: React + Vite + TypeScript
- Data: MongoDB
- Integrations: Discord OAuth2 and Discord bot APIs

## Branching

Recommended branch naming:

- feat/short-description
- fix/short-description
- docs/short-description
- refactor/short-description
- chore/short-description

## Backend Standards (Go)

- Run formatting before commit: go fmt ./...
- Resolve modules: go mod tidy
- Compile check: go build ./...
- Keep package boundaries clear (config, db, routes, schemas).
- Keep main entrypoint thin; app wiring belongs in app package.

## Frontend Standards (TypeScript)

- Lint before commit: npm run lint
- Build check: npm run build
- Use strict typing for API payloads and route params.

## Runtime Workflows

### Local backend

Run scripts/start_backend.sh from repo root.

### Full app (docker backend + local frontend)

Run scripts/start_app.sh from repo root.

## Code Review Checklist

1. Behavior matches issue scope.
2. API contract changes are documented in docs/API_REFERENCE.md.
3. Config changes are documented in docs/CONFIGURATION.md.
4. Legacy references are removed when touching migrated areas.
5. Tests or smoke checks are included in PR notes.

## Commit Guidance

Use clear, scoped commit messages:

- feat(auth): scaffold discord oauth flow
- fix(startup): force backend rebuild in start script
- docs(api): align endpoints with go migration

## Documentation Policy

- Keep docs aligned with the active Go runtime.
- Do not copy legacy Flask behavior into new docs unless marked as historical context.
