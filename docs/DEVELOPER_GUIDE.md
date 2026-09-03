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
- Event lifecycle logic belongs in `app/discord/event_service.go`, not in route or interaction
  handlers. Both transports must call the service so Discord and the web API cannot drift.
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

## Authentication Architecture

GuildLogger uses a two-layer authentication model. The layers are complementary
and operate independently — neither replaces the other.

### Frontend: `ProtectedRoute` (UX guard)

`ProtectedRoute` checks `AuthContext` (sourced from `localStorage`) before
rendering any protected page. If no session is found, the user is redirected to
`/auth` before any API call is made. This is a convenience guard only — it can
be bypassed by a user who manually sets `localStorage`.

### Backend: JWT Middleware (security boundary)

Every request to a protected API route is validated by the JWT middleware in
`backend/app/middleware/jwt.go`. The middleware verifies the token signature and
expiry on every request, independent of what the frontend believes. A missing,
forged, or expired token always results in `401 Unauthorized`.

### Request flow

User navigates to protected route
│
▼
ProtectedRoute checks localStorage
│
No token? ──────────────────► redirect to /auth (no API call made)
│
Token exists
▼
Page renders and calls backend API
│
▼
Backend JWT middleware validates token
│
Invalid/expired? ──────────────► 401 → http.ts clears localStorage
│
Valid
▼
Handler runs, returns data


### 401 handling

`frontend/src/api/http.ts` contains a response interceptor that clears
`localStorage` on a `401` response. This ensures stale or expired tokens are
removed automatically. Any component that receives a `401` should also call
`logout()` from `useAuth()` to clear the in-memory auth state and trigger a
redirect to `/auth`.