package image

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"waverless/pkg/config"
	"waverless/pkg/logger"
)

// Checker checks Docker images for updates
type Checker struct {
	client       *http.Client
	dockerConfig *config.DockerConfig
}

// NewChecker creates a new image checker
func NewChecker(dockerConfig *config.DockerConfig) *Checker {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Configure proxy for HTTP client
	// Priority: config.yaml proxy_url > environment variables (HTTP_PROXY, HTTPS_PROXY, http_proxy, https_proxy)
	var transport *http.Transport
	if dockerConfig != nil && dockerConfig.ProxyURL != "" {
		// Use proxy from config file
		proxyURL, err := url.Parse(dockerConfig.ProxyURL)
		if err == nil {
			transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	} else {
		// Use proxy from environment variables (HTTP_PROXY, HTTPS_PROXY, etc.)
		// http.ProxyFromEnvironment automatically reads HTTP_PROXY, HTTPS_PROXY, NO_PROXY, etc.
		transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}
	}

	if transport != nil {
		client.Transport = transport
	}

	return &Checker{
		client:       client,
		dockerConfig: dockerConfig,
	}
}

// ImageInfo represents Docker image information
type ImageInfo struct {
	Repository string
	Tag        string
	Digest     string
}

// ParseImageName parses a Docker image name into repository and tag
func ParseImageName(imageName string) (repository, tag string) {
	parts := strings.Split(imageName, ":")
	if len(parts) == 1 {
		return imageName, "latest"
	}
	return parts[0], parts[1]
}

// getDockerHubToken gets authentication token for Docker Hub API
func (c *Checker) getDockerHubToken(ctx context.Context, repository string) (string, error) {
	// Check if we have credentials for Docker Hub
	var auth *config.DockerRegistryAuth
	var foundKey string
	if c.dockerConfig != nil && c.dockerConfig.Registries != nil {
		// Try multiple possible registry keys
		for _, key := range []string{"https://index.docker.io/v1/", "index.docker.io", "docker.io"} {
			if a, ok := c.dockerConfig.Registries[key]; ok {
				auth = &a
				foundKey = key
				break
			}
		}
	}

	if auth == nil {
		// No authentication configured, return empty token (for public images)
		logger.InfoCtx(ctx, "No Docker Hub authentication configured")
		return "", nil
	}

	logger.InfoCtx(ctx, "Found Docker Hub auth config for registry: %s", foundKey)

	// Request authentication token from Docker Hub
	tokenURL := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", repository)

	req, err := http.NewRequestWithContext(ctx, "GET", tokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	// Set Basic Auth
	if auth.Username != "" && auth.Password != "" {
		req.SetBasicAuth(auth.Username, auth.Password)
		logger.InfoCtx(ctx, "Using username/password authentication for Docker Hub")
	} else if auth.Auth != "" {
		// Use pre-encoded auth
		req.Header.Set("Authorization", "Basic "+auth.Auth)
		logger.InfoCtx(ctx, "Using pre-encoded auth for Docker Hub")
	} else {
		logger.WarnCtx(ctx, "Docker Hub auth configured but no credentials provided")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get auth token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.ErrorCtx(ctx, "Docker Hub token request failed with status %d: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("auth token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	logger.InfoCtx(ctx, "Successfully obtained Docker Hub token")

	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return tokenResp.Token, nil
}

// GetImageDigest gets the digest of a Docker image from DockerHub
func (c *Checker) GetImageDigest(ctx context.Context, repository, tag string) (string, error) {
	// For public images, use Docker Hub Registry API v2
	// Format: https://registry.hub.docker.com/v2/{namespace}/{repository}/manifests/{tag}

	// Parse repository to extract namespace and repo name
	parts := strings.Split(repository, "/")
	var namespace, repoName string

	if len(parts) == 1 {
		// No namespace, use "library" (official images)
		namespace = "library"
		repoName = parts[0]
	} else {
		namespace = parts[0]
		repoName = strings.Join(parts[1:], "/")
	}

	fullRepo := fmt.Sprintf("%s/%s", namespace, repoName)

	// Get authentication token
	token, err := c.getDockerHubToken(ctx, fullRepo)
	if err != nil {
		// If we have credentials configured but failed to get token, return error
		if c.dockerConfig != nil && c.dockerConfig.Registries != nil {
			hasAuth := false
			for _, key := range []string{"https://index.docker.io/v1/", "index.docker.io", "docker.io"} {
				if _, ok := c.dockerConfig.Registries[key]; ok {
					hasAuth = true
					break
				}
			}
			if hasAuth {
				return "", fmt.Errorf("failed to get authentication token: %w", err)
			}
		}
		// No credentials configured, try without auth (for public images)
		logger.InfoCtx(ctx, "No Docker Hub credentials configured, trying to access as public image")
	}

	url := fmt.Sprintf("https://registry.hub.docker.com/v2/%s/manifests/%s", fullRepo, tag)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set Accept header to get the digest
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	// Set authorization if we have a token
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get image manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("image not found: %s:%s", repository, tag)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("DockerHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	// The digest is in the Docker-Content-Digest header
	digest := resp.Header.Get("Docker-Content-Digest")
	if digest == "" {
		// If not in header, try to parse from response body
		var manifest map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
			return "", fmt.Errorf("failed to decode manifest: %w", err)
		}

		// Try to get config digest as fallback
		if config, ok := manifest["config"].(map[string]interface{}); ok {
			if d, ok := config["digest"].(string); ok {
				digest = d
			}
		}
	}

	if digest == "" {
		return "", fmt.Errorf("failed to get digest for image: %s:%s", repository, tag)
	}

	logger.InfoCtx(ctx, "Got digest for image %s:%s: %s", repository, tag, digest)
	return digest, nil
}

// CheckImageUpdate checks if there's an update available for an image
// Returns: (hasUpdate, newDigest, error)
func (c *Checker) CheckImageUpdate(ctx context.Context, currentImage, currentDigest string) (bool, string, error) {
	repository, tag := ParseImageName(currentImage)

	newDigest, err := c.GetImageDigest(ctx, repository, tag)
	if err != nil {
		return false, "", err
	}

	// If current digest is empty, we can't compare
	if currentDigest == "" {
		logger.InfoCtx(ctx, "Current digest is empty for %s, treating as no update", currentImage)
		return false, newDigest, nil
	}

	hasUpdate := currentDigest != newDigest
	logger.InfoCtx(ctx, "Image update check for %s: hasUpdate=%v (current=%s, new=%s)",
		currentImage, hasUpdate, currentDigest, newDigest)

	return hasUpdate, newDigest, nil
}

// MatchImageByPrefix checks if an image matches a given prefix
// For example, if prefix is "wavespeed/model-deploy:wan_i2v-default-"
// it will match "wavespeed/model-deploy:wan_i2v-default-202511051642"
func MatchImageByPrefix(imageName, prefix string) bool {
	if prefix == "" {
		return false
	}
	return strings.HasPrefix(imageName, prefix)
}

// ExtractImageTag extracts the tag from a full image name
func ExtractImageTag(imageName string) string {
	_, tag := ParseImageName(imageName)
	return tag
}

// TagsListResponse represents the response from DockerHub tags list API
type TagsListResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// GetLatestImageByPrefix finds the latest image tag matching the given prefix
// Returns: (latestTag, latestDigest, error)
func (c *Checker) GetLatestImageByPrefix(ctx context.Context, imagePrefix string) (string, string, error) {
	// Parse image prefix to get repository
	// Format: "wavespeed/model-deploy:wan_i2v-default-"
	parts := strings.Split(imagePrefix, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid image prefix format, expected repository:tag-prefix, got: %s", imagePrefix)
	}
	repository := parts[0]
	tagPrefix := parts[1]

	// Parse repository to extract namespace and repo name
	repoParts := strings.Split(repository, "/")
	var namespace, repoName string
	if len(repoParts) == 1 {
		namespace = "library"
		repoName = repoParts[0]
	} else {
		namespace = repoParts[0]
		repoName = strings.Join(repoParts[1:], "/")
	}

	fullRepo := fmt.Sprintf("%s/%s", namespace, repoName)

	// Get authentication token
	token, err := c.getDockerHubToken(ctx, fullRepo)
	if err != nil {
		logger.WarnCtx(ctx, "Failed to get Docker Hub token for tag listing: %v", err)
	}

	// List all tags
	url := fmt.Sprintf("https://registry.hub.docker.com/v2/%s/tags/list", fullRepo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization if we have a token
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to get tags list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("DockerHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var tagsResp TagsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return "", "", fmt.Errorf("failed to decode tags response: %w", err)
	}

	// Filter tags by prefix and validate date format
	// Date format after prefix must be exactly 12 digits: YYYYMMDDHHMM
	var matchingTags []string
	for _, tag := range tagsResp.Tags {
		if strings.HasPrefix(tag, tagPrefix) {
			// Extract the suffix after the prefix
			suffix := strings.TrimPrefix(tag, tagPrefix)

			// Validate that suffix is exactly 12 digits (YYYYMMDDHHMM)
			if len(suffix) == 12 {
				allDigits := true
				for _, c := range suffix {
					if c < '0' || c > '9' {
						allDigits = false
						break
					}
				}
				if allDigits {
					matchingTags = append(matchingTags, tag)
				} else {
					logger.InfoCtx(ctx, "Skipping tag %s: suffix contains non-digit characters", tag)
				}
			} else {
				logger.InfoCtx(ctx, "Skipping tag %s: suffix length is %d, expected 12 (YYYYMMDDHHMM)", tag, len(suffix))
			}
		}
	}

	if len(matchingTags) == 0 {
		return "", "", fmt.Errorf("no tags found matching prefix %s with valid date format (YYYYMMDDHHMM)", imagePrefix)
	}

	// Sort tags in descending order (lexicographically, assuming timestamp format)
	// For timestamps like "wan_i2v-default-202512011030", lexicographic sort works
	sort.Slice(matchingTags, func(i, j int) bool {
		return matchingTags[i] > matchingTags[j]
	})

	latestTag := matchingTags[0]
	logger.InfoCtx(ctx, "Found %d matching tags for prefix %s, latest: %s", len(matchingTags), imagePrefix, latestTag)

	// Get digest for the latest tag
	latestImage := fmt.Sprintf("%s:%s", repository, latestTag)
	digest, err := c.GetImageDigest(ctx, repository, latestTag)
	if err != nil {
		return "", "", fmt.Errorf("failed to get digest for %s: %w", latestImage, err)
	}

	return latestImage, digest, nil
}
