# GuildLogger2

A full-stack web application for managing Discord guild activity, member roles, and hardware/resource tracking.

## Tech Stack

- **Backend**: Go 1.26+ + Echo v4 + MongoDB
- **Frontend**: React + Vite + TypeScript
- **Development**: Scripts for automated startup and dependency management
- **Deployment**: Docker Compose (backend + database containers)

## Prerequisites

- **Go 1.26+** — [golang.org/dl](https://golang.org/dl/)
- **Node.js** (latest LTS) — [nodejs.org](https://nodejs.org/)
- **Docker Desktop** — [docs.docker.com/desktop](https://docs.docker.com/desktop/)
- **MongoDB** connection (provided by Docker Compose or a remote URI)

## Development Setup

### Backend Setup

The `backend/` folder is a Go module (`github.com/P-wig/GuildLogger2/backend`).

1. **Install Go dependencies**:

   ```bash
   cd backend
   go mod tidy
   ```

2. **Configure environment variables**:

   ```bash
   cp .env.example .env
   # Edit .env — set MONGO_URI, MONGO_DB, CORS_ORIGINS as needed
   ```

### Frontend Setup

The frontend uses React + Vite for fast development and hot module replacement (HMR).

1. **Navigate to frontend directory and install dependencies**:
   ```bash
   cd frontend
   npm install
   ```
### Start Individual Components

#### Backend Only

```bash
./scripts/start_backend.sh
```

This script will:

- Check for a Go installation
- Run `go mod tidy`
- Start the Go backend with `go run .`

The backend will be available at `http://localhost:5001` and assumes a separate MongoDB instance is already available.

#### Frontend Only

```bash
./scripts/start_frontend.sh
```

This script will:

- Check if dependencies are installed (runs `npm install` if needed)
- Start the Vite development server

The frontend will be available at `http://localhost:5173`

### Start Full Application

Launch the entire application locally using a convenient script:

```bash
./scripts/start_app.sh
```
This assumes you have Docker and `docker-compose` installed on your machine. 
For Mac users, install from [here](https://docs.docker.com/desktop/setup/install/mac-install/)
and use `brew` to install `docker-compose`. 

>__NOTE__: Ensure the docker socket is available by first running Docker Desktop.

This orchestrator script will:

- Start the backend services in the background using `docker compose`
- Start the frontend server in the background
- Monitor both processes for errors
- Provide log file locations for debugging
- Gracefully shut down both services when you press `Ctrl+C`

**Log files**:

- Backend: `/tmp/backend.log`
- Frontend: `/tmp/frontend.log`

You can tail the logs in separate terminal windows:

```bash
tail -f /tmp/backend.log
tail -f /tmp/frontend.log
```

## API Endpoints

Once running, the backend provides the following endpoints:

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Service info |
| `GET` | `/api/health` | Health check |
| `POST` | `/api/auth/register` | Register a new user |
| `POST` | `/api/auth/login` | Log in |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `5001` | Backend listen port |
| `MONGO_URI` | `mongodb://localhost:27017` | MongoDB connection URI |
| `MONGO_DB` | `guildlogger` | Database name |
| `CORS_ORIGINS` | `http://localhost:5173` | Allowed CORS origins (comma-separated) |

## Running the Application


## Troubleshooting

### Port Already in Use

If you see "Address already in use" for port 5001:

```bash
lsof -ti:5001 | xargs kill -9
```

### Vite Command Not Found

```bash
cd frontend
npm install
```

### Docker Container Running Stale Image

```bash
docker compose down --remove-orphans
docker compose build --no-cache backend
docker compose up -d
```

## Alternative Way to Start the Application

In one terminal, run from the repo root:

```bash
docker compose up -d --build --remove-orphans
```

In a second terminal, run:

```bash
cd frontend
npm run dev
```
