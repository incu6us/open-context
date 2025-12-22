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
├── main.go              # Entry point & CLI
├── server/
│   ├── server.go        # MCP protocol & tool handlers
│   └── http.go          # HTTP transport
├── docs/
│   └── provider.go      # Documentation search & retrieval
├── fetcher/             # External source fetchers
│   ├── base_fetcher.go  # Common fetcher functionality
│   ├── go_fetcher.go    # Go packages (pkg.go.dev)
│   ├── npm_fetcher.go   # npm packages
│   ├── python_fetcher.go # PyPI packages
│   ├── rust_fetcher.go  # crates.io packages
│   ├── docker_fetcher.go # Docker Hub images
│   └── ...              # Other tool fetchers
├── cache/
│   └── cache.go         # Cache management
├── data/                # Local documentation storage
│   └── <language>/
│       ├── metadata.json
│       └── topics/
├── .github/
│   └── workflows/
│       └── ci.yml       # CI/CD pipeline
└── test.sh              # Integration tests
```

## Code Style

- Follow standard Go conventions
- Use `go fmt` for formatting
- Add comments for exported functions
- Keep functions focused and small
- Use meaningful variable names

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
1. Parse conventional commits since the last release
2. Determine the next version based on commit types
3. Create a git tag in semver format (e.g., `v0.1.0`)
4. Generate a GitHub release with changelog

## Questions?

Feel free to open an issue for any questions or clarifications needed.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.