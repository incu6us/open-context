#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO_URL="https://github.com/incu6us/open-context"
BINARY_NAME="open-context"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
CONFIG_DIR="$HOME/.open-context"

echo "========================================"
echo "  Open Context MCP Server Installer"
echo "========================================"
echo ""

# Function to print colored messages
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.23 or later from https://golang.org/dl/"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go $GO_VERSION is installed"
}

# Check Go version
check_go_version() {
    REQUIRED_VERSION="1.23"
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')

    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        print_error "Go version $GO_VERSION is too old. Please install Go $REQUIRED_VERSION or later"
        exit 1
    fi
}

# Install from source
install_from_source() {
    print_info "Installing open-context from source..."

    # Use go install to download and install
    if go install github.com/incu6us/open-context@latest; then
        print_success "Successfully installed open-context"
    else
        print_error "Failed to install from source"
        exit 1
    fi
}

# Create config directory
create_config_dir() {
    if [ ! -d "$CONFIG_DIR" ]; then
        mkdir -p "$CONFIG_DIR"
        print_success "Created configuration directory: $CONFIG_DIR"
    else
        print_info "Configuration directory already exists: $CONFIG_DIR"
    fi
}

# Create default config file
create_default_config() {
    CONFIG_FILE="$CONFIG_DIR/config.yaml"

    if [ ! -f "$CONFIG_FILE" ]; then
        cat > "$CONFIG_FILE" << 'EOF'
# Open Context Configuration
# Cache TTL (time to live) - how long cached data remains valid
# Format: duration string (e.g., "24h", "7d", "168h")
# Default: 168h (7 days)
cache_ttl: 168h
EOF
        print_success "Created default configuration: $CONFIG_FILE"
    else
        print_info "Configuration file already exists: $CONFIG_FILE"
    fi
}

# Check if binary is in PATH
check_path() {
    GOBIN="${GOBIN:-$(go env GOPATH)/bin}"

    if [[ ":$PATH:" != *":$GOBIN:"* ]]; then
        print_info "Go bin directory is not in your PATH"
        print_info "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo ""
        echo "    export PATH=\"\$PATH:$GOBIN\""
        echo ""
    else
        print_success "Go bin directory is in PATH"
    fi
}

# Verify installation
verify_installation() {
    GOBIN="${GOBIN:-$(go env GOPATH)/bin}"
    BINARY_PATH="$GOBIN/$BINARY_NAME"

    if [ -f "$BINARY_PATH" ]; then
        print_success "Binary installed at: $BINARY_PATH"

        # Test if it's executable
        if "$BINARY_PATH" --version &> /dev/null; then
            VERSION=$("$BINARY_PATH" --version)
            print_success "Installation verified: $VERSION"
        else
            print_error "Binary is not executable"
            exit 1
        fi
    else
        print_error "Binary not found at expected location: $BINARY_PATH"
        exit 1
    fi
}

# Print usage instructions
print_usage() {
    echo ""
    echo "========================================"
    echo "  Installation Complete!"
    echo "========================================"
    echo ""
    echo "Usage:"
    echo "  $BINARY_NAME --help              # Show help"
    echo "  $BINARY_NAME --version           # Show version"
    echo "  $BINARY_NAME                     # Start MCP server"
    echo "  $BINARY_NAME --clear-cache       # Clear cache"
    echo ""
    echo "Configuration:"
    echo "  Config file: $CONFIG_DIR/config.yaml"
    echo "  Cache dir:   $CONFIG_DIR/cache/"
    echo ""
    echo "Available Tools:"
    echo "  - Go package/module/library documentation"
    echo "  - npm package information"
    echo "  - Node.js version information"
    echo "  - TypeScript version information"
    echo "  - React version information"
    echo "  - Next.js version information"
    echo "  - Ansible version information"
    echo "  - Terraform version information"
    echo "  - Jenkins version information"
    echo "  - Kubernetes version information"
    echo "  - Helm version information"
    echo "  - Docker image information"
    echo ""
    echo "For more information, visit: $REPO_URL"
    echo ""
}

# Main installation process
main() {
    echo "Starting installation..."
    echo ""

    # Check prerequisites
    check_go
    check_go_version

    # Install
    install_from_source

    # Setup
    create_config_dir
    create_default_config

    # Verify
    verify_installation
    check_path

    # Done
    print_usage
}

# Run main installation
main
