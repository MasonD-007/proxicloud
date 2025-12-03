#!/bin/bash
# ProxiCloud Multi-Architecture Build Script
# Builds backend and frontend for multiple platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_DIR="build"
BACKEND_DIR="backend"
FRONTEND_DIR="frontend"

# Supported platforms
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

echo -e "${GREEN}ProxiCloud Build Script${NC}"
echo -e "Version: ${YELLOW}${VERSION}${NC}"
echo ""

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

# Clean previous builds
clean_builds() {
    print_info "Cleaning previous builds..."
    rm -rf "${BUILD_DIR}"
    mkdir -p "${BUILD_DIR}"
    print_status "Build directory cleaned"
}

# Build backend for all platforms
build_backend() {
    print_info "Building backend binaries..."
    
    cd "${BACKEND_DIR}"
    
    for platform in "${PLATFORMS[@]}"; do
        IFS="/" read -r os arch <<< "${platform}"
        output="../${BUILD_DIR}/proxicloud-api-${os}-${arch}"
        
        if [ "$os" = "windows" ]; then
            output="${output}.exe"
        fi
        
        print_info "Building for ${os}/${arch}..."
        
        GOOS=$os GOARCH=$arch CGO_ENABLED=1 go build \
            -ldflags "-X main.Version=${VERSION} -w -s" \
            -o "${output}" \
            cmd/api/main.go
        
        if [ $? -eq 0 ]; then
            size=$(ls -lh "${output}" | awk '{print $5}')
            print_status "Built ${os}/${arch} (${size})"
        else
            print_error "Failed to build ${os}/${arch}"
            exit 1
        fi
    done
    
    cd ..
    print_status "Backend builds complete"
}

# Build frontend
build_frontend() {
    print_info "Building frontend..."
    
    cd "${FRONTEND_DIR}"
    
    # Install dependencies if needed
    if [ ! -d "node_modules" ]; then
        print_info "Installing dependencies..."
        npm install
    fi
    
    # Build frontend
    print_info "Running Next.js build..."
    npm run build
    
    if [ $? -eq 0 ]; then
        print_status "Frontend build complete"
        
        # Copy standalone output to build directory
        print_info "Packaging frontend..."
        mkdir -p "../${BUILD_DIR}/frontend"
        cp -r .next/standalone/* "../${BUILD_DIR}/frontend/"
        cp -r .next/static "../${BUILD_DIR}/frontend/.next/"
        cp -r public "../${BUILD_DIR}/frontend/" 2>/dev/null || true
        
        print_status "Frontend packaged"
    else
        print_error "Frontend build failed"
        exit 1
    fi
    
    cd ..
}

# Create release archives
create_archives() {
    print_info "Creating release archives..."
    
    cd "${BUILD_DIR}"
    
    for platform in "${PLATFORMS[@]}"; do
        IFS="/" read -r os arch <<< "${platform}"
        binary="proxicloud-api-${os}-${arch}"
        
        if [ ! -f "${binary}" ]; then
            continue
        fi
        
        archive_name="proxicloud-${VERSION}-${os}-${arch}.tar.gz"
        
        print_info "Creating ${archive_name}..."
        
        # Create temporary directory for archive contents
        tmp_dir="proxicloud-${os}-${arch}"
        mkdir -p "${tmp_dir}"
        
        # Copy binary
        cp "${binary}" "${tmp_dir}/proxicloud-api"
        chmod +x "${tmp_dir}/proxicloud-api"
        
        # Copy frontend
        cp -r frontend "${tmp_dir}/"
        
        # Copy config example
        cp ../deploy/config/config.example.yaml "${tmp_dir}/config.yaml"
        
        # Copy systemd services
        mkdir -p "${tmp_dir}/systemd"
        cp ../deploy/systemd/*.service "${tmp_dir}/systemd/"
        
        # Create README
        cat > "${tmp_dir}/README.txt" << EOF
ProxiCloud ${VERSION} - ${os}/${arch}

Quick Start:
1. Edit config.yaml with your Proxmox details
2. Run backend: ./proxicloud-api
3. Run frontend: cd frontend && node server.js
4. Access: http://localhost:3000

For full documentation, visit:
https://github.com/MasonD-007/proxicloud

Installation:
See deploy/install.sh for automated installation
EOF
        
        # Create archive
        tar czf "${archive_name}" "${tmp_dir}"
        rm -rf "${tmp_dir}"
        
        size=$(ls -lh "${archive_name}" | awk '{print $5}')
        print_status "Created ${archive_name} (${size})"
    done
    
    cd ..
}

# Generate checksums
generate_checksums() {
    print_info "Generating checksums..."
    
    cd "${BUILD_DIR}"
    
    if command -v shasum &> /dev/null; then
        shasum -a 256 *.tar.gz > SHA256SUMS
        print_status "SHA256 checksums generated"
    else
        print_error "shasum not found, skipping checksums"
    fi
    
    cd ..
}

# Display build summary
show_summary() {
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo -e "${GREEN}Build Complete!${NC}"
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo ""
    echo "Version: ${VERSION}"
    echo "Output directory: ${BUILD_DIR}/"
    echo ""
    echo "Built artifacts:"
    ls -lh "${BUILD_DIR}"/*.tar.gz 2>/dev/null || echo "No archives found"
    echo ""
    echo "To upload to GitHub Releases:"
    echo "  gh release create v${VERSION} ${BUILD_DIR}/*.tar.gz ${BUILD_DIR}/SHA256SUMS"
    echo ""
}

# Main execution
main() {
    # Check prerequisites
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        exit 1
    fi
    
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed"
        exit 1
    fi
    
    # Check if we're in the project root
    if [ ! -d "${BACKEND_DIR}" ] || [ ! -d "${FRONTEND_DIR}" ]; then
        print_error "Must be run from project root"
        exit 1
    fi
    
    # Run build steps
    clean_builds
    build_backend
    build_frontend
    create_archives
    generate_checksums
    show_summary
}

# Run main function
main
