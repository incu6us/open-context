# Open Context

An MCP (Model Context Protocol) server that provides up-to-date documentation for programming languages and frameworks. Similar to context7, but built in Go with a simple, extensible architecture.

## Features

- **MCP Protocol Support**: Seamless integration with Claude, Cursor, and other MCP clients
- **Auto-Fetch from Official Sources**: Automatically download Go standard library docs from pkg.go.dev
- **On-Demand Fetching**: Use `get_go_info` tool to fetch Go versions and library docs from official sources
- **Easy to Extend**: Simple JSON-based structure for adding new languages (TypeScript, Jenkins, etc.)
- **Fast & Lightweight**: Written in Go for optimal performance
- **Intelligent Search**: Keyword-based search across all documentation with scoring
- **Simple Maintenance**: JSON-based documentation storage with local caching

## Installation

### Quick Install (Recommended)

Install directly from GitHub using the installation script:

```bash
curl -fsSL https://raw.githubusercontent.com/incu6us/open-context/master/install.sh | bash
```

Or download and run manually:

```bash
wget https://raw.githubusercontent.com/incu6us/open-context/master/install.sh
chmod +x install.sh
./install.sh
```

The installation script will:
- Check Go prerequisites (requires Go 1.23 or higher)
- Install the latest version using `go install`
- Create configuration directory (`~/.open-context/`)
- Generate default configuration file
- Verify the installation

### Prerequisites

- Go 1.23 or higher

### Manual Setup

```bash
# Clone the repository
git clone https://github.com/incu6us/open-context
cd open-context

# Build everything and fetch Go documentation
make setup

# Or build manually
go build -o open-context
```

**Note**: On first run, open-context automatically creates:
- `~/.open-context/cache/` - Cache directory for downloaded documentation
- `~/.open-context/config.yaml` - Configuration file with default settings (7-day cache TTL)

This works regardless of installation method (local build, `go install`, or package managers like Homebrew).

### On-Demand Documentation Fetching

Open Context uses the `get_go_info` tool to fetch documentation on-demand from official sources:

- **Go versions**: Fetches release notes from go.dev
- **Go libraries**: Fetches package documentation from pkg.go.dev

All fetched documentation is automatically cached locally in `~/.open-context/cache/` for fast retrieval.

## Configuration

Open Context uses a configuration file located at `~/.open-context/config.yaml`. This file is automatically created on first run with sensible defaults.

### Default Configuration

```yaml
# Cache TTL (Time To Live) - How long cached data should be kept
# Supported formats:
#   - "0" or empty: No expiration (cache never expires)
#   - "7d": 7 days (default)
#   - "24h": 24 hours
#   - "30m": 30 minutes
#   - "1w": 1 week (same as 7d)

cache_ttl: 7d
```

### Customizing Configuration

Edit `~/.open-context/config.yaml` to change settings:

```bash
# Open the config file in your editor
nano ~/.open-context/config.yaml

# Or use your preferred editor
code ~/.open-context/config.yaml
```

**Cache TTL Examples**:
- `cache_ttl: 0` - Cache never expires (useful for offline work)
- `cache_ttl: 24h` - Refresh cache daily
- `cache_ttl: 1w` - Refresh cache weekly

Changes take effect on the next run of open-context.

## Usage

### CLI Usage

#### Running the server

The server supports two transport modes:

**1. stdio transport (default)**

Uses standard input/output for MCP communication. This is the default mode used by most MCP clients like Claude Desktop and Cursor:

```bash
./open-context
# or explicitly specify stdio transport
./open-context --transport stdio
```

**2. HTTP transport**

Runs as an HTTP server with REST API and Server-Sent Events (SSE) support. This mode allows installation via `claude mcp add`:

```bash
# Start HTTP server on default port (9011)
./open-context --transport http

# Specify custom host and port
./open-context --transport http --host 0.0.0.0 --port 3000

# Short flags
./open-context -t http -H localhost -p 8081
```

HTTP endpoints:
- `GET /health` - Health check endpoint
- `POST /message` - MCP JSON-RPC messages
- `GET /sse` - Server-Sent Events stream

#### Managing cache

Clear all cached data using the `--clear-cache` flag:

```bash
# Clear cache using full flag
./open-context --clear-cache

# Or use the short alias
./open-context --cc
```

This removes the entire `~/.open-context/cache` directory. Cached data will be automatically refetched on next use.

#### Other options

```bash
# Show help
./open-context --help

# Show version
./open-context --version
```

### Configuration for MCP Clients

#### Claude Desktop

Edit your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

Add the following configuration:

```json
{
  "mcpServers": {
    "open-context": {
      "command": "/path/to/open-context"
    }
  }
}
```

#### Cursor

Go to Cursor Settings > Tools & Integrations > MCP Servers, and add:

```json
{
  "open-context": {
    "command": "/path/to/open-context"
  }
}
```

#### Claude Code

**Option 1: stdio transport (local binary)**

Add to your MCP configuration:

```bash
claude-code mcp add open-context /path/to/open-context
```

**Option 2: HTTP transport (recommended for remote servers)**

First, start the server in HTTP mode:

```bash
./open-context --transport http --host 0.0.0.0 --port 9011
```

Then install via Claude CLI:

```bash
claude mcp add --transport http open-context http://localhost:9011
```

For remote servers, use the server's public IP or domain:

```bash
claude mcp add --transport http open-context http://your-server.com:9011
```

## Quick Start with Prompts

Once configured, simply type in your conversation with Claude:

```
use open-context for go
```

Claude will automatically activate the documentation and use the tools to answer your questions! See [USING_PROMPTS.md](USING_PROMPTS.md) for details.

## Available Tools

The server provides 15 MCP tools for accessing documentation and version information:

### 1. search_docs

Search for documentation topics across all available languages.

**Parameters:**
- `query` (required): Search query for documentation topics
- `language` (optional): Filter by programming language (e.g., "go", "typescript")

**Example:**
```
Search for "goroutines" in Go documentation
```

### 2. get_docs

Get detailed documentation for a specific topic.

**Parameters:**
- `id` (optional): Documentation topic ID from search results
- `language` (optional): Programming language
- `topic` (optional): Topic name (alternative to ID)

**Example:**
```
Get documentation for topic "basics" in Go
```

### 3. list_docs

List all available documentation languages and their topics.

**Example:**
```
List all available documentation
```

### 4. get_go_info

Fetch and cache information about specific Go versions or Go libraries from official sources (go.dev and pkg.go.dev).

**Parameters:**
- `type` (required): Type of information - "version" for Go releases or "library" for Go packages
- `version` (conditional): Go version (e.g., "1.21") when type is "version", or library version when type is "library"
- `importPath` (conditional): Import path (e.g., "github.com/gin-gonic/gin") when type is "library"

**Examples:**
```
Get information about Go version 1.21
Get information about the github.com/gin-gonic/gin library
Get information about github.com/spf13/cobra version v1.8.0
```

**Features:**
- Fetches release notes from go.dev for Go versions
- Fetches package documentation from pkg.go.dev for libraries
- Automatically caches results locally for fast retrieval
- Supports version-specific queries
- Returns markdown-formatted documentation

See [GO_VERSION_LIBRARY_FEATURE.md](GO_VERSION_LIBRARY_FEATURE.md) for detailed documentation.

### 5-15. Additional Version Fetchers

The following tools fetch version information from official sources:

**5. get_npm_info** - Fetch npm package information
- `packageName` (required): npm package name (e.g., "express", "react")

**6. get_node_info** - Fetch Node.js version information
- `version` (required): Node.js version (e.g., "20.0.0", "18.17.0")

**7. get_typescript_info** - Fetch TypeScript version information
- `version` (required): TypeScript version (e.g., "5.0.0", "4.9.5")

**8. get_react_info** - Fetch React version information
- `version` (required): React version (e.g., "18.0.0", "17.0.2")

**9. get_nextjs_info** - Fetch Next.js version information
- `version` (required): Next.js version (e.g., "14.0.0", "13.5.0")

**10. get_ansible_info** - Fetch Ansible version information
- `version` (required): Ansible version (e.g., "2.15.0", "2.14.0")

**11. get_terraform_info** - Fetch Terraform version information
- `version` (required): Terraform version (e.g., "1.6.0", "1.5.7")

**12. get_jenkins_info** - Fetch Jenkins version information
- `version` (required): Jenkins version (e.g., "2.420", "2.401.3")

**13. get_kubernetes_info** - Fetch Kubernetes version information
- `version` (required): Kubernetes version (e.g., "1.28.0", "1.27.5")

**14. get_helm_info** - Fetch Helm version information
- `version` (required): Helm version (e.g., "3.13.0", "3.12.3")

**15. get_docker_image** - Fetch Docker image information from Docker Hub
- `image` (required): Docker image name (e.g., "golang", "node", "nginx", "myuser/myapp")
- `tag` (required): Docker image tag (e.g., "1.23.4-bookworm", "latest", "20-alpine")

**Features:**
- Fetches from official sources (GitHub releases, Docker Hub, npm registry)
- Shows available versions and tags
- Provides usage examples and installation instructions
- Automatically caches results locally
- Returns markdown-formatted documentation

## Available Prompts

The server provides MCP prompts for easy activation:

### use-docs

Automatically activates documentation in your conversation with Claude.

**Usage:**
```
use open-context
use open-context for go
use open-context for typescript
```

When you type this in your conversation with Claude, it will:
- Understand which documentation is available
- Automatically use the appropriate tools
- Provide context-aware answers from the documentation

See [USING_PROMPTS.md](USING_PROMPTS.md) for complete guide and examples.

## Adding New Documentation

Documentation is stored in the `data/` directory with a simple, extensible structure.

### Quick Start: Add a New Language

1. **Create the directory structure:**

```bash
mkdir -p data/jenkins/topics
```

2. **Create metadata file** (`data/jenkins/metadata.json`):

```json
{
  "name": "jenkins",
  "displayName": "Jenkins",
  "description": "Jenkins CI/CD automation documentation"
}
```

3. **Add documentation topics** (`data/jenkins/topics/pipeline-basics.json`):

```json
{
  "id": "pipeline-basics",
  "title": "Jenkins Pipeline Basics",
  "description": "Introduction to Jenkins declarative pipelines",
  "keywords": ["pipeline", "jenkinsfile", "ci", "cd", "declarative"],
  "content": "# Jenkins Pipeline Basics\n\n## Declarative Pipeline\n\n```groovy\npipeline {\n    agent any\n    \n    stages {\n        stage('Build') {\n            steps {\n                echo 'Building...'\n            }\n        }\n        stage('Test') {\n            steps {\n                echo 'Testing...'\n            }\n        }\n    }\n}\n```"
}
```

4. **Restart the server** to load new documentation

See [data/README.md](data/README.md) for detailed documentation on the file format and best practices.

## Architecture

```
open-context/
├── main.go              # Entry point
├── server/              # MCP server implementation
│   └── server.go        # Request handling and tool implementations
├── docs/                # Documentation provider
│   └── provider.go      # Search and retrieval logic
└── data/                # Documentation storage
    ├── README.md        # Documentation format guide
    ├── go/              # Go language docs
    │   └── topics/
    ├── typescript/      # TypeScript docs
    │   ├── metadata.json
    │   └── topics/
    └── ...              # Add more languages here
```

### Built-in Documentation

The server includes built-in Go documentation that loads automatically if the `data/` directory doesn't exist. This includes:

- **Go Basics**: Variables, functions, and core syntax
- **Goroutines**: Concurrent programming with goroutines and channels
- **HTTP**: Building HTTP servers and clients

You can override or extend these by creating the `data/go/` directory.

## Testing

### Quick Test

Run the automated test script:

```bash
./test.sh
```

This tests all MCP tools including the new `get_go_info` feature.

### Manual Testing

See [TESTING.md](TESTING.md) for comprehensive testing guide including:
- Manual JSON-RPC testing
- Testing with MCP clients (Claude Desktop, Cursor)
- Performance testing
- Error handling tests

### Unit Tests

```bash
go test ./...
```

## Development

### Adding a new MCP tool

1. Add the tool definition in `server.go` `handleToolsList()`
2. Implement the handler in `server.go` `handleToolCall()`
3. Add the corresponding method in `docs/provider.go` if needed

## Roadmap

- [x] Support for fetching documentation from external sources (Go stdlib from pkg.go.dev)
- [ ] Fetch all Go packages (currently ~100 most common)
- [ ] Add fetchers for more languages (Python, Rust, TypeScript, etc.)
- [ ] Version-specific documentation
- [ ] Extract and include code examples from official docs
- [ ] Incremental updates (only fetch changed documentation)
- [ ] Web interface for browsing documentation
- [ ] Parallel fetching for faster downloads

## Contributing

Contributions are welcome! Please feel free to submit pull requests with:

- New language documentation
- Bug fixes
- Feature improvements
- Documentation updates

## License

MIT License - see LICENSE file for details

## Acknowledgments

Inspired by [context7](https://github.com/upstash/context7) by Upstash.
