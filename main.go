package main

import (
	"log"
	"os"

	"github.com/incu6us/open-context/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	mcpServer := server.NewMCPServer()

	// Start the MCP server using stdio transport
	return mcpServer.Serve(os.Stdin, os.Stdout, os.Stderr)
}