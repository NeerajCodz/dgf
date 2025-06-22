package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/NeerajCodz/dgf/types"
)

var ErrPathNotFound = fmt.Errorf("path not found")

// ProcessGitHubURL processes a GitHub URL by parsing it and fetching its structure
func ProcessGitHubURL(url string, token string, branch string, commit string, path string, platform types.Platform) (types.ParsedURL, types.RepositoryStructure, error) {
	// Parse basic URL components
	parsed, err := ParseGitHubURL(url, platform)
	if err != nil {
		return parsed, types.RepositoryStructure{}, fmt.Errorf("failed to parse URL: %v", err)
	}

	// Override with provided values
	if path != "" {
		parsed.Path = path
		pathSegments := strings.Split(path, "/")
		if len(pathSegments) > 1 {
			parsed.ParentPath = strings.Join(pathSegments[:len(pathSegments)-1], "/")
			parsed.RequestPath = pathSegments[len(pathSegments)-1]
		} else {
			parsed.ParentPath = ""
			parsed.RequestPath = path
		}
	}

	// Determine ref
	var ref string
	if commit != "" {
		ref = commit
		parsed.Commit = commit
		parsed.Branch = ""
	} else if branch != "" {
		ref = branch
		parsed.Branch = branch
	} else if parsed.Branch != "" {
		ref = parsed.Branch
	} else {
		defaultBranch, err := fetchDefaultBranch(parsed.Username, parsed.Repo, token)
		if err != nil {
			return parsed, types.RepositoryStructure{}, fmt.Errorf("failed to fetch default branch: %v", err)
		}
		ref = defaultBranch
		parsed.Branch = defaultBranch
	}

	// Reconstruct parsed.url with the final ref and path
	if parsed.Path != "" {
		parsed.URL = fmt.Sprintf("https://github.com/%s/%s/tree/%s/%s", parsed.Username, parsed.Repo, ref, parsed.Path)
	} else {
		parsed.URL = fmt.Sprintf("https://github.com/%s/%s/tree/%s", parsed.Username, parsed.Repo, ref)
	}

	// Determine RequestType if Path is set
	if parsed.Path != "" {
		requestType, err := getRequestType(parsed.Username, parsed.Repo, ref, parsed.ParentPath, parsed.RequestPath, token)
		if err != nil {
			return parsed, types.RepositoryStructure{}, fmt.Errorf("failed to determine request type: %v", err)
		}
		parsed.RequestType = requestType
	}

	// Fetch repository structure
	structure, err := FetchGitHubStructure(parsed.Username, parsed.Repo, ref, parsed.Path, parsed.RequestType, token)
	if err != nil {
		return parsed, structure, err
	}

	return parsed, structure, nil
}

// fetchDefaultBranch fetches the default branch of a repository using the GitHub API
func fetchDefaultBranch(owner, repo, token string) (string, error) {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch repo info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch repo info: status %d - check repository owner (%s), repo (%s), or token permissions", resp.StatusCode, owner, repo)
	}

	var repoInfo struct {
		DefaultBranch string `json:"default_branch"`
	}
	err = json.NewDecoder(resp.Body).Decode(&repoInfo)
	if err != nil {
		return "", fmt.Errorf("failed to decode repo info: %v", err)
	}

	if repoInfo.DefaultBranch == "" {
		return "", fmt.Errorf("no default branch found for %s/%s", owner, repo)
	}

	return repoInfo.DefaultBranch, nil
}

// PrintStructure prints the repository structure
func PrintStructure(structure types.RepositoryStructure) {
	fmt.Println("Files:")
	for i, file := range structure.Files {
		fmt.Printf("  %s (Name: %s, Size: %d, SHA: %s, URL: %s, HTML URL: %s, Git URL: %s, Download URL: %s, Request Path: %s)\n",
			file, structure.FilesName[i], structure.FilesSize[i], structure.FilesSha[i],
			structure.FilesURL[i], structure.FilesHTMLURL[i], structure.FilesGitURL[i], structure.DownloadURLs[i], structure.FilesRequest[i])
	}
	fmt.Println("Folders:")
	for _, folder := range structure.Folders {
		fmt.Printf("  %s\n", folder)
	}
}
