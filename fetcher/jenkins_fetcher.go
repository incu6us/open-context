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

	yaml "gopkg.in/yaml.v3"

	"github.com/incu6us/open-context/cache"
	"github.com/incu6us/open-context/config"
)

type JenkinsFetcher struct {
	client *http.Client
	cache  *cache.Manager
}

type JenkinsVersionInfo struct {
	Version     string `yaml:"version"`
	ReleaseDate string `yaml:"releaseDate"`
	ReleaseURL  string `yaml:"releaseURL"`
	Content     string `yaml:"-"`
}

func NewJenkinsFetcher(cacheDir string) *JenkinsFetcher {
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

	return &JenkinsFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: cacheManager,
	}
}

// FetchJenkinsVersion fetches information about a specific Jenkins version
func (f *JenkinsFetcher) FetchJenkinsVersion(version string) (*JenkinsVersionInfo, error) {
	// Normalize version (Jenkins uses format like "2.440.3" or with "jenkins-" prefix)
	githubVersion := version
	if !strings.HasPrefix(version, "jenkins-") {
		githubVersion = "jenkins-" + version
	}

	// Check cache first
	cachedPath := f.cache.GetFilePath("jenkins", "versions", fmt.Sprintf("%s.md", version))
	versionInfo, err := f.loadVersionInfoFromMarkdown(cachedPath)
	if err == nil && versionInfo != nil {
		fmt.Fprintf(os.Stderr, "Loaded Jenkins version '%s' from cache\n", version)
		return versionInfo, nil
	}

	// Fetch from GitHub API
	fmt.Fprintf(os.Stderr, "Fetching Jenkins version '%s' from GitHub...\n", version)

	// Fetch release information from GitHub
	apiURL := fmt.Sprintf("https://api.github.com/repos/jenkinsci/jenkins/releases/tags/%s", githubVersion)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent for GitHub API
	req.Header.Set("User-Agent", "open-context-mcp-server")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Jenkins release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("jenkins version %s not found", version)
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
	versionInfo = &JenkinsVersionInfo{
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

func (f *JenkinsFetcher) buildVersionContent(info *JenkinsVersionInfo, releaseNotes string) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# Jenkins %s\n\n", info.Version))

	if info.ReleaseDate != "" {
		content.WriteString(fmt.Sprintf("**Release Date:** %s\n\n", info.ReleaseDate))
	}

	if info.ReleaseURL != "" {
		content.WriteString(fmt.Sprintf("**Release Notes:** [%s](%s)\n\n", info.Version, info.ReleaseURL))
	}

	content.WriteString("## Installation\n\n")
	content.WriteString("### Download WAR File\n\n")
	content.WriteString(fmt.Sprintf("Download from [Jenkins Downloads](https://get.jenkins.io/war/%s/jenkins.war)\n\n", info.Version))

	content.WriteString("### Using Docker\n\n")
	content.WriteString("```bash\n")
	content.WriteString(fmt.Sprintf("docker pull jenkins/jenkins:%s\n", info.Version))
	content.WriteString(fmt.Sprintf("docker run -p 8080:8080 -p 50000:50000 jenkins/jenkins:%s\n", info.Version))
	content.WriteString("```\n\n")

	content.WriteString("### Using Package Manager (Debian/Ubuntu)\n\n")
	content.WriteString("```bash\n")
	content.WriteString("wget -q -O - https://pkg.jenkins.io/debian-stable/jenkins.io.key | sudo apt-key add -\n")
	content.WriteString("sudo sh -c 'echo deb https://pkg.jenkins.io/debian-stable binary/ > /etc/apt/sources.list.d/jenkins.list'\n")
	content.WriteString("sudo apt update\n")
	content.WriteString("sudo apt install jenkins\n")
	content.WriteString("```\n\n")

	content.WriteString("### Using Package Manager (RHEL/CentOS)\n\n")
	content.WriteString("```bash\n")
	content.WriteString("sudo wget -O /etc/yum.repos.d/jenkins.repo https://pkg.jenkins.io/redhat-stable/jenkins.repo\n")
	content.WriteString("sudo rpm --import https://pkg.jenkins.io/redhat-stable/jenkins.io.key\n")
	content.WriteString("sudo yum install jenkins\n")
	content.WriteString("```\n\n")

	if releaseNotes != "" {
		content.WriteString("## Release Notes\n\n")
		content.WriteString(releaseNotes)
		content.WriteString("\n\n")
	}

	content.WriteString("## Documentation\n\n")
	content.WriteString("For detailed documentation, visit:\n\n")
	content.WriteString("- [Jenkins Documentation](https://www.jenkins.io/doc/)\n")
	content.WriteString("- [Jenkins User Handbook](https://www.jenkins.io/doc/book/)\n")
	content.WriteString("- [Jenkins Plugins](https://plugins.jenkins.io/)\n")
	content.WriteString("- [Jenkins GitHub Repository](https://github.com/jenkinsci/jenkins)\n")
	content.WriteString("- [Jenkins Changelog](https://www.jenkins.io/changelog/)\n")

	return content.String()
}

func (f *JenkinsFetcher) saveVersionInfoAsMarkdown(filePath string, info *JenkinsVersionInfo) error {
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

func (f *JenkinsFetcher) loadVersionInfoFromMarkdown(filePath string) (*JenkinsVersionInfo, error) {
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

	return &JenkinsVersionInfo{
		Version:     meta.Version,
		ReleaseDate: meta.ReleaseDate,
		ReleaseURL:  meta.ReleaseURL,
		Content:     strings.TrimSpace(parts[2]),
	}, nil
}
