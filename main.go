package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	cli "github.com/urfave/cli/v3"

	"github.com/incu6us/open-context/config"
	"github.com/incu6us/open-context/server"
)

// Project build specific vars, set during build time via ldflags
var (
	Tag       string
	Commit    string
	SourceURL string
	GoVersion string
)

func main() {
	cmd := &cli.Command{
		Name:    "open-context",
		Usage:   "MCP server providing documentation for programming languages and frameworks",
		Version: getVersion(),
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "clear-cache",
				Aliases: []string{"cc"},
				Usage:   "Clear all cached data and exit",
			},
			&cli.StringFlag{
				Name:    "transport",
				Aliases: []string{"t"},
				Usage:   "Transport mode: 'stdio' or 'http'",
				Value:   "stdio",
			},
			&cli.StringFlag{
				Name:    "host",
				Aliases: []string{"H"},
				Usage:   "Host address for HTTP transport (e.g., '0.0.0.0' or 'localhost')",
				Value:   "localhost",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "Port for HTTP transport",
				Value:   9011,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Check if clear-cache flag is set
			if cmd.Bool("clear-cache") {
				return clearCache()
			}

			// Get transport mode
			transport := cmd.String("transport")
			host := cmd.String("host")
			port := cmd.Int("port")

			// Run the MCP server with specified transport
			return runServer(transport, host, port)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func runServer(transport, host string, port int) error {
	mcpServer, err := server.NewMCPServer()
	if err != nil {
		return err
	}

	switch transport {
	case "stdio":
		log.Println("Starting MCP server with stdio transport")
		return mcpServer.Serve(os.Stdin, os.Stdout, os.Stderr)

	case "http":
		addr := fmt.Sprintf("%s:%d", host, port)
		log.Printf("Starting MCP server with HTTP transport on %s", addr)
		httpServer := server.NewHTTPServer(mcpServer)
		return httpServer.ServeHTTP(addr)

	default:
		return fmt.Errorf("invalid transport mode: %s (must be 'stdio' or 'http')", transport)
	}
}

func clearCache() error {
	// Get cache directory
	cacheDir, err := config.GetCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get cache directory: %w", err)
	}

	// Check if cache directory exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		fmt.Printf("Cache directory does not exist: %s\n", cacheDir)
		return nil
	}

	// Remove cache directory
	fmt.Printf("Removing cache directory: %s\n", cacheDir)
	if err := os.RemoveAll(cacheDir); err != nil {
		return fmt.Errorf("failed to remove cache directory: %w", err)
	}

	fmt.Println("✓ Cache cleared successfully!")
	return nil
}

// getVersion returns the version string, trimmed of "v" prefix
func getVersion() string {
	if Tag != "" {
		return strings.TrimPrefix(Tag, "v")
	}
	return "dev"
}
