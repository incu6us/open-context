# Using Open Context with Prompts

This guide explains how to use the MCP prompts feature to automatically activate documentation in your conversations with Claude.

## What are MCP Prompts?

MCP prompts are pre-configured messages that help Claude understand when and how to use the open-context documentation server. When you invoke a prompt, it tells Claude to actively use the documentation tools.

## Available Prompts

### `use-docs` - Use Documentation

Activates the open-context documentation for your conversation.

**Usage in Claude Code or Claude Desktop:**

Simply type in your conversation:
```
use open-context
```

or

```
use open context
```

Claude will recognize this and invoke the `use-docs` prompt, which makes it aware of the available documentation and tools.

## How It Works

### Step 1: Invoke the Prompt

When you type "use open-context" or similar phrases, Claude will:
1. Check available prompts from the MCP server
2. Find the `use-docs` prompt
3. Invoke it to get instructions

### Step 2: Choose Documentation (Optional)

You can specify which documentation to use:

- **"use open-context for go"** - Use Go documentation
- **"use open-context for typescript"** - Use TypeScript documentation
- **"use open-context"** - See all available documentation

### Step 3: Ask Questions

After activating the documentation, Claude will automatically use the appropriate tools to answer your questions:

**Examples:**

```
You: use open-context for go

Claude: I will help you using the **go** documentation...
        What would you like to know?

You: What's new in Go 1.25?

Claude: [Uses open-context_get_go_info tool to fetch Go 1.25 release notes]
```

```
You: use open-context

Claude: Available documentation:
        - TypeScript - JavaScript with syntax for types
        - Go - Documentation for go

        What would you like to know?

You: Tell me about TypeScript generics

Claude: [Uses open-context_search_docs and open-context_get_docs tools]
```

## What Happens When You Use the Prompt

### Without Specifying Documentation

When you invoke the prompt without specifying documentation, Claude receives:

- List of all available documentation
- Their topics and descriptions
- Information about the `open-context_get_go_info` tool for on-demand fetching
- Instructions on how to specify documentation

### With Specific Documentation (e.g., "go")

When you specify a documentation type, Claude receives:

- Confirmation of which documentation is active
- List of available tools:
  - `open-context_search_docs` - Search for topics
  - `open-context_get_docs` - Get specific content
  - `open-context_list_docs` - List all documentation
  - `open-context_get_go_info` - Fetch Go versions/libraries (for Go documentation)
- Specific examples for that documentation type
- Instructions to use these tools automatically

## Benefits

✅ **Automatic Tool Usage**: Claude knows to use the documentation tools without you asking explicitly

✅ **Context Aware**: Claude understands which documentation is available

✅ **Flexible**: You can switch between different documentation types

✅ **On-Demand Fetching**: For Go, Claude can fetch any version or library from official sources

## Example Conversations

### Example 1: Go Version Information

```
You: use open-context for go

Claude: I will help you using the go documentation. What would you like to know?

You: What's new in Go 1.25?

Claude: [Automatically uses open-context_get_go_info tool]

        Let me fetch the Go 1.25 release notes for you...

        [Returns formatted release notes]
```

### Example 2: Library Information

```
You: use open-context for go

Claude: Ready to help with Go documentation.

You: Tell me about the Gin web framework

Claude: [Automatically uses open-context_get_go_info tool]

        [Returns library information with installation instructions, docs link, etc.]
```

### Example 3: TypeScript Basics

```
You: use open-context for typescript

Claude: I will help you using the typescript documentation.

You: How do generics work?

Claude: [Automatically uses open-context_search_docs to find "generics" topic]
        [Uses open-context_get_docs to retrieve the content]

        [Returns TypeScript generics documentation]
```

## Advanced Usage

### Checking Available Prompts

You can ask Claude to list available prompts:

```
You: What prompts are available from open-context?

Claude: [Lists available prompts including use-docs]
```

### Switching Documentation Mid-Conversation

```
You: use open-context for go

Claude: Ready to help with Go documentation.

You: [asks Go questions]

You: Now switch to TypeScript

Claude: [Invokes use-docs prompt with typescript]

        Now using TypeScript documentation. What would you like to know?
```

## How Prompts Differ from Direct Tool Use

### Without Prompts (Manual)

```
You: Search for goroutines in the documentation

Claude: [Uses open-context_search_docs tool]
        [Returns search results]

You: Get the goroutines documentation

Claude: [Uses open-context_get_docs tool]
        [Returns documentation]
```

### With Prompts (Automatic)

```
You: use open-context for go

Claude: Ready to help with Go documentation.

You: Tell me about goroutines

Claude: [Automatically uses open-context_search_docs]
        [Automatically uses open-context_get_docs]
        [Provides comprehensive answer]
```

The prompt makes Claude proactive about using the tools!

## Trigger Phrases

Claude will recognize various phrases to invoke the prompt:

- "use open-context"
- "use open context"
- "use open-context for go"
- "use the open-context documentation"
- "activate open-context"

Note: The exact trigger phrases depend on Claude's understanding. The prompt name is `use-docs`.

## Technical Details

### Prompt Structure

The `use-docs` prompt has one optional argument:

- **documentation** (optional): The documentation name (e.g., "go", "typescript")

### MCP Protocol

The prompt uses the MCP prompts protocol:
- `prompts/list` - Lists available prompts
- `prompts/get` - Retrieves a specific prompt with arguments

### Response Format

The prompt returns a message that Claude sees as user input, making it aware of:
1. Available documentation and tools
2. How to use them
3. Examples of queries

## Troubleshooting

### "Prompt not found"

**Problem:** Claude says the prompt doesn't exist

**Solution:** Make sure the MCP server is properly configured and restart Claude

### "No documentation available"

**Problem:** The prompt shows no documentation

**Solution:** Use `open-context_get_go_info` tool for on-demand fetching of Go documentation

### "Claude doesn't use tools automatically"

**Problem:** Even after using the prompt, Claude doesn't use tools

**Solution:** Be more specific in your questions, or explicitly mention using the documentation

## Configuration

The MCP server must be configured in your MCP client. See the main [README.md](README.md) for configuration instructions.

### Claude Desktop

```json
{
  "mcpServers": {
    "open-context": {
      "command": "/path/to/open-context"
    }
  }
}
```

### Cursor

Add to Cursor Settings > MCP Servers

### Claude Code

```bash
claude mcp add open-context /path/to/open-context
```

## Summary

The prompts feature makes using open-context documentation natural and automatic:

1. Type "use open-context" or similar
2. Optionally specify documentation type
3. Ask your questions
4. Claude automatically uses the right tools

This provides a seamless experience where documentation is available exactly when you need it!
