# Manual Testing Guide

Quick reference for testing the MCP server manually.

## Method 1: Using the Test Script

**Run all tests:**
```bash
./test.sh
```

This will test all 8 features including the new `get_go_info` tool.

## Method 2: Direct JSON-RPC Testing

**Step 1: Start the server in a terminal**
```bash
./open-context
```

The server will wait for JSON-RPC requests on stdin.

**Step 2: In the same terminal, type these JSON requests:**

### Initialize
```json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
```

### List Tools
```json
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
```

### Get Go Version Info
```json
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"version","version":"1.21"}}}
```

### Get Library Info
```json
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"library","importPath":"github.com/gin-gonic/gin"}}}
```

### Get Library with Version
```json
{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"library","importPath":"github.com/spf13/cobra","version":"v1.8.0"}}}
```

**Step 3: Press Ctrl+D to exit**

## Method 3: Echo Pipe (Quick Single Test)

```bash
# Test Go version
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"version","version":"1.21"}}}' | ./open-context

# Test library
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"library","importPath":"github.com/gin-gonic/gin"}}}' | ./open-context
```

## Method 4: Using a File

Create a test file with multiple requests:

**Create `requests.jsonl`:**
```bash
cat > requests.jsonl << 'EOF'
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"version","version":"1.21"}}}
EOF
```

**Run:**
```bash
./open-context < requests.jsonl
```

## Method 5: With Claude Desktop

1. **Configure** `~/Library/Application Support/Claude/claude_desktop_config.json`:
   ```json
   {
     "mcpServers": {
       "open-context": {
         "command": "/full/path/to/open-context"
       }
     }
   }
   ```

2. **Restart Claude Desktop**

3. **Ask questions:**
   - "What's new in Go 1.22?"
   - "Tell me about the Gin web framework for Go"
   - "Show me information about github.com/spf13/cobra v1.8.0"

## Quick Tests

### Test 1: Version Info (First Fetch)
```bash
time (echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"version","version":"1.22"}}}' | ./open-context > /dev/null)
```
Should take ~1-2 seconds (fetches from go.dev)

### Test 2: Version Info (Cached)
```bash
time (echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"version","version":"1.22"}}}' | ./open-context > /dev/null)
```
Should be instant (<100ms - loads from cache)

### Test 3: Check Cache
```bash
ls -lh data/go/versions/
ls -lh data/go/libraries/
```

### Test 4: View Cached Data
```bash
cat data/go/versions/1.22.json | jq .
```

## Common Test Queries

### Search Documentation
```json
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"search_docs","arguments":{"query":"http"}}}
```

### Get Documentation
```json
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_docs","arguments":{"id":"basics","language":"go"}}}
```

### List Languages
```json
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_languages","arguments":{}}}
```

## Debugging

**Enable logging:**
```bash
./open-context 2> debug.log &
# Run your tests
cat debug.log
```

**Check for errors:**
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"invalid"}}}' | ./open-context
```

## Expected Results

✅ **Success Response Format:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "... documentation content ..."
      }
    ]
  }
}
```

❌ **Error Response Format:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32000,
    "message": "error description"
  }
}
```

## Performance Benchmarks

| Operation | First Fetch | Cached |
|-----------|-------------|--------|
| Go Version | ~1.5s | <10ms |
| Library Info | ~1.2s | <5ms |
| Search Docs | N/A | ~50ms |
| Get Docs | N/A | ~5ms |
