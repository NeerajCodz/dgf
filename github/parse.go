package github

import (
	"fmt"
	"strings"

	"github.com/NeerajCodz/dgf/types"
)

// ParseGitHubURL parses a GitHub URL for username, repo, branch, parent_path, request_path, and request_type
func ParseGitHubURL(url string, platform types.Platform) (types.ParsedURL, error) {
	result := types.ParsedURL{
		URL:  url,
		Name: platform.Name,
		ID:   platform.ID,
	}

	// Normalize URL by ensuring it has a scheme
	normalizedURL := url
	if strings.HasPrefix(url, "github.com/") {
		normalizedURL = "https://" + url
	} else if strings.HasPrefix(url, "http://github.com/") || strings.HasPrefix(url, "https://github.com/") {
		// Already has a scheme
	} else {
		return result, fmt.Errorf("invalid URL format: does not match platform site")
	}

	// Find matching base URL
	var baseURL string
	for _, site := range platform.URL.Site {
		if strings.HasPrefix(normalizedURL, site) {
			baseURL = site
			break
		}
	}
	if baseURL == "" {
		return result, fmt.Errorf("invalid URL format: does not match platform site")
	}

	// Extract path after base URL
	path := strings.TrimPrefix(normalizedURL, baseURL)
	segments := strings.Split(strings.Trim(path, "/"), "/")

	// Minimum requirement: <username>/<repo>
	if len(segments) < 2 {
		return result, fmt.Errorf("invalid GitHub URL structure: missing username or repo")
	}

	result.Username = segments[0]
	result.Repo = segments[1]

	// Check for branch and path (e.g., /blob/<branch> or /tree/<branch>)
	if len(segments) >= 4 && (segments[2] == "blob" || segments[2] == "tree") {
		result.Branch = segments[3]
		// Extract path segments after branch
		if len(segments) > 4 {
			fullPath := strings.Join(segments[4:], "/")
			// Split into ParentPath and RequestPath
			pathSegments := strings.Split(fullPath, "/")
			if len(pathSegments) > 1 {
				result.ParentPath = strings.Join(pathSegments[:len(pathSegments)-1], "/")
				result.RequestPath = pathSegments[len(pathSegments)-1]
			} else {
				result.ParentPath = ""
				result.RequestPath = pathSegments[0]
			}
			// Set Path as ParentPath + "/" + RequestPath
			if result.ParentPath != "" {
				result.Path = result.ParentPath + "/" + result.RequestPath
			} else {
				result.Path = result.RequestPath
			}
		}
	}

	return result, nil
}
