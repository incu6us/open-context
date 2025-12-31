# Open Context

A high-performance MCP (Model Context Protocol) server that provides up-to-date documentation for programming languages, frameworks, and tools. Built in Go for speed and simplicity.

## What is Open Context?

Open Context fetches and caches documentation from official sources, making it instantly available to AI assistants like Claude. Instead of relying on outdated training data, get real-time access to:

- **Go**: Standard library docs, third-party packages, version release notes
- **JavaScript/TypeScript**: npm packages, Node.js, React, Next.js versions
- **Python**: PyPI packages with installation instructions and package metadata
- **Rust**: Crates.io packages with version info and documentation links
- **DevOps Tools**: Docker, Kubernetes, Helm, Terraform, Ansible, Jenkins, GitHub Actions
- **And more**: Easy to extend with any language or framework

## Key Features

- **Always Up-to-Date**: Fetches from official sources (pkg.go.dev, npm registry, PyPI, crates.io, GitHub releases, GitHub API, Docker Hub)
- **Smart Caching**: Local cache with configurable TTL (default: 7 days)
- **Fast & Lightweight**: Written in Go, starts in milliseconds
- **Two Transport Modes**: stdio for local use, HTTP for remote servers
- **MCP Native**: Seamless integration with Claude Desktop, Cursor, and Claude Code

---

## Quick Start

### 1. Install

**Using the install script (recommended):**

```bash
curl -fsSL https://raw.githubusercontent.com/incu6us/open-context/master/install.sh | bash
```

**Or build from source:**

```bash
git clone https://github.com/incu6us/open-context
cd open-context
go build -o open-context
```

**Prerequisites**: Go 1.23 or higher

### 2. Configure Your MCP Client

Choose your client and follow the setup instructions:

<details>
<summary><b>Claude Desktop</b></summary>

1. Find your config file:
   - **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

2. Add open-context:
   ```json
   {
     "mcpServers": {
       "open-context": {
         "command": "/path/to/open-context"
       }
     }
   }
   ```

3. Restart Claude Desktop

</details>

<details>
<summary><b>Cursor</b></summary>

1. Go to **Settings** > **Tools & Integrations** > **MCP Servers**

2. Add this configuration:
   ```json
   {
     "open-context": {
       "command": "/path/to/open-context"
     }
   }
   ```

3. Restart Cursor

</details>

<details>
<summary><b>Claude Code</b></summary>

**Option A: Local binary (stdio transport)**

```bash
claude-code mcp add open-context /path/to/open-context
```

**Option B: Remote server (HTTP transport)**

1. Start the server on your remote machine:
   ```bash
   ./open-context --transport http --host 0.0.0.0 --port 9011
   ```

2. Install via Claude CLI:
   ```bash
   claude mcp add --transport http open-context http://your-server.com:9011
   ```

</details>

### 3. Start Using It

In your conversation with Claude, just type:

```
use open-context for go
```

or

```
create dockerfile using latest alpine image. use open-context
```

Claude will automatically fetch and use the documentation to answer your questions!

---

## Installation Options

### Quick Install Script

The installation script automates everything:

```bash
curl -fsSL https://raw.githubusercontent.com/incu6us/open-context/master/install.sh | bash
```

### Manual Installation

```bash
# Clone repository
git clone https://github.com/incu6us/open-context
cd open-context

# Option 1: Build manually
go build -o open-context

# Option 2: Use make
make build
# Or: make setup (builds + shows next steps)
```

### Install via go install

```bash
go install github.com/incu6us/open-context@latest
```

The binary will be in `$GOPATH/bin/open-context` (usually `~/go/bin/open-context`).

---

## Configuration

Open Context creates a config file at `~/.open-context/config.yaml` on first run.

### Cache Configuration

```yaml
# How long to keep cached documentation
# Formats: "7d" (days), "24h" (hours), "30m" (minutes), "0" (never expire)
cache_ttl: 7d
```

**Example configurations:**

- **Offline work**: `cache_ttl: 0` (never expires)
- **Daily updates**: `cache_ttl: 24h`
- **Weekly updates**: `cache_ttl: 7d` (default)

### Edit Configuration

```bash
# Use your preferred editor
nano ~/.open-context/config.yaml
code ~/.open-context/config.yaml
vim ~/.open-context/config.yaml
```

Changes take effect on next server start.

---

## Usage

### Server Modes

**stdio transport (default)** - For local MCP clients:

```bash
./open-context
```

**HTTP transport** - For remote access or HTTP-based clients:

```bash
# Default (localhost:9011)
./open-context --transport http

# Custom host and port
./open-context --transport http --host 0.0.0.0 --port 3000

# Short flags
./open-context -t http -H 0.0.0.0 -p 9011
```

HTTP endpoints:
- `GET /health` - Health check
- `POST /message` - MCP JSON-RPC messages
- `GET /sse` - Server-Sent Events stream

### Cache Management

```bash
# Clear cache (full flag)
./open-context --clear-cache

# Clear cache (short alias)
./open-context --cc
```

This removes `~/.open-context/cache/`. Data will be refetched on next use.

### Other Commands

```bash
# Show help
./open-context --help

# Show version
./open-context --version
```

---

## Using with Claude

### Simple Activation

Once configured, activate documentation in your conversation:

```
use open-context for go
use open-context for typescript
use open-context
```

Claude will automatically:
1. Discover available documentation
2. Use the appropriate tools
3. Fetch information from official sources
4. Cache results locally

### Example Queries

**Go package documentation:**
```
What's new in Go 1.21?
Show me how to use github.com/gin-gonic/gin
```

**npm packages:**
```
Get the latest version of express
Show me React 18 features
```

**DevOps tools:**
```
What's in Kubernetes 1.28?
Show me Terraform 1.6 changes
Get the golang:1.23-alpine Docker image details
```

See [USING_PROMPTS.md](USING_PROMPTS.md) for more examples.

---

## Available Tools

The server provides 15 MCP tools for fetching documentation:

### Documentation Tools

| Tool | Description |
|------|-------------|
| `open-context_search_docs` | Search across all documentation |
| `open-context_get_docs` | Get specific documentation topic |
| `open-context_list_docs` | List all available documentation |

### Version & Package Fetchers

| Tool | What it Fetches | Example |
|------|-----------------|---------|
| `open-context_get_go_info` | Go versions & packages | Go 1.21, github.com/gin-gonic/gin |
| `open-context_get_npm_info` | npm packages | express, react |
| `open-context_get_python_info` | Python packages (PyPI) | requests, django, numpy |
| `open-context_get_rust_info` | Rust crates (crates.io) | serde, tokio, actix-web |
| `open-context_get_node_info` | Node.js versions | 20.0.0, 18.17.0 |
| `open-context_get_typescript_info` | TypeScript versions | 5.0.0, 4.9.5 |
| `open-context_get_react_info` | React versions | 18.0.0, 17.0.2 |
| `open-context_get_nextjs_info` | Next.js versions | 14.0.0, 13.5.0 |
| `open-context_get_ansible_info` | Ansible versions | 2.15.0 |
| `open-context_get_terraform_info` | Terraform versions | 1.6.0 |
| `open-context_get_jenkins_info` | Jenkins versions | 2.420 |
| `open-context_get_kubernetes_info` | Kubernetes versions | 1.28.0 |
| `open-context_get_helm_info` | Helm versions | 3.13.0 |
| `open-context_get_docker_image` | Docker Hub images | golang:1.23-alpine |
| `open-context_get_github_action` | GitHub Actions | actions/checkout, docker/setup-buildx-action |

**All tools automatically:**
- Fetch from official sources
- Cache results locally
- Return markdown-formatted documentation
- Include installation/usage examples

For detailed tool documentation, see the [Tools Reference](#tools-reference) below.

---

## Advanced Topics

### Adding Custom Documentation

You can add custom documentation for any language or framework.

**1. Create directory structure:**

```bash
mkdir -p data/jenkins/topics
```

**2. Create metadata file** (`data/jenkins/metadata.json`):

```json
{
  "name": "jenkins",
  "displayName": "Jenkins",
  "description": "Jenkins CI/CD automation documentation"
}
```

**3. Add documentation topics** (`data/jenkins/topics/pipeline-basics.json`):

```json
{
  "id": "pipeline-basics",
  "title": "Jenkins Pipeline Basics",
  "description": "Introduction to Jenkins declarative pipelines",
  "keywords": ["pipeline", "jenkinsfile", "ci", "cd"],
  "content": "# Jenkins Pipeline Basics\n\n[Your markdown content here]"
}
```

**4. Restart the server**

See [data/README.md](data/README.md) for complete documentation format guide.

### Development

**Run tests:**

```bash
# Quick automated test
./test.sh

# Unit tests
go test ./...

# With race detector
go test -race ./...
```

**Add a new MCP tool:**

1. Add tool definition in `server/server.go` → `handleToolsList()`
2. Implement handler in `server/server.go` → `handleToolCall()`
3. Add provider method in `docs/provider.go` (if needed)

**Project structure:**

```
open-context/
├── main.go              # Entry point & CLI
├── server/
│   ├── server.go        # MCP protocol & tool handlers
│   └── http.go          # HTTP transport
├── docs/
│   └── provider.go      # Documentation search & retrieval
├── fetcher/             # External source fetchers
│   ├── go_fetcher.go
│   ├── npm_fetcher.go
│   └── ...
├── cache/               # Cache management
└── data/                # Local documentation storage
```

---

## Tools Reference

### open-context_search_docs

Search for documentation topics across all languages.

**Parameters:**
- `query` (required): Search query
- `language` (optional): Filter by language (e.g., "go", "typescript")

**Example:**
```
Search for "goroutines" in Go documentation
```

### open-context_get_docs

Get detailed documentation for a specific topic.

**Parameters:**
- `id` (optional): Topic ID from search results
- `language` (optional): Programming language
- `topic` (optional): Topic name (alternative to ID)

**Example:**
```
Get documentation for topic "basics" in Go
```

### open-context_list_docs

List all available documentation languages and topics.

**Example:**
```
List all available documentation
```

### open-context_get_go_info

Fetch Go version information or package documentation.

**Parameters:**
- `type` (required): `"version"` or `"library"`
- `version` (conditional): Go version (e.g., "1.21") for versions, or library version
- `importPath` (conditional): Import path (e.g., "github.com/gin-gonic/gin") for libraries

**Examples:**
```
Get information about Go version 1.21
Get information about github.com/gin-gonic/gin
Get information about github.com/spf13/cobra version v1.8.0
```

**Sources:**
- Versions: go.dev release notes
- Libraries: pkg.go.dev package documentation

See [GO_VERSION_LIBRARY_FEATURE.md](GO_VERSION_LIBRARY_FEATURE.md) for details.

### open-context_get_npm_info

Fetch npm package information.

**Parameters:**
- `packageName` (required): Package name (e.g., "express", "react")

**Source:** npm registry

### open-context_get_python_info

Fetch Python package information from PyPI.

**Parameters:**
- `packageName` (required): Package name (e.g., "requests", "django", "numpy")
- `version` (optional): Specific version (defaults to latest)

**Source:** PyPI (Python Package Index)

### open-context_get_rust_info

Fetch Rust crate information from crates.io.

**Parameters:**
- `crateName` (required): Crate name (e.g., "serde", "tokio", "actix-web")
- `version` (optional): Specific version (defaults to latest)

**Source:** crates.io

### open-context_get_node_info

Fetch Node.js version information.

**Parameters:**
- `version` (required): Node.js version (e.g., "20.0.0")

**Source:** GitHub releases

### open-context_get_typescript_info

Fetch TypeScript version information.

**Parameters:**
- `version` (required): TypeScript version (e.g., "5.0.0")

**Source:** GitHub releases

### open-context_get_react_info

Fetch React version information.

**Parameters:**
- `version` (required): React version (e.g., "18.0.0")

**Source:** GitHub releases

### open-context_get_nextjs_info

Fetch Next.js version information.

**Parameters:**
- `version` (required): Next.js version (e.g., "14.0.0")

**Source:** GitHub releases

### open-context_get_ansible_info

Fetch Ansible version information.

**Parameters:**
- `version` (required): Ansible version (e.g., "2.15.0")

**Source:** GitHub releases

### open-context_get_terraform_info

Fetch Terraform version information.

**Parameters:**
- `version` (required): Terraform version (e.g., "1.6.0")

**Source:** GitHub releases

### open-context_get_jenkins_info

Fetch Jenkins version information.

**Parameters:**
- `version` (required): Jenkins version (e.g., "2.420")

**Source:** GitHub releases

### open-context_get_kubernetes_info

Fetch Kubernetes version information.

**Parameters:**
- `version` (required): Kubernetes version (e.g., "1.28.0")

**Source:** GitHub releases

### open-context_get_helm_info

Fetch Helm version information.

**Parameters:**
- `version` (required): Helm version (e.g., "3.13.0")

**Source:** GitHub releases

### open-context_get_docker_image

Fetch Docker image information from Docker Hub.

**Parameters:**
- `image` (required): Image name (e.g., "golang", "nginx", "myuser/myapp")
- `tag` (required): Image tag (e.g., "1.23-alpine", "latest")

**Example:**
```
Get Docker image golang:1.23-alpine
```

**Source:** Docker Hub API

### open-context_get_github_action

Fetch GitHub Action information from GitHub API.

**Parameters:**
- `repository` (required): GitHub repository in format "owner/repo" (e.g., "actions/checkout", "docker/setup-buildx-action")
- `version` (optional): Specific version/tag of the action (defaults to latest release)

**Example:**
```
Get GitHub Action actions/checkout
Get GitHub Action docker/setup-buildx-action with version v2.10.0
```

**Source:** GitHub API

---

## Roadmap

- [x] Go package fetching from pkg.go.dev
- [x] HTTP transport for remote servers
- [x] Version fetchers for major tools
- [x] Python packages (PyPI)
- [x] Rust crates (crates.io)
- [ ] Version-specific documentation
- [ ] Web UI for browsing docs

## Contributing

Contributions welcome! Areas where you can help:

- Add new language/framework fetchers
- Improve documentation
- Add test coverage
- Report bugs or suggest features
- Share your use cases

## Testing

See [TESTING.md](docs/TESTING.md) for comprehensive testing guide.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

Inspired by [context7](https://github.com/upstash/context7) by Upstash.

---

**Questions?** Open an issue on [GitHub](https://github.com/incu6us/open-context/issues)
