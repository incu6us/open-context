# Usage Examples: Go Version & Library Information

This document provides practical examples of using the `get_go_info` tool.

## Example 1: Getting Go Version Release Notes

**Query:**
"What are the new features in Go 1.21?"

**MCP Tool Call:**
```json
{
  "name": "get_go_info",
  "arguments": {
    "type": "version",
    "version": "1.21"
  }
}
```

**Response:**
The tool will return markdown-formatted release notes including:
- Introduction to Go 1.21
- New language features
- Standard library changes
- Toolchain improvements
- Performance enhancements
- Links to official documentation

**First call:** Fetches from go.dev and caches locally
**Subsequent calls:** Loads instantly from cache

## Example 2: Exploring a Popular Go Library

**Query:**
"Tell me about the Gin web framework for Go"

**MCP Tool Call:**
```json
{
  "name": "get_go_info",
  "arguments": {
    "type": "library",
    "importPath": "github.com/gin-gonic/gin"
  }
}
```

**Response:**
Returns information about the latest version including:
- Package synopsis
- Installation command: `go get github.com/gin-gonic/gin`
- Import statement
- Repository URL
- License information
- Link to pkg.go.dev documentation

## Example 3: Checking a Specific Library Version

**Query:**
"Show me documentation for Cobra v1.8.0"

**MCP Tool Call:**
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

**Response:**
Returns version-specific information:
- Version: v1.8.0
- Installation: `go get github.com/spf13/cobra@v1.8.0`
- Documentation link with version
- All other metadata

## Example 4: Comparing Go Versions

**User conversation:**

User: "What's the difference between Go 1.20 and 1.21?"

AI Assistant:
1. Calls `get_go_info` with `{"type": "version", "version": "1.20"}`
2. Calls `get_go_info` with `{"type": "version", "version": "1.21"}`
3. Compares the release notes
4. Provides a summary of key differences

## Example 5: Library Recommendations

**User conversation:**

User: "I need a good HTTP router for Go. What do you recommend?"

AI Assistant:
1. Calls `get_go_info` for various routers:
   - `github.com/gorilla/mux`
   - `github.com/gin-gonic/gin`
   - `github.com/go-chi/chi`
2. Compares features, licenses, and documentation
3. Provides informed recommendation based on official data

## Example 6: Dependency Version Selection

**User conversation:**

User: "Should I upgrade to the latest version of viper?"

AI Assistant:
1. Calls `get_go_info` with `{"type": "library", "importPath": "github.com/spf13/viper"}`
2. Reviews the latest version information
3. Calls with specific older version if needed for comparison
4. Provides upgrade recommendation based on changes

## Cache Behavior

### First Request
```bash
# Fetches from network
→ Fetching Go 1.21 information from official source...
✓ Cached to: data/go/versions/1.21.json
```

### Second Request
```bash
# Loads from cache
→ Loaded Go 1.21 info from cache
✓ Instant response
```

### Cache Location

```
data/
└── go/
    ├── versions/
    │   ├── 1.20.json
    │   ├── 1.21.json
    │   └── 1.22.json
    └── libraries/
        ├── github.com_gin-gonic_gin.json
        ├── github.com_spf13_cobra_v1.8.0.json
        └── github.com_gorilla_mux.json
```

## Integration with Existing Tools

The `get_go_info` tool works alongside existing documentation tools:

1. **search_docs** - Search cached documentation
2. **get_docs** - Get specific documentation topics
3. **list_languages** - List available languages
4. **get_go_info** - Fetch fresh Go version/library info from official sources

## Error Handling

### Invalid Version
```json
{
  "type": "version",
  "version": "99.99"
}
```
Returns: Error with message about unavailable version

### Invalid Library
```json
{
  "type": "library",
  "importPath": "github.com/nonexistent/package"
}
```
Returns: Error from pkg.go.dev (404 status)

### Network Issues
If offline and no cache exists, returns network error.
If offline but cache exists, returns cached data.

## Best Practices

1. **Version format**: Use semantic version format (e.g., "1.21", "v1.8.0")
2. **Import paths**: Use full import paths (e.g., "github.com/org/repo")
3. **Cache management**: Delete cache files to force refresh
4. **Batch queries**: Make multiple calls for comparisons
5. **Offline usage**: Pre-cache important versions/libraries

## Performance Characteristics

| Scenario | Time | Network |
|----------|------|---------|
| First fetch (version) | ~1-2s | Required |
| First fetch (library) | ~1-2s | Required |
| Cached fetch | <10ms | None |
| Offline (cached) | <10ms | None |
| Offline (not cached) | Error | None |

## Use Cases Summary

✅ Learning about new Go versions
✅ Researching Go libraries
✅ Version-specific documentation
✅ License checking
✅ Dependency evaluation
✅ Offline development reference
✅ Automated documentation gathering
