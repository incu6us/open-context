.PHONY: build clean fetch-go test install help

# Build the MCP server
build:
	@echo "Building open-context MCP server..."
	@go build -o open-context .
	@echo "✓ Built: ./open-context"

# Build the documentation fetcher
build-fetcher:
	@echo "Building documentation fetcher..."
	@go build -o fetch-docs ./cmd/fetch
	@echo "✓ Built: ./fetch-docs"

# Build everything
all: build build-fetcher

# Fetch Go standard library documentation
fetch-go: build-fetcher
	@echo "Fetching Go standard library documentation..."
	@./fetch-docs -language=go -output=./data
	@echo "✓ Go documentation fetched to ./data/go/"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f open-context fetch-docs
	@echo "✓ Cleaned"

# Clean all generated documentation
clean-docs:
	@echo "Cleaning generated documentation..."
	@rm -rf ./data/go/topics/*.json
	@echo "✓ Documentation cleaned"

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Run the MCP server
run: build
	@./open-context

# Quick test of the server
quick-test: build
	@echo "Testing MCP server initialization..."
	@echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./open-context | head -1
	@echo ""
	@echo "✓ Server is working!"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

# Full setup: build everything and fetch Go docs
setup: all fetch-go
	@echo ""
	@echo "✓ Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Configure your MCP client (see QUICKSTART.md)"
	@echo "  2. Run: make run (or ./open-context)"

help:
	@echo "Open Context - MCP Documentation Server"
	@echo ""
	@echo "Available targets:"
	@echo "  make build         - Build the MCP server"
	@echo "  make build-fetcher - Build the documentation fetcher"
	@echo "  make all           - Build everything"
	@echo "  make fetch-go      - Fetch Go standard library documentation"
	@echo "  make setup         - Build and fetch all documentation"
	@echo "  make run           - Run the MCP server"
	@echo "  make test          - Run tests"
	@echo "  make quick-test    - Quick test of the server"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make clean-docs    - Clean generated documentation"
	@echo "  make fmt           - Format code"
	@echo "  make deps          - Install dependencies"
	@echo "  make help          - Show this help"