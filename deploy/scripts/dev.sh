#!/bin/bash
# ProxiCloud Development Script
# Run backend and frontend from source for debugging (no binaries)
# This script is designed to run on your Proxmox node for development/testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "${SCRIPT_DIR}/../.." && pwd )"
BACKEND_DIR="${PROJECT_ROOT}/backend"
FRONTEND_DIR="${PROJECT_ROOT}/frontend"
CONFIG_FILE="${CONFIG_FILE:-${PROJECT_ROOT}/config.test.yaml}"
CACHE_PATH="${CACHE_PATH:-/tmp/proxicloud-dev/cache.db}"
ANALYTICS_PATH="${ANALYTICS_PATH:-/tmp/proxicloud-dev/analytics.db}"
BACKEND_PORT="${BACKEND_PORT:-8080}"
FRONTEND_PORT="${FRONTEND_PORT:-3000}"

# PIDs for cleanup
BACKEND_PID=""
FRONTEND_PID=""

# Function to print status
print_status() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_info() {
    echo -e "${YELLOW}[i]${NC} $1"
}

print_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# Cleanup function
cleanup() {
    echo ""
    print_info "Shutting down services..."
    
    if [ -n "$BACKEND_PID" ]; then
        print_info "Stopping backend (PID: $BACKEND_PID)..."
        kill $BACKEND_PID 2>/dev/null || true
        wait $BACKEND_PID 2>/dev/null || true
    fi
    
    if [ -n "$FRONTEND_PID" ]; then
        print_info "Stopping frontend (PID: $FRONTEND_PID)..."
        kill $FRONTEND_PID 2>/dev/null || true
        wait $FRONTEND_PID 2>/dev/null || true
    fi
    
    print_status "Cleanup complete"
    exit 0
}

# Register cleanup on exit
trap cleanup SIGINT SIGTERM EXIT

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    local missing_deps=()
    
    # Check Go
    if ! command -v go &> /dev/null; then
        missing_deps+=("Go")
    else
        GO_VERSION=$(go version | awk '{print $3}')
        print_status "Found Go: $GO_VERSION"
    fi
    
    # Check Node.js
    if ! command -v node &> /dev/null; then
        missing_deps+=("Node.js")
    else
        NODE_VERSION=$(node --version)
        print_status "Found Node.js: $NODE_VERSION"
    fi
    
    # Check npm
    if ! command -v npm &> /dev/null; then
        missing_deps+=("npm")
    else
        NPM_VERSION=$(npm --version)
        print_status "Found npm: v$NPM_VERSION"
    fi
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        print_error "Missing dependencies: ${missing_deps[*]}"
        echo ""
        echo "Installation instructions:"
        echo "  Go:      https://golang.org/doc/install"
        echo "  Node.js: https://nodejs.org/ or use your package manager"
        exit 1
    fi
    
    # Check if we're in the project root
    if [ ! -d "${BACKEND_DIR}" ] || [ ! -d "${FRONTEND_DIR}" ]; then
        print_error "Script must be in project structure"
        print_error "Backend dir: ${BACKEND_DIR}"
        print_error "Frontend dir: ${FRONTEND_DIR}"
        exit 1
    fi
    
    print_status "All prerequisites met"
}

# Setup development environment
setup_environment() {
    print_info "Setting up development environment..."
    
    # Create temp directories for dev data
    mkdir -p "$(dirname "$CACHE_PATH")"
    mkdir -p "$(dirname "$ANALYTICS_PATH")"
    
    # Check for config file
    if [ ! -f "$CONFIG_FILE" ]; then
        print_error "Configuration file not found: $CONFIG_FILE"
        echo ""
        echo "Please create a configuration file. You can copy from example:"
        echo "  cp ${PROJECT_ROOT}/deploy/config/config.example.yaml $CONFIG_FILE"
        echo ""
        echo "Or set CONFIG_FILE environment variable:"
        echo "  export CONFIG_FILE=/path/to/your/config.yaml"
        exit 1
    fi
    
    print_status "Using config: $CONFIG_FILE"
    print_status "Cache path: $CACHE_PATH"
    print_status "Analytics path: $ANALYTICS_PATH"
}

# Install backend dependencies
setup_backend() {
    print_info "Setting up backend..."
    
    cd "${BACKEND_DIR}"
    
    # Download Go dependencies
    print_info "Downloading Go modules..."
    go mod download
    
    if [ $? -eq 0 ]; then
        print_status "Backend dependencies installed"
    else
        print_error "Failed to install backend dependencies"
        exit 1
    fi
    
    cd "${PROJECT_ROOT}"
}

# Install frontend dependencies
setup_frontend() {
    print_info "Setting up frontend..."
    
    cd "${FRONTEND_DIR}"
    
    # Install npm dependencies if needed
    if [ ! -d "node_modules" ]; then
        print_info "Installing npm packages..."
        npm install
        
        if [ $? -eq 0 ]; then
            print_status "Frontend dependencies installed"
        else
            print_error "Failed to install frontend dependencies"
            exit 1
        fi
    else
        print_status "Frontend dependencies already installed"
    fi
    
    cd "${PROJECT_ROOT}"
}

# Start backend in development mode
start_backend() {
    print_info "Starting backend in development mode..."
    
    cd "${BACKEND_DIR}"
    
    # Export environment variables
    export CONFIG_PATH="$CONFIG_FILE"
    export CACHE_PATH="$CACHE_PATH"
    export ANALYTICS_PATH="$ANALYTICS_PATH"
    
    print_debug "CONFIG_PATH=$CONFIG_PATH"
    print_debug "CACHE_PATH=$CACHE_PATH"
    print_debug "ANALYTICS_PATH=$ANALYTICS_PATH"
    
    # Run backend with go run (shows all errors and allows live reload)
    echo ""
    echo -e "${BLUE}========== BACKEND OUTPUT ==========${NC}"
    
    go run cmd/api/main.go &
    BACKEND_PID=$!
    
    # Wait a moment to check if it started successfully
    sleep 2
    
    if kill -0 $BACKEND_PID 2>/dev/null; then
        print_status "Backend started (PID: $BACKEND_PID)"
        print_info "Backend API: http://localhost:$BACKEND_PORT"
    else
        print_error "Backend failed to start"
        exit 1
    fi
    
    cd "${PROJECT_ROOT}"
}

# Start frontend in development mode
start_frontend() {
    print_info "Starting frontend in development mode..."
    
    cd "${FRONTEND_DIR}"
    
    # Set API URL
    export NEXT_PUBLIC_API_URL="http://localhost:$BACKEND_PORT"
    
    print_debug "NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL"
    
    # Run frontend in dev mode (hot reload enabled)
    echo ""
    echo -e "${BLUE}========== FRONTEND OUTPUT ==========${NC}"
    
    npm run dev -- -p $FRONTEND_PORT &
    FRONTEND_PID=$!
    
    # Wait a moment to check if it started successfully
    sleep 3
    
    if kill -0 $FRONTEND_PID 2>/dev/null; then
        print_status "Frontend started (PID: $FRONTEND_PID)"
        print_info "Frontend UI: http://localhost:$FRONTEND_PORT"
    else
        print_error "Frontend failed to start"
        exit 1
    fi
    
    cd "${PROJECT_ROOT}"
}

# Display summary
show_summary() {
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo -e "${GREEN}ProxiCloud Development Server Running${NC}"
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo ""
    echo -e "Backend:  ${BLUE}http://localhost:$BACKEND_PORT${NC} (PID: $BACKEND_PID)"
    echo -e "Frontend: ${BLUE}http://localhost:$FRONTEND_PORT${NC} (PID: $FRONTEND_PID)"
    echo ""
    echo "Features:"
    echo "  • Live reload on file changes"
    echo "  • Detailed error messages and stack traces"
    echo "  • Source maps for debugging"
    echo "  • No compilation required"
    echo ""
    echo "Configuration:"
    echo "  • Config:    $CONFIG_FILE"
    echo "  • Cache:     $CACHE_PATH"
    echo "  • Analytics: $ANALYTICS_PATH"
    echo ""
    echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
    echo ""
}

# Monitor processes
monitor() {
    while true; do
        # Check if backend is still running
        if [ -n "$BACKEND_PID" ] && ! kill -0 $BACKEND_PID 2>/dev/null; then
            print_error "Backend process died unexpectedly"
            cleanup
        fi
        
        # Check if frontend is still running
        if [ -n "$FRONTEND_PID" ] && ! kill -0 $FRONTEND_PID 2>/dev/null; then
            print_error "Frontend process died unexpectedly"
            cleanup
        fi
        
        sleep 5
    done
}

# Main execution
main() {
    echo -e "${GREEN}ProxiCloud Development Server${NC}"
    echo ""
    
    # Check environment
    check_prerequisites
    setup_environment
    
    # Setup dependencies
    setup_backend
    setup_frontend
    
    # Start services
    start_backend
    start_frontend
    
    # Show summary
    show_summary
    
    # Monitor processes
    monitor
}

# Parse command line arguments
case "${1:-}" in
    --help|-h)
        echo "ProxiCloud Development Script"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h    Show this help message"
        echo ""
        echo "Environment Variables:"
        echo "  CONFIG_FILE      Path to config file (default: ./config.test.yaml)"
        echo "  CACHE_PATH       Path to cache database (default: /tmp/proxicloud-dev/cache.db)"
        echo "  ANALYTICS_PATH   Path to analytics database (default: /tmp/proxicloud-dev/analytics.db)"
        echo "  BACKEND_PORT     Backend port (default: 8080)"
        echo "  FRONTEND_PORT    Frontend port (default: 3000)"
        echo ""
        echo "Example:"
        echo "  # Basic usage"
        echo "  $0"
        echo ""
        echo "  # Custom config and ports"
        echo "  CONFIG_FILE=/etc/proxicloud/config.yaml FRONTEND_PORT=4000 $0"
        echo ""
        exit 0
        ;;
    "")
        main
        ;;
    *)
        print_error "Unknown option: $1"
        echo "Use --help for usage information"
        exit 1
        ;;
esac
