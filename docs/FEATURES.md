# Open Context Features

## Overview

Open Context is an MCP server that provides intelligent documentation access for programming languages and frameworks. Unlike static documentation, Open Context can automatically fetch and update documentation from official sources.

## Core Features

### 1. MCP Protocol Integration

Full implementation of the Model Context Protocol (MCP) for seamless integration with:
- Claude Desktop
- Cursor IDE
- Claude Code CLI
- Any MCP-compatible client

**Four MCP Tools:**
- `open-context_search_docs`: Intelligent search across all documentation
- `open-context_get_docs`: Retrieve specific documentation topics
- `open-context_list_docs`: Browse available documentation and topics
- `open-context_get_go_info`: Fetch Go version info and library documentation from official sources (go.dev, pkg.go.dev)

### 2. Automatic Documentation Fetching

**On-Demand Documentation Fetching**

Documentation is fetched automatically from official sources when requested via MCP tools:

```bash
# No pre-fetching needed!
# Use MCP tools like get_go_info to fetch on-demand
```

**Features:**
- Smart package selection (prioritizes commonly used packages)
- Extracts synopsis and descriptions from official pages
- Generates clean, searchable markdown documentation
- Rate-limited to respect pkg.go.dev servers
- Caches results locally in `~/.open-context/cache/`
- Automatic cache expiration based on configurable TTL

**How it works:**
1. Scrapes pkg.go.dev/std for package list
2. Prioritizes essential packages (fmt, io, net/http, etc.)
3. Downloads each package page
4. Extracts metadata and creates markdown
5. Saves in Open Context JSON format

### 3. Configuration & Cache Management

**Auto-Configuration**

On first run, Open Context automatically creates:
- `~/.open-context/config.yaml` - Configuration file with sensible defaults
- `~/.open-context/cache/` - Cache directory for downloaded documentation

**Cache TTL (Time To Live)**

Configure how long cached data should be kept:

```yaml
# ~/.open-context/config.yaml
cache_ttl: 7d  # Expire after 7 days (default)
```

**Supported Duration Formats:**
- `0` - Never expire (permanent cache)
- `7d` - 7 days
- `24h` - 24 hours
- `30m` - 30 minutes
- `1w` - 1 week

**Cache Clearing**

Clear all cached data using the CLI:

```bash
# Clear cache
./open-context --clear-cache

# Or use the short alias
./open-context --cc
```

**Cross-Platform Support**

Works seamlessly regardless of installation method:
- Direct build (`go build`)
- Go install (`go install`)
- Package managers (Homebrew, etc.)

Configuration and cache are always stored in `~/.open-context/` on all platforms (macOS, Linux, Windows).

### 4. Intelligent Search System

**Scored Search Results:**
- Title matches: 10 points
- Keyword matches: 5 points each
- Description matches: 3 points
- Content matches: 1 point

**Features:**
- Search across all languages or filter by specific language
- Results sorted by relevance
- Keyword-based matching
- Fast in-memory search

### 5. Extensible Architecture

**Adding Documentation:**

**Manual Documentation (For Any Language/Framework)**
```bash
# Create structure
mkdir -p data/jenkins/topics

# Add metadata
cat > data/jenkins/metadata.json << 'EOF'
{
  "name": "jenkins",
  "displayName": "Jenkins",
  "description": "Jenkins CI/CD documentation"
}
EOF

# Add topics (JSON files)
cat > data/jenkins/topics/pipeline.json << 'EOF'
{
  "id": "pipeline",
  "title": "Jenkins Pipelines",
  "description": "CI/CD pipelines with Jenkinsfile",
  "keywords": ["pipeline", "ci", "cd", "jenkinsfile"],
  "content": "# Jenkins Pipelines\n\n..."
}
EOF
```

### 6. Built-in Documentation

Ships with curated documentation:

**Go:**
- Basics (variables, functions, syntax)
- Goroutines and concurrency
- HTTP servers and clients

**TypeScript:**
- Type system basics
- Generics and advanced types

These serve as examples and fallback if data directory doesn't exist.

## Technical Architecture

### Components

```
┌─────────────────────────────────────────────┐
│           MCP Client (Claude/Cursor)        │
└─────────────────┬───────────────────────────┘
                  │ JSON-RPC over stdio
                  ▼
┌─────────────────────────────────────────────┐
│         MCP Server (server/server.go)       │
│  - Protocol handling                        │
│  - Tool registration                        │
│  - Request routing                          │
└─────────────────┬───────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────┐
│    Documentation Provider (docs/provider.go) │
│  - Load documentation                       │
│  - Search & scoring                         │
│  - Topic retrieval                          │
└─────────────────┬───────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────┐
│         Documentation Storage (data/)       │
│  - JSON files organized by language        │
│  - Metadata + topics structure             │
│  - Auto-loaded on server start             │
└─────────────────────────────────────────────┘
```

### Fetcher System

```
┌─────────────────────────────────────────────┐
│      Fetch Command (cmd/fetch/main.go)      │
│  - CLI interface                            │
│  - Language selection                       │
│  - Output directory config                  │
└─────────────────┬───────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────┐
│    Go Fetcher (fetcher/go_fetcher.go)      │
│  - Scrape pkg.go.dev                       │
│  - Parse HTML                               │
│  - Extract documentation                    │
│  - Generate markdown                        │
│  - Save to JSON                             │
└─────────────────┬───────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────┐
│         Official Documentation Sources      │
│  - pkg.go.dev (Go)                         │
│  - Future: Python, Rust, etc.              │
└─────────────────────────────────────────────┘
```

## Usage Patterns

### 1. Quick Setup
```bash
make setup  # Build the server
```

### 2. Clear Cache (optional)
```bash
./open-context --clear-cache  # Or: make clean-cache
```

### 3. Fetch Documentation On-Demand
Documentation is fetched automatically when you ask Claude:
```
"What's new in Go 1.21?"
"Show me how to use github.com/gin-gonic/gin"
"Get the latest TypeScript 5.0 features"
```

### 4. Search Existing Docs
```
"Search open-context for http client in Go"
```

### 5. Add Custom Docs
```bash
# Manual: Create JSON files in data/<language>/topics/
# Or: Write a new fetcher in fetcher/<language>_fetcher.go
```

## Data Format

### Metadata (`data/<language>/metadata.json`)
```json
{
  "name": "go",
  "displayName": "Go",
  "description": "Go programming language documentation"
}
```

### Topic (`data/<language>/topics/<topic>.json`)
```json
{
  "id": "unique-id",
  "title": "Human Readable Title",
  "description": "Brief description",
  "keywords": ["search", "terms", "here"],
  "content": "# Markdown Content\n\n..."
}
```

## Performance

- **Startup**: Fast (loads all docs into memory)
- **Search**: Instant (in-memory search with scoring)
- **Fetch**: ~100 packages in ~1 minute (rate-limited)
- **Memory**: Minimal (only metadata and text)

## Extensibility

### Add a New Language Fetcher

Fetchers can be created to support additional languages using the on-demand fetching pattern:

1. Create `fetcher/<language>_fetcher.go`
2. Implement fetching logic similar to `go_fetcher.go`
3. Add new MCP tool to expose the fetcher functionality
4. Documentation is automatically cached on first request

### Example Structure
```go
type PythonFetcher struct {
    client   *http.Client
    cache    *cache.Cache
}

func (f *PythonFetcher) FetchPackageInfo(packageName, version string) (*PackageInfo, error) {
    // 1. Check cache first
    // 2. If not cached:
    //    - Fetch from official source (e.g., pypi.org)
    //    - Parse content
    //    - Convert to markdown with YAML frontmatter
    //    - Save to cache
    // 3. Return package info
    return packageInfo, nil
}
```

## Comparison with Context7

| Feature | Open Context | Context7 |
|---------|-------------|----------|
| Language | Go | Node.js/TypeScript |
| Documentation | File-based + Auto-fetch | API-based |
| Extensibility | JSON files + Fetchers | Proprietary |
| Languages | Go, TypeScript (+ manual) | 100+ (via API) |
| Self-hosted | Yes | Yes |
| Backend | Local files | Private API |
| Customization | Full control | Limited |

## Future Enhancements

See [README.md](README.md#roadmap) for full roadmap.

**Upcoming:**
- Python, Rust, Java fetchers
- Version-specific documentation
- Code example extraction
- Incremental updates
- Parallel fetching
- Web UI for browsing

## License

MIT License - see [LICENSE](LICENSE)