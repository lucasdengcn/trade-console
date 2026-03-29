#!/bin/bash

# Trade Manager Build Script
# This script automates the build process for the trade-manager project

set -e  # Exit on any error

# Configuration
PROJECT_NAME="trade-manager"
BINARY_NAME="trade-cli"
VERSION="1.0.0"
BUILD_DIR="dist"
LOG_FILE="build.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    local level=$1
    local message=$2
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "${timestamp} [${level}] ${message}" | tee -a "${LOG_FILE}"
}

# Print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    print_info "Found Go version: ${go_version}"
    
    # Check if Go version is at least 1.21
    if [[ "$(printf '%s\n' "1.21" "${go_version}" | sort -V | head -n1)" != "1.21" ]]; then
        print_error "Go version 1.21 or higher is required"
        exit 1
    fi
}

# Clean build directory
clean() {
    print_info "Cleaning build directory..."
    rm -rf "${BUILD_DIR}"
    rm -f "${BINARY_NAME}" "${BINARY_NAME}.exe"
    rm -f "${LOG_FILE}"
    print_success "Clean completed"
}

# Download dependencies
download_deps() {
    print_info "Downloading dependencies..."
    if go mod download; then
        print_success "Dependencies downloaded successfully"
    else
        print_error "Failed to download dependencies"
        exit 1
    fi
}

# Verify dependencies
verify_deps() {
    print_info "Verifying dependencies..."
    if go mod verify; then
        print_success "Dependencies verified successfully"
    else
        print_error "Dependency verification failed"
        exit 1
    fi
}

# Run tests
run_tests() {
    print_info "Running tests..."
    if go test -v ./... 2>&1 | tee -a "${LOG_FILE}"; then
        print_success "All tests passed"
    else
        print_error "Tests failed"
        exit 1
    fi
}

# Build for current platform
build_current() {
    local platform=$1
    local optimization=$2
    
    print_info "Building for ${platform}..."
    
    local ldflags="-X main.version=${VERSION}"
    if [[ "${optimization}" == "true" ]]; then
        ldflags="-s -w ${ldflags}"
    fi
    
    local output="${BINARY_NAME}"
    if [[ "${platform}" == "windows" ]]; then
        output="${BINARY_NAME}.exe"
    fi
    
    if go build -ldflags="${ldflags}" -o "${output}" cmd/cli/main.go; then
        print_success "Build completed: ${output}"
        
        # Show binary information
        if command -v file &> /dev/null; then
            file "${output}"
        fi
        
        if command -v du &> /dev/null; then
            local size=$(du -h "${output}" | cut -f1)
            print_info "Binary size: ${size}"
        fi
    else
        print_error "Build failed"
        exit 1
    fi
}

# Cross-compile for multiple platforms
cross_compile() {
    print_info "Starting cross-compilation..."
    
    # Create build directory
    mkdir -p "${BUILD_DIR}"
    
    # Platforms to build for
    local platforms=(
        "linux/amd64"
        "linux/arm64"
        "windows/amd64"
        "darwin/amd64"
        "darwin/arm64"
    )
    
    for platform in "${platforms[@]}"; do
        local os=$(echo "${platform}" | cut -d'/' -f1)
        local arch=$(echo "${platform}" | cut -d'/' -f2)
        
        print_info "Building for ${os}/${arch}..."
        
        local output="${BUILD_DIR}/${BINARY_NAME}-${os}-${arch}"
        if [[ "${os}" == "windows" ]]; then
            output="${output}.exe"
        fi
        
        if GOOS="${os}" GOARCH="${arch}" go build -ldflags="-s -w -X main.version=${VERSION}" -o "${output}" cmd/cli/main.go; then
            print_success "Built: ${output}"
        else
            print_error "Failed to build for ${os}/${arch}"
        fi
    done
    
    # Create checksums
    print_info "Creating checksums..."
    cd "${BUILD_DIR}" && sha256sum * > "checksums.txt" && cd ..
    print_success "Checksums created: ${BUILD_DIR}/checksums.txt"
}

# Create Docker image
docker_build() {
    if ! command -v docker &> /dev/null; then
        print_warning "Docker not found, skipping Docker build"
        return
    fi
    
    print_info "Building Docker image..."
    
    # Create Dockerfile if it doesn't exist
    if [[ ! -f "Dockerfile" ]]; then
        cat > Dockerfile << 'EOF'
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o trade-cli cmd/cli/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/trade-cli .
COPY --from=builder /app/config.json .
CMD ["./trade-cli"]
EOF
        print_info "Dockerfile created"
    fi
    
    if docker build -t "${PROJECT_NAME}:${VERSION}" .; then
        print_success "Docker image built: ${PROJECT_NAME}:${VERSION}"
        
        # Also tag as latest
        docker tag "${PROJECT_NAME}:${VERSION}" "${PROJECT_NAME}:latest"
        print_success "Also tagged as: ${PROJECT_NAME}:latest"
    else
        print_error "Docker build failed"
    fi
}

# Show build information
show_info() {
    print_info "=== Build Information ==="
    print_info "Project: ${PROJECT_NAME}"
    print_info "Version: ${VERSION}"
    print_info "Binary: ${BINARY_NAME}"
    print_info "Build Directory: ${BUILD_DIR}"
    print_info "Log File: ${LOG_FILE}"
    
    if command -v go &> /dev/null; then
        print_info "Go Version: $(go version | awk '{print $3}' | sed 's/go//')"
    fi
    
    if command -v docker &> /dev/null; then
        print_info "Docker: Available"
    else
        print_info "Docker: Not available"
    fi
}

# Show usage information
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  build           Build for current platform (default)"
    echo "  build-optimized Build with optimizations"
    echo "  cross           Cross-compile for multiple platforms"
    echo "  docker          Build Docker image"
    echo "  test            Run tests only"
    echo "  clean           Clean build artifacts"
    echo "  info            Show build information"
    echo "  all             Run full build process (clean, deps, test, build)"
    echo "  help            Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 build           # Standard build"
    echo "  $0 build-optimized # Optimized build"
    echo "  $0 cross           # Cross-compile"
    echo "  $0 all             # Full build process"
}

# Main function
main() {
    local command=${1:-"build"}
    
    case "${command}" in
        "build")
            check_go
            download_deps
            verify_deps
            build_current "$(uname -s | tr '[:upper:]' '[:lower:]')" "false"
            ;;
        "build-optimized")
            check_go
            download_deps
            verify_deps
            build_current "$(uname -s | tr '[:upper:]' '[:lower:]')" "true"
            ;;
        "cross")
            check_go
            download_deps
            verify_deps
            cross_compile
            ;;
        "docker")
            docker_build
            ;;
        "test")
            check_go
            download_deps
            run_tests
            ;;
        "clean")
            clean
            ;;
        "info")
            show_info
            ;;
        "all")
            check_go
            clean
            download_deps
            verify_deps
            run_tests
            build_current "$(uname -s | tr '[:upper:]' '[:lower:]')" "true"
            cross_compile
            docker_build
            ;;
        "help"|"-h"|"--help")
            usage
            ;;
        *)
            print_error "Unknown command: ${command}"
            usage
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"