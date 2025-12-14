# Testing Guide

This guide explains how to test the Open Context MCP server.

## Prerequisites

1. **Build the server:**
   ```bash
   go build -o open-context
   ```

2. **Install jq (optional but recommended):**
   ```bash
   # macOS
   brew install jq

   # Linux
   apt-get install jq
   ```

## Quick Test

Run the automated test script:

```bash
./test.sh
```

This will test all MCP tools including:
- Server initialization
- Tools listing
- Documentation search
- Documentation retrieval
- Language listing
- Go version information fetching
- Go library information fetching

## Manual Testing

### 1. Test Server Initialization

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./open-context | jq .
```

**Expected output:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {}
    },
    "serverInfo": {
      "name": "open-context",
      "version": "0.1.0"
    }
  }
}
```

### 2. List Available Tools

```bash
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ./open-context | jq .
```

**Expected:** 4 tools listed (search_docs, get_docs, list_languages, get_go_info)

### 3. Search Documentation

```bash
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search_docs","arguments":{"query":"goroutines"}}}' | ./open-context | jq .
```

### 4. Get Specific Documentation

```bash
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_docs","arguments":{"id":"basics","language":"go"}}}' | ./open-context | jq .
```

### 5. Test Go Version Info (New Feature)

```bash
echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"version","version":"1.21"}}}' | ./open-context | jq .
```

**First run:** Fetches from go.dev and caches
**Second run:** Loads from cache instantly

### 6. Test Library Info (New Feature)

**Latest version:**
```bash
echo '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"library","importPath":"github.com/gin-gonic/gin"}}}' | ./open-context | jq .
```

**Specific version:**
```bash
echo '{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"library","importPath":"github.com/gin-gonic/gin","version":"v1.9.1"}}}' | ./open-context | jq .
```

## Interactive Testing

For interactive testing, you can use a simple test client:

```bash
# Create a test file
cat > test_interactive.sh << 'EOF'
#!/bin/bash

echo "Open Context Interactive Tester"
echo "==============================="
echo ""

# Start server in background
./open-context &
SERVER_PID=$!

# Give it a moment to start
sleep 1

# Send requests
echo "1. Initializing..."
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'

echo ""
echo "2. Listing tools..."
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'

echo ""
echo "3. Searching for HTTP..."
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search_docs","arguments":{"query":"http"}}}'

# Kill server
kill $SERVER_PID
EOF

chmod +x test_interactive.sh
./test_interactive.sh
```

## Testing with MCP Clients

### Claude Desktop

1. **Configure Claude Desktop:**

   Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

   ```json
   {
     "mcpServers": {
       "open-context": {
         "command": "/absolute/path/to/open-context"
       }
     }
   }
   ```

2. **Restart Claude Desktop**

3. **Test queries:**
   - "Search for Go concurrency documentation"
   - "What's new in Go 1.21?"
   - "Tell me about the Gin web framework"
   - "Get information about github.com/spf13/cobra"

### Cursor IDE

1. **Add to Cursor settings:**

   Go to Settings > Tools & Integrations > MCP Servers

   ```json
   {
     "open-context": {
       "command": "/absolute/path/to/open-context"
     }
   }
   ```

2. **Test in Cursor chat**

## Checking Cache

After testing the `get_go_info` tool, verify cache files were created:

```bash
# Check version cache
ls -lh data/go/versions/

# Check library cache
ls -lh data/go/libraries/

# View cached data
cat data/go/versions/1.21.json | jq .
cat data/go/libraries/github.com_gin-gonic_gin_v1.9.1.json | jq .
```

## Performance Testing

### Cache Performance

Test cache performance with a simple script:

```bash
cat > perf_test.sh << 'EOF'
#!/bin/bash

echo "Performance Test: Cache vs Network"
echo "==================================="

# First fetch (network)
echo "First fetch (should fetch from network):"
time echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"version","version":"1.22"}}}' | ./open-context > /dev/null

# Second fetch (cache)
echo ""
echo "Second fetch (should load from cache):"
time echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"version","version":"1.22"}}}' | ./open-context > /dev/null
EOF

chmod +x perf_test.sh
./perf_test.sh
```

**Expected:** Second fetch should be significantly faster (~100x)

## Error Testing

### Invalid Version

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"version","version":"99.99"}}}' | ./open-context | jq .
```

**Expected:** Error response with appropriate message

### Invalid Library

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"library","importPath":"github.com/nonexistent/package"}}}' | ./open-context | jq .
```

**Expected:** Error response about package not found

### Missing Parameters

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"version"}}}' | ./open-context | jq .
```

**Expected:** Error about missing version parameter

## Debugging

### Enable Verbose Logging

The server logs to stderr. Redirect to see detailed logs:

```bash
./open-context 2>server.log &
# Run tests
cat server.log
```

### Check Server Health

```bash
# Check if server is running
ps aux | grep open-context

# Check for network issues
ping go.dev
ping pkg.go.dev
```

## Common Issues

### 1. Server Not Starting

**Problem:** Server exits immediately

**Solution:**
- Check if binary exists: `ls -lh open-context`
- Rebuild: `go build -o open-context`
- Check permissions: `chmod +x open-context`

### 2. jq Not Found

**Problem:** `jq: command not found`

**Solution:**
```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# Or test without jq formatting
echo '...' | ./open-context | cat
```

### 3. Cache Not Working

**Problem:** Every request fetches from network

**Solution:**
- Check cache directory exists: `ls -la data/go/`
- Check permissions: `ls -la data/`
- Manually create: `mkdir -p data/go/versions data/go/libraries`

### 4. Network Errors

**Problem:** Failed to fetch from go.dev or pkg.go.dev

**Solution:**
- Check internet connection
- Verify URLs are accessible: `curl -I https://go.dev`
- Check firewall/proxy settings

## Test Coverage

The test suite covers:

✅ MCP protocol initialization
✅ Tool discovery (tools/list)
✅ Documentation search
✅ Documentation retrieval
✅ Language listing
✅ Go version fetching (with caching)
✅ Go library fetching (with caching)
✅ Version-specific library fetching
✅ Error handling

## Continuous Testing

For development, you can set up a watch script:

```bash
# Install entr (file watcher)
brew install entr  # macOS

# Watch for changes and rebuild+test
find . -name "*.go" | entr -c sh -c 'go build && ./test.sh'
```

## Integration Testing

Test with actual MCP client:

```bash
# Using stdio transport
cat << EOF | ./open-context
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"version","version":"1.21"}}}
EOF
```

## Success Criteria

A successful test run should show:
- ✅ All 8 tests pass
- ✅ No error responses
- ✅ Cache files created in data/go/
- ✅ Second fetch is significantly faster
- ✅ Valid JSON responses
- ✅ Proper MCP protocol compliance

## Next Steps

After successful testing:
1. Configure in your MCP client (Claude Desktop, Cursor, etc.)
2. Try real-world queries
3. Monitor cache growth
4. Report any issues on GitHub

## Additional Resources

- [MCP Protocol Specification](https://modelcontextprotocol.io/)
- [GO_VERSION_LIBRARY_FEATURE.md](GO_VERSION_LIBRARY_FEATURE.md) - Feature documentation
- [examples/go_version_library_usage.md](examples/go_version_library_usage.md) - Usage examples
