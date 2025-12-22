package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type RustCrateInfo struct {
	Name          string `yaml:"name"`
	Version       string `yaml:"version"`
	Description   string `yaml:"description"`
	Homepage      string `yaml:"homepage"`
	Repository    string `yaml:"repository"`
	Documentation string `yaml:"documentation"`
	License       string `yaml:"license"`
	Downloads     int64  `yaml:"downloads"`
	Content       string `yaml:"-"`
}

type RustFetcher struct {
	*BaseFetcher
}

func NewRustFetcher(cacheDir string) *RustFetcher {
	return &RustFetcher{
		BaseFetcher: NewBaseFetcher(cacheDir),
	}
}

// FetchCrateInfo fetches information about a Rust crate from crates.io
func (f *RustFetcher) FetchCrateInfo(crateName, version string) (*RustCrateInfo, error) {
	// Sanitize crate name for file system
	safeName := strings.ReplaceAll(crateName, "/", "_")
	if version != "" {
		safeName = fmt.Sprintf("%s_%s", safeName, version)
	}

	// Check cache first
	cachedPath := f.getCache().GetFilePath("rust", "crates", fmt.Sprintf("%s.md", safeName))
	crateInfo, err := f.loadCrateInfoFromMarkdown(cachedPath)
	if err == nil && crateInfo != nil {
		fmt.Fprintf(os.Stderr, "Loaded Rust crate '%s' from cache\n", crateName)
		return crateInfo, nil
	}

	// Fetch from crates.io
	fmt.Fprintf(os.Stderr, "Fetching Rust crate '%s' from crates.io...\n", crateName)

	var url string
	if version != "" {
		url = fmt.Sprintf("https://crates.io/api/v1/crates/%s/%s", crateName, version)
	} else {
		url = fmt.Sprintf("https://crates.io/api/v1/crates/%s", crateName)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent header as required by crates.io API
	req.Header.Set("User-Agent", "open-context-mcp-server")
	req.Header.Set("Accept", "application/json")

	resp, err := f.getClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch crate info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("rust crate %s not found", crateName)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crates.io API returned status %d for crate %s", resp.StatusCode, crateName)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse crates.io response
	var cratesData map[string]interface{}
	if err := json.Unmarshal(body, &cratesData); err != nil {
		return nil, fmt.Errorf("failed to parse crates.io data: %w", err)
	}

	// Extract crate information
	crateInfo = &RustCrateInfo{}

	// Get crate section
	if crate, ok := cratesData["crate"].(map[string]interface{}); ok {
		crateInfo.Name = getStringFromCrateMap(crate, "name")
		crateInfo.Description = getStringFromCrateMap(crate, "description")
		crateInfo.Homepage = getStringFromCrateMap(crate, "homepage")
		crateInfo.Repository = getStringFromCrateMap(crate, "repository")
		crateInfo.Documentation = getStringFromCrateMap(crate, "documentation")

		if downloads, ok := crate["downloads"].(float64); ok {
			crateInfo.Downloads = int64(downloads)
		}
	}

	// Get version section
	if versionData, ok := cratesData["version"].(map[string]interface{}); ok {
		crateInfo.Version = getStringFromCrateMap(versionData, "num")
		crateInfo.License = getStringFromCrateMap(versionData, "license")
	} else if versions, ok := cratesData["versions"].([]interface{}); ok && len(versions) > 0 {
		// If no specific version requested, get the latest
		if latestVersion, ok := versions[0].(map[string]interface{}); ok {
			crateInfo.Version = getStringFromCrateMap(latestVersion, "num")
			crateInfo.License = getStringFromCrateMap(latestVersion, "license")
		}
	}

	// Build content
	crateInfo.Content = f.buildCrateContent(crateInfo)

	// Cache the result
	if err := f.saveCrateInfoAsMarkdown(cachedPath, crateInfo); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache crate info: %v\n", err)
	}

	return crateInfo, nil
}

func (f *RustFetcher) buildCrateContent(info *RustCrateInfo) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# %s\n\n", info.Name))

	if info.Description != "" {
		content.WriteString(fmt.Sprintf("**Description:** %s\n\n", info.Description))
	}

	content.WriteString(fmt.Sprintf("**Version:** %s\n\n", info.Version))

	if info.License != "" {
		content.WriteString(fmt.Sprintf("**License:** %s\n\n", info.License))
	}

	if info.Downloads > 0 {
		content.WriteString(fmt.Sprintf("**Downloads:** %d\n\n", info.Downloads))
	}

	if info.Homepage != "" {
		content.WriteString(fmt.Sprintf("**Homepage:** %s\n\n", info.Homepage))
	}

	if info.Repository != "" {
		content.WriteString(fmt.Sprintf("**Repository:** %s\n\n", info.Repository))
	}

	if info.Documentation != "" {
		content.WriteString(fmt.Sprintf("**Documentation:** %s\n\n", info.Documentation))
	}

	content.WriteString("## Installation\n\n")
	content.WriteString("### Using Cargo\n\n")
	content.WriteString("```bash\n")
	content.WriteString(fmt.Sprintf("cargo add %s", info.Name))
	if info.Version != "" {
		content.WriteString(fmt.Sprintf("@%s", info.Version))
	}
	content.WriteString("\n```\n\n")

	content.WriteString("### Adding to Cargo.toml\n\n")
	content.WriteString("```toml\n")
	content.WriteString("[dependencies]\n")
	content.WriteString(info.Name)
	if info.Version != "" {
		content.WriteString(fmt.Sprintf(" = \"%s\"", info.Version))
	}
	content.WriteString("\n```\n\n")

	content.WriteString("## Links\n\n")
	content.WriteString(fmt.Sprintf("- [Crates.io](https://crates.io/crates/%s)\n", info.Name))
	if info.Documentation != "" {
		content.WriteString(fmt.Sprintf("- [Documentation](%s)\n", info.Documentation))
	} else {
		content.WriteString(fmt.Sprintf("- [Documentation](https://docs.rs/%s)\n", info.Name))
	}
	if info.Repository != "" {
		content.WriteString(fmt.Sprintf("- [Repository](%s)\n", info.Repository))
	}

	return content.String()
}

func (f *RustFetcher) saveCrateInfoAsMarkdown(filePath string, info *RustCrateInfo) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("name: \"%s\"\n", info.Name))
	content.WriteString(fmt.Sprintf("version: \"%s\"\n", info.Version))
	if info.Description != "" {
		content.WriteString(fmt.Sprintf("description: \"%s\"\n", escapeYAMLCrateString(info.Description)))
	}
	if info.Homepage != "" {
		content.WriteString(fmt.Sprintf("homepage: \"%s\"\n", info.Homepage))
	}
	if info.Repository != "" {
		content.WriteString(fmt.Sprintf("repository: \"%s\"\n", info.Repository))
	}
	if info.Documentation != "" {
		content.WriteString(fmt.Sprintf("documentation: \"%s\"\n", info.Documentation))
	}
	if info.License != "" {
		content.WriteString(fmt.Sprintf("license: \"%s\"\n", escapeYAMLCrateString(info.License)))
	}
	if info.Downloads > 0 {
		content.WriteString(fmt.Sprintf("downloads: %d\n", info.Downloads))
	}
	content.WriteString("---\n\n")

	// Markdown content
	content.WriteString(info.Content)

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

func (f *RustFetcher) loadCrateInfoFromMarkdown(filePath string) (*RustCrateInfo, error) {
	expired, err := f.getCache().IsExpired(filePath)
	if err != nil || expired {
		return nil, fmt.Errorf("file not found or expired")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	content := string(data)
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid markdown format: missing frontmatter")
	}

	var meta struct {
		Name          string `yaml:"name"`
		Version       string `yaml:"version"`
		Description   string `yaml:"description"`
		Homepage      string `yaml:"homepage"`
		Repository    string `yaml:"repository"`
		Documentation string `yaml:"documentation"`
		License       string `yaml:"license"`
		Downloads     int64  `yaml:"downloads"`
	}

	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return &RustCrateInfo{
		Name:          meta.Name,
		Version:       meta.Version,
		Description:   meta.Description,
		Homepage:      meta.Homepage,
		Repository:    meta.Repository,
		Documentation: meta.Documentation,
		License:       meta.License,
		Downloads:     meta.Downloads,
		Content:       strings.TrimSpace(parts[2]),
	}, nil
}

func getStringFromCrateMap(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func escapeYAMLCrateString(s string) string {
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
