# Open Context

An MCP (Model Context Protocol) server that provides up-to-date documentation for programming languages and frameworks. Similar to context7, but built in Go with a simple, extensible architecture.

## Features

- **MCP Protocol Support**: Seamless integration with Claude, Cursor, and other MCP clients
- **Auto-Fetch from Official Sources**: Automatically download Go standard library docs from pkg.go.dev
- **Built-in Documentation**: Comes with curated Go and TypeScript documentation
- **Easy to Extend**: Simple JSON-based structure for adding new languages (TypeScript, Jenkins, etc.)
- **Fast & Lightweight**: Written in Go for optimal performance
- **Intelligent Search**: Keyword-based search across all documentation with scoring
- **Simple Maintenance**: JSON-based documentation storage

## Installation

### Prerequisites

- Go 1.21 or higher

### Quick Setup

```bash
# Clone the repository
git clone <repository-url>
cd open-context

# Build everything and fetch Go documentation
make setup

# Or build manually
go build -o open-context
go build -o fetch-docs ./cmd/fetch
```

### Fetch Go Standard Library Documentation

Automatically download documentation from pkg.go.dev:

```bash
# Using make
make fetch-go

# Using the script
./scripts/fetch-go-docs.sh

# Or directly
./fetch-docs -language=go -output=./data
```

This will fetch ~100 commonly used Go standard library packages and make them searchable through the MCP server.

## Usage

### Running the server

The server uses stdio transport for MCP communication:

```bash
./open-context
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

Add to your MCP configuration:

```bash
claude-code mcp add open-context /path/to/open-context
```

## Available Tools

The server provides four MCP tools:

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

### 3. list_languages

List all available programming languages and their topics.

**Example:**
```
List all available documentation languages
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