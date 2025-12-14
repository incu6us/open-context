package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/incu6us/open-context/docs"
	"github.com/incu6us/open-context/fetcher"
)

type MCPServer struct {
	docProvider *docs.Provider
	goFetcher   *fetcher.GoFetcher
}

func NewMCPServer() *MCPServer {
	return &MCPServer{
		docProvider: docs.NewProvider(),
		goFetcher:   fetcher.NewGoFetcher("./data"),
	}
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
	Tools map[string]interface{} `json:"tools,omitempty"`
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
				Tools: map[string]interface{}{},
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
						"description": "Filter by programming language (e.g., 'go', 'typescript')",
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
						"description": "Programming language (e.g., 'go', 'typescript')",
					},
					"topic": map[string]interface{}{
						"type":        "string",
						"description": "Topic name (alternative to ID)",
					},
				},
			},
		},
		{
			Name:        "list_languages",
			Description: "List all available programming languages and frameworks",
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
	case "list_languages":
		result, err = s.listLanguages()
	case "get_go_info":
		result, err = s.getGoInfo(params.Arguments)
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

	language := ""
	if lang, ok := args["language"].(string); ok {
		language = lang
	}

	results := s.docProvider.Search(query, language)

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (s *MCPServer) getDocs(args map[string]interface{}) (string, error) {
	var id, language, topic string

	if v, ok := args["id"].(string); ok {
		id = v
	}
	if v, ok := args["language"].(string); ok {
		language = v
	}
	if v, ok := args["topic"].(string); ok {
		topic = v
	}

	doc, err := s.docProvider.GetDoc(id, language, topic)
	if err != nil {
		return "", err
	}

	return doc, nil
}

func (s *MCPServer) listLanguages() (string, error) {
	languages := s.docProvider.ListLanguages()

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