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

**Three MCP Tools:**
- `search_docs`: Intelligent search across all documentation
- `get_docs`: Retrieve specific documentation topics
- `list_languages`: Browse available languages and topics

### 2. Automatic Documentation Fetching

**Go Standard Library Fetcher**

Automatically downloads and parses documentation from pkg.go.dev:

```bash
# Fetch ~100 most common Go packages
make fetch-go
```

**Features:**
- Smart package selection (prioritizes commonly used packages)
- Extracts synopsis and descriptions from official pages
- Generates clean, searchable markdown documentation
- Rate-limited to respect pkg.go.dev servers
- Caches results locally in `data/go/topics/`

**How it works:**
1. Scrapes pkg.go.dev/std for package list
2. Prioritizes essential packages (fmt, io, net/http, etc.)
3. Downloads each package page
4. Extracts metadata and creates markdown
5. Saves in Open Context JSON format

### 3. Intelligent Search System

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

### 4. Extensible Architecture

**Two Ways to Add Documentation:**

**A. Auto-Fetch (For Supported Languages)**
```bash
./fetch-docs -language=go -output=./data
```

**B. Manual (For Any Language/Framework)**
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

### 5. Built-in Documentation

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
make setup  # Build + fetch Go docs
```

### 2. Update Documentation
```bash
make fetch-go  # Re-fetch Go stdlib
```

### 3. Search from AI
```
"Search open-context for http client in Go"
```

### 4. Get Specific Docs
```
"Get the net/http package documentation from open-context"
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

1. Create `fetcher/<language>_fetcher.go`
2. Implement fetching logic
3. Add to `cmd/fetch/main.go`
4. Run: `./fetch-docs -language=<language>`

### Example Structure
```go
type PythonFetcher struct {
    client   *http.Client
    cacheDir string
}

func (f *PythonFetcher) FetchDocs() error {
    // 1. Get package list
    // 2. For each package:
    //    - Fetch documentation
    //    - Parse content
    //    - Convert to topic format
    //    - Save JSON
    return nil
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