package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/NeerajCodz/dgf/types"
)

// FetchGitHubContents fetches directory contents from GitHub API
func FetchGitHubContents(owner, repo, ref, path, token string) ([]types.GitHubContent, error) {
	// Normalize repository name for API (case-insensitive)
	owner = strings.ToLower(owner)
	repo = strings.ToLower(repo)
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", owner, repo)
	if path != "" {
		api += "/" + path
	}
	if ref != "" {
		api += "?ref=" + ref
	}
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	if token != "" {
		req.Header.Add("Authorization", "token " + token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contents: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, ErrPathNotFound
	} else if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%d %s: %s", resp.StatusCode, http.StatusText(resp.StatusCode), string(body))
	}

	var contents []types.GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, fmt.Errorf("failed to decode contents: %v", err)
	}

	return contents, nil
}

// getRequestType determines whether a path is a file or directory
func getRequestType(url string, owner, repo, ref, parentPath, requestPath, token string) (string, error) {
	if requestPath == "" {
		return "", nil
	}

	// Normalize owner and repo for API
	owner = strings.ToLower(owner)
	repo = strings.ToLower(repo)

	// Construct full path
	fullPath := requestPath
	if parentPath != "" {
		fullPath = parentPath + "/" + requestPath
	}

	// If parentPath is provided, check its contents for requestPath
	if parentPath != "" {
		contents, err := FetchGitHubContents(owner, repo, ref, parentPath, token)
		if err != nil {
			if err == ErrPathNotFound {
				return "", fmt.Errorf("parent path %s not found", parentPath)
			}
			return "", fmt.Errorf("failed to fetch parent path %s: %v", parentPath, err)
		}
		for _, content := range contents {
			if content.Name == requestPath {
				return content.Type, nil
			}
		}
		return "", fmt.Errorf("request path %s not found in parent path %s", requestPath, parentPath)
	}

	// If no parentPath, check if fullPath is a directory
	contents, err := FetchGitHubContents(owner, repo, ref, fullPath, token)
	if err == nil && len(contents) > 0 {
		return "dir", nil
	} else if err != nil && err != ErrPathNotFound {
		return "", fmt.Errorf("failed to fetch directory contents for path %s: %v", fullPath, err)
	}

	// If not a directory, check if it's a file
	content, err := fetchSingleFile(owner, repo, ref, fullPath, token)
	if err == nil && content.Type == "file" {
		return "file", nil
	} else if err != ErrPathNotFound {
		return "", fmt.Errorf("failed to fetch file details for path %s: %v", fullPath, err)
	}

	return "", ErrPathNotFound
}