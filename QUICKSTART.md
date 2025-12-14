# Quick Start Guide

Get started with Open Context MCP server in 5 minutes.

## Build

```bash
# Quick setup (builds server + fetcher, downloads Go docs)
make setup

# Or build manually
go build -o open-context
go build -o fetch-docs ./cmd/fetch
```

## Fetch Go Documentation

Automatically download Go standard library documentation from pkg.go.dev:

```bash
# Fetch ~100 commonly used Go packages
make fetch-go

# Or use the script
./scripts/fetch-go-docs.sh
```

## Test Locally

```bash
# Test the server directly
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./open-context
```

## Configure with Claude Desktop

1. **Find your config file:**
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`

2. **Add Open Context:**
   ```json
   {
     "mcpServers": {
       "open-context": {
         "command": "/full/path/to/open-context/open-context"
       }
     }
   }
   ```

3. **Restart Claude Desktop**

4. **Test it:**
   In Claude, type: "Use open-context to search for goroutines documentation"

## Configure with Cursor

1. **Open Cursor Settings** > Tools & Integrations > MCP Servers

2. **Add configuration:**
   ```json
   {
     "open-context": {
       "command": "/full/path/to/open-context/open-context"
     }
   }
   ```

3. **Restart Cursor**

## Usage Examples

### Search for Documentation

In your AI assistant, try:
- "Search open-context for 'http server' in Go"
- "Find TypeScript generics documentation using open-context"
- "List all available documentation languages from open-context"

### Get Specific Documentation

- "Get Go basics documentation from open-context"
- "Show me the goroutines documentation using open-context"
- "Get TypeScript generics docs from open-context"

## Add More Documentation

### Quick Method: Use the Example Script

```bash
# Add Jenkins documentation
./examples/add-jenkins-docs.sh

# Restart the server
```

### Manual Method

```bash
# Create structure
mkdir -p data/python/topics

# Add metadata
cat > data/python/metadata.json << 'EOF'
{
  "name": "python",
  "displayName": "Python",
  "description": "Python programming language documentation"
}
EOF

# Add a topic
cat > data/python/topics/basics.json << 'EOF'
{
  "id": "basics",
  "title": "Python Basics",
  "description": "Introduction to Python programming",
  "keywords": ["basics", "introduction", "syntax"],
  "content": "# Python Basics\n\nPython is a high-level programming language.\n\n```python\nprint('Hello, World!')\n```"
}
EOF

# Restart the server
```

## Available Documentation

Out of the box, Open Context includes:

### Go
- **basics**: Variables, functions, syntax
- **goroutines**: Concurrency with goroutines and channels
- **http**: HTTP servers and clients

### TypeScript
- **basics**: Type system, interfaces, type inference
- **generics**: Generic types and constraints

## Troubleshooting

### Server won't start
- Check that the binary has execute permissions: `chmod +x open-context`
- Verify the full path is correct in your config

### No documentation appears
- The server loads from the `data/` directory
- If `data/` doesn't exist, it uses built-in Go docs
- Check file permissions in `data/` directory

### Changes not reflected
- Restart the MCP server
- Restart your AI assistant (Claude Desktop, Cursor, etc.)

## Next Steps

- Read [README.md](README.md) for full documentation
- See [data/README.md](data/README.md) for documentation format details
- Check [CONTRIBUTING.md](CONTRIBUTING.md) to contribute new docs