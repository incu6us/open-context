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

type NPMPackageInfo struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Homepage    string `yaml:"homepage"`
	Repository  string `yaml:"repository"`
	License     string `yaml:"license"`
	Author      string `yaml:"author"`
	Content     string `yaml:"-"`
}

type NPMFetcher struct {
	*BaseFetcher
}

func NewNPMFetcher(cacheDir string) *NPMFetcher {
	return &NPMFetcher{
		BaseFetcher: NewBaseFetcher(cacheDir),
	}
}

// FetchPackageInfo fetches information about an npm package
func (f *NPMFetcher) FetchPackageInfo(packageName, version string) (*NPMPackageInfo, error) {
	// Sanitize package name for file system
	safeName := strings.ReplaceAll(packageName, "/", "_")
	if version != "" {
		safeName = fmt.Sprintf("%s_%s", safeName, version)
	}

	// Check cache first
	cachedPath := f.getCache().GetFilePath("npm", "packages", fmt.Sprintf("%s.md", safeName))
	pkgInfo, err := f.loadPackageInfoFromMarkdown(cachedPath)
	if err == nil && pkgInfo != nil {
		fmt.Fprintf(os.Stderr, "Loaded npm package '%s' from cache\n", packageName)
		return pkgInfo, nil
	}

	// Fetch from npm registry
	fmt.Fprintf(os.Stderr, "Fetching npm package '%s' from registry.npmjs.org...\n", packageName)

	var url string
	if version != "" {
		url = fmt.Sprintf("https://registry.npmjs.org/%s/%s", packageName, version)
	} else {
		url = fmt.Sprintf("https://registry.npmjs.org/%s/latest", packageName)
	}

	resp, err := f.getClient().Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("npm registry returned status %d for package %s", resp.StatusCode, packageName)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse npm registry response
	var npmData map[string]interface{}
	if err := json.Unmarshal(body, &npmData); err != nil {
		return nil, fmt.Errorf("failed to parse npm data: %w", err)
	}

	// Extract package information
	pkgInfo = &NPMPackageInfo{
		Name:    getString(npmData, "name"),
		Version: getString(npmData, "version"),
	}

	// Extract description
	if desc, ok := npmData["description"].(string); ok {
		pkgInfo.Description = desc
	}

	// Extract homepage
	if homepage, ok := npmData["homepage"].(string); ok {
		pkgInfo.Homepage = homepage
	}

	// Extract repository
	if repo, ok := npmData["repository"].(map[string]interface{}); ok {
		if url, ok := repo["url"].(string); ok {
			// Clean up git+https:// prefix
			url = strings.TrimPrefix(url, "git+")
			url = strings.TrimSuffix(url, ".git")
			pkgInfo.Repository = url
		}
	}

	// Extract license
	if license, ok := npmData["license"].(string); ok {
		pkgInfo.License = license
	}

	// Extract author
	if author, ok := npmData["author"].(map[string]interface{}); ok {
		if name, ok := author["name"].(string); ok {
			pkgInfo.Author = name
		}
	} else if author, ok := npmData["author"].(string); ok {
		pkgInfo.Author = author
	}

	// Build content
	pkgInfo.Content = f.buildPackageContent(pkgInfo)

	// Cache the result
	if err := f.savePackageInfoAsMarkdown(cachedPath, pkgInfo); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache package info: %v\n", err)
	}

	return pkgInfo, nil
}

func (f *NPMFetcher) buildPackageContent(info *NPMPackageInfo) string {
	var content strings.Builder

	fmt.Fprintf(&content, "# %s\n\n", info.Name)

	if info.Description != "" {
		fmt.Fprintf(&content, "**Description:** %s\n\n", info.Description)
	}

	fmt.Fprintf(&content, "**Version:** %s\n\n", info.Version)

	if info.Author != "" {
		fmt.Fprintf(&content, "**Author:** %s\n\n", info.Author)
	}

	if info.License != "" {
		fmt.Fprintf(&content, "**License:** %s\n\n", info.License)
	}

	if info.Homepage != "" {
		fmt.Fprintf(&content, "**Homepage:** %s\n\n", info.Homepage)
	}

	if info.Repository != "" {
		fmt.Fprintf(&content, "**Repository:** %s\n\n", info.Repository)
	}

	content.WriteString("## Installation\n\n")
	content.WriteString("```bash\n")
	fmt.Fprintf(&content, "npm install %s", info.Name)
	if info.Version != "" {
		fmt.Fprintf(&content, "@%s", info.Version)
	}
	content.WriteString("\n```\n\n")

	content.WriteString("## Documentation\n\n")
	fmt.Fprintf(&content, "For detailed documentation, visit [npmjs.com](https://www.npmjs.com/package/%s)", info.Name)
	if info.Version != "" {
		fmt.Fprintf(&content, "/v/%s", info.Version)
	}
	content.WriteString("\n")

	return content.String()
}

func (f *NPMFetcher) savePackageInfoAsMarkdown(filePath string, info *NPMPackageInfo) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	fmt.Fprintf(&content, "name: \"%s\"\n", info.Name)
	fmt.Fprintf(&content, "version: \"%s\"\n", info.Version)
	if info.Description != "" {
		fmt.Fprintf(&content, "description: \"%s\"\n", escapeYAML(info.Description))
	}
	if info.Homepage != "" {
		fmt.Fprintf(&content, "homepage: \"%s\"\n", info.Homepage)
	}
	if info.Repository != "" {
		fmt.Fprintf(&content, "repository: \"%s\"\n", info.Repository)
	}
	if info.License != "" {
		fmt.Fprintf(&content, "license: \"%s\"\n", info.License)
	}
	if info.Author != "" {
		fmt.Fprintf(&content, "author: \"%s\"\n", escapeYAML(info.Author))
	}
	content.WriteString("---\n\n")

	// Markdown content
	content.WriteString(info.Content)

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

func (f *NPMFetcher) loadPackageInfoFromMarkdown(filePath string) (*NPMPackageInfo, error) {
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
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Description string `yaml:"description"`
		Homepage    string `yaml:"homepage"`
		Repository  string `yaml:"repository"`
		License     string `yaml:"license"`
		Author      string `yaml:"author"`
	}

	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return &NPMPackageInfo{
		Name:        meta.Name,
		Version:     meta.Version,
		Description: meta.Description,
		Homepage:    meta.Homepage,
		Repository:  meta.Repository,
		License:     meta.License,
		Author:      meta.Author,
		Content:     strings.TrimSpace(parts[2]),
	}, nil
}

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func escapeYAML(s string) string {
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
