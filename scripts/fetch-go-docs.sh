#!/bin/bash

# Fetch Go standard library documentation from pkg.go.dev
# This script builds the fetcher if needed and runs it

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

echo "Open Context - Go Documentation Fetcher"
echo "========================================"
echo ""

# Build the fetcher if it doesn't exist
if [ ! -f "./fetch-docs" ]; then
    echo "Building documentation fetcher..."
    go build -o fetch-docs ./cmd/fetch
    echo "✓ Fetcher built"
    echo ""
fi

# Run the fetcher
echo "Fetching Go standard library documentation from pkg.go.dev..."
echo "This may take a few minutes..."
echo ""

./fetch-docs -language=go -output=./data

echo ""
echo "========================================"
echo "✓ Go documentation fetched successfully!"
echo ""
echo "Documentation location: ./data/go/topics/"
echo ""
echo "To use the new documentation:"
echo "  1. Restart the open-context MCP server"
echo "  2. The Go stdlib packages will be searchable"
echo ""