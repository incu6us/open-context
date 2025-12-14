# Go Version and Library Information Feature

This document describes the new feature that allows fetching and caching Go version release notes and library information from official sources.

## Overview

The application now supports fetching detailed information about:
1. **Go versions** - Release notes, features, and changes from go.dev
2. **Go libraries** - Package information, documentation, and metadata from pkg.go.dev

All fetched data is **automatically cached locally** in the `data/` directory to avoid redundant network requests and provide faster responses.

## MCP Tool: `get_go_info`

A new MCP tool has been added that AI assistants can use to fetch Go-related information on demand.

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `type` | string | Yes | Type of information: `"version"` or `"library"` |
| `version` | string | Conditional | Go version (e.g., `"1.21"`, `"1.22"`) when type is `"version"`. Optional library version when type is `"library"` |
| `importPath` | string | Conditional | Import path (e.g., `"github.com/gin-gonic/gin"`) when type is `"library"` |

### Examples

#### Fetch Go Version Information

**Input:**
```json
{
  "type": "version",
  "version": "1.21"
}
```

**Output:**
Markdown formatted release notes including:
- Introduction and overview
- New features
- Changes and improvements
- Links to official documentation

#### Fetch Library Information (Latest Version)

**Input:**
```json
{
  "type": "library",
  "importPath": "github.com/gin-gonic/gin"
}
```

**Output:**
Markdown formatted library information including:
- Package synopsis
- Installation instructions
- Import statement
- Repository and license information
- Links to pkg.go.dev documentation

#### Fetch Library Information (Specific Version)

**Input:**
```json
{
  "type": "library",
  "importPath": "github.com/gin-gonic/gin",
  "version": "v1.9.1"
}
```

**Output:**
Same as above, but for the specific version.

## Caching Mechanism

### Cache Structure

Fetched data is stored in the following structure:

```
data/
└── go/
    ├── versions/
    │   ├── 1.21.json       # Go 1.21 release notes
    │   └── 1.22.json       # Go 1.22 release notes
    └── libraries/
        ├── github.com_gin-gonic_gin.json           # Latest version
        └── github.com_gin-gonic_gin_v1.9.1.json   # Specific version
```

### Cache Behavior

1. **First Request**: Data is fetched from official sources (go.dev, pkg.go.dev) and cached locally
2. **Subsequent Requests**: Data is loaded from the local cache instantly
3. **Cache Invalidation**: To refresh data, simply delete the corresponding JSON file

### Cache Files

Each cache file contains structured JSON with:
- **Version files**: Release notes content, release URL, release date
- **Library files**: Import path, version, synopsis, description, repository, license

## API Usage (Programmatic)

### Fetch Go Version Information

```go
import "github.com/incu6us/open-context/fetcher"

f := fetcher.NewGoFetcher("./data")
versionInfo, err := f.FetchGoVersion("1.21")
if err != nil {
    log.Fatal(err)
}

fmt.Println(versionInfo.Content)
```

### Fetch Library Information

```go
import "github.com/incu6us/open-context/fetcher"

f := fetcher.NewGoFetcher("./data")

// Latest version
libInfo, err := f.FetchLibraryInfo("github.com/gin-gonic/gin", "")

// Specific version
libInfo, err := f.FetchLibraryInfo("github.com/gin-gonic/gin", "v1.9.1")

if err != nil {
    log.Fatal(err)
}

fmt.Println(libInfo.Description)
```

## Use Cases for AI Assistants

This feature enables AI assistants to:

1. **Answer version-specific questions**
   - "What's new in Go 1.21?"
   - "What are the breaking changes in Go 1.22?"

2. **Provide library recommendations**
   - "Tell me about the Gin web framework"
   - "What's the latest version of github.com/gorilla/mux?"

3. **Help with dependency selection**
   - "Should I use this library version?"
   - "What's the license for this package?"

4. **Offer accurate, up-to-date information**
   - Data comes directly from official sources
   - Cached for performance and offline access

## Benefits

✅ **Always up-to-date** - Fetches from official Go sources
✅ **Fast responses** - Local caching for instant retrieval
✅ **Offline capable** - Works with cached data when offline
✅ **Bandwidth efficient** - Fetches once, uses many times
✅ **Version-aware** - Supports both latest and specific versions

## Implementation Details

### Data Sources

- **Go Version Info**: https://go.dev/doc/go{version}
- **Library Info**: https://pkg.go.dev/{importPath}[@version]

### HTTP Client

- 30-second timeout for requests
- Proper error handling for network issues
- Respects official source rate limits

### HTML Parsing

- Uses `golang.org/x/net/html` for robust HTML parsing
- Extracts relevant content from structured documentation
- Converts to Markdown for AI-friendly format

## Future Enhancements

Potential improvements for this feature:

- [ ] Support for more programming languages (Python, Rust, TypeScript)
- [ ] Automatic cache invalidation/refresh
- [ ] Batch fetching for multiple libraries
- [ ] Enhanced metadata extraction (contributors, stars, downloads)
- [ ] Support for Go modules and dependencies
