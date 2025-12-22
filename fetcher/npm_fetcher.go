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

	content.WriteString(fmt.Sprintf("# %s\n\n", info.Name))

	if info.Description != "" {
		content.WriteString(fmt.Sprintf("**Description:** %s\n\n", info.Description))
	}

	content.WriteString(fmt.Sprintf("**Version:** %s\n\n", info.Version))

	if info.Author != "" {
		content.WriteString(fmt.Sprintf("**Author:** %s\n\n", info.Author))
	}

	if info.License != "" {
		content.WriteString(fmt.Sprintf("**License:** %s\n\n", info.License))
	}

	if info.Homepage != "" {
		content.WriteString(fmt.Sprintf("**Homepage:** %s\n\n", info.Homepage))
	}

	if info.Repository != "" {
		content.WriteString(fmt.Sprintf("**Repository:** %s\n\n", info.Repository))
	}

	content.WriteString("## Installation\n\n")
	content.WriteString("```bash\n")
	content.WriteString(fmt.Sprintf("npm install %s", info.Name))
	if info.Version != "" {
		content.WriteString(fmt.Sprintf("@%s", info.Version))
	}
	content.WriteString("\n```\n\n")

	content.WriteString("## Documentation\n\n")
	content.WriteString(fmt.Sprintf("For detailed documentation, visit [npmjs.com](https://www.npmjs.com/package/%s)", info.Name))
	if info.Version != "" {
		content.WriteString(fmt.Sprintf("/v/%s", info.Version))
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
	content.WriteString(fmt.Sprintf("name: \"%s\"\n", info.Name))
	content.WriteString(fmt.Sprintf("version: \"%s\"\n", info.Version))
	if info.Description != "" {
		content.WriteString(fmt.Sprintf("description: \"%s\"\n", escapeYAML(info.Description)))
	}
	if info.Homepage != "" {
		content.WriteString(fmt.Sprintf("homepage: \"%s\"\n", info.Homepage))
	}
	if info.Repository != "" {
		content.WriteString(fmt.Sprintf("repository: \"%s\"\n", info.Repository))
	}
	if info.License != "" {
		content.WriteString(fmt.Sprintf("license: \"%s\"\n", info.License))
	}
	if info.Author != "" {
		content.WriteString(fmt.Sprintf("author: \"%s\"\n", escapeYAML(info.Author)))
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
