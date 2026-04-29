#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo -e "${BLUE}═══════════════════════════════════════${NC}"
echo -e "${BLUE}  Starting Full Stack Application${NC}"
echo -e "${BLUE}═══════════════════════════════════════${NC}"
echo ""

# Function to cleanup background processes on exit
cleanup() {
    echo -e "\n${YELLOW}Shutting down services...${NC}"
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null
        echo -e "${GREEN}Frontend stopped${NC}"
    fi
    docker compose down --remove-orphans > /tmp/backend.log 2>&1 || true
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

if ! docker compose up -d --build --force-recreate mongo backend > /tmp/backend.log 2>&1; then
    echo -e "${RED}Backend failed to start. Reading /tmp/backend.log...${NC}"
    cat /tmp/backend.log
    exit 1
fi

# Wait for backend health endpoint (max 30s)
HEALTH_URL="http://localhost:5001/api/health"
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
    cat /tmp/backend.log
    exit 1
fi

echo -e "${GREEN}Backend is healthy${NC}"
echo ""

# Start frontend server
echo -e "${GREEN}🎨 Frontend ...${NC}"
"$SCRIPT_DIR/start_frontend.sh" > /tmp/frontend.log 2>&1 &
FRONTEND_PID=$!

# Wait a moment for frontend to initialize
sleep 2

# Check if frontend is still running
if ! kill -0 $FRONTEND_PID 2>/dev/null; then
    echo -e "${RED}❌ Frontend failed to start. Reading /tmp/frontend.log for details${NC}"
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
    echo -e "${BLUE}  FRONTEND LOG${NC}"
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
    cat /tmp/frontend.log
    cleanup
fi

echo -e "${GREEN}✓ Frontend started (PID: $FRONTEND_PID)${NC}"
echo ""

echo -e "${BLUE}═══════════════════════════════════════${NC}"
echo -e "${GREEN}✨ Application is running!${NC}"
echo -e "${BLUE}═══════════════════════════════════════${NC}"
echo ""
echo -e "${CYAN}🌐 Application URLs:${NC}"
echo -e "${YELLOW}Frontend:${NC} http://localhost:5173"
echo -e "${YELLOW}Backend:${NC}  http://localhost:5001"
echo -e "${YELLOW}API Health:${NC} http://localhost:5001/api/health"
echo ""
echo -e "${CYAN}📋 Logs:${NC}"
echo -e "${YELLOW}Backend logs:${NC}  tail -f /tmp/backend.log"
echo -e "${YELLOW}Frontend logs:${NC} tail -f /tmp/frontend.log"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
echo ""

# Wait for both processes
wait