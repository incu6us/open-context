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

type GitHubActionInfo struct {
	Repository  string `yaml:"repository"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
	Version     string `yaml:"version"`
	Stars       int    `yaml:"stars"`
	License     string `yaml:"license"`
	Homepage    string `yaml:"homepage"`
	Content     string `yaml:"-"`
}

type GitHubActionsFetcher struct {
	*BaseFetcher
}

func NewGitHubActionsFetcher(cacheDir string) *GitHubActionsFetcher {
	return &GitHubActionsFetcher{
		BaseFetcher: NewBaseFetcher(cacheDir),
	}
}

// FetchActionInfo fetches information about a GitHub Action
// repository should be in format "owner/repo" (e.g., "actions/checkout")
func (f *GitHubActionsFetcher) FetchActionInfo(repository, version string) (*GitHubActionInfo, error) {
	// Sanitize repository name for file system
	safeName := strings.ReplaceAll(repository, "/", "_")
	if version != "" {
		safeName = fmt.Sprintf("%s_%s", safeName, version)
	}

	// Check cache first
	cachedPath := f.getCache().GetFilePath("github-actions", "actions", fmt.Sprintf("%s.md", safeName))
	actionInfo, err := f.loadActionInfoFromMarkdown(cachedPath)
	if err == nil && actionInfo != nil {
		fmt.Fprintf(os.Stderr, "Loaded GitHub Action '%s' from cache\n", repository)
		return actionInfo, nil
	}

	// Fetch from GitHub API
	fmt.Fprintf(os.Stderr, "Fetching GitHub Action '%s' from GitHub API...\n", repository)

	// Fetch repository information
	repoURL := fmt.Sprintf("https://api.github.com/repos/%s", repository)
	req, err := http.NewRequest("GET", repoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers as recommended by GitHub API
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "open-context-mcp-server")

	resp, err := f.getClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch action info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("github action repository %s not found", repository)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned status %d for repository %s", resp.StatusCode, repository)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse GitHub API response
	var repoData map[string]interface{}
	if err := json.Unmarshal(body, &repoData); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub API data: %w", err)
	}

	// Extract repository information
	actionInfo = &GitHubActionInfo{
		Repository:  repository,
		Name:        getStringFromActionMap(repoData, "name"),
		Description: getStringFromActionMap(repoData, "description"),
		Homepage:    getStringFromActionMap(repoData, "html_url"),
	}

	if stars, ok := repoData["stargazers_count"].(float64); ok {
		actionInfo.Stars = int(stars)
	}

	if license, ok := repoData["license"].(map[string]interface{}); ok {
		actionInfo.License = getStringFromActionMap(license, "spdx_id")
	}

	if owner, ok := repoData["owner"].(map[string]interface{}); ok {
		actionInfo.Author = getStringFromActionMap(owner, "login")
	}

	// Fetch action.yml or action.yaml to get action metadata
	actionYml := f.fetchActionYaml(repository, version)
	if actionYml != nil {
		if name, ok := actionYml["name"].(string); ok && name != "" {
			actionInfo.Name = name
		}
		if desc, ok := actionYml["description"].(string); ok && desc != "" {
			actionInfo.Description = desc
		}
	}

	// If version is not specified, try to get the latest release
	if version == "" {
		latestRelease := f.fetchLatestRelease(repository)
		if latestRelease != "" {
			actionInfo.Version = latestRelease
		} else {
			actionInfo.Version = "latest"
		}
	} else {
		actionInfo.Version = version
	}

	// Build content
	actionInfo.Content = f.buildActionContent(actionInfo, actionYml)

	// Cache the result
	if err := f.saveActionInfoAsMarkdown(cachedPath, actionInfo); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache action info: %v\n", err)
	}

	return actionInfo, nil
}

func (f *GitHubActionsFetcher) fetchActionYaml(repository, version string) map[string]interface{} {
	// Try action.yml first, then action.yaml
	ref := "main"
	if version != "" {
		ref = version
	}

	for _, filename := range []string{"action.yml", "action.yaml"} {
		url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", repository, ref, filename)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}

		req.Header.Set("User-Agent", "open-context-mcp-server")
		resp, err := f.getClient().Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if err != nil {
				continue
			}

			var actionYml map[string]interface{}
			if err := yaml.Unmarshal(body, &actionYml); err == nil {
				return actionYml
			}
		} else {
			_ = resp.Body.Close()
		}
	}

	return nil
}

func (f *GitHubActionsFetcher) fetchLatestRelease(repository string) string {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repository)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "open-context-mcp-server")

	resp, err := f.getClient().Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var release map[string]interface{}
	if err := json.Unmarshal(body, &release); err != nil {
		return ""
	}

	return getStringFromActionMap(release, "tag_name")
}

func (f *GitHubActionsFetcher) buildActionContent(info *GitHubActionInfo, actionYml map[string]interface{}) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# %s\n\n", info.Name))

	if info.Description != "" {
		content.WriteString(fmt.Sprintf("**Description:** %s\n\n", info.Description))
	}

	content.WriteString(fmt.Sprintf("**Repository:** %s\n\n", info.Repository))
	content.WriteString(fmt.Sprintf("**Version:** %s\n\n", info.Version))

	if info.Author != "" {
		content.WriteString(fmt.Sprintf("**Author:** %s\n\n", info.Author))
	}

	if info.License != "" {
		content.WriteString(fmt.Sprintf("**License:** %s\n\n", info.License))
	}

	if info.Stars > 0 {
		content.WriteString(fmt.Sprintf("**Stars:** %d\n\n", info.Stars))
	}

	if info.Homepage != "" {
		content.WriteString(fmt.Sprintf("**Homepage:** %s\n\n", info.Homepage))
	}

	// Add inputs if available
	if actionYml != nil {
		if inputs, ok := actionYml["inputs"].(map[string]interface{}); ok && len(inputs) > 0 {
			content.WriteString("## Inputs\n\n")
			for inputName, inputDef := range inputs {
				content.WriteString(fmt.Sprintf("### `%s`\n\n", inputName))
				if inputMap, ok := inputDef.(map[string]interface{}); ok {
					if desc, ok := inputMap["description"].(string); ok {
						content.WriteString(fmt.Sprintf("%s\n\n", desc))
					}
					if required, ok := inputMap["required"].(bool); ok && required {
						content.WriteString("**Required:** Yes\n\n")
					}
					if defaultVal, ok := inputMap["default"]; ok {
						content.WriteString(fmt.Sprintf("**Default:** `%v`\n\n", defaultVal))
					}
				}
			}
		}

		// Add outputs if available
		if outputs, ok := actionYml["outputs"].(map[string]interface{}); ok && len(outputs) > 0 {
			content.WriteString("## Outputs\n\n")
			for outputName, outputDef := range outputs {
				content.WriteString(fmt.Sprintf("### `%s`\n\n", outputName))
				if outputMap, ok := outputDef.(map[string]interface{}); ok {
					if desc, ok := outputMap["description"].(string); ok {
						content.WriteString(fmt.Sprintf("%s\n\n", desc))
					}
				}
			}
		}
	}

	content.WriteString("## Usage Example\n\n")
	content.WriteString("```yaml\n")
	content.WriteString("name: Example Workflow\n\n")
	content.WriteString("on: [push]\n\n")
	content.WriteString("jobs:\n")
	content.WriteString("  example:\n")
	content.WriteString("    runs-on: ubuntu-latest\n")
	content.WriteString("    steps:\n")
	content.WriteString(fmt.Sprintf("      - uses: %s@%s\n", info.Repository, info.Version))

	// Add example inputs if available
	if actionYml != nil {
		if inputs, ok := actionYml["inputs"].(map[string]interface{}); ok && len(inputs) > 0 {
			content.WriteString("        with:\n")
			count := 0
			for inputName, inputDef := range inputs {
				if count >= 2 { // Show max 2 example inputs
					break
				}
				if inputMap, ok := inputDef.(map[string]interface{}); ok {
					if defaultVal, ok := inputMap["default"]; ok {
						content.WriteString(fmt.Sprintf("          %s: %v\n", inputName, defaultVal))
						count++
					}
				}
			}
		}
	}
	content.WriteString("```\n\n")

	content.WriteString("## Links\n\n")
	content.WriteString(fmt.Sprintf("- [GitHub Repository](https://github.com/%s)\n", info.Repository))
	content.WriteString(fmt.Sprintf("- [GitHub Marketplace](https://github.com/marketplace/actions/%s)\n", strings.ReplaceAll(info.Repository, "/", "-")))

	return content.String()
}

func (f *GitHubActionsFetcher) saveActionInfoAsMarkdown(filePath string, info *GitHubActionInfo) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("repository: \"%s\"\n", info.Repository))
	content.WriteString(fmt.Sprintf("name: \"%s\"\n", escapeYAMLActionString(info.Name)))
	if info.Description != "" {
		content.WriteString(fmt.Sprintf("description: \"%s\"\n", escapeYAMLActionString(info.Description)))
	}
	if info.Author != "" {
		content.WriteString(fmt.Sprintf("author: \"%s\"\n", info.Author))
	}
	if info.Version != "" {
		content.WriteString(fmt.Sprintf("version: \"%s\"\n", info.Version))
	}
	if info.Stars > 0 {
		content.WriteString(fmt.Sprintf("stars: %d\n", info.Stars))
	}
	if info.License != "" {
		content.WriteString(fmt.Sprintf("license: \"%s\"\n", info.License))
	}
	if info.Homepage != "" {
		content.WriteString(fmt.Sprintf("homepage: \"%s\"\n", info.Homepage))
	}
	content.WriteString("---\n\n")

	// Markdown content
	content.WriteString(info.Content)

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

func (f *GitHubActionsFetcher) loadActionInfoFromMarkdown(filePath string) (*GitHubActionInfo, error) {
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
		Repository  string `yaml:"repository"`
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Author      string `yaml:"author"`
		Version     string `yaml:"version"`
		Stars       int    `yaml:"stars"`
		License     string `yaml:"license"`
		Homepage    string `yaml:"homepage"`
	}

	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return &GitHubActionInfo{
		Repository:  meta.Repository,
		Name:        meta.Name,
		Description: meta.Description,
		Author:      meta.Author,
		Version:     meta.Version,
		Stars:       meta.Stars,
		License:     meta.License,
		Homepage:    meta.Homepage,
		Content:     strings.TrimSpace(parts[2]),
	}, nil
}

func getStringFromActionMap(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func escapeYAMLActionString(s string) string {
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
