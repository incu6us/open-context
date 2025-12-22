package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// HTTPServer wraps MCPServer to provide HTTP/SSE transport
type HTTPServer struct {
	mcp     *MCPServer
	clients map[string]*sseClient
	mu      sync.RWMutex
}

type sseClient struct {
	id       string
	messages chan []byte
	done     chan struct{}
}

func NewHTTPServer(mcp *MCPServer) *HTTPServer {
	return &HTTPServer{
		mcp:     mcp,
		clients: make(map[string]*sseClient),
	}
}

// ServeHTTP starts the HTTP server on the specified address
func (h *HTTPServer) ServeHTTP(addr string) error {
	mux := http.NewServeMux()

	// CORS middleware
	corsHandler := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}

	// Health check endpoint
	mux.HandleFunc("/health", corsHandler(h.handleHealth))

	// MCP message endpoint
	mux.HandleFunc("/message", corsHandler(h.handleMessage))

	// SSE endpoint for streaming responses
	mux.HandleFunc("/sse", corsHandler(h.handleSSE))

	log.Printf("Starting HTTP server on %s", addr)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return server.ListenAndServe()
}

func (h *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok","server":"open-context","version":"0.1.0"}`))
}

func (h *HTTPServer) handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request: %v", err), http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse MCP request
	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Handle the request
	resp := h.mcp.handleRequest(req)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (h *HTTPServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Create client
	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
	client := &sseClient{
		id:       clientID,
		messages: make(chan []byte, 10),
		done:     make(chan struct{}),
	}

	h.mu.Lock()
	h.clients[clientID] = client
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, clientID)
		h.mu.Unlock()
		close(client.done)
	}()

	// Send initial connection message
	if _, err := fmt.Fprintf(w, "event: connected\ndata: {\"clientId\":\"%s\"}\n\n", clientID); err != nil {
		log.Printf("Error sending connected event: %v", err)
		return
	}
	flusher.Flush()

	// Keep connection alive and send messages
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-client.done:
			return
		case msg := <-client.messages:
			if _, err := fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg); err != nil {
				log.Printf("Error sending message event: %v", err)
				return
			}
			flusher.Flush()
		case <-ticker.C:
			// Send heartbeat
			if _, err := fmt.Fprintf(w, "event: heartbeat\ndata: {\"timestamp\":%d}\n\n", time.Now().Unix()); err != nil {
				log.Printf("Error sending heartbeat event: %v", err)
				return
			}
			flusher.Flush()
		}
	}
}

// SendToClient sends a message to a specific client via SSE
func (h *HTTPServer) SendToClient(clientID string, message interface{}) error {
	h.mu.RLock()
	client, ok := h.clients[clientID]
	h.mu.RUnlock()

	if !ok {
		return fmt.Errorf("client not found: %s", clientID)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	select {
	case client.messages <- data:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending to client %s", clientID)
	}
}

// BroadcastToAll sends a message to all connected clients
func (h *HTTPServer) BroadcastToAll(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.messages <- data:
		case <-time.After(1 * time.Second):
			log.Printf("Warning: timeout sending to client %s", client.id)
		}
	}

	return nil
}
