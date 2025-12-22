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

type NodeVersionInfo struct {
	Version     string `yaml:"version"`
	ReleaseDate string `yaml:"releaseDate"`
	LTS         string `yaml:"lts"`
	Content     string `yaml:"-"`
}

type NodeFetcher struct {
	*BaseFetcher
}

func NewNodeFetcher(cacheDir string) *NodeFetcher {
	return &NodeFetcher{
		BaseFetcher: NewBaseFetcher(cacheDir),
	}
}

// FetchNodeVersion fetches information about a specific Node.js version
func (f *NodeFetcher) FetchNodeVersion(version string) (*NodeVersionInfo, error) {
	// Normalize version (add 'v' prefix if missing)
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	// Check cache first
	cachedPath := f.cache.GetFilePath("node", "versions", fmt.Sprintf("%s.md", version))
	versionInfo, err := f.loadVersionInfoFromMarkdown(cachedPath)
	if err == nil && versionInfo != nil {
		fmt.Fprintf(os.Stderr, "Loaded Node.js version '%s' from cache\n", version)
		return versionInfo, nil
	}

	// Fetch from Node.js distribution API
	fmt.Fprintf(os.Stderr, "Fetching Node.js version '%s' from nodejs.org...\n", version)

	// First, get the version list to find details
	resp, err := f.client.Get("https://nodejs.org/dist/index.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Node.js version list: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nodejs.org returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse version list
	var versions []map[string]interface{}
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, fmt.Errorf("failed to parse version data: %w", err)
	}

	// Find the requested version
	var versionData map[string]interface{}
	for _, v := range versions {
		if vStr, ok := v["version"].(string); ok && vStr == version {
			versionData = v
			break
		}
	}

	if versionData == nil {
		return nil, fmt.Errorf("node.js version %s not found", version)
	}

	// Extract version information
	versionInfo = &NodeVersionInfo{
		Version: version,
	}

	// Extract release date
	if date, ok := versionData["date"].(string); ok {
		versionInfo.ReleaseDate = date
	}

	// Extract LTS information
	if lts, ok := versionData["lts"].(string); ok && lts != "" {
		versionInfo.LTS = lts
	} else if lts, ok := versionData["lts"].(bool); ok && lts {
		versionInfo.LTS = "Yes"
	}

	// Build content
	versionInfo.Content = f.buildVersionContent(versionInfo, versionData)

	// Cache the result
	if err := f.saveVersionInfoAsMarkdown(cachedPath, versionInfo); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache version info: %v\n", err)
	}

	return versionInfo, nil
}

func (f *NodeFetcher) buildVersionContent(info *NodeVersionInfo, data map[string]interface{}) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# Node.js %s\n\n", info.Version))

	if info.ReleaseDate != "" {
		content.WriteString(fmt.Sprintf("**Release Date:** %s\n\n", info.ReleaseDate))
	}

	if info.LTS != "" && info.LTS != "false" {
		content.WriteString(fmt.Sprintf("**LTS:** %s\n\n", info.LTS))
	}

	// Add modules information if available
	if modules, ok := data["modules"].(string); ok {
		content.WriteString(fmt.Sprintf("**Modules Version:** %s\n\n", modules))
	}

	// Add V8 version if available
	if v8, ok := data["v8"].(string); ok {
		content.WriteString(fmt.Sprintf("**V8 Version:** %s\n\n", v8))
	}

	// Add npm version if available
	if npm, ok := data["npm"].(string); ok {
		content.WriteString(fmt.Sprintf("**npm Version:** %s\n\n", npm))
	}

	content.WriteString("## Installation\n\n")
	content.WriteString("### Using nvm\n\n")
	content.WriteString("```bash\n")
	content.WriteString(fmt.Sprintf("nvm install %s\n", info.Version))
	content.WriteString("```\n\n")

	content.WriteString("### Download\n\n")
	content.WriteString(fmt.Sprintf("Download from [nodejs.org](https://nodejs.org/dist/%s/)\n\n", info.Version))

	content.WriteString("## Documentation\n\n")
	// Extract major version for documentation link
	majorVersion := strings.TrimPrefix(info.Version, "v")
	if idx := strings.Index(majorVersion, "."); idx > 0 {
		majorVersion = majorVersion[:idx]
	}
	content.WriteString(fmt.Sprintf("For detailed documentation, visit [Node.js v%s Documentation](https://nodejs.org/docs/latest-v%s.x/api/)\n", majorVersion, majorVersion))

	return content.String()
}

func (f *NodeFetcher) saveVersionInfoAsMarkdown(filePath string, info *NodeVersionInfo) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("version: \"%s\"\n", info.Version))
	if info.ReleaseDate != "" {
		content.WriteString(fmt.Sprintf("releaseDate: \"%s\"\n", info.ReleaseDate))
	}
	if info.LTS != "" {
		content.WriteString(fmt.Sprintf("lts: \"%s\"\n", info.LTS))
	}
	content.WriteString("---\n\n")

	// Markdown content
	content.WriteString(info.Content)

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

func (f *NodeFetcher) loadVersionInfoFromMarkdown(filePath string) (*NodeVersionInfo, error) {
	expired, err := f.cache.IsExpired(filePath)
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
		Version     string `yaml:"version"`
		ReleaseDate string `yaml:"releaseDate"`
		LTS         string `yaml:"lts"`
	}

	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return &NodeVersionInfo{
		Version:     meta.Version,
		ReleaseDate: meta.ReleaseDate,
		LTS:         meta.LTS,
		Content:     strings.TrimSpace(parts[2]),
	}, nil
}
