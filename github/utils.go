package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/NeerajCodz/dgf/types"
)

// FetchGitHubContents fetches contents from a GitHub API URL and returns the JSON data
func FetchGitHubContents(owner, repo, ref, path, token string) ([]types.GitHubContent, error) {
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
		req.Header.Add("Authorization", "token "+token)
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
	err = json.NewDecoder(resp.Body).Decode(&contents)
	if err != nil {
		return nil, fmt.Errorf("failed to decode contents: %v", err)
	}

	return contents, nil
}

// getRequestType determines if the request path is a file or directory
func getRequestType(owner, repo, ref, parentPath, requestPath, token string) (string, error) {
	if requestPath == "" {
		return "", nil // No request path, so no type
	}

	// Fetch contents of the parent path to find the request path
	fetchPath := parentPath
	if fetchPath == "" {
		fetchPath = requestPath // If no parent path, check root
	}
	contents, err := FetchGitHubContents(owner, repo, ref, fetchPath, token)
	if err != nil {
		if err == ErrPathNotFound {
			return "", ErrPathNotFound
		}
		return "", fmt.Errorf("failed to fetch contents for type check: %v", err)
	}

	for _, content := range contents {
		comparePath := content.Path
		if parentPath != "" {
			if strings.HasPrefix(content.Path, parentPath+"/") {
				comparePath = strings.TrimPrefix(content.Path, parentPath+"/")
			} else {
				continue
			}
		}
		if comparePath == requestPath {
			return content.Type, nil
		}
	}

	return "", ErrPathNotFound
}
