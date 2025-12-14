package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/incu6us/open-context/fetcher"
)

func main() {
	var (
		language = flag.String("language", "go", "Language to fetch documentation for (currently only 'go' supported)")
		output   = flag.String("output", "./data", "Output directory for documentation")
	)

	flag.Parse()

	if err := run(*language, *output); err != nil {
		log.Fatal(err)
	}
}

func run(language, output string) error {
	switch language {
	case "go":
		return fetchGoDocs(output)
	default:
		return fmt.Errorf("unsupported language: %s (currently only 'go' is supported)", language)
	}
}

func fetchGoDocs(output string) error {
	fmt.Println("Fetching Go standard library documentation from pkg.go.dev...")
	fmt.Printf("Output directory: %s\n", output)
	fmt.Println()

	f := fetcher.NewGoFetcher(output)

	if err := f.FetchStdLib(); err != nil {
		return fmt.Errorf("failed to fetch Go stdlib: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Documentation fetched successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Review the generated files in", output+"/go/topics/")
	fmt.Println("  2. Restart the open-context MCP server")
	fmt.Println("  3. The new documentation will be available for search")

	return nil
}