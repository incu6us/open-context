.PHONY: build clean test install help

# Build the MCP server
build:
	@echo "Building open-context MCP server..."
	@go build -o open-context .
	@echo "✓ Built: ./open-context"

# Build everything (kept for compatibility)
all: build

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f open-context
	@echo "✓ Cleaned"

# Clean cache directory
clean-cache:
	@echo "Cleaning cache directory..."
	@rm -rf ~/.open-context/cache
	@echo "✓ Cache cleaned"

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

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run --timeout=5m
	@echo "✓ Linting complete"

# Full setup: build and run initial test
setup: build
	@echo ""
	@echo "✓ Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Configure your MCP client (see README.md)"
	@echo "  2. Run: make run (or ./open-context)"
	@echo ""
	@echo "Note: Documentation is fetched on-demand from official sources"
	@echo "      when you use the MCP tools (no pre-fetching needed)"

help:
	@echo "Open Context - MCP Documentation Server"
	@echo ""
	@echo "Available targets:"
	@echo "  make build        - Build the MCP server"
	@echo "  make all          - Build everything (same as build)"
	@echo "  make setup        - Build and show next steps"
	@echo "  make run          - Run the MCP server"
	@echo "  make test         - Run tests"
	@echo "  make quick-test   - Quick test of the server"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make clean-cache  - Clean cache directory (~/.open-context/cache)"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Run linter (requires golangci-lint)"
	@echo "  make deps         - Install dependencies"
	@echo "  make help         - Show this help"