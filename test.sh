#!/bin/bash

# Test script for the MCP server
# Requires: jq (for JSON formatting)

set -e

echo "Testing Open Context MCP Server"
echo "================================"
echo ""

# Check if server binary exists
if [ ! -f "./open-context" ]; then
    echo "Error: ./open-context binary not found. Run 'make' or 'go build' first."
    exit 1
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Warning: jq not found. Output will not be formatted."
    echo "Install with: brew install jq (macOS) or apt-get install jq (Linux)"
    echo ""
    JQ_CMD="cat"
else
    JQ_CMD="jq ."
fi

# Test 1: Initialize
echo "Test 1: Initialize"
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./open-context | $JQ_CMD
echo ""

# Test 2: List tools
echo "Test 2: List tools (should show 4 tools)"
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ./open-context | $JQ_CMD
echo ""

# Test 3: Search docs
echo "Test 3: Search for 'goroutines'"
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search_docs","arguments":{"query":"goroutines"}}}' | ./open-context | $JQ_CMD
echo ""

# Test 4: Get docs
echo "Test 4: Get Go basics documentation"
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_docs","arguments":{"id":"basics","language":"go"}}}' | ./open-context | $JQ_CMD
echo ""

# Test 5: List languages
echo "Test 5: List all languages"
echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"list_languages","arguments":{}}}' | ./open-context | $JQ_CMD
echo ""

# Test 6: Get Go version info (new feature)
echo "Test 6: Get Go version info (1.21)"
echo '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"version","version":"1.21"}}}' | ./open-context | $JQ_CMD
echo ""

# Test 7: Get library info (new feature)
echo "Test 7: Get library info (github.com/spf13/cobra)"
echo '{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"library","importPath":"github.com/spf13/cobra"}}}' | ./open-context | $JQ_CMD
echo ""

# Test 8: Get library info with version (new feature)
echo "Test 8: Get library info with specific version"
echo '{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"get_go_info","arguments":{"type":"library","importPath":"github.com/spf13/cobra","version":"v1.8.0"}}}' | ./open-context | $JQ_CMD
echo ""

echo "================================"
echo "All tests completed!"
echo ""
echo "Check data/go/versions/ and data/go/libraries/ for cached results."