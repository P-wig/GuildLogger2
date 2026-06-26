#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
FRONTEND_DIR="$REPO_ROOT/frontend"

if ! pushd "$FRONTEND_DIR" > /dev/null 2>&1; then
  echo -e "${RED}Error: Could not navigate to $FRONTEND_DIR${NC}"
  exit 1
fi

if [ ! -d "node_modules" ] || [ "package-lock.json" -nt "node_modules" ]; then
  echo -e "${YELLOW}Installing dependencies...${NC}"
  npm install
else
  echo -e "${GREEN}Dependencies already installed${NC}"
fi

echo -e "${GREEN}Starting Vite Dev server...${NC}"
npm run dev -- --host 127.0.0.1 --port 5173 --strictPort
