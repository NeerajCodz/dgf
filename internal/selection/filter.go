package selection

import (
	"strings"

	"github.com/NeerajCodz/dgf/pkg/types"
)

// Filter filters items based on a search query
func Filter(items []types.RepoItem, query string) []types.RepoItem {
	if query == "" {
		return items
	}

	query = strings.ToLower(query)
	filtered := make([]types.RepoItem, 0)

	for _, item := range items {
		if matchesQuery(item.Name, query) || matchesQuery(item.Path, query) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// matchesQuery performs case-insensitive substring matching
func matchesQuery(text, query string) bool {
	return strings.Contains(strings.ToLower(text), query)
}

// HighlightMatches returns indices of matching characters for highlighting
func HighlightMatches(text, query string) []int {
	if query == "" {
		return nil
	}

	textLower := strings.ToLower(text)
	queryLower := strings.ToLower(query)

	idx := strings.Index(textLower, queryLower)
	if idx == -1 {
		return nil
	}

	indices := make([]int, len(query))
	for i := range query {
		indices[i] = idx + i
	}
	return indices
}

// FuzzyMatch performs fuzzy matching and returns a score
func FuzzyMatch(text, query string) (bool, int) {
	if query == "" {
		return true, 0
	}

	textLower := strings.ToLower(text)
	queryLower := strings.ToLower(query)

	// Exact match gets highest score
	if strings.Contains(textLower, queryLower) {
		return true, 100 - len(text) // Shorter matches score higher
	}

	// Character-by-character fuzzy matching
	textIdx := 0
	matches := 0
	for _, qc := range queryLower {
		for textIdx < len(textLower) {
			if rune(textLower[textIdx]) == qc {
				matches++
				textIdx++
				break
			}
			textIdx++
		}
	}

	if matches == len(query) {
		return true, matches * 2 // Score based on how many matched
	}

	return false, 0
}

// FilterByExtension filters items by file extension
func FilterByExtension(items []types.RepoItem, extensions []string) []types.RepoItem {
	if len(extensions) == 0 {
		return items
	}

	// Normalize extensions
	extMap := make(map[string]bool)
	for _, ext := range extensions {
		extMap[strings.ToLower(strings.TrimPrefix(ext, "."))] = true
	}

	filtered := make([]types.RepoItem, 0)
	for _, item := range items {
		if item.IsDir() {
			filtered = append(filtered, item)
			continue
		}

		ext := item.Extension()
		if extMap[strings.ToLower(ext)] {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// FilterDirectoriesOnly returns only directories
func FilterDirectoriesOnly(items []types.RepoItem) []types.RepoItem {
	filtered := make([]types.RepoItem, 0)
	for _, item := range items {
		if item.IsDir() {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// FilterFilesOnly returns only files
func FilterFilesOnly(items []types.RepoItem) []types.RepoItem {
	filtered := make([]types.RepoItem, 0)
	for _, item := range items {
		if item.IsFile() {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
