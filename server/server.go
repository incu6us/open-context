package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/incu6us/open-context/config"
	"github.com/incu6us/open-context/docs"
	"github.com/incu6us/open-context/fetcher"
)

type MCPServer struct {
	docProvider       *docs.Provider
	goFetcher         *fetcher.GoFetcher
	npmFetcher        *fetcher.NPMFetcher
	nodeFetcher       *fetcher.NodeFetcher
	typescriptFetcher *fetcher.TypeScriptFetcher
	nextjsFetcher     *fetcher.NextJSFetcher
	reactFetcher      *fetcher.ReactFetcher
	ansibleFetcher    *fetcher.AnsibleFetcher
	terraformFetcher  *fetcher.TerraformFetcher
	jenkinsFetcher    *fetcher.JenkinsFetcher
	kubernetesFetcher *fetcher.KubernetesFetcher
	helmFetcher       *fetcher.HelmFetcher
}

func NewMCPServer() (*MCPServer, error) {
	// Get cache directory path and ensure it exists
	cacheDir, err := config.GetCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache directory: %w", err)
	}

	// Create doc provider with cache directory
	docProvider, err := docs.NewProvider(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize doc provider: %w", err)
	}

	return &MCPServer{
		docProvider:       docProvider,
		goFetcher:         fetcher.NewGoFetcher(cacheDir),
		npmFetcher:        fetcher.NewNPMFetcher(cacheDir),
		nodeFetcher:       fetcher.NewNodeFetcher(cacheDir),
		typescriptFetcher: fetcher.NewTypeScriptFetcher(cacheDir),
		nextjsFetcher:     fetcher.NewNextJSFetcher(cacheDir),
		reactFetcher:      fetcher.NewReactFetcher(cacheDir),
		ansibleFetcher:    fetcher.NewAnsibleFetcher(cacheDir),
		terraformFetcher:  fetcher.NewTerraformFetcher(cacheDir),
		jenkinsFetcher:    fetcher.NewJenkinsFetcher(cacheDir),
		kubernetesFetcher: fetcher.NewKubernetesFetcher(cacheDir),
		helmFetcher:       fetcher.NewHelmFetcher(cacheDir),
	}, nil
}

// MCP Protocol structures
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ToolInfo struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
}

type Capabilities struct {
	Tools   map[string]interface{} `json:"tools,omitempty"`
	Prompts map[string]interface{} `json:"prompts,omitempty"`
}

func (s *MCPServer) Serve(stdin io.Reader, stdout, stderr io.Writer) error {
	scanner := bufio.NewScanner(stdin)
	encoder := json.NewEncoder(stdout)

	for scanner.Scan() {
		line := scanner.Bytes()

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("Error parsing request: %v", err)
			continue
		}

		resp := s.handleRequest(req)
		if err := encoder.Encode(resp); err != nil {
			log.Printf("Error encoding response: %v", err)
			return err
		}
	}

	return scanner.Err()
}

func (s *MCPServer) handleRequest(req Request) Response {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolCall(req)
	case "prompts/list":
		return s.handlePromptsList(req)
	case "prompts/get":
		return s.handlePromptsGet(req)
	default:
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}
	}
}

func (s *MCPServer) handleInitialize(req Request) Response {
	return Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: Capabilities{
				Tools:   map[string]interface{}{},
				Prompts: map[string]interface{}{},
			},
			ServerInfo: ServerInfo{
				Name:    "open-context",
				Version: "0.1.0",
			},
		},
	}
}

func (s *MCPServer) handleToolsList(req Request) Response {
	tools := []ToolInfo{
		{
			Name:        "search_docs",
			Description: "Search for documentation topics across all available documentation sources",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query for documentation topics",
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "Filter by documentation name (e.g., 'go', 'typescript')",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "get_docs",
			Description: "Get detailed documentation for a specific topic or library",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Documentation topic ID (from search results)",
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "Documentation name (e.g., 'go', 'typescript')",
					},
					"topic": map[string]interface{}{
						"type":        "string",
						"description": "Topic name (alternative to ID)",
					},
				},
			},
		},
		{
			Name:        "list_docs",
			Description: "List all available documentation languages and their topics",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
		},
		{
			Name:        "get_go_info",
			Description: "Fetch and cache information about specific Go versions or Go libraries from official sources",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Type of information to fetch: 'version' for Go release info or 'library' for Go package/library info",
						"enum":        []string{"version", "library"},
					},
					"version": map[string]interface{}{
						"type":        "string",
						"description": "Go version to fetch (e.g., '1.21', '1.22') when type is 'version', or library version when type is 'library'",
					},
					"importPath": map[string]interface{}{
						"type":        "string",
						"description": "Import path of the Go library (e.g., 'github.com/gin-gonic/gin') when type is 'library'",
					},
				},
				"required": []string{"type"},
			},
		},
		{
			Name:        "get_npm_info",
			Description: "Fetch and cache information about npm packages from the npm registry",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"packageName": map[string]interface{}{
						"type":        "string",
						"description": "Name of the npm package (e.g., 'express', 'react', '@types/node')",
					},
					"version": map[string]interface{}{
						"type":        "string",
						"description": "Specific version of the package (optional, defaults to latest)",
					},
				},
				"required": []string{"packageName"},
			},
		},
		{
			Name:        "get_node_info",
			Description: "Fetch and cache information about Node.js versions from nodejs.org",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version": map[string]interface{}{
						"type":        "string",
						"description": "Node.js version to fetch (e.g., '18.17.0', 'v20.0.0')",
					},
				},
				"required": []string{"version"},
			},
		},
		{
			Name:        "get_typescript_info",
			Description: "Fetch and cache information about TypeScript versions from GitHub releases",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version": map[string]interface{}{
						"type":        "string",
						"description": "TypeScript version to fetch (e.g., '5.0.0', '4.9.5')",
					},
				},
				"required": []string{"version"},
			},
		},
		{
			Name:        "get_nextjs_info",
			Description: "Fetch and cache information about Next.js versions from GitHub releases",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version": map[string]interface{}{
						"type":        "string",
						"description": "Next.js version to fetch (e.g., '13.0.0', '14.0.0')",
					},
				},
				"required": []string{"version"},
			},
		},
		{
			Name:        "get_react_info",
			Description: "Fetch and cache information about React versions from GitHub releases",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version": map[string]interface{}{
						"type":        "string",
						"description": "React version to fetch (e.g., '18.0.0', '19.0.0')",
					},
				},
				"required": []string{"version"},
			},
		},
		{
			Name:        "get_ansible_info",
			Description: "Fetch and cache information about Ansible versions from GitHub releases",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version": map[string]interface{}{
						"type":        "string",
						"description": "Ansible version to fetch (e.g., '2.15.0', '2.16.0')",
					},
				},
				"required": []string{"version"},
			},
		},
		{
			Name:        "get_terraform_info",
			Description: "Fetch and cache information about Terraform versions from GitHub releases",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version": map[string]interface{}{
						"type":        "string",
						"description": "Terraform version to fetch (e.g., '1.5.0', '1.6.0')",
					},
				},
				"required": []string{"version"},
			},
		},
		{
			Name:        "get_jenkins_info",
			Description: "Fetch and cache information about Jenkins versions from GitHub releases",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version": map[string]interface{}{
						"type":        "string",
						"description": "Jenkins version to fetch (e.g., '2.440.3', '2.450.0')",
					},
				},
				"required": []string{"version"},
			},
		},
		{
			Name:        "get_kubernetes_info",
			Description: "Fetch and cache information about Kubernetes versions from GitHub releases",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version": map[string]interface{}{
						"type":        "string",
						"description": "Kubernetes version to fetch (e.g., '1.28.0', '1.29.0')",
					},
				},
				"required": []string{"version"},
			},
		},
		{
			Name:        "get_helm_info",
			Description: "Fetch and cache information about Helm versions from GitHub releases",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version": map[string]interface{}{
						"type":        "string",
						"description": "Helm version to fetch (e.g., '3.12.0', '3.13.0')",
					},
				},
				"required": []string{"version"},
			},
		},
	}

	return Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

func (s *MCPServer) handleToolCall(req Request) Response {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32602,
				Message: fmt.Sprintf("Invalid params: %v", err),
			},
		}
	}

	var result interface{}
	var err error

	switch params.Name {
	case "search_docs":
		result, err = s.searchDocs(params.Arguments)
	case "get_docs":
		result, err = s.getDocs(params.Arguments)
	case "list_docs":
		result, err = s.listDocs()
	case "get_go_info":
		result, err = s.getGoInfo(params.Arguments)
	case "get_npm_info":
		result, err = s.getNPMInfo(params.Arguments)
	case "get_node_info":
		result, err = s.getNodeInfo(params.Arguments)
	case "get_typescript_info":
		result, err = s.getTypeScriptInfo(params.Arguments)
	case "get_nextjs_info":
		result, err = s.getNextJSInfo(params.Arguments)
	case "get_react_info":
		result, err = s.getReactInfo(params.Arguments)
	case "get_ansible_info":
		result, err = s.getAnsibleInfo(params.Arguments)
	case "get_terraform_info":
		result, err = s.getTerraformInfo(params.Arguments)
	case "get_jenkins_info":
		result, err = s.getJenkinsInfo(params.Arguments)
	case "get_kubernetes_info":
		result, err = s.getKubernetesInfo(params.Arguments)
	case "get_helm_info":
		result, err = s.getHelmInfo(params.Arguments)
	default:
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32601,
				Message: fmt.Sprintf("Unknown tool: %s", params.Name),
			},
		}
	}

	if err != nil {
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32000,
				Message: err.Error(),
			},
		}
	}

	return Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": result,
				},
			},
		},
	}
}

func (s *MCPServer) searchDocs(args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok {
		return "", fmt.Errorf("query parameter is required")
	}

	documentation := ""
	if langArg, ok := args["language"].(string); ok {
		documentation = langArg
	}

	results := s.docProvider.Search(query, documentation)

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (s *MCPServer) getDocs(args map[string]interface{}) (string, error) {
	var id, documentation, topic string

	if v, ok := args["id"].(string); ok {
		id = v
	}
	if v, ok := args["language"].(string); ok {
		documentation = v
	}
	if v, ok := args["topic"].(string); ok {
		topic = v
	}

	doc, err := s.docProvider.GetDoc(id, documentation, topic)
	if err != nil {
		return "", err
	}

	return doc, nil
}

func (s *MCPServer) listDocs() (string, error) {
	languages := s.docProvider.ListDocumentations()

	data, err := json.MarshalIndent(languages, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (s *MCPServer) getGoInfo(args map[string]interface{}) (string, error) {
	infoType, ok := args["type"].(string)
	if !ok {
		return "", fmt.Errorf("type parameter is required (must be 'version' or 'library')")
	}

	switch infoType {
	case "version":
		version, ok := args["version"].(string)
		if !ok || version == "" {
			return "", fmt.Errorf("version parameter is required when type is 'version'")
		}

		versionInfo, err := s.goFetcher.FetchGoVersion(version)
		if err != nil {
			return "", fmt.Errorf("failed to fetch Go version info: %w", err)
		}

		return versionInfo.Content, nil

	case "library":
		importPath, ok := args["importPath"].(string)
		if !ok || importPath == "" {
			return "", fmt.Errorf("importPath parameter is required when type is 'library'")
		}

		version := ""
		if v, ok := args["version"].(string); ok {
			version = v
		}

		libInfo, err := s.goFetcher.FetchLibraryInfo(importPath, version)
		if err != nil {
			return "", fmt.Errorf("failed to fetch library info: %w", err)
		}

		return libInfo.Description, nil

	default:
		return "", fmt.Errorf("invalid type: %s (must be 'version' or 'library')", infoType)
	}
}

func (s *MCPServer) getNPMInfo(args map[string]interface{}) (string, error) {
	packageName, ok := args["packageName"].(string)
	if !ok || packageName == "" {
		return "", fmt.Errorf("packageName parameter is required")
	}

	version := ""
	if v, ok := args["version"].(string); ok {
		version = v
	}

	pkgInfo, err := s.npmFetcher.FetchPackageInfo(packageName, version)
	if err != nil {
		return "", fmt.Errorf("failed to fetch npm package info: %w", err)
	}

	return pkgInfo.Content, nil
}

func (s *MCPServer) getNodeInfo(args map[string]interface{}) (string, error) {
	version, ok := args["version"].(string)
	if !ok || version == "" {
		return "", fmt.Errorf("version parameter is required")
	}

	versionInfo, err := s.nodeFetcher.FetchNodeVersion(version)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Node.js version info: %w", err)
	}

	return versionInfo.Content, nil
}

func (s *MCPServer) getTypeScriptInfo(args map[string]interface{}) (string, error) {
	version, ok := args["version"].(string)
	if !ok || version == "" {
		return "", fmt.Errorf("version parameter is required")
	}

	versionInfo, err := s.typescriptFetcher.FetchTypeScriptVersion(version)
	if err != nil {
		return "", fmt.Errorf("failed to fetch TypeScript version info: %w", err)
	}

	return versionInfo.Content, nil
}

func (s *MCPServer) getNextJSInfo(args map[string]interface{}) (string, error) {
	version, ok := args["version"].(string)
	if !ok || version == "" {
		return "", fmt.Errorf("version parameter is required")
	}

	versionInfo, err := s.nextjsFetcher.FetchNextJSVersion(version)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Next.js version info: %w", err)
	}

	return versionInfo.Content, nil
}

func (s *MCPServer) getReactInfo(args map[string]interface{}) (string, error) {
	version, ok := args["version"].(string)
	if !ok || version == "" {
		return "", fmt.Errorf("version parameter is required")
	}

	versionInfo, err := s.reactFetcher.FetchReactVersion(version)
	if err != nil {
		return "", fmt.Errorf("failed to fetch React version info: %w", err)
	}

	return versionInfo.Content, nil
}

func (s *MCPServer) getAnsibleInfo(args map[string]interface{}) (string, error) {
	version, ok := args["version"].(string)
	if !ok || version == "" {
		return "", fmt.Errorf("version parameter is required")
	}

	versionInfo, err := s.ansibleFetcher.FetchAnsibleVersion(version)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Ansible version info: %w", err)
	}

	return versionInfo.Content, nil
}

func (s *MCPServer) getTerraformInfo(args map[string]interface{}) (string, error) {
	version, ok := args["version"].(string)
	if !ok || version == "" {
		return "", fmt.Errorf("version parameter is required")
	}

	versionInfo, err := s.terraformFetcher.FetchTerraformVersion(version)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Terraform version info: %w", err)
	}

	return versionInfo.Content, nil
}

func (s *MCPServer) getJenkinsInfo(args map[string]interface{}) (string, error) {
	version, ok := args["version"].(string)
	if !ok || version == "" {
		return "", fmt.Errorf("version parameter is required")
	}

	versionInfo, err := s.jenkinsFetcher.FetchJenkinsVersion(version)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Jenkins version info: %w", err)
	}

	return versionInfo.Content, nil
}

func (s *MCPServer) getKubernetesInfo(args map[string]interface{}) (string, error) {
	version, ok := args["version"].(string)
	if !ok || version == "" {
		return "", fmt.Errorf("version parameter is required")
	}

	versionInfo, err := s.kubernetesFetcher.FetchKubernetesVersion(version)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Kubernetes version info: %w", err)
	}

	return versionInfo.Content, nil
}

func (s *MCPServer) getHelmInfo(args map[string]interface{}) (string, error) {
	version, ok := args["version"].(string)
	if !ok || version == "" {
		return "", fmt.Errorf("version parameter is required")
	}

	versionInfo, err := s.helmFetcher.FetchHelmVersion(version)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Helm version info: %w", err)
	}

	return versionInfo.Content, nil
}

func (s *MCPServer) handlePromptsList(req Request) Response {
	prompts := []map[string]interface{}{
		{
			"name":        "use-docs",
			"description": "Use open-context documentation for the current conversation",
			"arguments": []map[string]interface{}{
				{
					"name":        "documentation",
					"description": "Documentation to use (e.g., 'go', 'typescript'). Leave empty to list available docs.",
					"required":    false,
				},
			},
		},
	}

	return Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"prompts": prompts,
		},
	}
}

func (s *MCPServer) handlePromptsGet(req Request) Response {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32602,
				Message: fmt.Sprintf("Invalid params: %v", err),
			},
		}
	}

	if params.Name != "use-docs" {
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32602,
				Message: fmt.Sprintf("Unknown prompt: %s", params.Name),
			},
		}
	}

	// Get documentation parameter if provided
	documentation := ""
	if params.Arguments != nil {
		if doc, ok := params.Arguments["documentation"].(string); ok {
			documentation = doc
		}
	}

	// Build the prompt message
	var promptText string
	if documentation == "" {
		// List available documentation
		docs := s.docProvider.ListDocumentations()
		promptText = "Available documentation:\n\n"
		for _, doc := range docs {
			promptText += fmt.Sprintf("**%s** - %s\n", doc.DisplayName, doc.Description)
			if len(doc.Topics) > 0 {
				promptText += "  Topics: "
				topicNames := []string{}
				for _, topic := range doc.Topics {
					topicNames = append(topicNames, topic.Title)
				}
				promptText += fmt.Sprintf("%s\n", topicNames)
			}
			promptText += "\n"
		}
		promptText += "\nYou can also use the `get_go_info` tool to fetch any Go version or library documentation on demand from official sources.\n\n"
		promptText += "To use specific documentation, invoke this prompt with the documentation name (e.g., 'go' or 'typescript')."
	} else {
		// Provide instructions for using specific documentation
		promptText = fmt.Sprintf("I will help you using the **%s** documentation available in the open-context server.\n\n", documentation)
		promptText += "I have access to the following tools:\n\n"
		promptText += "1. **search_docs** - Search for topics in the documentation\n"
		promptText += "2. **get_docs** - Get specific documentation content\n"
		promptText += "3. **list_docs** - List all available documentation\n"

		if documentation == "go" {
			promptText += "4. **get_go_info** - Fetch Go version release notes or library information from official sources (go.dev, pkg.go.dev)\n\n"
			promptText += "For Go-specific queries:\n"
			promptText += "- Ask about Go versions: \"What's new in Go 1.25?\"\n"
			promptText += "- Ask about Go libraries: \"Tell me about github.com/gin-gonic/gin\"\n"
			promptText += "- Ask about standard library: \"How does net/http work?\"\n"
		}

		promptText += "\nI will automatically use these tools to provide accurate, up-to-date information from the official documentation.\n\n"
		promptText += "What would you like to know?"
	}

	return Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"description": "Use open-context documentation",
			"messages": []map[string]interface{}{
				{
					"role": "user",
					"content": map[string]interface{}{
						"type": "text",
						"text": promptText,
					},
				},
			},
		},
	}
}
