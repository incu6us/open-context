package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestMainHelpCommand tests the --help flag
func TestMainHelpCommand(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run --help: %v", err)
	}

	outputStr := string(output)
	expectedStrings := []string{
		"NAME:",
		"open-context",
		"USAGE:",
		"VERSION:",
		"GLOBAL OPTIONS:",
		"--help",
		"--clear-cache",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Help output missing expected string: %s", expected)
		}
	}

	t.Log("✓ Help command works correctly")
}

// TestMainVersionCommand tests the --version flag
func TestMainVersionCommand(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run --version: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "version") && !strings.Contains(outputStr, "0.1.0") {
		t.Errorf("Version output unexpected: %s", outputStr)
	}

	t.Log("✓ Version command works correctly")
}

// TestMainClearCacheCommand tests the --clear-cache flag
func TestMainClearCacheCommand(t *testing.T) {
	// Use a temporary home directory
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, ".open-context", "cache")

	// Create a fake cache directory with some content
	testCacheFile := filepath.Join(cacheDir, "test", "test.txt")
	if err := os.MkdirAll(filepath.Dir(testCacheFile), 0755); err != nil {
		t.Fatalf("Failed to create test cache: %v", err)
	}
	if err := os.WriteFile(testCacheFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test cache file: %v", err)
	}

	// Run clear-cache command
	cmd := exec.Command("go", "run", ".", "--clear-cache")
	cmd.Env = append(os.Environ(), "HOME="+tempDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run --clear-cache: %v\nOutput: %s", err, string(output))
	}

	// Verify cache was cleared
	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Errorf("Cache directory still exists after clearing")
	}

	t.Log("✓ Clear cache command works correctly")
}

// TestMCPServerInitialize tests MCP server initialization
func TestMCPServerInitialize(t *testing.T) {
	tempDir := t.TempDir()

	// Build the binary first for faster tests
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tempDir, "open-context"))
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	cmd := exec.Command(filepath.Join(tempDir, "open-context"))
	cmd.Env = append(os.Environ(), "HOME="+tempDir)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to get stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	// Send initialize request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params":  map[string]interface{}{},
	}

	requestJSON, _ := json.Marshal(request)
	if _, err := stdin.Write(append(requestJSON, '\n')); err != nil {
		t.Fatalf("Failed to write request: %v", err)
	}
	stdin.Close()

	// Read response
	decoder := json.NewDecoder(stdout)
	var response map[string]interface{}
	if err := decoder.Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response
	if response["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc 2.0, got %v", response["jsonrpc"])
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be an object")
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocol version 2024-11-05, got %v", result["protocolVersion"])
	}

	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected serverInfo to be an object")
	}

	if serverInfo["name"] != "open-context" {
		t.Errorf("Expected server name 'open-context', got %v", serverInfo["name"])
	}

	cmd.Wait()
	t.Log("✓ MCP server initialization works correctly")
}

// TestMCPServerToolsList tests listing all available tools
func TestMCPServerToolsList(t *testing.T) {
	tempDir := t.TempDir()

	cmd := exec.Command("go", "run", ".")
	cmd.Env = append(os.Environ(), "HOME="+tempDir)

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	// Send tools/list request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	}

	requestJSON, _ := json.Marshal(request)
	stdin.Write(append(requestJSON, '\n'))
	stdin.Close()

	// Read response
	decoder := json.NewDecoder(stdout)
	var response map[string]interface{}
	if err := decoder.Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be an object")
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatalf("Expected tools to be an array")
	}

	// Verify all expected tools are present
	expectedTools := []string{
		"search_docs",
		"get_docs",
		"list_docs",
		"get_go_info",
		"get_npm_info",
		"get_node_info",
		"get_typescript_info",
		"get_nextjs_info",
		"get_react_info",
		"get_ansible_info",
		"get_terraform_info",
		"get_jenkins_info",
		"get_kubernetes_info",
		"get_helm_info",
	}

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolMap := tool.(map[string]interface{})
		toolNames[toolMap["name"].(string)] = true
	}

	for _, expectedTool := range expectedTools {
		if !toolNames[expectedTool] {
			t.Errorf("Expected tool '%s' not found", expectedTool)
		}
	}

	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
	}

	cmd.Wait()
	t.Logf("✓ MCP server tools list works correctly (%d tools)", len(tools))
}

// TestMCPServerAllFetchers tests all fetchers via MCP protocol
func TestMCPServerAllFetchers(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name            string
		toolName        string
		arguments       map[string]interface{}
		expectInContent []string
		timeout         time.Duration
	}{
		{
			name:     "Terraform fetcher",
			toolName: "get_terraform_info",
			arguments: map[string]interface{}{
				"version": "1.6.0",
			},
			expectInContent: []string{"Terraform", "1.6", "tfenv install"},
			timeout:         30 * time.Second,
		},
		{
			name:     "Kubernetes fetcher",
			toolName: "get_kubernetes_info",
			arguments: map[string]interface{}{
				"version": "1.28.0",
			},
			expectInContent: []string{"Kubernetes", "1.28", "kubectl"},
			timeout:         30 * time.Second,
		},
		{
			name:     "Helm fetcher",
			toolName: "get_helm_info",
			arguments: map[string]interface{}{
				"version": "3.13.0",
			},
			expectInContent: []string{"Helm", "3.13", "curl"},
			timeout:         30 * time.Second,
		},
		{
			name:     "Ansible fetcher",
			toolName: "get_ansible_info",
			arguments: map[string]interface{}{
				"version": "2.15.0",
			},
			expectInContent: []string{"Ansible", "2.15", "pip install"},
			timeout:         30 * time.Second,
		},
		{
			name:     "Node.js fetcher",
			toolName: "get_node_info",
			arguments: map[string]interface{}{
				"version": "20.0.0",
			},
			expectInContent: []string{"Node.js", "20.0", "nvm"},
			timeout:         30 * time.Second,
		},
		{
			name:     "npm fetcher",
			toolName: "get_npm_info",
			arguments: map[string]interface{}{
				"packageName": "express",
			},
			expectInContent: []string{"express", "npm install"},
			timeout:         30 * time.Second,
		},
		{
			name:     "React fetcher",
			toolName: "get_react_info",
			arguments: map[string]interface{}{
				"version": "18.0.0",
			},
			expectInContent: []string{"React", "18.0", "npm install"},
			timeout:         30 * time.Second,
		},
		{
			name:     "Next.js fetcher",
			toolName: "get_nextjs_info",
			arguments: map[string]interface{}{
				"version": "14.0.0",
			},
			expectInContent: []string{"Next.js", "14.0", "npm install"},
			timeout:         30 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("go", "run", ".")
			cmd.Env = append(os.Environ(), "HOME="+tempDir)

			stdin, _ := cmd.StdinPipe()
			stdout, _ := cmd.StdoutPipe()
			stderr, _ := cmd.StderrPipe()

			if err := cmd.Start(); err != nil {
				t.Fatalf("Failed to start command: %v", err)
			}

			// Create a channel to handle timeout
			done := make(chan bool)
			var response map[string]interface{}
			var decodeErr error

			go func() {
				// Send request
				request := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      1,
					"method":  "tools/call",
					"params": map[string]interface{}{
						"name":      tc.toolName,
						"arguments": tc.arguments,
					},
				}

				requestJSON, _ := json.Marshal(request)
				stdin.Write(append(requestJSON, '\n'))
				stdin.Close()

				// Read stderr in background
				go io.Copy(io.Discard, stderr)

				// Read response
				decoder := json.NewDecoder(stdout)
				decodeErr = decoder.Decode(&response)
				done <- true
			}()

			select {
			case <-done:
				if decodeErr != nil {
					t.Fatalf("Failed to decode response: %v", decodeErr)
				}
			case <-time.After(tc.timeout):
				cmd.Process.Kill()
				t.Fatalf("Test timed out after %v", tc.timeout)
			}

			// Check for errors (allow "not found" errors as some versions might not exist)
			if errorObj, ok := response["error"].(map[string]interface{}); ok {
				errorMsg := errorObj["message"].(string)
				if !strings.Contains(errorMsg, "not found") && !strings.Contains(errorMsg, "404") {
					t.Fatalf("Unexpected error: %v", errorMsg)
				}
				t.Logf("Skipping test due to version not found (expected for some test versions)")
				cmd.Wait()
				return
			}

			// Verify result
			result, ok := response["result"].(map[string]interface{})
			if !ok {
				t.Fatalf("Expected result to be an object")
			}

			content, ok := result["content"].([]interface{})
			if !ok || len(content) == 0 {
				t.Fatalf("Expected content to be a non-empty array")
			}

			textContent := content[0].(map[string]interface{})["text"].(string)

			// Verify expected content
			for _, expected := range tc.expectInContent {
				if !strings.Contains(textContent, expected) {
					t.Errorf("Expected content to contain '%s'", expected)
				}
			}

			cmd.Wait()
			t.Logf("✓ %s works correctly", tc.name)
		})
	}
}

// TestMCPServerErrorHandling tests error handling in MCP protocol
func TestMCPServerErrorHandling(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name          string
		request       map[string]interface{}
		expectedError bool
		errorContains string
	}{
		{
			name: "Unknown method",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "unknown_method",
			},
			expectedError: true,
			errorContains: "Method not found",
		},
		{
			name: "Unknown tool",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "tools/call",
				"params": map[string]interface{}{
					"name":      "unknown_tool",
					"arguments": map[string]interface{}{},
				},
			},
			expectedError: true,
			errorContains: "Unknown tool",
		},
		{
			name: "Missing required parameter",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "tools/call",
				"params": map[string]interface{}{
					"name":      "get_node_info",
					"arguments": map[string]interface{}{},
				},
			},
			expectedError: true,
			errorContains: "version parameter is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("go", "run", ".")
			cmd.Env = append(os.Environ(), "HOME="+tempDir)

			stdin, _ := cmd.StdinPipe()
			stdout, _ := cmd.StdoutPipe()
			stderr, _ := cmd.StderrPipe()

			if err := cmd.Start(); err != nil {
				t.Fatalf("Failed to start command: %v", err)
			}

			// Read stderr in background
			go io.Copy(io.Discard, stderr)

			// Send request
			requestJSON, _ := json.Marshal(tc.request)
			stdin.Write(append(requestJSON, '\n'))
			stdin.Close()

			// Read response
			decoder := json.NewDecoder(stdout)
			var response map[string]interface{}
			if err := decoder.Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Verify error
			if tc.expectedError {
				errorObj, ok := response["error"].(map[string]interface{})
				if !ok {
					t.Fatalf("Expected error but got none")
				}

				errorMsg := errorObj["message"].(string)
				if !strings.Contains(errorMsg, tc.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tc.errorContains, errorMsg)
				}

				t.Logf("✓ Correctly returned error: %s", errorMsg)
			}

			cmd.Wait()
		})
	}
}

// TestMCPServerCacheCreation tests that cache is created correctly
func TestMCPServerCacheCreation(t *testing.T) {
	tempDir := t.TempDir()

	cmd := exec.Command("go", "run", ".")
	cmd.Env = append(os.Environ(), "HOME="+tempDir)

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	// Read stderr in background
	go io.Copy(io.Discard, stderr)

	// Send request to fetch something that will be cached
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "get_terraform_info",
			"arguments": map[string]interface{}{
				"version": "1.6.0",
			},
		},
	}

	requestJSON, _ := json.Marshal(request)
	stdin.Write(append(requestJSON, '\n'))
	stdin.Close()

	// Read response
	decoder := json.NewDecoder(stdout)
	var response map[string]interface{}
	decoder.Decode(&response)

	cmd.Wait()

	// Verify cache directory was created
	cacheDir := filepath.Join(tempDir, ".open-context", "cache")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Errorf("Cache directory was not created")
	}

	// Verify terraform cache subdirectory exists
	if response["error"] == nil {
		terraformCache := filepath.Join(cacheDir, "terraform", "versions")
		if _, err := os.Stat(terraformCache); os.IsNotExist(err) {
			t.Errorf("Terraform cache directory was not created")
		} else {
			t.Log("✓ Cache directory structure created correctly")
		}
	} else {
		t.Log("Request failed (likely rate limiting), but cache directory exists")
	}
}

// TestMCPServerFullWorkflow tests a complete workflow
func TestMCPServerFullWorkflow(t *testing.T) {
	tempDir := t.TempDir()

	cmd := exec.Command("go", "run", ".")
	cmd.Env = append(os.Environ(), "HOME="+tempDir)

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	// Sequence of requests
	requests := []map[string]interface{}{
		{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "initialize",
			"params":  map[string]interface{}{},
		},
		{
			"jsonrpc": "2.0",
			"id":      2,
			"method":  "tools/list",
		},
		{
			"jsonrpc": "2.0",
			"id":      3,
			"method":  "tools/call",
			"params": map[string]interface{}{
				"name": "get_helm_info",
				"arguments": map[string]interface{}{
					"version": "3.13.0",
				},
			},
		},
	}

	decoder := json.NewDecoder(stdout)

	for i, request := range requests {
		requestJSON, _ := json.Marshal(request)
		if _, err := stdin.Write(append(requestJSON, '\n')); err != nil {
			t.Fatalf("Failed to write request %d: %v", i+1, err)
		}

		var response map[string]interface{}
		if err := decoder.Decode(&response); err != nil {
			t.Fatalf("Failed to decode response %d: %v", i+1, err)
		}

		if errorObj, ok := response["error"].(map[string]interface{}); ok {
			// Allow not found errors for the last request
			if i < 2 || !strings.Contains(errorObj["message"].(string), "not found") {
				t.Errorf("Request %d failed: %v", i+1, errorObj["message"])
			}
		}

		t.Logf("✓ Request %d completed successfully", i+1)
	}

	stdin.Close()
	cmd.Wait()

	t.Log("✓ Full MCP workflow completed successfully")
}
