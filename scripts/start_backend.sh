#!/bin/bash
set -e

# Config
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BACKEND_DIR="./backend"

# Ensure Go is installed
if ! command -v go >/dev/null 2>&1; then
  echo -e "${RED}Error: Go is not installed or not on PATH${NC}"
  exit 1
fi

# Navigate to backend directory
if ! pushd "$BACKEND_DIR" > /dev/null 2>&1; then
  echo -e "${RED}Error: Could not navigate to $BACKEND_DIR${NC}"
  exit 1
fi

echo -e "${YELLOW}Resolving Go modules...${NC}"
go mod tidy

echo -e "${GREEN}Starting Go backend server...${NC}"
go run . &
GO_PID=$!

sleep 2

if ! kill -0 $GO_PID 2>/dev/null; then
  echo -e "${RED}Error: Failed to start backend server${NC}"
  popd > /dev/null
  exit 1
fi

echo "$GO_PID"
popd > /dev/null