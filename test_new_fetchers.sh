#!/bin/bash

echo "=== Testing New Fetchers ==="
echo ""

# Test 1: Test npm fetcher
echo "1. Testing npm fetcher (express package):"
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./open-context > /dev/null
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_npm_info","arguments":{"packageName":"express"}}}' | ./open-context 2>&1 | grep -E "(Fetching|Loaded|express)" | head -5
echo ""

# Test 2: Test Node.js fetcher
echo "2. Testing Node.js fetcher (v20.0.0):"
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./open-context > /dev/null
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_node_info","arguments":{"version":"20.0.0"}}}' | ./open-context 2>&1 | grep -E "(Fetching|Loaded|Node)" | head -5
echo ""

# Test 3: Test TypeScript fetcher
echo "3. Testing TypeScript fetcher (5.0.0):"
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./open-context > /dev/null
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_typescript_info","arguments":{"version":"5.0.0"}}}' | ./open-context 2>&1 | grep -E "(Fetching|Loaded|TypeScript)" | head -5
echo ""

# Test 4: Verify tools are listed
echo "4. Verifying all tools are listed:"
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./open-context > /dev/null
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | ./open-context 2>&1 | grep -E "(get_npm_info|get_node_info|get_typescript_info)" | head -3
echo ""

# Test 5: Check cache structure
echo "5. Checking cache structure:"
echo "NPM cache:"
ls -la ~/.open-context/cache/npm/packages/ 2>/dev/null | tail -3
echo ""
echo "Node cache:"
ls -la ~/.open-context/cache/node/versions/ 2>/dev/null | tail -3
echo ""
echo "TypeScript cache:"
ls -la ~/.open-context/cache/typescript/versions/ 2>/dev/null | tail -3
echo ""

echo "=== All tests completed ==="
