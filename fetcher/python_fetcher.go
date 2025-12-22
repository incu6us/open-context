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

type PythonPackageInfo struct {
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	Summary    string `yaml:"summary"`
	Homepage   string `yaml:"homepage"`
	Repository string `yaml:"repository"`
	License    string `yaml:"license"`
	Author     string `yaml:"author"`
	Content    string `yaml:"-"`
}

type PythonFetcher struct {
	*BaseFetcher
}

func NewPythonFetcher(cacheDir string) *PythonFetcher {
	return &PythonFetcher{
		BaseFetcher: NewBaseFetcher(cacheDir),
	}
}

// FetchPackageInfo fetches information about a Python package from PyPI
func (f *PythonFetcher) FetchPackageInfo(packageName, version string) (*PythonPackageInfo, error) {
	// Sanitize package name for file system
	safeName := strings.ReplaceAll(packageName, "/", "_")
	if version != "" {
		safeName = fmt.Sprintf("%s_%s", safeName, version)
	}

	// Check cache first
	cachedPath := f.getCache().GetFilePath("python", "packages", fmt.Sprintf("%s.md", safeName))
	pkgInfo, err := f.loadPackageInfoFromMarkdown(cachedPath)
	if err == nil && pkgInfo != nil {
		fmt.Fprintf(os.Stderr, "Loaded Python package '%s' from cache\n", packageName)
		return pkgInfo, nil
	}

	// Fetch from PyPI
	fmt.Fprintf(os.Stderr, "Fetching Python package '%s' from pypi.org...\n", packageName)

	var url string
	if version != "" {
		url = fmt.Sprintf("https://pypi.org/pypi/%s/%s/json", packageName, version)
	} else {
		url = fmt.Sprintf("https://pypi.org/pypi/%s/json", packageName)
	}

	resp, err := f.getClient().Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("python package %s not found", packageName)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PyPI API returned status %d for package %s", resp.StatusCode, packageName)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse PyPI response
	var pypiData map[string]interface{}
	if err := json.Unmarshal(body, &pypiData); err != nil {
		return nil, fmt.Errorf("failed to parse PyPI data: %w", err)
	}

	// Extract package information
	pkgInfo = &PythonPackageInfo{}

	// Get info section
	if info, ok := pypiData["info"].(map[string]interface{}); ok {
		pkgInfo.Name = getStringFromMap(info, "name")
		pkgInfo.Version = getStringFromMap(info, "version")
		pkgInfo.Summary = getStringFromMap(info, "summary")
		pkgInfo.Homepage = getStringFromMap(info, "home_page")
		pkgInfo.License = getStringFromMap(info, "license")
		pkgInfo.Author = getStringFromMap(info, "author")

		// Try to get repository from project_urls
		if projectURLs, ok := info["project_urls"].(map[string]interface{}); ok {
			// Try common repository keys
			for _, key := range []string{"Repository", "Source", "Source Code", "GitHub", "GitLab"} {
				if repoURL, ok := projectURLs[key].(string); ok && repoURL != "" {
					pkgInfo.Repository = repoURL
					break
				}
			}
		}
	}

	// Build content
	pkgInfo.Content = f.buildPackageContent(pkgInfo)

	// Cache the result
	if err := f.savePackageInfoAsMarkdown(cachedPath, pkgInfo); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache package info: %v\n", err)
	}

	return pkgInfo, nil
}

func (f *PythonFetcher) buildPackageContent(info *PythonPackageInfo) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# %s\n\n", info.Name))

	if info.Summary != "" {
		content.WriteString(fmt.Sprintf("**Summary:** %s\n\n", info.Summary))
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
	content.WriteString(fmt.Sprintf("pip install %s", info.Name))
	if info.Version != "" {
		content.WriteString(fmt.Sprintf("==%s", info.Version))
	}
	content.WriteString("\n```\n\n")

	content.WriteString("### Using requirements.txt\n\n")
	content.WriteString("```\n")
	content.WriteString(info.Name)
	if info.Version != "" {
		content.WriteString(fmt.Sprintf("==%s", info.Version))
	}
	content.WriteString("\n```\n\n")

	content.WriteString("### Using Poetry\n\n")
	content.WriteString("```bash\n")
	content.WriteString(fmt.Sprintf("poetry add %s", info.Name))
	if info.Version != "" {
		content.WriteString(fmt.Sprintf("@%s", info.Version))
	}
	content.WriteString("\n```\n\n")

	content.WriteString("## Documentation\n\n")
	content.WriteString(fmt.Sprintf("For detailed documentation, visit [PyPI](https://pypi.org/project/%s/)\n", info.Name))

	return content.String()
}

func (f *PythonFetcher) savePackageInfoAsMarkdown(filePath string, info *PythonPackageInfo) error {
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
	if info.Summary != "" {
		content.WriteString(fmt.Sprintf("summary: \"%s\"\n", escapeYAMLString(info.Summary)))
	}
	if info.Homepage != "" {
		content.WriteString(fmt.Sprintf("homepage: \"%s\"\n", info.Homepage))
	}
	if info.Repository != "" {
		content.WriteString(fmt.Sprintf("repository: \"%s\"\n", info.Repository))
	}
	if info.License != "" {
		content.WriteString(fmt.Sprintf("license: \"%s\"\n", escapeYAMLString(info.License)))
	}
	if info.Author != "" {
		content.WriteString(fmt.Sprintf("author: \"%s\"\n", escapeYAMLString(info.Author)))
	}
	content.WriteString("---\n\n")

	// Markdown content
	content.WriteString(info.Content)

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

func (f *PythonFetcher) loadPackageInfoFromMarkdown(filePath string) (*PythonPackageInfo, error) {
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
		Name       string `yaml:"name"`
		Version    string `yaml:"version"`
		Summary    string `yaml:"summary"`
		Homepage   string `yaml:"homepage"`
		Repository string `yaml:"repository"`
		License    string `yaml:"license"`
		Author     string `yaml:"author"`
	}

	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return &PythonPackageInfo{
		Name:       meta.Name,
		Version:    meta.Version,
		Summary:    meta.Summary,
		Homepage:   meta.Homepage,
		Repository: meta.Repository,
		License:    meta.License,
		Author:     meta.Author,
		Content:    strings.TrimSpace(parts[2]),
	}, nil
}

func getStringFromMap(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func escapeYAMLString(s string) string {
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
