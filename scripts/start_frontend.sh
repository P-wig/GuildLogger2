#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

FRONTEND_DIR="./frontend"

# Navigate to frontend directory
if ! pushd "$FRONTEND_DIR" > /dev/null 2>&1; then
  echo -e "${RED}❌ Error: Could not navigate to $FRONTEND_DIR${NC}"
  exit 1
fi

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
  echo -e "${YELLOW}⚙️  Installing dependencies...${NC}"
  if ! npm install; then
    echo -e "${RED}❌ Error: npm install failed${NC}"
    popd > /dev/null
    exit 1
  fi
else
  echo -e "${GREEN}✓ Dependencies already installed${NC}"
fi

# Start Vite dev server
echo -e "${GREEN}🚀 Starting Vite Dev server...${NC}"
if ! npm run dev; then
  echo -e "${RED}❌ Error: Failed to start Vite dev server${NC}"
  popd > /dev/null
  exit 1
fi

popd > /dev/null
