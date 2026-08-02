#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
 
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_FILE="$REPO_ROOT/docker-compose.yml"
ENV_FILE="$REPO_ROOT/backend/.env"
BACKEND_LOG="/tmp/backend.log"
FRONTEND_LOG="/tmp/frontend.log"
 
 echo -e "${BLUE}═══════════════════════════════════════${NC}"
 echo -e "${BLUE}  Starting Full Stack Application${NC}"
 echo -e "${BLUE}═══════════════════════════════════════${NC}"
 echo ""

CLOUDFLARE_LOG="/tmp/cloudflare.log"
CLOUDFLARE_PID=""
TUNNEL_URL=""

# Function to cleanup background processes on exit
cleanup() {
    echo -e "\n${YELLOW}Shutting down services...${NC}"
    if [ ! -z "$CLOUDFLARE_PID" ]; then
        kill $CLOUDFLARE_PID 2>/dev/null
        echo -e "${GREEN}Cloudflare tunnel stopped${NC}"
    fi
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null
        echo -e "${GREEN}Frontend stopped${NC}"
    fi
    docker compose --project-directory "$REPO_ROOT" --env-file "$ENV_FILE" -f "$COMPOSE_FILE" down --remove-orphans > "$BACKEND_LOG" 2>&1 || true
    echo -e "${GREEN}Backend services stopped${NC}"
    exit 0
 }
 
trap cleanup SIGINT SIGTERM

# Start backend server
echo -e "${GREEN}Backend...${NC}"

if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Docker Engine is not running. Start Docker Desktop and retry.${NC}"
    exit 1
 fi
 
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}Missing $ENV_FILE${NC}"
    echo -e "${YELLOW}Create it with: DISCORD_CLIENT_ID, DISCORD_CLIENT_SECRET, DISCORD_REDIRECT_URI, SECRET_KEY${NC}"
    exit 1
 fi
 
# Load repo-root .env into this shell so checks below and child processes are consistent.
set -a
. "$ENV_FILE"
set +a
 
MISSING=()
[ -z "$DISCORD_CLIENT_ID" ] && MISSING+=("DISCORD_CLIENT_ID")
[ -z "$DISCORD_CLIENT_SECRET" ] && MISSING+=("DISCORD_CLIENT_SECRET")
[ -z "$DISCORD_REDIRECT_URI" ] && MISSING+=("DISCORD_REDIRECT_URI")
[ -z "$SECRET_KEY" ] && MISSING+=("SECRET_KEY")
if [ ${#MISSING[@]} -ne 0 ]; then
    echo -e "${RED}Missing required OAuth env vars in $ENV_FILE:${NC} ${MISSING[*]}"
    exit 1
 fi
 
if ! docker compose --project-directory "$REPO_ROOT" --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d --build --force-recreate mongo backend > "$BACKEND_LOG" 2>&1; then
    echo -e "${RED}Backend failed to start. Reading $BACKEND_LOG...${NC}"
    cat "$BACKEND_LOG"
    echo -e "${BLUE}Container logs:${NC}"
    docker compose --project-directory "$REPO_ROOT" --env-file "$ENV_FILE" -f "$COMPOSE_FILE" logs backend --tail=200 || true
    exit 1
 fi
 
 # Wait for backend health endpoint (max 30s)
 HEALTH_URL="http://127.0.0.1:5001/api/health"
 READY=0
 for i in $(seq 1 30); do
    if curl -fsS "$HEALTH_URL" > /dev/null 2>&1; then
        READY=1
        break
    fi
    sleep 1
 done

 if [ "$READY" -ne 1 ]; then
    echo -e "${RED}Backend did not become healthy at ${HEALTH_URL}${NC}"
    echo -e "${BLUE}BACKEND LOG${NC}"
    cat "$BACKEND_LOG"
    docker compose --project-directory "$REPO_ROOT" --env-file "$ENV_FILE" -f "$COMPOSE_FILE" logs backend --tail=200 || true
    exit 1
 fi
 
 echo -e "${GREEN}Backend is healthy${NC}"
 echo ""

 # Start Cloudflare tunnel for Discord slash command interactions.
 # The tunnel URL changes each session — it is printed in the URLs block below.
 echo -e "${GREEN}⛅ Cloudflare tunnel...${NC}"
 CLOUDFLARED_BIN=""
 if command -v cloudflared >/dev/null 2>&1; then
     CLOUDFLARED_BIN="cloudflared"
 elif [ -f "/c/Program Files (x86)/cloudflared/cloudflared.exe" ]; then
     CLOUDFLARED_BIN="/c/Program Files (x86)/cloudflared/cloudflared.exe"
 fi

 if [ -n "$CLOUDFLARED_BIN" ]; then
     "$CLOUDFLARED_BIN" tunnel --url http://localhost:5001 > "$CLOUDFLARE_LOG" 2>&1 &
     CLOUDFLARE_PID=$!
     for i in $(seq 1 30); do
         TUNNEL_URL=$(grep -o 'https://[^[:space:]]*.trycloudflare.com' "$CLOUDFLARE_LOG" 2>/dev/null | head -1)
         [ -n "$TUNNEL_URL" ] && break
         sleep 1
     done
     if [ -z "$TUNNEL_URL" ]; then
         echo -e "${YELLOW}⚠ Tunnel did not produce a URL within 30s — slash commands unavailable.${NC}"
         echo -e "${YELLOW}  Check /tmp/cloudflare.log for details.${NC}"
     else
         echo -e "${GREEN}✓ Cloudflare tunnel active${NC}"
     fi
 else
     echo -e "${YELLOW}⚠ cloudflared not found — skipping tunnel. Slash commands will be unavailable.${NC}"
     echo -e "${YELLOW}  Install: winget install Cloudflare.cloudflared${NC}"
 fi
 echo ""

 # Start frontend server
 echo -e "${GREEN}🎨 Frontend ...${NC}"
 "$SCRIPT_DIR/start_frontend.sh" > "$FRONTEND_LOG" 2>&1 &
 FRONTEND_PID=$!
 
 # Wait a moment for frontend to initialize
 sleep 2

 # Check if frontend is still running
 if ! kill -0 $FRONTEND_PID 2>/dev/null; then
    echo -e "${RED}❌ Frontend failed to start. Reading /tmp/frontend.log for details${NC}"
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
    echo -e "${BLUE}  FRONTEND LOG${NC}"
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
    cat "$FRONTEND_LOG"
    cleanup
 fi

FRONTEND_URL="http://127.0.0.1:5173"
FRONTEND_READY=0
for i in $(seq 1 30); do
  if curl -fsS "$FRONTEND_URL" > /dev/null 2>&1; then
    FRONTEND_READY=1
    break
  fi
  if ! kill -0 $FRONTEND_PID 2>/dev/null; then
    break
  fi
  sleep 1
done

if [ "$FRONTEND_READY" -ne 1 ]; then
  echo -e "${RED}Frontend did not become healthy at ${FRONTEND_URL}${NC}"
  echo -e "${BLUE}FRONTEND LOG${NC}"
  cat "$FRONTEND_LOG"
  cleanup
fi

 echo -e "${GREEN}✓ Frontend started (PID: $FRONTEND_PID)${NC}"
 echo ""

 echo -e "${BLUE}═══════════════════════════════════════${NC}"
 echo -e "${GREEN}✨ Application is running!${NC}"
 echo -e "${BLUE}═══════════════════════════════════════${NC}"
 echo ""
 echo -e "${CYAN}🌐 Application URLs:${NC}"
 echo -e "${YELLOW}Frontend:${NC}   http://127.0.0.1:5173"
 echo -e "${YELLOW}Backend:${NC}    http://127.0.0.1:5001"
 echo -e "${YELLOW}API Health:${NC} http://127.0.0.1:5001/api/health"
 if [ -n "$TUNNEL_URL" ]; then
     echo -e "${YELLOW}Interactions:${NC} ${TUNNEL_URL}/api/interactions"
     echo -e "${CYAN}  ↑ Paste into Discord Developer Portal → General Information → Interactions Endpoint URL${NC}"
 fi
 echo ""
 echo -e "${CYAN}📋 Logs:${NC}"
 echo -e "${YELLOW}Backend logs:${NC}  tail -f /tmp/backend.log"
 echo -e "${YELLOW}Frontend logs:${NC} tail -f /tmp/frontend.log"
 echo ""
 echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
 echo ""

 # Wait for both processes
 wait