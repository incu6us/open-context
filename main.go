package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/incu6us/open-context/config"
	"github.com/incu6us/open-context/server"
)

func main() {
	cmd := &cli.Command{
		Name:    "open-context",
		Usage:   "MCP server providing documentation for programming languages and frameworks",
		Version: "0.1.0",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "clear-cache",
				Aliases: []string{"cc"},
				Usage:   "Clear all cached data and exit",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Check if clear-cache flag is set
			if cmd.Bool("clear-cache") {
				return clearCache()
			}

			// Default action - run the MCP server
			return runServer()
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func runServer() error {
	mcpServer, err := server.NewMCPServer()
	if err != nil {
		return err
	}

	return mcpServer.Serve(os.Stdin, os.Stdout, os.Stderr)
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
