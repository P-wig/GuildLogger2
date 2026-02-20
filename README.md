# Monolith Full-Stack Application

This repository contains a monolithic codebase for Part 1 of the class project in ECE 382V: Cloud Native App Development.

## Tech Stack

- **Backend**: Flask + MongoDB (Python 3.12+)
- **Frontend**: React + Vite + TypeScript
- **Development**: Scripts for automated startup and dependency management

## Prerequisites

- **Python 3.12+** (managed via `pyenv`)
- **Node.js** (latest LTS version recommended)
- **MongoDB** connection (local or remote)

## Development Setup

### Backend Setup

The `backend` folder contains the Flask Python application. It's recommended to use Python 3.12+ with a virtual environment.

#### Step-by-Step (macOS/Linux)

> **Note**: If you're on Linux or Windows, the steps are similar except for how you install `pyenv`. Use your system's package manager (e.g., `apt` for Ubuntu, `chocolatey` for Windows).

1. **Install and configure Python 3.12.8 using `pyenv`**:

   ```bash
   brew install pyenv
   pyenv install 3.12.8
   cd backend
   pyenv local 3.12.8
   ```

   This creates a `.python-version` file in the backend directory, ensuring `pyenv` automatically uses Python 3.12.8 when you're in that directory.

2. **Create virtual environment and install dependencies**:

   ```bash
   python -m venv .venv
   source .venv/bin/activate
   pip install -e .
   ```

3. **Configure environment variables**:

   ```bash
   cp .env.example .env
   # Edit .env with your MongoDB connection string and other settings
   ```

4. **Initialize the database** (optional, for first-time setup):
   ```bash
   flask --app app:create_app init-db
   ```

### Frontend Setup

The frontend uses React + Vite for fast development and hot module replacement (HMR).

1. **Verify Node.js installation**:

   ```bash
   node --version
   ```

   If not installed, download from [nodejs.org](https://nodejs.org/) or use your system's package manager.

2. **Navigate to frontend directory and install dependencies**:
   ```bash
   cd frontend/frontend
   npm install
   ```

## Running the Application

### Start Individual Components

#### Backend Only

```bash
./scripts/start_backend.sh
```

This script will:

- Check for the virtual environment
- Activate it automatically
- Start the Flask development server
- Output the process PID for tracking

The backend will be available at `http://localhost:5001`

#### Frontend Only

```bash
./scripts/start_frontend.sh
```

This script will:

- Check if dependencies are installed (runs `npm install` if needed)
- Start the Vite development server

The frontend will be available at `http://localhost:5173`

### Start Full Application

To run both backend and frontend simultaneously:

```bash
./scripts/start_app.sh
```

This orchestrator script will:

- Start the backend server in the background
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

- `GET /api/health` - Health check endpoint
- `GET /api/users` - List all users
- `POST /api/users` - Create a new user

## Troubleshooting

### Port Already in Use

If you see "Address already in use" for port 5001:

- On macOS, disable 'AirPlay Receiver' in System Settings
- Or kill the process using the port: `lsof -ti:5001 | xargs kill -9`

### Vite Command Not Found

If you see "vite: command not found":

- Run `npm install` in the `frontend/` directory
- The `start_frontend.sh` script handles this automatically

## Alternative way to start the application

In one termianl run from dir root:

```bash
docker compose up -d --build --remove-orphans
```

In a second terminal, run:

```bash
cd frontend
npm run dev
```
