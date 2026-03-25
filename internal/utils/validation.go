package utils

import (
	"regexp"
	"strings"
)

// ValidateGitHubURL validates if a string is a valid GitHub URL
func ValidateGitHubURL(url string) bool {
	if url == "" {
		return false
	}

	// Normalize URL
	normalized := url
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		normalized = "https://" + url
	}

	// Match GitHub URL patterns
	patterns := []string{
		`^https?://github\.com/[\w\-\.]+/[\w\-\.]+/?.*$`,
		`^https?://www\.github\.com/[\w\-\.]+/[\w\-\.]+/?.*$`,
	}

	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, normalized)
		if err == nil && matched {
			return true
		}
	}

	return false
}

// NormalizePath normalizes a repository path
func NormalizePath(path string) string {
	// Remove leading and trailing slashes
	path = strings.Trim(path, "/")
	// Remove duplicate slashes
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	return path
}

// NormalizeURL normalizes a GitHub URL
func NormalizeURL(url string) string {
	// Add https:// if missing
	if strings.HasPrefix(url, "github.com/") {
		url = "https://" + url
	}
	// Remove trailing slash
	url = strings.TrimSuffix(url, "/")
	return url
}

// ExtractExtension extracts file extension from filename
func ExtractExtension(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return strings.ToLower(filename[i+1:])
		}
		if filename[i] == '/' {
			break
		}
	}
	return ""
}

// TruncateString truncates a string to maxLen with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// ContainsIgnoreCase checks if slice contains item (case-insensitive)
func ContainsIgnoreCase(slice []string, item string) bool {
	item = strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == item {
			return true
		}
	}
	return false
}
