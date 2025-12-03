#!/bin/bash

# ProxiCloud Quick Start Script
# This script helps you quickly test ProxiCloud locally

set -e

echo "╔════════════════════════════════════════════════╗"
echo "║         ProxiCloud Quick Start                 ║"
echo "╚════════════════════════════════════════════════╝"
echo ""

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BACKEND_DIR="$SCRIPT_DIR/backend"
FRONTEND_DIR="$SCRIPT_DIR/frontend"
CONFIG_FILE="$SCRIPT_DIR/config.test.yaml"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check prerequisites
echo "📋 Checking prerequisites..."

if ! command -v go &> /dev/null; then
    echo -e "${RED}✗${NC} Go is not installed. Please install Go 1.21+"
    exit 1
fi
echo -e "${GREEN}✓${NC} Go found: $(go version)"

if ! command -v node &> /dev/null; then
    echo -e "${RED}✗${NC} Node.js is not installed. Please install Node.js 18+"
    exit 1
fi
echo -e "${GREEN}✓${NC} Node.js found: $(node --version)"

if ! command -v npm &> /dev/null; then
    echo -e "${RED}✗${NC} npm is not installed"
    exit 1
fi
echo -e "${GREEN}✓${NC} npm found: $(npm --version)"

echo ""

# Check config file
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}✗${NC} Config file not found: $CONFIG_FILE"
    exit 1
fi
echo -e "${GREEN}✓${NC} Config file found"

# Check if config is updated
if grep -q "your-secret-here" "$CONFIG_FILE"; then
    echo -e "${YELLOW}⚠${NC}  Warning: Config file appears to have default values"
    echo "   Please update $CONFIG_FILE with your Proxmox credentials"
    echo ""
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo ""

# Build backend
echo "🔨 Building backend..."
cd "$BACKEND_DIR"
if [ ! -f "proxicloud-api" ]; then
    go build -o proxicloud-api ./cmd/api
    echo -e "${GREEN}✓${NC} Backend built successfully"
else
    echo -e "${YELLOW}→${NC} Backend binary already exists (skipping build)"
fi

echo ""

# Check frontend dependencies
echo "📦 Checking frontend dependencies..."
cd "$FRONTEND_DIR"
if [ ! -d "node_modules" ]; then
    echo "   Installing dependencies..."
    npm install --silent
    echo -e "${GREEN}✓${NC} Dependencies installed"
else
    echo -e "${YELLOW}→${NC} Dependencies already installed"
fi

echo ""

# Create cache directory
CACHE_DIR="/tmp/proxicloud"
mkdir -p "$CACHE_DIR"
echo -e "${GREEN}✓${NC} Cache directory ready: $CACHE_DIR"

echo ""
echo "╔════════════════════════════════════════════════╗"
echo "║            Starting ProxiCloud                 ║"
echo "╚════════════════════════════════════════════════╝"
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Shutting down ProxiCloud..."
    kill $BACKEND_PID 2>/dev/null || true
    kill $FRONTEND_PID 2>/dev/null || true
    exit 0
}

trap cleanup SIGINT SIGTERM

# Start backend
echo "🚀 Starting backend..."
cd "$BACKEND_DIR"
CONFIG_PATH="$CONFIG_FILE" CACHE_PATH="$CACHE_DIR/cache.db" ./proxicloud-api > /tmp/proxicloud-backend.log 2>&1 &
BACKEND_PID=$!

# Wait for backend to start
sleep 2

# Check if backend is running
if ! kill -0 $BACKEND_PID 2>/dev/null; then
    echo -e "${RED}✗${NC} Backend failed to start. Check logs:"
    tail -20 /tmp/proxicloud-backend.log
    exit 1
fi

# Test backend health
if curl -s http://localhost:8080/api/health > /dev/null; then
    echo -e "${GREEN}✓${NC} Backend is running on http://localhost:8080"
else
    echo -e "${RED}✗${NC} Backend health check failed"
    tail -20 /tmp/proxicloud-backend.log
    kill $BACKEND_PID
    exit 1
fi

echo ""

# Start frontend
echo "🚀 Starting frontend..."
cd "$FRONTEND_DIR"
NEXT_PUBLIC_API_URL=http://localhost:8080/api npm run dev > /tmp/proxicloud-frontend.log 2>&1 &
FRONTEND_PID=$!

# Wait for frontend to start
echo "   Waiting for frontend to start..."
for i in {1..30}; do
    if curl -s http://localhost:3000 > /dev/null 2>&1; then
        break
    fi
    sleep 1
    if [ $i -eq 30 ]; then
        echo -e "${RED}✗${NC} Frontend failed to start. Check logs:"
        tail -20 /tmp/proxicloud-frontend.log
        kill $BACKEND_PID
        kill $FRONTEND_PID
        exit 1
    fi
done

echo -e "${GREEN}✓${NC} Frontend is running on http://localhost:3000"

echo ""
echo "╔════════════════════════════════════════════════╗"
echo "║         ProxiCloud is Ready! 🎉                ║"
echo "╚════════════════════════════════════════════════╝"
echo ""
echo "📍 Access Points:"
echo "   Frontend:  http://localhost:3000"
echo "   Backend:   http://localhost:8080/api"
echo ""
echo "📊 Features Available:"
echo "   ✓ Dashboard with container statistics"
echo "   ✓ Container list and management"
echo "   ✓ Create new containers"
echo "   ✓ Container detail view"
echo "   ✓ Start/Stop/Reboot/Delete containers"
echo "   ✓ Offline mode with caching"
echo ""
echo "📝 Logs:"
echo "   Backend:   tail -f /tmp/proxicloud-backend.log"
echo "   Frontend:  tail -f /tmp/proxicloud-frontend.log"
echo "   Cache DB:  sqlite3 $CACHE_DIR/cache.db"
echo ""
echo "🧪 Test Offline Mode:"
echo "   1. Load a page (e.g., Dashboard)"
echo "   2. Stop backend: kill $BACKEND_PID"
echo "   3. Refresh page - offline banner should appear"
echo "   4. Restart: cd backend && CONFIG_PATH=$CONFIG_FILE ./proxicloud-api"
echo ""
echo "Press Ctrl+C to stop all services"
echo ""

# Keep script running
wait $BACKEND_PID $FRONTEND_PID
