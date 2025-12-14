package docs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Provider struct {
	languages map[string]*Language
}

type Language struct {
	Name        string            `json:"name"`
	DisplayName string            `json:"displayName"`
	Description string            `json:"description"`
	Topics      map[string]*Topic `json:"topics"`
}

type Topic struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Keywords    []string `json:"keywords"`
	Language    string   `json:"language"`
}

type SearchResult struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Language    string  `json:"language"`
	Score       float64 `json:"score"`
}

func NewProvider() *Provider {
	p := &Provider{
		languages: make(map[string]*Language),
	}

	// Load all documentation from data directory
	if err := p.loadDocumentation(); err != nil {
		// If loading fails, initialize with empty data
		// This allows the server to run even without documentation
		fmt.Fprintf(os.Stderr, "Warning: failed to load documentation: %v\n", err)
	}

	return p
}

func (p *Provider) loadDocumentation() error {
	dataDir := filepath.Join(".", "data")

	// Check if data directory exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		// Initialize with built-in Go documentation
		p.initializeDefaultDocs()
		return nil
	}

	// Walk through language directories
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		langName := entry.Name()
		langDir := filepath.Join(dataDir, langName)

		// Load language metadata
		metadataPath := filepath.Join(langDir, "metadata.json")
		var lang Language

		if data, err := os.ReadFile(metadataPath); err == nil {
			if err := json.Unmarshal(data, &lang); err != nil {
				return fmt.Errorf("failed to parse metadata for %s: %w", langName, err)
			}
		} else {
			// Default metadata if file doesn't exist
			lang = Language{
				Name:        langName,
				DisplayName: strings.Title(langName),
				Description: fmt.Sprintf("Documentation for %s", langName),
			}
		}

		// Ensure Topics map is initialized
		if lang.Topics == nil {
			lang.Topics = make(map[string]*Topic)
		}

		// Load topics
		topicsDir := filepath.Join(langDir, "topics")
		if topicEntries, err := os.ReadDir(topicsDir); err == nil {
			for _, topicEntry := range topicEntries {
				if topicEntry.IsDir() || !strings.HasSuffix(topicEntry.Name(), ".json") {
					continue
				}

				topicPath := filepath.Join(topicsDir, topicEntry.Name())
				data, err := os.ReadFile(topicPath)
				if err != nil {
					continue
				}

				var topic Topic
				if err := json.Unmarshal(data, &topic); err != nil {
					continue
				}

				topic.Language = langName
				lang.Topics[topic.ID] = &topic
			}
		}

		p.languages[langName] = &lang
	}

	// If no languages were loaded, initialize with defaults
	if len(p.languages) == 0 {
		p.initializeDefaultDocs()
	}

	return nil
}

func (p *Provider) initializeDefaultDocs() {
	// Initialize with built-in Go documentation
	goLang := &Language{
		Name:        "go",
		DisplayName: "Go",
		Description: "Go programming language documentation",
		Topics:      make(map[string]*Topic),
	}

	// Add some basic Go topics
	goLang.Topics["basics"] = &Topic{
		ID:          "basics",
		Title:       "Go Basics",
		Description: "Introduction to Go programming language",
		Keywords:    []string{"basics", "introduction", "getting started", "hello world"},
		Language:    "go",
		Content: `# Go Basics

Go is a statically typed, compiled programming language designed at Google.

## Hello World

` + "```go" + `
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

## Key Features

- **Simple and Clean**: Go has a simple syntax that's easy to learn
- **Fast Compilation**: Go compiles quickly to machine code
- **Built-in Concurrency**: Goroutines and channels make concurrent programming easy
- **Strong Standard Library**: Comprehensive standard library for common tasks
- **Static Typing**: Catch errors at compile time
- **Garbage Collection**: Automatic memory management

## Variables

` + "```go" + `
// Declare and initialize
var name string = "John"
var age int = 30

// Short declaration
message := "Hello"
count := 42
` + "```" + `

## Functions

` + "```go" + `
func add(a, b int) int {
    return a + b
}

// Multiple return values
func swap(a, b string) (string, string) {
    return b, a
}
` + "```",
	}

	goLang.Topics["goroutines"] = &Topic{
		ID:          "goroutines",
		Title:       "Goroutines and Concurrency",
		Description: "Concurrent programming with goroutines and channels",
		Keywords:    []string{"concurrency", "goroutines", "channels", "async", "parallel"},
		Language:    "go",
		Content: `# Goroutines and Concurrency

Go makes concurrent programming simple with goroutines and channels.

## Goroutines

A goroutine is a lightweight thread managed by the Go runtime.

` + "```go" + `
package main

import (
    "fmt"
    "time"
)

func say(s string) {
    for i := 0; i < 3; i++ {
        time.Sleep(100 * time.Millisecond)
        fmt.Println(s)
    }
}

func main() {
    go say("world")  // Start a new goroutine
    say("hello")     // Run in the main goroutine
}
` + "```" + `

## Channels

Channels are typed conduits for sending and receiving values with the channel operator <-.

` + "```go" + `
package main

import "fmt"

func sum(s []int, c chan int) {
    sum := 0
    for _, v := range s {
        sum += v
    }
    c <- sum  // Send sum to channel
}

func main() {
    s := []int{7, 2, 8, -9, 4, 0}

    c := make(chan int)
    go sum(s[:len(s)/2], c)
    go sum(s[len(s)/2:], c)

    x, y := <-c, <-c  // Receive from channel
    fmt.Println(x, y, x+y)
}
` + "```" + `

## Select Statement

The select statement lets a goroutine wait on multiple communication operations.

` + "```go" + `
func fibonacci(c, quit chan int) {
    x, y := 0, 1
    for {
        select {
        case c <- x:
            x, y = y, x+y
        case <-quit:
            fmt.Println("quit")
            return
        }
    }
}
` + "```",
	}

	goLang.Topics["http"] = &Topic{
		ID:          "http",
		Title:       "HTTP Server and Client",
		Description: "Building HTTP servers and making HTTP requests",
		Keywords:    []string{"http", "server", "client", "web", "api", "rest"},
		Language:    "go",
		Content: `# HTTP Server and Client

Go's net/http package provides excellent support for HTTP.

## HTTP Server

` + "```go" + `
package main

import (
    "fmt"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
` + "```" + `

## JSON API Example

` + "```go" + `
package main

import (
    "encoding/json"
    "net/http"
)

type Response struct {
    Message string ` + "`json:\"message\"`" + `
    Status  string ` + "`json:\"status\"`" + `
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    resp := Response{
        Message: "Hello, API!",
        Status:  "success",
    }

    json.NewEncoder(w).Encode(resp)
}

func main() {
    http.HandleFunc("/api", apiHandler)
    http.ListenAndServe(":8080", nil)
}
` + "```" + `

## HTTP Client

` + "```go" + `
package main

import (
    "fmt"
    "io"
    "net/http"
)

func main() {
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }

    fmt.Println(string(body))
}
` + "```",
	}

	p.languages["go"] = goLang
}

func (p *Provider) Search(query string, language string) []SearchResult {
	query = strings.ToLower(query)
	var results []SearchResult

	for langName, lang := range p.languages {
		// Skip if language filter is specified and doesn't match
		if language != "" && langName != language {
			continue
		}

		for _, topic := range lang.Topics {
			score := p.calculateScore(query, topic)
			if score > 0 {
				results = append(results, SearchResult{
					ID:          topic.ID,
					Title:       topic.Title,
					Description: topic.Description,
					Language:    langName,
					Score:       score,
				})
			}
		}
	}

	// Sort by score (simple bubble sort for small datasets)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

func (p *Provider) calculateScore(query string, topic *Topic) float64 {
	score := 0.0

	// Check title
	if strings.Contains(strings.ToLower(topic.Title), query) {
		score += 10.0
	}

	// Check keywords
	for _, keyword := range topic.Keywords {
		if strings.Contains(strings.ToLower(keyword), query) {
			score += 5.0
		}
	}

	// Check description
	if strings.Contains(strings.ToLower(topic.Description), query) {
		score += 3.0
	}

	// Check content
	if strings.Contains(strings.ToLower(topic.Content), query) {
		score += 1.0
	}

	return score
}

func (p *Provider) GetDoc(id, language, topic string) (string, error) {
	// If ID is provided, use it directly
	if id != "" {
		for langName, lang := range p.languages {
			if language != "" && langName != language {
				continue
			}

			if doc, ok := lang.Topics[id]; ok {
				return doc.Content, nil
			}
		}
		return "", fmt.Errorf("documentation not found for id: %s", id)
	}

	// If topic is provided, search by title
	if topic != "" {
		for langName, lang := range p.languages {
			if language != "" && langName != language {
				continue
			}

			for _, doc := range lang.Topics {
				if strings.EqualFold(doc.Title, topic) || strings.EqualFold(doc.ID, topic) {
					return doc.Content, nil
				}
			}
		}
		return "", fmt.Errorf("documentation not found for topic: %s", topic)
	}

	return "", fmt.Errorf("either id or topic must be provided")
}

func (p *Provider) ListLanguages() []Language {
	var languages []Language

	for _, lang := range p.languages {
		// Create a copy without the full topic content
		langCopy := Language{
			Name:        lang.Name,
			DisplayName: lang.DisplayName,
			Description: lang.Description,
			Topics:      make(map[string]*Topic),
		}

		// Add topic summaries
		for id, topic := range lang.Topics {
			langCopy.Topics[id] = &Topic{
				ID:          topic.ID,
				Title:       topic.Title,
				Description: topic.Description,
				Keywords:    topic.Keywords,
				Language:    topic.Language,
			}
		}

		languages = append(languages, langCopy)
	}

	return languages
}