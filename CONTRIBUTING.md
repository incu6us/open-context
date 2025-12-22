# Contributing to Open Context

Thank you for your interest in contributing to Open Context! This document provides guidelines for contributing to the project.

## How to Contribute

### Adding Documentation

The easiest way to contribute is by adding documentation for new languages or frameworks.

1. **Fork the repository**

2. **Create a new branch**
   ```bash
   git checkout -b add-<language>-docs
   ```

3. **Add your documentation**

   Follow the structure in `data/README.md`:

   ```bash
   mkdir -p data/<language>/topics
   ```

   Create `data/<language>/metadata.json`:
   ```json
   {
     "name": "language-name",
     "displayName": "Language Display Name",
     "description": "Brief description"
   }
   ```

   Add topics in `data/<language>/topics/<topic>.json`:
   ```json
   {
     "id": "topic-id",
     "title": "Topic Title",
     "description": "Brief description",
     "keywords": ["keyword1", "keyword2"],
     "content": "# Markdown content here..."
   }
   ```

4. **Test your changes**
   ```bash
   go build -o open-context
   ./test.sh
   ```

5. **Submit a pull request**

### Code Contributions

For code changes:

1. **Open an issue first** to discuss the change
2. **Fork and create a branch**
3. **Make your changes** following Go best practices
4. **Add tests** if applicable
5. **Ensure code builds**: `go build`
6. **Format code**: `go fmt ./...`
7. **Submit a pull request**

## Documentation Standards

### Content Quality

- **Accurate**: All code examples should be tested and working
- **Current**: Use latest stable versions of languages/frameworks
- **Complete**: Include both explanations and code examples
- **Clear**: Write for developers of all skill levels

### Formatting

- Use Markdown for all content
- Include syntax-highlighted code blocks
- Add comments to complex code examples
- Keep examples concise but practical

### Keywords

Choose keywords that developers would naturally search for:
- Technical terms (e.g., "async", "promises", "goroutines")
- Common use cases (e.g., "http server", "database connection")
- Related concepts (e.g., "concurrency", "parallelism")

## Development Setup

### Prerequisites

- Go 1.23 or higher
- Git

### Building

```bash
go build -o open-context
```

### Testing

```bash
# Run Go tests
go test ./...

# Test MCP server
./test.sh
```

### Project Structure

```
open-context/
├── main.go              # Entry point, CLI, and version variables
├── server/
│   ├── server.go        # MCP protocol & tool handlers
│   └── http.go          # HTTP transport
├── provider/
│   └── provider.go      # Documentation search & retrieval
├── config/
│   └── config.go        # Configuration management
├── fetcher/             # External source fetchers
│   ├── base.go          # BaseFetcher with common functionality
│   ├── go_fetcher.go    # Go packages (pkg.go.dev)
│   ├── npm_fetcher.go   # npm packages
│   ├── python_fetcher.go # PyPI packages
│   ├── rust_fetcher.go  # crates.io packages
│   ├── docker_image_fetcher.go # Docker Hub images
│   ├── github_actions_fetcher.go # GitHub Actions
│   ├── node_fetcher.go  # Node.js versions
│   ├── typescript_fetcher.go # TypeScript versions
│   ├── react_fetcher.go # React versions
│   ├── nextjs_fetcher.go # Next.js versions
│   ├── ansible_fetcher.go # Ansible versions
│   ├── terraform_fetcher.go # Terraform versions
│   ├── jenkins_fetcher.go # Jenkins versions
│   ├── kubernetes_fetcher.go # Kubernetes versions
│   └── helm_fetcher.go  # Helm versions
├── cache/
│   └── cache.go         # Cache management with TTL
├── data/                # Local documentation storage
│   └── <language>/
│       ├── metadata.json
│       └── topics/
├── .github/
│   └── workflows/
│       └── ci.yml       # CI/CD with automated releases
└── test.sh              # Integration tests
```

## Code Style

- Follow standard Go conventions
- Use `go fmt` for formatting
- Add comments for exported functions
- Keep functions focused and small
- Use meaningful variable names

## Adding New Fetchers

To add a new fetcher for external data sources:

### 1. Create the Fetcher File

Create `fetcher/<name>_fetcher.go`:

```go
package fetcher

import (
	"encoding/json"
	"fmt"
)

type YourServiceInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	// Add fields specific to your service
}

type YourServiceFetcher struct {
	*BaseFetcher
}

func NewYourServiceFetcher(cacheDir string) *YourServiceFetcher {
	return &YourServiceFetcher{
		BaseFetcher: NewBaseFetcher(cacheDir),
	}
}

func (f *YourServiceFetcher) FetchInfo(name, version string) (*YourServiceInfo, error) {
	cacheKey := fmt.Sprintf("yourservice/%s/%s", name, version)

	// Try to get from cache first
	if cachedData, err := f.getCache().Get(cacheKey); err == nil {
		var info YourServiceInfo
		if err := json.Unmarshal(cachedData, &info); err == nil {
			return &info, nil
		}
	}

	// Fetch from external API
	url := fmt.Sprintf("https://api.yourservice.com/%s/%s", name, version)
	resp, err := f.getClient().Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var info YourServiceInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(info); err == nil {
		_ = f.getCache().Set(cacheKey, data)
	}

	return &info, nil
}
```

### 2. Add to Server

In `server/server.go`:

**a) Add fetcher field to MCPServer:**
```go
type MCPServer struct {
	// ... existing fields
	yourServiceFetcher *fetcher.YourServiceFetcher
}
```

**b) Initialize in NewMCPServer():**
```go
yourServiceFetcher: fetcher.NewYourServiceFetcher(cacheDir),
```

**c) Add tool definition in handleToolsList():**
```go
{
	Name:        "get_yourservice_info",
	Description: "Fetch information about YourService packages",
	InputSchema: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Package name",
			},
			"version": map[string]interface{}{
				"type":        "string",
				"description": "Package version (optional)",
			},
		},
		"required": []string{"name"},
	},
},
```

**d) Add handler in handleToolCall():**
```go
case "get_yourservice_info":
	return s.getYourServiceInfo(args)
```

**e) Implement handler method:**
```go
func (s *MCPServer) getYourServiceInfo(args map[string]interface{}) (string, error) {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("name parameter is required")
	}

	version := ""
	if v, ok := args["version"].(string); ok {
		version = v
	}

	info, err := s.yourServiceFetcher.FetchInfo(name, version)
	if err != nil {
		return "", fmt.Errorf("failed to fetch info: %w", err)
	}

	// Format response as markdown
	response := fmt.Sprintf("# %s\n\nVersion: %s\n\n%s\n",
		info.Name, info.Version, info.Description)

	return response, nil
}
```

### 3. Add Tests

In `main_test.go`:

**a) Add to expected tools list:**
```go
expectedTools := []string{
	// ... existing tools
	"get_yourservice_info",
}
```

**b) Add test case:**
```go
{
	name:     "YourService fetcher",
	toolName: "get_yourservice_info",
	arguments: map[string]interface{}{
		"name": "example-package",
	},
	expectInContent: []string{"example-package"},
	timeout:         defaultTimeout,
},
```

### 4. Update Documentation

Add your new tool to `README.md`:
- Update the tools table
- Add tool reference documentation
- Add usage examples

### Key Patterns

- **Always use BaseFetcher**: Embed `*BaseFetcher` to get HTTP client and cache
- **Cache everything**: Use `f.getCache()` to cache API responses
- **Error handling**: Return descriptive errors with context
- **Markdown output**: Format responses as markdown for better readability
- **Optional parameters**: Handle optional parameters gracefully

## Conventional Commits

This project uses [Conventional Commits](https://www.conventionalcommits.org/) for all changes. This enables automated versioning and changelog generation.

### Commit Message Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

- **feat**: A new feature (triggers MINOR version bump)
- **fix**: A bug fix (triggers PATCH version bump)
- **docs**: Documentation only changes
- **style**: Changes that don't affect code meaning (formatting, whitespace)
- **refactor**: Code change that neither fixes a bug nor adds a feature
- **perf**: Performance improvements
- **test**: Adding or updating tests
- **build**: Changes to build system or dependencies
- **ci**: Changes to CI configuration files and scripts
- **chore**: Other changes that don't modify src or test files

### Breaking Changes

Add `!` after type/scope or `BREAKING CHANGE:` in footer for breaking changes (triggers MAJOR version bump):

```
feat!: remove support for Go 1.21

BREAKING CHANGE: minimum Go version is now 1.23
```

### Examples

**New feature:**
```
feat(fetcher): add Python package fetcher

- Fetch package info from PyPI
- Cache results with TTL
- Include installation instructions

Closes #45
```

**Bug fix:**
```
fix(cache): handle expired cache entries correctly

Previously, expired cache entries were not properly cleaned up,
leading to stale data being returned.
```

**Documentation:**
```
docs: update installation instructions for Windows

Add detailed steps for Windows users including PATH configuration.
```

**Breaking change:**
```
feat(server)!: change HTTP transport default port to 9011

BREAKING CHANGE: The default HTTP port changed from 8080 to 9011.
Update your configurations accordingly.
```

### Scopes

Common scopes in this project:
- `fetcher`: Changes to fetcher implementations
- `server`: MCP server changes
- `cache`: Cache management
- `docs`: Documentation provider
- `cli`: Command-line interface
- `ci`: CI/CD pipeline

### Guidelines

- Keep the subject line under 72 characters
- Use imperative mood ("add" not "adds" or "added")
- Don't capitalize the first letter of the description
- Don't end the subject line with a period
- Reference issues and PRs in the footer

### Automated Releases

When commits are merged to `master` branch, the CI pipeline will:
1. Run tests, linting, and format checks
2. Parse conventional commits since the last release using semantic-release
3. Determine the next version based on commit types (feat=minor, fix=patch, BREAKING=major)
4. Build cross-platform binaries for 7 platforms with version injection:
   - Linux: amd64, arm64, armv7
   - macOS: amd64, arm64
   - Windows: amd64, arm64
5. Inject version metadata via ldflags:
   - `main.Tag` - The release version (e.g., v0.1.0)
   - `main.Commit` - The full commit hash
   - `main.SourceURL` - The repository URL
   - `main.GoVersion` - The Go compiler version used
6. Create a git tag in semver format (e.g., `v0.1.0`)
7. Generate a GitHub release with changelog
8. Attach all binaries to the release
9. Update CHANGELOG.md automatically

The version information is displayed when running `./open-context --version`

## Questions?

Feel free to open an issue for any questions or clarifications needed.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.