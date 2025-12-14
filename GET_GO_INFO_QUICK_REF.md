# get_go_info - Quick Reference

## Overview

The `get_go_info` MCP tool fetches and caches information about Go versions and libraries from official sources (go.dev and pkg.go.dev).

## Usage

### 1. Get Go Version Information

**Fetch release notes for any Go version:**

```json
{
  "name": "get_go_info",
  "arguments": {
    "type": "version",
    "version": "1.25"
  }
}
```

**Examples:**
- `"version": "1.25"` - Latest Go version (Dec 2025)
- `"version": "1.24"` - Previous version
- `"version": "1.21"` - Earlier version

**Returns:**
- Release notes from go.dev
- New features and changes
- Links to official documentation
- Markdown formatted

### 2. Get Library Information (Latest)

**Fetch info about any Go library:**

```json
{
  "name": "get_go_info",
  "arguments": {
    "type": "library",
    "importPath": "github.com/gin-gonic/gin"
  }
}
```

**Popular libraries to try:**
- `github.com/gin-gonic/gin` - Web framework
- `github.com/spf13/cobra` - CLI library
- `github.com/gorilla/mux` - HTTP router
- `github.com/sirupsen/logrus` - Logger
- `gorm.io/gorm` - ORM library

**Returns:**
- Synopsis and description
- Installation instructions
- Import statement
- Repository URL
- License information
- Link to pkg.go.dev

### 3. Get Library Information (Specific Version)

**Fetch version-specific library info:**

```json
{
  "name": "get_go_info",
  "arguments": {
    "type": "library",
    "importPath": "github.com/spf13/cobra",
    "version": "v1.8.0"
  }
}
```

**Returns:**
- Same as latest, but for specific version
- Installation command includes `@version`
- Documentation link includes version

## Command Line Examples

### Using echo pipe:

```bash
# Get Go 1.25 info
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"version","version":"1.25"}}}' | ./open-context

# Get Gin framework info
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"library","importPath":"github.com/gin-gonic/gin"}}}' | ./open-context

# Get Cobra v1.8.0 info
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"library","importPath":"github.com/spf13/cobra","version":"v1.8.0"}}}' | ./open-context
```

### Using file input:

```bash
# Create request file
cat > request.jsonl << 'EOF'
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"version","version":"1.25"}}}
EOF

# Execute
./open-context < request.jsonl
```

## Caching Behavior

### First Request
- Fetches from official source (go.dev or pkg.go.dev)
- Takes ~1-2 seconds
- Saves to cache
- Returns markdown content

### Subsequent Requests
- Loads from local cache
- Takes <10ms (100x faster!)
- Returns same content instantly

### Cache Locations

```
data/
└── go/
    ├── versions/
    │   ├── 1.21.json
    │   ├── 1.24.json
    │   └── 1.25.json
    └── libraries/
        ├── github.com_gin-gonic_gin.json
        ├── github.com_spf13_cobra.json
        └── github.com_spf13_cobra_v1.8.0.json
```

### Clear Cache

```bash
# Clear all version cache
rm -rf data/go/versions/

# Clear all library cache
rm -rf data/go/libraries/

# Clear specific version
rm data/go/versions/1.25.json

# Clear specific library
rm data/go/libraries/github.com_gin-gonic_gin.json
```

## Use Cases

### For AI Assistants

**When user asks:**
- "What's new in Go 1.25?"
  → Use `{"type": "version", "version": "1.25"}`

- "Tell me about the Gin web framework"
  → Use `{"type": "library", "importPath": "github.com/gin-gonic/gin"}`

- "Should I use Cobra v1.8.0?"
  → Use `{"type": "library", "importPath": "github.com/spf13/cobra", "version": "v1.8.0"}`

### For Developers

- Research libraries before adding dependencies
- Check version-specific changes
- Compare library licenses
- Get installation commands
- Find official documentation

## Error Handling

### Invalid Version

```json
{"type": "version", "version": "99.99"}
```
**Returns:** Error - Version not found on go.dev

### Invalid Library

```json
{"type": "library", "importPath": "github.com/nonexistent/package"}
```
**Returns:** Error - 404 from pkg.go.dev

### Missing Parameters

```json
{"type": "version"}
```
**Returns:** Error - version parameter required

```json
{"type": "library"}
```
**Returns:** Error - importPath parameter required

## Integration with Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "open-context": {
      "command": "/full/path/to/open-context"
    }
  }
}
```

Then ask Claude:
- "What's new in the latest Go version?"
- "Tell me about popular Go web frameworks"
- "Get information about github.com/spf13/cobra"

## Performance

| Operation | First Fetch | Cached |
|-----------|-------------|--------|
| Go Version | ~1.5s | <10ms |
| Library (latest) | ~1.2s | <5ms |
| Library (versioned) | ~1.2s | <5ms |

## Data Sources

- **Go Versions:** https://go.dev/doc/go{version}
- **Libraries:** https://pkg.go.dev/{importPath}[@version]

## Output Format

Returns MCP-formatted response:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "# Markdown formatted content here..."
      }
    ]
  }
}
```

## See Also

- [GO_VERSION_LIBRARY_FEATURE.md](GO_VERSION_LIBRARY_FEATURE.md) - Complete feature documentation
- [examples/go_version_library_usage.md](examples/go_version_library_usage.md) - Usage examples
- [TESTING.md](TESTING.md) - Testing guide
- [README.md](README.md) - Main documentation
