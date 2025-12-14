# Documentation Fetcher

The fetcher module automatically downloads documentation from official sources and converts it into the Open Context format.

## Supported Languages

### Go
Fetches Go standard library documentation from [pkg.go.dev](https://pkg.go.dev).

## Usage

### Command Line

Build and run the fetcher:

```bash
# Build the fetcher
go build -o fetch-docs ./cmd/fetch

# Fetch Go documentation
./fetch-docs -language=go -output=./data
```

### Using Make

```bash
# Fetch Go docs
make fetch-go

# Or as part of setup
make setup
```

### Using the Script

```bash
./scripts/fetch-go-docs.sh
```

## How It Works

### Go Fetcher

1. **Discovers Packages**: Scrapes pkg.go.dev/std to find all standard library packages
2. **Prioritizes**: Selects ~100 most commonly used packages
3. **Fetches Documentation**: For each package:
   - Downloads the pkg.go.dev page
   - Extracts synopsis from meta tags
   - Generates structured markdown documentation
   - Creates links to official documentation
4. **Saves**: Converts to Open Context JSON format in `data/go/topics/`

### Priority Packages

The fetcher prioritizes commonly used packages:
- **I/O & Formatting**: fmt, io, os, strings, strconv, bufio, bytes
- **Networking**: net/http, net/url
- **Encoding**: encoding/json, html/template, text/template
- **Concurrency**: context, sync
- **Crypto**: crypto/sha256, crypto/md5, crypto/tls
- **Time & Math**: time, math
- **Testing**: testing
- **And more...**

## Output Format

Each package generates a JSON file like:

```json
{
  "id": "net_http",
  "title": "Go Package: net/http",
  "description": "Package http provides HTTP client and server implementations.",
  "keywords": ["http", "net/http", "standard library", "stdlib", "go", "net"],
  "content": "# Package net/http\n\nImport path: `net/http`\n\n..."
}
```

## Configuration

### Fetch Limits

By default, the fetcher limits to 100 packages to avoid overwhelming the system. You can modify this in `fetcher/go_fetcher.go`:

```go
if len(result) >= 100 { // Adjust this limit
    break
}
```

### Rate Limiting

The fetcher waits 500ms between requests to be respectful to pkg.go.dev:

```go
time.Sleep(500 * time.Millisecond)
```

## Adding More Languages

To add a fetcher for a new language:

1. Create `<language>_fetcher.go` in the `fetcher/` directory
2. Implement the fetching logic (web scraping, API calls, etc.)
3. Convert to the Open Context topic format
4. Add to `cmd/fetch/main.go`

Example structure:

```go
type TypeScriptFetcher struct {
    client   *http.Client
    cacheDir string
}

func (f *TypeScriptFetcher) FetchDocs() error {
    // Implementation
}
```

## Best Practices

1. **Respect Rate Limits**: Always add delays between requests
2. **Handle Errors Gracefully**: Log warnings but continue processing
3. **Validate Output**: Ensure generated JSON is valid
4. **Cache Results**: Don't re-fetch unless necessary
5. **Attribution**: Include links to original documentation

## Troubleshooting

### Network Errors

If you encounter network errors:
- Check your internet connection
- Verify pkg.go.dev is accessible
- Increase the timeout in the HTTP client

### Missing Packages

If some packages aren't fetched:
- Check the logs for specific errors
- Some packages may fail due to parsing issues
- Review the priority list to ensure your package is included

### Parse Errors

If HTML parsing fails:
- The website structure may have changed
- Update the extraction functions in `go_fetcher.go`

## New Features

### Go Version Information
The fetcher now supports fetching Go version release notes from go.dev:

```go
f := fetcher.NewGoFetcher("./data")
versionInfo, err := f.FetchGoVersion("1.21")
// Returns release notes, features, and documentation
```

### Go Library Information
The fetcher can fetch information about any Go library from pkg.go.dev:

```go
f := fetcher.NewGoFetcher("./data")

// Latest version
libInfo, err := f.FetchLibraryInfo("github.com/gin-gonic/gin", "")

// Specific version
libInfo, err := f.FetchLibraryInfo("github.com/gin-gonic/gin", "v1.9.1")
```

Both features include automatic caching:
- **First request**: Fetches from official sources and caches locally
- **Subsequent requests**: Loads instantly from cache
- **Cache locations**:
  - Versions: `data/go/versions/{version}.json`
  - Libraries: `data/go/libraries/{importPath}_{version}.json`

See [GO_VERSION_LIBRARY_FEATURE.md](../GO_VERSION_LIBRARY_FEATURE.md) for detailed documentation.

## Future Enhancements

- [ ] Fetch all packages (not just priority ones)
- [x] Support for specific Go versions (✓ Implemented)
- [ ] Extract code examples from documentation
- [ ] Parse function signatures and types
- [x] Support for third-party packages (✓ Implemented via FetchLibraryInfo)
- [ ] Incremental updates (only fetch changed docs)
- [ ] Parallel fetching for faster downloads
- [ ] Multi-language support (Python, Rust, TypeScript, etc.)