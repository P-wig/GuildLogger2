# Testing Strategy

GuildLogger2 testing guidance for the Go migration phase.

## Testing Goals

- Prevent regressions while moving from Flask to Go.
- Verify route behavior and payload contracts.
- Validate database interactions and retry-safe logic.

## Test Layers

### Unit Tests

- Go utility functions and validation helpers
- Request/response schema validation
- Event rules and stats calculations

### Integration Tests

- Echo route handlers with test Mongo database
- Auth and authorization boundaries
- Member sync idempotency
- Notification scheduling outcomes

### End-to-End Tests

- Critical UI journeys (auth, event creation, registration)
- Cross-service behavior with backend + database running

## Current Practical Smoke Checks

### Backend local

1. go fmt ./...
2. go build ./...
3. go run .
4. GET /api/health returns ok

### Docker backend

1. docker compose up -d --build mongo backend
2. verify backend command uses Go binary
3. GET /api/health returns ok

## Coverage Priorities

1. Auth flows
2. Guild/member sync logic
3. Event capacity and cutoff validation
4. Notifications and retries

## Tooling Direction

- Backend: go test with table-driven tests
- Frontend: Vitest + Testing Library
- E2E: Playwright (planned)

## Exit Criteria for Migration Milestones

- Build passes
- Health and core route smoke checks pass
- Updated docs reflect current implementation
