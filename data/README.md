# Documentation Data Directory

This directory contains documentation for various programming languages and frameworks.

## Directory Structure

```
data/
├── <language>/
│   ├── metadata.json        # Language metadata
│   └── topics/              # Documentation topics
│       ├── topic1.json
│       ├── topic2.json
│       └── ...
```

## Adding a New Language

To add documentation for a new language or framework:

### 1. Create the language directory

```bash
mkdir -p data/<language>/topics
```

### 2. Create metadata.json

Create `data/<language>/metadata.json` with the following structure:

```json
{
  "name": "typescript",
  "displayName": "TypeScript",
  "description": "TypeScript programming language documentation"
}
```

### 3. Add documentation topics

Create JSON files in `data/<language>/topics/` with this structure:

```json
{
  "id": "topic-id",
  "title": "Topic Title",
  "description": "Brief description of this topic",
  "keywords": ["keyword1", "keyword2", "keyword3"],
  "content": "# Topic Title\n\nYour markdown content here...\n\n```typescript\n// code examples\n```"
}
```

### Field Descriptions

- **id**: Unique identifier for the topic (lowercase, use hyphens)
- **title**: Human-readable title
- **description**: Short description (1-2 sentences)
- **keywords**: Array of searchable keywords
- **content**: Full documentation in Markdown format

## Example: Adding TypeScript Documentation

### Create metadata

```bash
mkdir -p data/typescript/topics
cat > data/typescript/metadata.json << 'EOF'
{
  "name": "typescript",
  "displayName": "TypeScript",
  "description": "TypeScript - JavaScript with syntax for types"
}
EOF
```

### Add a topic

```bash
cat > data/typescript/topics/basics.json << 'EOF'
{
  "id": "basics",
  "title": "TypeScript Basics",
  "description": "Introduction to TypeScript type system",
  "keywords": ["basics", "types", "introduction", "getting started"],
  "content": "# TypeScript Basics\n\nTypeScript is JavaScript with syntax for types.\n\n## Type Annotations\n\n```typescript\nlet name: string = \"John\";\nlet age: number = 30;\nlet isActive: boolean = true;\n```\n\n## Interfaces\n\n```typescript\ninterface User {\n  name: string;\n  age: number;\n  email?: string;  // optional property\n}\n\nconst user: User = {\n  name: \"John\",\n  age: 30\n};\n```"
}
EOF
```

## Best Practices

1. **Use Markdown**: Write content in Markdown format for consistency
2. **Include Code Examples**: Add practical code examples with syntax highlighting
3. **Add Keywords**: Include comprehensive keywords for better search results
4. **Keep Topics Focused**: Each topic should cover a specific concept
5. **Update Regularly**: Keep documentation up-to-date with latest versions

## Built-in Documentation

If the data directory doesn't exist or is empty, the server will use built-in Go documentation as a fallback.

## Loading Documentation

The server automatically loads all documentation from this directory on startup. No restart is needed after adding new files - just reload the server.