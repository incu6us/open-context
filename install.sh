#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO_OWNER="incu6us"
REPO_NAME="open-context"
REPO_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}"
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

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        msys*|mingw*|cygwin*)
            OS="windows"
            ;;
        *)
            print_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="armv7"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    PLATFORM="${OS}_${ARCH}"
    print_info "Detected platform: $PLATFORM"
}

# Get latest release version from GitHub API
get_latest_version() {
    print_info "Fetching latest release version..."

    LATEST_VERSION=$(curl -sL "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$LATEST_VERSION" ]; then
        print_error "Failed to fetch latest version from GitHub"
        exit 1
    fi

    print_success "Latest version: $LATEST_VERSION"
}

# Get current installed version
get_installed_version() {
    BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"

    if [ -f "$BINARY_PATH" ]; then
        INSTALLED_VERSION=$("$BINARY_PATH" --version 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' || echo "")

        if [ -n "$INSTALLED_VERSION" ]; then
            print_info "Installed version: $INSTALLED_VERSION"
            return 0
        fi
    fi

    INSTALLED_VERSION=""
    print_info "No existing installation found"
    return 1
}

# Compare versions
compare_versions() {
    if [ "$INSTALLED_VERSION" = "$LATEST_VERSION" ]; then
        print_success "Already running the latest version ($LATEST_VERSION)"
        return 0
    else
        print_info "Update available: $INSTALLED_VERSION -> $LATEST_VERSION"
        return 1
    fi
}

# Download and install binary
download_and_install() {
    # Determine binary name based on OS
    if [ "$OS" = "windows" ]; then
        BINARY_FILE="${BINARY_NAME}.exe"
    else
        BINARY_FILE="$BINARY_NAME"
    fi

    # Construct download URL
    ASSET_NAME="${BINARY_NAME}_${PLATFORM}"
    if [ "$OS" = "windows" ]; then
        ASSET_NAME="${ASSET_NAME}.exe"
    fi

    DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${LATEST_VERSION}/${ASSET_NAME}"

    print_info "Downloading from: $DOWNLOAD_URL"

    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"

    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    TMP_FILE="$TMP_DIR/$BINARY_FILE"

    # Download binary
    if curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"; then
        print_success "Downloaded successfully"
    else
        print_error "Failed to download binary from $DOWNLOAD_URL"
        print_info "Please check if the release exists at: ${REPO_URL}/releases"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Make binary executable
    chmod +x "$TMP_FILE"

    # Move to install directory
    BINARY_PATH="$INSTALL_DIR/$BINARY_FILE"
    mv "$TMP_FILE" "$BINARY_PATH"

    # Clean up
    rm -rf "$TMP_DIR"

    print_success "Installed to: $BINARY_PATH"
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
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        print_info "Installation directory is not in your PATH"
        print_info "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo ""
        echo "    export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
    else
        print_success "Installation directory is in PATH"
    fi
}

# Verify installation
verify_installation() {
    if [ "$OS" = "windows" ]; then
        BINARY_PATH="$INSTALL_DIR/${BINARY_NAME}.exe"
    else
        BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"
    fi

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
    echo ""
    echo "Configuration:"
    echo "  Config file: $CONFIG_DIR/config.yaml"
    echo "  Cache dir:   $CONFIG_DIR/cache/"
    echo ""
    echo "For more information, visit: $REPO_URL"
    echo ""
}

# Main installation process
main() {
    echo "Starting installation..."
    echo ""

    # Detect platform
    detect_platform

    # Get versions
    get_latest_version
    get_installed_version || true

    # Check if update is needed
    if [ -n "$INSTALLED_VERSION" ]; then
        if compare_versions; then
            # Already up to date, but ensure config exists
            create_config_dir
            create_default_config
            check_path
            echo ""
            print_success "No action needed - already running latest version"
            exit 0
        fi
    fi

    # Download and install
    download_and_install

    # Setup configuration
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
