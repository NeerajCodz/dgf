package github

import (
	"strings"

	"github.com/NeerajCodz/dgf/pkg/types"
)

// ParseURL parses a GitHub URL into its components
func ParseURL(url string, platform types.Platform) (types.ParsedURL, error) {
	result := types.ParsedURL{
		URL:  url,
		Name: platform.Name,
		ID:   platform.ID,
	}

	// Normalize URL
	normalizedURL := url
	if strings.HasPrefix(url, "github.com/") {
		normalizedURL = "https://" + url
	} else if !strings.HasPrefix(url, "http://github.com/") && !strings.HasPrefix(url, "https://github.com/") {
		return result, ErrPathNotFound
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
		return result, ErrPathNotFound
	}

	// Parse path segments
	path := strings.TrimPrefix(normalizedURL, baseURL)
	segments := strings.Split(strings.Trim(path, "/"), "/")

	if len(segments) < 2 {
		return result, ErrPathNotFound
	}

	result.Username = segments[0]
	result.Repo = segments[1]

	// Parse branch/commit and path if present
	if len(segments) >= 4 && (segments[2] == "blob" || segments[2] == "tree") {
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

	return result, nil
}

// ParseFromArgs constructs a ParsedURL from command-line arguments
func ParseFromArgs(args types.Args, platform types.Platform) (types.ParsedURL, error) {
	result := types.ParsedURL{
		Name:     platform.Name,
		ID:       platform.ID,
		Username: args.Username,
		Repo:     args.Repo,
	}

	// Construct URL from template
	result.URL = platform.URLStruc.Site
	result.URL = strings.ReplaceAll(result.URL, "<username>", args.Username)
	result.URL = strings.ReplaceAll(result.URL, "<repo>", args.Repo)

	return result, nil
}

// isPotentialCommitHash checks if a string could be a Git commit hash
func isPotentialCommitHash(s string) bool {
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

// BuildURL constructs a GitHub URL from components
func BuildURL(username, repo, ref, path string) string {
	url := "https://github.com/" + username + "/" + repo
	if ref != "" {
		url += "/tree/" + ref
		if path != "" {
			url += "/" + path
		}
	}
	return url
}
