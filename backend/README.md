# Backend Service (Go + Echo)

This backend is being migrated from Flask to Go and now runs on Echo with MongoDB.

## Current Status
- Runtime: Go
- Web framework: Echo
- Database: MongoDB
- Available routes:
	- `GET /`
	- `GET /api/health`

## Requirements
- Go 1.26+
- Docker Desktop (recommended for local MongoDB)

## Project Structure
Key files and directories:
- `main.go` - backend entrypoint
- `app/app.go` - Echo app setup, middleware, and route registration
- `app/config/` - environment-based config loader
- `app/db/` - MongoDB connection and helpers
- `app/routes/` - route handlers

## Configuration
Environment variables used by the Go app:
- `PORT` - backend port (default: `5001`)
- `MONGO_URI` - Mongo connection string (default: `mongodb://localhost:27017`)
- `MONGO_DB` - Mongo database name (default: `test`)
- `CORS_ORIGINS` - comma-separated origins (default: `http://localhost:5173`)

Example (PowerShell):
```powershell
$env:PORT="5001"
$env:MONGO_URI="mongodb://localhost:27017"
$env:MONGO_DB="cloudnative"
$env:CORS_ORIGINS="http://localhost:5173"
```

## Run Locally (Go)
From `backend/`:
```bash
go mod tidy
go fmt ./...
go build ./...
go run .
```

Health check:
```bash
curl http://localhost:5001/api/health
```

## Run with Docker Compose
From repo root:
```bash
docker compose up -d --build mongo backend
```

Check backend logs:
```bash
docker compose logs backend --tail=100
```

Health check:
```bash
curl http://localhost:5001/api/health
```

Stop services:
```bash
docker compose down --remove-orphans
```

## Notes
- Python/Flask files may still exist in the repo as migration references.
- The active backend runtime path is Go (`main.go` + Echo app wiring).
