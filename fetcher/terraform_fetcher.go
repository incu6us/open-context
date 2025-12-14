package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/incu6us/open-context/cache"
	"github.com/incu6us/open-context/config"
)

type TerraformFetcher struct {
	client *http.Client
	cache  *cache.Manager
}

type TerraformVersionInfo struct {
	Version     string `yaml:"version"`
	ReleaseDate string `yaml:"releaseDate"`
	ReleaseURL  string `yaml:"releaseURL"`
	Content     string `yaml:"-"`
}

func NewTerraformFetcher(cacheDir string) *TerraformFetcher {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config, using defaults: %v\n", err)
		cfg = &config.Config{
			CacheTTL: config.Duration{Duration: 0},
		}
	}

	// Create cache manager
	cacheManager := cache.NewManager(cacheDir, cfg.CacheTTL.Duration)

	return &TerraformFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: cacheManager,
	}
}

// FetchTerraformVersion fetches information about a specific Terraform version
func (f *TerraformFetcher) FetchTerraformVersion(version string) (*TerraformVersionInfo, error) {
	// Normalize version (add 'v' prefix if missing for GitHub API)
	githubVersion := version
	if !strings.HasPrefix(version, "v") {
		githubVersion = "v" + version
	}

	// Check cache first
	cachedPath := f.cache.GetFilePath("terraform", "versions", fmt.Sprintf("%s.md", version))
	versionInfo, err := f.loadVersionInfoFromMarkdown(cachedPath)
	if err == nil && versionInfo != nil {
		fmt.Fprintf(os.Stderr, "Loaded Terraform version '%s' from cache\n", version)
		return versionInfo, nil
	}

	// Fetch from GitHub API
	fmt.Fprintf(os.Stderr, "Fetching Terraform version '%s' from GitHub...\n", version)

	// Fetch release information from GitHub
	apiURL := fmt.Sprintf("https://api.github.com/repos/hashicorp/terraform/releases/tags/%s", githubVersion)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent for GitHub API
	req.Header.Set("User-Agent", "open-context-mcp-server")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Terraform release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("Terraform version %s not found", version)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse GitHub release data
	var releaseData map[string]interface{}
	if err := json.Unmarshal(body, &releaseData); err != nil {
		return nil, fmt.Errorf("failed to parse release data: %w", err)
	}

	// Extract version information
	versionInfo = &TerraformVersionInfo{
		Version: version,
	}

	// Extract release URL
	if htmlURL, ok := releaseData["html_url"].(string); ok {
		versionInfo.ReleaseURL = htmlURL
	}

	// Extract release date
	if publishedAt, ok := releaseData["published_at"].(string); ok {
		// Parse and format the date
		if t, err := time.Parse(time.RFC3339, publishedAt); err == nil {
			versionInfo.ReleaseDate = t.Format("2006-01-02")
		} else {
			versionInfo.ReleaseDate = publishedAt
		}
	}

	// Extract release notes from body
	releaseNotes := ""
	if body, ok := releaseData["body"].(string); ok {
		releaseNotes = body
	}

	// Build content
	versionInfo.Content = f.buildVersionContent(versionInfo, releaseNotes)

	// Cache the result
	if err := f.saveVersionInfoAsMarkdown(cachedPath, versionInfo); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache version info: %v\n", err)
	}

	return versionInfo, nil
}

func (f *TerraformFetcher) buildVersionContent(info *TerraformVersionInfo, releaseNotes string) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# Terraform %s\n\n", info.Version))

	if info.ReleaseDate != "" {
		content.WriteString(fmt.Sprintf("**Release Date:** %s\n\n", info.ReleaseDate))
	}

	if info.ReleaseURL != "" {
		content.WriteString(fmt.Sprintf("**Release Notes:** [%s](%s)\n\n", info.Version, info.ReleaseURL))
	}

	content.WriteString("## Installation\n\n")
	content.WriteString("### Using tfenv (version manager)\n\n")
	content.WriteString("```bash\n")
	content.WriteString(fmt.Sprintf("tfenv install %s\n", info.Version))
	content.WriteString(fmt.Sprintf("tfenv use %s\n", info.Version))
	content.WriteString("```\n\n")

	content.WriteString("### Direct Download\n\n")
	content.WriteString(fmt.Sprintf("Download from [releases.hashicorp.com](https://releases.hashicorp.com/terraform/%s/)\n\n", info.Version))

	content.WriteString("### Using Homebrew (macOS)\n\n")
	content.WriteString("```bash\n")
	content.WriteString(fmt.Sprintf("brew install terraform@%s\n", info.Version))
	content.WriteString("```\n\n")

	content.WriteString("### Using Chocolatey (Windows)\n\n")
	content.WriteString("```powershell\n")
	content.WriteString(fmt.Sprintf("choco install terraform --version=%s\n", info.Version))
	content.WriteString("```\n\n")

	if releaseNotes != "" {
		content.WriteString("## Release Notes\n\n")
		content.WriteString(releaseNotes)
		content.WriteString("\n\n")
	}

	content.WriteString("## Documentation\n\n")
	content.WriteString("For detailed documentation, visit:\n\n")
	content.WriteString("- [Terraform Documentation](https://www.terraform.io/docs)\n")
	content.WriteString("- [Terraform Registry](https://registry.terraform.io/)\n")
	content.WriteString("- [Terraform GitHub Repository](https://github.com/hashicorp/terraform)\n")
	content.WriteString("- [Terraform Release Notes](https://github.com/hashicorp/terraform/releases)\n")

	return content.String()
}

func (f *TerraformFetcher) saveVersionInfoAsMarkdown(filePath string, info *TerraformVersionInfo) error {
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
	if info.ReleaseURL != "" {
		content.WriteString(fmt.Sprintf("releaseURL: \"%s\"\n", info.ReleaseURL))
	}
	content.WriteString("---\n\n")

	// Markdown content
	content.WriteString(info.Content)

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

func (f *TerraformFetcher) loadVersionInfoFromMarkdown(filePath string) (*TerraformVersionInfo, error) {
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
		ReleaseURL  string `yaml:"releaseURL"`
	}

	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return &TerraformVersionInfo{
		Version:     meta.Version,
		ReleaseDate: meta.ReleaseDate,
		ReleaseURL:  meta.ReleaseURL,
		Content:     strings.TrimSpace(parts[2]),
	}, nil
}
