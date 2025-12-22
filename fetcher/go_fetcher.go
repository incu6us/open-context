package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
	"gopkg.in/yaml.v3"

	"github.com/incu6us/open-context/cache"
	"github.com/incu6us/open-context/config"
)

const (
	pkgGoDevBaseURL = "https://pkg.go.dev"
	goStdLibURL     = "https://pkg.go.dev/std"
	goDevBaseURL    = "https://go.dev"
	goRelNotesURL   = "https://go.dev/doc/devel/release"
	goProxyBaseURL  = "https://proxy.golang.org"
)

type GoFetcher struct {
	client      *http.Client
	cacheDir    string
	packageList []string
	cache       *cache.Manager
}

type PackageDoc struct {
	Name        string   `json:"name"`
	ImportPath  string   `json:"importPath"`
	Synopsis    string   `json:"synopsis"`
	Description string   `json:"description"`
	Examples    []string `json:"examples,omitempty"`
}

type GoVersionInfo struct {
	Version     string `json:"version"`
	ReleaseDate string `json:"releaseDate"`
	ReleaseURL  string `json:"releaseURL"`
	Content     string `json:"content"`
}

type LibraryInfo struct {
	ImportPath  string `json:"importPath"`
	Version     string `json:"version"`
	Synopsis    string `json:"synopsis"`
	Description string `json:"description"`
	Repository  string `json:"repository,omitempty"`
	License     string `json:"license,omitempty"`
}

func NewGoFetcher(cacheDir string) *GoFetcher {
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

	return &GoFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cacheDir: cacheDir,
		cache:    cacheManager,
	}
}

// FetchStdLib fetches documentation for all Go standard library packages
func (f *GoFetcher) FetchStdLib() error {
	fmt.Println("Fetching Go standard library package list...")

	packages, err := f.getStdLibPackages()
	if err != nil {
		return fmt.Errorf("failed to get stdlib packages: %w", err)
	}

	f.packageList = packages
	fmt.Printf("Found %d standard library packages\n", len(packages))

	// Create output directory
	outputDir := filepath.Join(f.cacheDir, "go", "topics")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create metadata
	metadata := map[string]interface{}{
		"name":        "go",
		"displayName": "Go",
		"description": "Go standard library documentation (fetched from pkg.go.dev)",
	}

	metadataPath := filepath.Join(f.cacheDir, "go", "metadata.json")
	if err := writeJSON(metadataPath, metadata); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Fetch documentation for key packages (to avoid overwhelming the system)
	keyPackages := f.getKeyPackages(packages)

	fmt.Printf("Fetching documentation for %d key packages...\n", len(keyPackages))
	for i, pkg := range keyPackages {
		fmt.Printf("[%d/%d] Fetching %s...\n", i+1, len(keyPackages), pkg)

		doc, err := f.fetchPackageDoc(pkg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch %s: %v\n", pkg, err)
			continue
		}

		// Convert package name to safe filename
		filename := strings.ReplaceAll(pkg, "/", "_") + ".json"
		outputPath := filepath.Join(outputDir, filename)

		// Convert to topic format
		topic := f.packageDocToTopic(doc)
		if err := writeJSON(outputPath, topic); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to write %s: %v\n", pkg, err)
			continue
		}

		// Be nice to the server
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("Go standard library documentation fetched successfully!")
	return nil
}

// getStdLibPackages retrieves the list of all standard library packages
func (f *GoFetcher) getStdLibPackages() ([]string, error) {
	resp, err := f.client.Get(goStdLibURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var packages []string
	var findPackages func(*html.Node)
	findPackages = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && strings.HasPrefix(attr.Val, "/") {
					// Extract package path from href like "/fmt" or "/net/http"
					path := strings.TrimPrefix(attr.Val, "/")
					if path != "" && !strings.Contains(path, "?") && !strings.Contains(path, "#") {
						// Basic validation: looks like a package path
						if regexp.MustCompile(`^[a-z][a-z0-9/]*$`).MatchString(path) {
							packages = append(packages, path)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findPackages(c)
		}
	}

	findPackages(doc)
	return uniqueStrings(packages), nil
}

// getKeyPackages returns a curated list of important packages to fetch
func (f *GoFetcher) getKeyPackages(allPackages []string) []string {
	// Priority packages that are commonly used
	priority := []string{
		"fmt", "io", "os", "strings", "strconv", "time",
		"net/http", "encoding/json", "context", "sync",
		"bufio", "bytes", "errors", "log", "path/filepath",
		"regexp", "sort", "math", "crypto/sha256", "crypto/md5",
		"crypto/tls", "database/sql", "html/template", "text/template",
		"net/url", "io/ioutil", "reflect", "runtime", "testing",
	}

	var result []string
	packageSet := make(map[string]bool)

	// Add priority packages first
	for _, pkg := range priority {
		if contains(allPackages, pkg) {
			result = append(result, pkg)
			packageSet[pkg] = true
		}
	}

	// Add other packages up to a reasonable limit
	for _, pkg := range allPackages {
		if len(result) >= 100 { // Limit to 100 packages
			break
		}
		if !packageSet[pkg] && !strings.Contains(pkg, "internal") {
			result = append(result, pkg)
			packageSet[pkg] = true
		}
	}

	return result
}

// fetchPackageDoc fetches documentation for a specific package
func (f *GoFetcher) fetchPackageDoc(pkgPath string) (*PackageDoc, error) {
	url := fmt.Sprintf("%s/%s", pkgGoDevBaseURL, pkgPath)

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	pkgDoc := &PackageDoc{
		Name:       filepath.Base(pkgPath),
		ImportPath: pkgPath,
	}

	// Extract documentation
	pkgDoc.Synopsis = f.extractSynopsis(doc)
	pkgDoc.Description = f.extractDescription(doc, pkgPath)

	return pkgDoc, nil
}

// extractSynopsis extracts the package synopsis
func (f *GoFetcher) extractSynopsis(doc *html.Node) string {
	var synopsis string
	var findSynopsis func(*html.Node)
	findSynopsis = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Look for meta description or documentation overview
			if n.Data == "meta" {
				var isDescription bool
				var content string
				for _, attr := range n.Attr {
					if attr.Key == "name" && attr.Val == "description" {
						isDescription = true
					}
					if attr.Key == "content" {
						content = attr.Val
					}
				}
				if isDescription && content != "" {
					synopsis = content
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if synopsis != "" {
				return
			}
			findSynopsis(c)
		}
	}

	findSynopsis(doc)
	return synopsis
}

// extractDescription extracts the package description and creates markdown content
func (f *GoFetcher) extractDescription(doc *html.Node, pkgPath string) string {
	synopsis := f.extractSynopsis(doc)

	// Build markdown documentation
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# Package %s\n\n", pkgPath))
	content.WriteString(fmt.Sprintf("Import path: `%s`\n\n", pkgPath))

	if synopsis != "" {
		content.WriteString(fmt.Sprintf("%s\n\n", synopsis))
	}

	content.WriteString("## Overview\n\n")
	content.WriteString(fmt.Sprintf("The `%s` package provides functionality as documented at [pkg.go.dev/%s](%s/%s).\n\n",
		filepath.Base(pkgPath), pkgPath, pkgGoDevBaseURL, pkgPath))

	content.WriteString("## Import\n\n")
	content.WriteString("```go\n")
	content.WriteString(fmt.Sprintf("import \"%s\"\n", pkgPath))
	content.WriteString("```\n\n")

	content.WriteString("## Documentation\n\n")
	content.WriteString("For detailed documentation, examples, and API reference, visit:\n\n")
	content.WriteString(fmt.Sprintf("- [pkg.go.dev/%s](%s/%s)\n", pkgPath, pkgGoDevBaseURL, pkgPath))
	content.WriteString(fmt.Sprintf("- [Go Standard Library Documentation](https://golang.org/pkg/%s/)\n\n", pkgPath))

	// Add common usage note
	content.WriteString("## Usage\n\n")
	content.WriteString("This package is part of the Go standard library. ")
	content.WriteString("Check the official documentation for the most up-to-date information, ")
	content.WriteString("type definitions, functions, and examples.\n")

	return content.String()
}

// packageDocToTopic converts a PackageDoc to the topic format
func (f *GoFetcher) packageDocToTopic(doc *PackageDoc) map[string]interface{} {
	// Generate keywords
	keywords := []string{
		doc.Name,
		doc.ImportPath,
		"standard library",
		"stdlib",
		"go",
	}

	// Add path components as keywords
	parts := strings.Split(doc.ImportPath, "/")
	keywords = append(keywords, parts...)

	return map[string]interface{}{
		"id":          strings.ReplaceAll(doc.ImportPath, "/", "_"),
		"title":       fmt.Sprintf("Go Package: %s", doc.ImportPath),
		"description": doc.Synopsis,
		"keywords":    uniqueStrings(keywords),
		"content":     doc.Description,
	}
}

// Helper functions

func writeJSON(path string, data interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, s := range slice {
		if s != "" && !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += getText(c)
	}
	return text
}

// FetchGoVersion fetches and caches information about a specific Go version
func (f *GoFetcher) FetchGoVersion(version string) (*GoVersionInfo, error) {
	// Build cache path
	cachedPath := f.cache.GetFilePath("go", "versions", fmt.Sprintf("%s.md", version))

	// Try to load from cache
	versionInfo, err := f.loadVersionInfoFromMarkdown(cachedPath)
	if err == nil && versionInfo != nil {
		fmt.Printf("Loaded Go %s info from cache\n", version)
		return versionInfo, nil
	}

	// Fetch from official Go website
	fmt.Printf("Fetching Go %s information from official source...\n", version)

	releaseURL := fmt.Sprintf("%s/doc/go%s", goDevBaseURL, version)
	resp, err := f.client.Get(releaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Go %s info: %w", version, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for Go %s", resp.StatusCode, version)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract content
	content := f.extractReleaseNotes(doc, version)

	// Create version info
	versionInfo = &GoVersionInfo{
		Version:     version,
		ReleaseURL:  releaseURL,
		ReleaseDate: f.extractReleaseDate(doc),
		Content:     content,
	}

	// Cache the result
	if err := f.cacheVersionInfo(versionInfo); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache version info: %v\n", err)
	}

	return versionInfo, nil
}

// FetchLibraryInfo fetches and caches information about a Go library/package
func (f *GoFetcher) FetchLibraryInfo(importPath, version string) (*LibraryInfo, error) {
	// If no version specified, query the Go proxy to get the latest version
	if version == "" {
		latestVersion, err := f.getLatestVersion(importPath)
		if err != nil {
			// If we can't get the latest version, log a warning but continue without it
			fmt.Fprintf(os.Stderr, "Warning: failed to get latest version for %s: %v\n", importPath, err)
		} else {
			// Check if the result includes a path change (format: "version:path")
			if strings.Contains(latestVersion, ":") {
				parts := strings.SplitN(latestVersion, ":", 2)
				version = parts[0]
				importPath = parts[1]
				fmt.Printf("Resolved latest version for original path: %s@%s\n", importPath, version)
			} else {
				version = latestVersion
				fmt.Printf("Resolved latest version for %s: %s\n", importPath, version)
			}
		}
	}

	// Build cache path
	cacheKey := strings.ReplaceAll(importPath, "/", "_")
	if version != "" {
		cacheKey = fmt.Sprintf("%s_%s", cacheKey, version)
	}
	cachedPath := f.cache.GetFilePath("go", "libraries", fmt.Sprintf("%s.md", cacheKey))

	// Try to load from cache
	libInfo, err := f.loadLibraryInfoFromMarkdown(cachedPath)
	if err == nil && libInfo != nil {
		fmt.Printf("Loaded %s info from cache\n", importPath)
		return libInfo, nil
	}

	// Fetch from pkg.go.dev
	fmt.Printf("Fetching %s information from pkg.go.dev...\n", importPath)

	url := fmt.Sprintf("%s/%s", pkgGoDevBaseURL, importPath)
	if version != "" {
		url = fmt.Sprintf("%s/%s@%s", pkgGoDevBaseURL, importPath, version)
	}

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch library info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, importPath)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract library information
	libInfo = &LibraryInfo{
		ImportPath:  importPath,
		Version:     version,
		Synopsis:    f.extractSynopsis(doc),
		Description: f.extractLibraryDescription(doc, importPath, version),
		Repository:  f.extractRepository(doc),
		License:     f.extractLicense(doc),
	}

	// Cache the result
	if err := f.cacheLibraryInfo(libInfo); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache library info: %v\n", err)
	}

	return libInfo, nil
}

// Helper methods for extracting information

func (f *GoFetcher) extractReleaseNotes(doc *html.Node, version string) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# Go %s Release Notes\n\n", version))
	content.WriteString(fmt.Sprintf("Official release notes: [go.dev/doc/go%s](https://go.dev/doc/go%s)\n\n", version, version))

	// Extract main content
	var extractContent func(*html.Node, bool)
	var inArticle bool

	extractContent = func(n *html.Node, insideArticle bool) {
		if n.Type == html.ElementNode {
			if n.Data == "article" || (n.Data == "div" && hasClass(n, "Article")) {
				insideArticle = true
				inArticle = true
			}

			if insideArticle {
				switch n.Data {
				case "h2", "h3", "h4":
					text := strings.TrimSpace(getText(n))
					if text != "" {
						prefix := strings.Repeat("#", getHeadingLevel(n.Data))
						content.WriteString(fmt.Sprintf("%s %s\n\n", prefix, text))
					}
				case "p":
					text := strings.TrimSpace(getText(n))
					if text != "" {
						content.WriteString(fmt.Sprintf("%s\n\n", text))
					}
				case "pre":
					code := strings.TrimSpace(getText(n))
					if code != "" {
						content.WriteString(fmt.Sprintf("```\n%s\n```\n\n", code))
					}
				case "ul", "ol":
					f.extractList(n, &content)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractContent(c, insideArticle)
		}
	}

	extractContent(doc, false)

	// If no content was extracted, provide a default message
	if !inArticle {
		content.WriteString("## Overview\n\n")
		content.WriteString(fmt.Sprintf("Go %s introduces new features and improvements. ", version))
		content.WriteString(fmt.Sprintf("Visit the [official release notes](https://go.dev/doc/go%s) for complete details.\n", version))
	}

	return content.String()
}

func (f *GoFetcher) extractLibraryDescription(doc *html.Node, importPath, version string) string {
	var content strings.Builder

	versionStr := "latest"
	if version != "" {
		versionStr = version
	}

	content.WriteString(fmt.Sprintf("# %s\n\n", importPath))
	content.WriteString(fmt.Sprintf("**Version:** %s\n\n", versionStr))
	content.WriteString(fmt.Sprintf("**Import path:** `%s`\n\n", importPath))

	// Add synopsis
	synopsis := f.extractSynopsis(doc)
	if synopsis != "" {
		content.WriteString(fmt.Sprintf("%s\n\n", synopsis))
	}

	// Add repository and license info
	repo := f.extractRepository(doc)
	if repo != "" {
		content.WriteString(fmt.Sprintf("**Repository:** %s\n\n", repo))
	}

	license := f.extractLicense(doc)
	if license != "" {
		content.WriteString(fmt.Sprintf("**License:** %s\n\n", license))
	}

	content.WriteString("## Installation\n\n")
	content.WriteString("```bash\n")
	content.WriteString(fmt.Sprintf("go get %s", importPath))
	if version != "" {
		content.WriteString(fmt.Sprintf("@%s", version))
	}
	content.WriteString("\n```\n\n")

	content.WriteString("## Import\n\n")
	content.WriteString("```go\n")
	content.WriteString(fmt.Sprintf("import \"%s\"\n", importPath))
	content.WriteString("```\n\n")

	content.WriteString("## Documentation\n\n")
	url := fmt.Sprintf("%s/%s", pkgGoDevBaseURL, importPath)
	if version != "" {
		url = fmt.Sprintf("%s/%s@%s", pkgGoDevBaseURL, importPath, version)
	}
	content.WriteString(fmt.Sprintf("For detailed documentation and examples, visit [pkg.go.dev](%s)\n", url))

	return content.String()
}

func (f *GoFetcher) extractReleaseDate(doc *html.Node) string {
	// Try to find release date in the document
	var date string
	var findDate func(*html.Node)

	findDate = func(n *html.Node) {
		if date != "" {
			return
		}

		if n.Type == html.TextNode {
			text := n.Data
			// Look for date patterns like "2023-08-08" or "August 2023"
			if strings.Contains(text, "20") && (strings.Contains(text, "-") ||
				strings.Contains(strings.ToLower(text), "january") ||
				strings.Contains(strings.ToLower(text), "february") ||
				strings.Contains(strings.ToLower(text), "march")) {
				date = strings.TrimSpace(text)
				return
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findDate(c)
		}
	}

	findDate(doc)
	return date
}

func (f *GoFetcher) extractRepository(doc *html.Node) string {
	var repo string
	var findRepo func(*html.Node)

	findRepo = func(n *html.Node) {
		if repo != "" {
			return
		}

		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					if strings.Contains(attr.Val, "github.com") ||
						strings.Contains(attr.Val, "gitlab.com") ||
						strings.Contains(attr.Val, "bitbucket.org") {
						repo = attr.Val
						return
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findRepo(c)
		}
	}

	findRepo(doc)
	return repo
}

func (f *GoFetcher) extractLicense(doc *html.Node) string {
	var license string
	var findLicense func(*html.Node)

	findLicense = func(n *html.Node) {
		if license != "" {
			return
		}

		if n.Type == html.ElementNode && n.Data == "a" {
			text := strings.TrimSpace(getText(n))
			if strings.Contains(strings.ToLower(text), "license") {
				license = text
				return
			}
		}

		if n.Type == html.TextNode {
			text := n.Data
			// Common license patterns
			licenses := []string{"MIT", "Apache", "BSD", "GPL", "LGPL", "MPL"}
			for _, lic := range licenses {
				if strings.Contains(text, lic) {
					license = lic
					return
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findLicense(c)
		}
	}

	findLicense(doc)
	return license
}

func (f *GoFetcher) extractList(n *html.Node, content *strings.Builder) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "li" {
			text := strings.TrimSpace(getText(c))
			if text != "" {
				fmt.Fprintf(content, "- %s\n", text)
			}
		}
	}
	content.WriteString("\n")
}

func (f *GoFetcher) cacheVersionInfo(info *GoVersionInfo) error {
	outputPath := f.cache.GetFilePath("go", "versions", fmt.Sprintf("%s.md", info.Version))
	return f.saveVersionInfoAsMarkdown(outputPath, info)
}

func (f *GoFetcher) cacheLibraryInfo(info *LibraryInfo) error {
	filename := strings.ReplaceAll(info.ImportPath, "/", "_")
	if info.Version != "" {
		filename = fmt.Sprintf("%s_%s", filename, info.Version)
	}

	outputPath := f.cache.GetFilePath("go", "libraries", fmt.Sprintf("%s.md", filename))
	return f.saveLibraryInfoAsMarkdown(outputPath, info)
}

func hasClass(n *html.Node, className string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" && strings.Contains(attr.Val, className) {
			return true
		}
	}
	return false
}

func getHeadingLevel(tag string) int {
	switch tag {
	case "h1":
		return 1
	case "h2":
		return 2
	case "h3":
		return 3
	case "h4":
		return 4
	case "h5":
		return 5
	case "h6":
		return 6
	default:
		return 2
	}
}

// getLatestVersion queries the Go module proxy to find the latest version of a module
// It checks for major versions v2, v3, v4, etc. to find the truly latest version
func (f *GoFetcher) getLatestVersion(importPath string) (string, error) {
	// Try to find the latest major version by checking v2, v3, v4, etc.
	// We check up to v10 which should be sufficient for most packages
	var latestVersion string
	var latestPath string

	// First, check the base path (v0 or v1)
	baseVersion, err := f.queryProxyLatest(importPath)
	if err == nil {
		latestVersion = baseVersion
		latestPath = importPath
	}

	// Check for v2, v3, v4, ... v10
	for major := 2; major <= 10; major++ {
		testPath := fmt.Sprintf("%s/v%d", importPath, major)
		version, err := f.queryProxyLatest(testPath)
		if err == nil {
			// Found a newer major version
			latestVersion = version
			latestPath = testPath
		}
	}

	if latestVersion == "" {
		return "", fmt.Errorf("no versions found for %s", importPath)
	}

	// If we found a different path (with major version suffix), update the import path
	if latestPath != importPath {
		fmt.Printf("Found newer major version at %s (%s)\n", latestPath, latestVersion)
		// Return the path with major version so caller knows to use it
		return latestVersion + ":" + latestPath, nil
	}

	return latestVersion, nil
}

// queryProxyLatest queries the Go proxy for the latest version of a specific module path
func (f *GoFetcher) queryProxyLatest(importPath string) (string, error) {
	url := fmt.Sprintf("%s/%s/@latest", goProxyBaseURL, importPath)

	resp, err := f.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	var result struct {
		Version string `json:"Version"`
		Time    string `json:"Time"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Version, nil
}

// Markdown conversion helpers

func (f *GoFetcher) saveVersionInfoAsMarkdown(filePath string, info *GoVersionInfo) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("version: \"%s\"\n", info.Version))
	content.WriteString(fmt.Sprintf("releaseURL: \"%s\"\n", info.ReleaseURL))
	content.WriteString("---\n\n")

	// Markdown content
	content.WriteString(info.Content)

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

func (f *GoFetcher) saveLibraryInfoAsMarkdown(filePath string, info *LibraryInfo) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("importPath: \"%s\"\n", info.ImportPath))
	if info.Version != "" {
		content.WriteString(fmt.Sprintf("version: \"%s\"\n", info.Version))
	}
	if info.Synopsis != "" {
		content.WriteString(fmt.Sprintf("synopsis: \"%s\"\n", strings.ReplaceAll(info.Synopsis, "\"", "\\\"")))
	}
	if info.Repository != "" {
		content.WriteString(fmt.Sprintf("repository: \"%s\"\n", info.Repository))
	}
	if info.License != "" {
		content.WriteString(fmt.Sprintf("license: \"%s\"\n", info.License))
	}
	content.WriteString("---\n\n")

	// Markdown content
	content.WriteString(info.Description)

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

func (f *GoFetcher) loadVersionInfoFromMarkdown(filePath string) (*GoVersionInfo, error) {
	// Check if file is expired first
	expired, err := f.cache.IsExpired(filePath)
	if err != nil || expired {
		return nil, fmt.Errorf("file not found or expired")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	content := string(data)

	// Split frontmatter and content
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid markdown format: missing frontmatter")
	}

	// Parse YAML frontmatter
	var meta struct {
		Version    string `yaml:"version"`
		ReleaseURL string `yaml:"releaseURL"`
	}

	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return &GoVersionInfo{
		Version:    meta.Version,
		ReleaseURL: meta.ReleaseURL,
		Content:    strings.TrimSpace(parts[2]),
	}, nil
}

func (f *GoFetcher) loadLibraryInfoFromMarkdown(filePath string) (*LibraryInfo, error) {
	// Check if file is expired first
	expired, err := f.cache.IsExpired(filePath)
	if err != nil || expired {
		return nil, fmt.Errorf("file not found or expired")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	content := string(data)

	// Split frontmatter and content
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid markdown format: missing frontmatter")
	}

	// Parse YAML frontmatter
	var meta struct {
		ImportPath string `yaml:"importPath"`
		Version    string `yaml:"version"`
		Synopsis   string `yaml:"synopsis"`
		Repository string `yaml:"repository"`
		License    string `yaml:"license"`
	}

	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return &LibraryInfo{
		ImportPath:  meta.ImportPath,
		Version:     meta.Version,
		Synopsis:    meta.Synopsis,
		Description: strings.TrimSpace(parts[2]),
		Repository:  meta.Repository,
		License:     meta.License,
	}, nil
}
