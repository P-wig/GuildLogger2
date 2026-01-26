#!/bin/bash
set -e 

# Config
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BACKEND_DIR="./backend"

# Navigate to backend directory
if ! pushd "$BACKEND_DIR" > /dev/null 2>&1; then
  echo -e "${RED}❌ Error: Could not navigate to $BACKEND_DIR${NC}"
  exit 1
fi

# Check for python virtual env
echo -e "${YELLOW}⚙️  Checking for virtual environment...${NC}"
if [ ! -d .venv ]; then 
    echo -e "${RED}❌ Error: Virtual environment not found!${NC}"
    echo -e "${YELLOW}Read the docs to learn about a proper backend setup.${NC}"
    popd > /dev/null
    exit 1
else
    echo -e "${GREEN}✓ Virtual environment found${NC}"
    # Activate the env
    echo -e "${YELLOW}⚙️  Activating virtual environment...${NC}"
    source .venv/bin/activate
    echo -e "${YELLOW}⚠️  This script will not install packages into the virtual environment.${NC}"
    # Start the app
    echo -e "${GREEN}🚀 Starting backend server...${NC}"
    python run.py &
    PYTHON_PID=$!
    
    # Wait a moment to check if it started successfully
    sleep 2
    
    # Check if process is still running
    if ! kill -0 $PYTHON_PID 2>/dev/null; then
      echo -e "${RED}❌ Error: Failed to start backend server${NC}"
      popd > /dev/null
      exit 1
    fi
    
    # Output the PID to stdout
    echo "$PYTHON_PID"
fi

popd > /dev/null

