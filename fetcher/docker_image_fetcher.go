package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/incu6us/open-context/cache"
	"github.com/incu6us/open-context/config"
)

type DockerImageFetcher struct {
	client *http.Client
	cache  *cache.Manager
}

type DockerImageInfo struct {
	Image       string   `yaml:"image"`
	Tag         string   `yaml:"tag"`
	Digest      string   `yaml:"digest"`
	LastUpdated string   `yaml:"lastUpdated"`
	FullImage   string   `yaml:"fullImage"`
	Content     string   `yaml:"-"`
	Tags        []string `yaml:"-"`
}

type DockerHubTagResponse struct {
	Count   int `json:"count"`
	Results []struct {
		Name        string    `json:"name"`
		FullSize    int64     `json:"full_size"`
		LastUpdated time.Time `json:"last_updated"`
		Digest      string    `json:"digest,omitempty"`
		Images      []struct {
			Architecture string `json:"architecture"`
			OS           string `json:"os"`
			Size         int64  `json:"size"`
		} `json:"images"`
	} `json:"results"`
}

func NewDockerImageFetcher(cacheDir string) *DockerImageFetcher {
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

	return &DockerImageFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: cacheManager,
	}
}

// FetchDockerImage fetches information about a specific Docker image and tag
func (f *DockerImageFetcher) FetchDockerImage(image, tag string) (*DockerImageInfo, error) {
	// Normalize image name (handle official images)
	namespace, repository := parseImageName(image)

	// Check cache first
	cacheKey := fmt.Sprintf("%s_%s_%s", namespace, repository, tag)
	cachedPath := f.cache.GetFilePath("docker", "images", fmt.Sprintf("%s.md", cacheKey))
	imageInfo, err := f.loadImageInfoFromMarkdown(cachedPath)
	if err == nil && imageInfo != nil {
		fmt.Fprintf(os.Stderr, "Loaded Docker image '%s:%s' from cache\n", image, tag)
		return imageInfo, nil
	}

	// Fetch from Docker Hub API
	fmt.Fprintf(os.Stderr, "Fetching Docker image '%s:%s' from Docker Hub...\n", image, tag)

	// First, get the specific tag information
	tagInfo, err := f.fetchTagInfo(namespace, repository, tag)
	if err != nil {
		return nil, err
	}

	// Get available tags for context
	tags, err := f.fetchAvailableTags(namespace, repository, 20)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to fetch available tags: %v\n", err)
		tags = []string{}
	}

	// Build image info
	imageInfo = &DockerImageInfo{
		Image:       image,
		Tag:         tag,
		Digest:      tagInfo.Results[0].Digest,
		LastUpdated: tagInfo.Results[0].LastUpdated.Format("2006-01-02"),
		FullImage:   fmt.Sprintf("%s:%s", image, tag),
		Tags:        tags,
	}

	// Build content
	imageInfo.Content = f.buildImageContent(imageInfo, tagInfo)

	// Cache the result
	if err := f.saveImageInfoAsMarkdown(cachedPath, imageInfo); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache image info: %v\n", err)
	}

	return imageInfo, nil
}

func (f *DockerImageFetcher) fetchTagInfo(namespace, repository, tag string) (*DockerHubTagResponse, error) {
	apiURL := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags/%s", namespace, repository, tag)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "open-context-mcp-server")
	req.Header.Set("Accept", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Docker image tag: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("docker image tag %s/%s:%s not found", namespace, repository, tag)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docker Hub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tagData DockerHubTagResponse
	// The single tag endpoint returns just the tag object, not a results array
	var singleTag struct {
		Name        string    `json:"name"`
		FullSize    int64     `json:"full_size"`
		LastUpdated time.Time `json:"last_updated"`
		Digest      string    `json:"digest,omitempty"`
		Images      []struct {
			Architecture string `json:"architecture"`
			OS           string `json:"os"`
			Size         int64  `json:"size"`
		} `json:"images"`
	}

	if err := json.Unmarshal(body, &singleTag); err != nil {
		return nil, fmt.Errorf("failed to parse tag data: %w", err)
	}

	tagData.Results = []struct {
		Name        string    `json:"name"`
		FullSize    int64     `json:"full_size"`
		LastUpdated time.Time `json:"last_updated"`
		Digest      string    `json:"digest,omitempty"`
		Images      []struct {
			Architecture string `json:"architecture"`
			OS           string `json:"os"`
			Size         int64  `json:"size"`
		} `json:"images"`
	}{singleTag}

	return &tagData, nil
}

func (f *DockerImageFetcher) fetchAvailableTags(namespace, repository string, limit int) ([]string, error) {
	apiURL := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags?page_size=%d", namespace, repository, limit)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "open-context-mcp-server")
	req.Header.Set("Accept", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docker Hub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tagsData DockerHubTagResponse
	if err := json.Unmarshal(body, &tagsData); err != nil {
		return nil, fmt.Errorf("failed to parse tags data: %w", err)
	}

	tags := make([]string, 0, len(tagsData.Results))
	for _, tag := range tagsData.Results {
		tags = append(tags, tag.Name)
	}

	// Sort tags in reverse order (newest first, assuming semantic versioning)
	sort.Sort(sort.Reverse(sort.StringSlice(tags)))

	return tags, nil
}

func (f *DockerImageFetcher) buildImageContent(info *DockerImageInfo, tagData *DockerHubTagResponse) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# Docker Image: %s\n\n", info.FullImage))

	if len(tagData.Results) > 0 {
		tag := tagData.Results[0]

		content.WriteString("## Image Information\n\n")
		content.WriteString(fmt.Sprintf("**Tag:** %s\n\n", info.Tag))
		content.WriteString(fmt.Sprintf("**Last Updated:** %s\n\n", info.LastUpdated))

		if tag.FullSize > 0 {
			sizeMB := float64(tag.FullSize) / (1024 * 1024)
			content.WriteString(fmt.Sprintf("**Size:** %.2f MB\n\n", sizeMB))
		}

		if info.Digest != "" {
			content.WriteString(fmt.Sprintf("**Digest:** `%s`\n\n", info.Digest))
		}

		// Architecture information
		if len(tag.Images) > 0 {
			content.WriteString("## Available Architectures\n\n")
			archMap := make(map[string]bool)
			for _, img := range tag.Images {
				arch := fmt.Sprintf("%s/%s", img.OS, img.Architecture)
				archMap[arch] = true
			}
			for arch := range archMap {
				content.WriteString(fmt.Sprintf("- %s\n", arch))
			}
			content.WriteString("\n")
		}
	}

	// Usage examples
	content.WriteString("## Usage\n\n")
	content.WriteString("### Pull the image\n\n")
	content.WriteString("```bash\n")
	content.WriteString(fmt.Sprintf("docker pull %s\n", info.FullImage))
	content.WriteString("```\n\n")

	content.WriteString("### Run a container\n\n")
	content.WriteString("```bash\n")
	content.WriteString(fmt.Sprintf("docker run -it %s\n", info.FullImage))
	content.WriteString("```\n\n")

	content.WriteString("### Use in Dockerfile\n\n")
	content.WriteString("```dockerfile\n")
	content.WriteString(fmt.Sprintf("FROM %s\n", info.FullImage))
	content.WriteString("```\n\n")

	// Available tags
	if len(info.Tags) > 0 {
		content.WriteString("## Recent Tags\n\n")
		content.WriteString(fmt.Sprintf("For image `%s`, the following tags are available:\n\n", info.Image))
		for i, tag := range info.Tags {
			if i >= 10 {
				content.WriteString(fmt.Sprintf("\n...and %d more tags\n", len(info.Tags)-10))
				break
			}
			content.WriteString(fmt.Sprintf("- `%s`\n", tag))
		}
		content.WriteString("\n")
	}

	content.WriteString("## Documentation\n\n")
	content.WriteString(fmt.Sprintf("- [Docker Hub Repository](https://hub.docker.com/r/%s)\n", info.Image))
	content.WriteString("- [Docker Documentation](https://docs.docker.com/)\n")

	return content.String()
}

func (f *DockerImageFetcher) saveImageInfoAsMarkdown(filePath string, info *DockerImageInfo) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("image: \"%s\"\n", info.Image))
	content.WriteString(fmt.Sprintf("tag: \"%s\"\n", info.Tag))
	if info.Digest != "" {
		content.WriteString(fmt.Sprintf("digest: \"%s\"\n", info.Digest))
	}
	if info.LastUpdated != "" {
		content.WriteString(fmt.Sprintf("lastUpdated: \"%s\"\n", info.LastUpdated))
	}
	content.WriteString(fmt.Sprintf("fullImage: \"%s\"\n", info.FullImage))
	content.WriteString("---\n\n")

	// Markdown content
	content.WriteString(info.Content)

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

func (f *DockerImageFetcher) loadImageInfoFromMarkdown(filePath string) (*DockerImageInfo, error) {
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
		Image       string `yaml:"image"`
		Tag         string `yaml:"tag"`
		Digest      string `yaml:"digest"`
		LastUpdated string `yaml:"lastUpdated"`
		FullImage   string `yaml:"fullImage"`
	}

	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return &DockerImageInfo{
		Image:       meta.Image,
		Tag:         meta.Tag,
		Digest:      meta.Digest,
		LastUpdated: meta.LastUpdated,
		FullImage:   meta.FullImage,
		Content:     strings.TrimSpace(parts[2]),
	}, nil
}

// parseImageName parses a Docker image name into namespace and repository
// For official images like "golang", it returns ("library", "golang")
// For user images like "myuser/myapp", it returns ("myuser", "myapp")
func parseImageName(image string) (namespace, repository string) {
	parts := strings.Split(image, "/")
	if len(parts) == 1 {
		// Official image
		return "library", parts[0]
	}
	// User/org image
	return parts[0], parts[1]
}
