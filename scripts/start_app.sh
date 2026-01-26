#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo -e "${BLUE}═══════════════════════════════════════${NC}"
echo -e "${BLUE}  Starting Full Stack Application${NC}"
echo -e "${BLUE}═══════════════════════════════════════${NC}"
echo ""

# Function to cleanup background processes on exit
cleanup() {
    echo -e "\n${YELLOW}🛑 Shutting down services...${NC}"
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null
        echo -e "${GREEN}✓ Backend stopped${NC}"
    fi
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null
        echo -e "${GREEN}✓ Frontend stopped${NC}"
    fi
    exit 0
}

trap cleanup SIGINT SIGTERM

# Start backend server
echo -e "${GREEN}🔧 Backend...${NC}"
"$SCRIPT_DIR/start_backend.sh" > /tmp/backend.log 2>&1 &
BACKEND_PID=$!

# Wait a moment for backend to initialize
sleep 2

# Check if backend is still running
if ! kill -0 $BACKEND_PID 2>/dev/null; then
    echo -e "${RED}❌ Backend failed to start. Reading /tmp/backend.log for details...${NC}"
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
    echo -e "${BLUE}  BACKEND LOG${NC}"
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
    cat /tmp/backend.log
    exit 1
fi

echo -e "${GREEN}✓ Backend started (PID: $BACKEND_PID)${NC}"
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
echo -e "${YELLOW}Backend logs:${NC}  tail -f /tmp/backend.log"
echo -e "${YELLOW}Frontend logs:${NC} tail -f /tmp/frontend.log"
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
echo ""

# Wait for both processes
wait
