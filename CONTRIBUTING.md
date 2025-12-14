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

- Go 1.21 or higher
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
├── main.go              # Entry point
├── server/              # MCP server implementation
│   └── server.go
├── docs/                # Documentation provider
│   └── provider.go
├── data/                # Documentation content
│   └── <language>/
│       ├── metadata.json
│       └── topics/
└── test.sh             # Integration tests
```

## Code Style

- Follow standard Go conventions
- Use `go fmt` for formatting
- Add comments for exported functions
- Keep functions focused and small
- Use meaningful variable names

## Commit Messages

- Use present tense ("Add feature" not "Added feature")
- Keep first line under 50 characters
- Add detailed description if needed
- Reference issues: "Fixes #123"

Example:
```
Add TypeScript generics documentation

- Created topics/generics.json with comprehensive examples
- Added keywords for better searchability
- Included utility types and constraints

Fixes #42
```

## Questions?

Feel free to open an issue for any questions or clarifications needed.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.