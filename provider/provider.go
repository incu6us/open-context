package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Documentation struct {
	Name        string            `json:"name"`
	DisplayName string            `json:"displayName"`
	Description string            `json:"description"`
	Topics      map[string]*Topic `json:"topics"`
}

type Topic struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Content       string   `json:"content"`
	Keywords      []string `json:"keywords"`
	Documentation string   `json:"documentation"`
}

type SearchResult struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Documentation string  `json:"documentation"`
	Score         float64 `json:"score"`
}

type Provider struct {
	documentations map[string]*Documentation
	cacheDir       string
}

func NewProvider(cacheDir string) (*Provider, error) {
	p := &Provider{
		documentations: make(map[string]*Documentation),
		cacheDir:       cacheDir,
	}

	// Load all documentation from cache directory
	if err := p.loadDocumentation(); err != nil {
		// If loading fails, initialize with empty data
		// This allows the server to run even without documentation
		fmt.Fprintf(os.Stderr, "Warning: failed to load documentation: %v\n", err)
	}

	return p, nil
}

func (p *Provider) loadDocumentation() error {
	// Use the cache directory provided during initialization
	dataDir := p.cacheDir

	// Check if data directory exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Info: Cache directory is empty. Use 'get_go_info' tool for on-demand fetching.\n")
		return nil
	}

	// Walk through documentation directories
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		docName := entry.Name()
		docDir := filepath.Join(dataDir, docName)

		// Load Documentation metadata
		metadataPath := filepath.Join(docDir, "metadata.json")
		var documentation Documentation

		if data, err := os.ReadFile(metadataPath); err == nil {
			if err := json.Unmarshal(data, &documentation); err != nil {
				return fmt.Errorf("failed to parse metadata for %s: %w", docName, err)
			}
		} else {
			// Default metadata if file doesn't exist
			displayName := docName
			if len(docName) > 0 {
				displayName = strings.ToUpper(docName[:1]) + docName[1:]
			}
			documentation = Documentation{
				Name:        docName,
				DisplayName: displayName,
				Description: fmt.Sprintf("Documentation for %s", docName),
			}
		}

		// Ensure Topics map is initialized
		if documentation.Topics == nil {
			documentation.Topics = make(map[string]*Topic)
		}

		// Load topics
		topicsDir := filepath.Join(docDir, "topics")
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

				topic.Documentation = docName
				documentation.Topics[topic.ID] = &topic
			}
		}

		p.documentations[docName] = &documentation
	}

	// Info message if no documentations were loaded
	if len(p.documentations) == 0 {
		fmt.Fprintf(os.Stderr, "Info: No documentation loaded. Use 'get_go_info' tool for on-demand fetching.\n")
	}

	return nil
}

func (p *Provider) Search(query string, documentation string) []SearchResult {
	query = strings.ToLower(query)
	var results []SearchResult

	for docName, doc := range p.documentations {
		// Skip if documentation filter is specified and doesn't match
		if documentation != "" && docName != documentation {
			continue
		}

		for _, topic := range doc.Topics {
			score := p.calculateScore(query, topic)
			if score > 0 {
				results = append(results, SearchResult{
					ID:            topic.ID,
					Title:         topic.Title,
					Description:   topic.Description,
					Documentation: docName,
					Score:         score,
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

func (p *Provider) GetDoc(id, doc, topic string) (string, error) {
	// If ID is provided, use it directly
	if id != "" {
		for docName, documentation := range p.documentations {
			if doc != "" && docName != doc {
				continue
			}

			if foundDoc, ok := documentation.Topics[id]; ok {
				return foundDoc.Content, nil
			}
		}
		return "", fmt.Errorf("documentation not found for id: %s", id)
	}

	// If topic is provided, search by title
	if topic != "" {
		for docName, documentation := range p.documentations {
			if doc != "" && docName != doc {
				continue
			}

			for _, foundDoc := range documentation.Topics {
				if strings.EqualFold(foundDoc.Title, topic) || strings.EqualFold(foundDoc.ID, topic) {
					return foundDoc.Content, nil
				}
			}
		}
		return "", fmt.Errorf("documentation not found for topic: %s", topic)
	}

	return "", fmt.Errorf("either id or topic must be provided")
}

func (p *Provider) ListDocumentations() []Documentation {
	var documentations []Documentation

	for _, documentation := range p.documentations {
		// Create a copy without the full topic content
		docCopy := Documentation{
			Name:        documentation.Name,
			DisplayName: documentation.DisplayName,
			Description: documentation.Description,
			Topics:      make(map[string]*Topic),
		}

		// Add topic summaries
		for id, topic := range documentation.Topics {
			docCopy.Topics[id] = &Topic{
				ID:            topic.ID,
				Title:         topic.Title,
				Description:   topic.Description,
				Keywords:      topic.Keywords,
				Documentation: topic.Documentation,
			}
		}

		documentations = append(documentations, docCopy)
	}

	return documentations
}
