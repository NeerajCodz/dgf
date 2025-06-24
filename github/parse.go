package github

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/NeerajCodz/dgf/types"
)

// ParseGitHubURL parses a GitHub URL or constructs one from site arguments
func ParseGitHubURL(url string, platform types.Platform, args types.Args) (types.ParsedURL, error) {
	result := types.ParsedURL{
		URL:  url,
		Name: platform.Name,
		ID:   platform.ID,
	}

	// Check if site arguments are provided
	hasSiteArgs := args.Site != "" || args.Username != "" || args.Repo != ""
	hasURL := url != ""

	// If site args are provided, construct the URL
	if hasSiteArgs {
		if args.Site == "" || args.Username == "" || args.Repo == "" {
			return result, fmt.Errorf("must provide all of --site, --username, and --repo")
		}

		// Read config/git.json to validate platform
		configData, err := os.ReadFile("config/git.json")
		if err != nil {
			return result, fmt.Errorf("failed to read config file: %v", err)
		}

		var platforms []types.Platform
		if err := json.Unmarshal(configData, &platforms); err != nil {
			return result, fmt.Errorf("failed to parse config file: %v", err)
		}

		// Find matching platform (case-insensitive)
		siteID := strings.ToLower(args.Site)
		var selectedPlatform types.Platform
		for _, p := range platforms {
			if p.ID == siteID {
				selectedPlatform = p
				break
			}
		}
		if selectedPlatform.ID == "" {
			return result, fmt.Errorf("invalid site ID '%s'", args.Site)
		}

		// Construct base URL
		result.URL = selectedPlatform.URLStruc.Site
		result.URL = strings.ReplaceAll(result.URL, "<username>", args.Username)
		result.URL = strings.ReplaceAll(result.URL, "<repo>", args.Repo)
		result.Name = selectedPlatform.Name
		result.ID = selectedPlatform.ID
		result.Username = args.Username
		result.Repo = args.Repo
		return result, nil
	}

	// If URL is provided, parse it
	if hasURL {
		normalizedURL := url
		if strings.HasPrefix(url, "github.com/") {
			normalizedURL = "https://" + url
		} else if !strings.HasPrefix(url, "http://github.com/") && !strings.HasPrefix(url, "https://github.com/") {
			return result, fmt.Errorf("invalid URL format: does not match GitHub site")
		}

		var baseURL string
		for _, site := range platform.URL.Site {
			if strings.HasPrefix(normalizedURL, site) {
				baseURL = site
				break
			}
		}
		if baseURL == "" {
			return result, fmt.Errorf("invalid URL format: does not match GitHub site")
		}

		path := strings.TrimPrefix(normalizedURL, baseURL)
		segments := strings.Split(strings.Trim(path, "/"), "/")

		if len(segments) < 2 {
			return result, fmt.Errorf("invalid GitHub URL structure: missing username or repo")
		}

		result.Username = segments[0]
		result.Repo = segments[1]

		// Parse branch or commit and path if present
		if len(segments) >= 4 && (segments[2] == "blob" || segments[2] == "tree") {
			// Check if segments[3] is a commit hash (e.g., 40 characters for SHA-1)
			if len(segments[3]) >= 7 && isPotentialCommitHash(segments[3]) {
				result.Commit = segments[3]
			} else {
				result.Branch = segments[3]
			}
			if len(segments) > 4 {
				fullPath := strings.Join(segments[4:], "/")
				pathSegments := strings.Split(fullPath, "/")
				if len(pathSegments) > 1 {
					result.ParentPath = strings.Join(pathSegments[:len(pathSegments)-1], "/")
					result.RequestPath = pathSegments[len(pathSegments)-1]
				} else {
					result.ParentPath = ""
					result.RequestPath = pathSegments[0]
				}
				result.Path = fullPath
			}
		}
	}

	return result, nil
}

// isPotentialCommitHash checks if a string could be a Git commit hash
func isPotentialCommitHash(s string) bool {
	// Git commit hashes are typically 40 characters (SHA-1) or at least 7 characters for short hashes
	if len(s) < 7 || len(s) > 40 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}