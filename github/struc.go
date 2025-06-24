package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/NeerajCodz/dgf/types"
)

// FetchGitHubStructure fetches the repository structure, filtering files by format if specified
func FetchGitHubStructure(owner, repo, ref, path, requestType, token string, args types.Args) (types.RepositoryStructure, error) {
	// Normalize owner and repo for API
	owner = strings.ToLower(owner)
	repo = strings.ToLower(repo)

	// Determine parent path for relative path construction
	var parentPath string
	if path != "" {
		pathSegments := strings.Split(path, "/")
		if len(pathSegments) > 1 {
			parentPath = strings.Join(pathSegments[:len(pathSegments)-1], "/")
		}
	}

	// Initialize an empty repository structure
	structure := types.RepositoryStructure{
		Files:        []string{},
		FilesName:    []string{},
		FilesSha:     []string{},
		FilesHTMLURL: []string{},
		FilesGitURL:  []string{},
		FilesURL:     []string{},
		FilesSize:    []int{},
		Folders:      []string{},
		DownloadURLs: []string{},
		FilesRequest: []string{},
	}

	// Handle single file request
	if requestType == "file" && path != "" {
		content, err := fetchSingleFile(owner, repo, ref, path, token)
		if err != nil {
			if err == ErrPathNotFound {
				return structure, ErrPathNotFound
			}
			return structure, fmt.Errorf("failed to fetch file details for %s: %v", path, err)
		}

		// Apply format filtering
		if len(args.Formats) == 1 && args.Formats[0] == "" {
			// -f "" means only files with no extension
			if filepath.Ext(content.Name) != "" {
				return structure, nil
			}
		} else if len(args.Formats) > 0 {
			// -f image or -f [jpg,pdf]
			ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(content.Name), "."))
			if !contains(args.Formats, ext) {
				return structure, nil
			}
		}

		// Populate structure with file details
		requestPath := content.Name
		structure.Files = []string{path}
		structure.FilesName = []string{content.Name}
		structure.FilesSha = []string{content.Sha}
		structure.FilesHTMLURL = []string{content.HTMLURL}
		structure.FilesGitURL = []string{content.GitURL}
		structure.FilesURL = []string{content.URL}
		structure.FilesSize = []int{content.Size}
		if content.DownloadURL != nil {
			structure.DownloadURLs = []string{*content.DownloadURL}
		} else {
			structure.DownloadURLs = []string{""}
		}
		structure.FilesRequest = []string{requestPath}
		return structure, nil
	}

	// Fetch contents (root or specified path)
	contents, err := FetchGitHubContents(owner, repo, ref, path, token)
	if err != nil {
		if err == ErrPathNotFound {
			return structure, ErrPathNotFound
		}
		return structure, fmt.Errorf("failed to fetch contents for path %s: %v", path, err)
	}

	// Process each item in the directory
	for _, content := range contents {
		itemPath := content.Path
		var requestItemPath string
		if parentPath != "" && strings.HasPrefix(itemPath, parentPath+"/") {
			requestItemPath = strings.TrimPrefix(itemPath, parentPath+"/")
		} else {
			requestItemPath = itemPath
		}

		if content.Type == "file" {
			// Apply format filtering
			if len(args.Formats) == 1 && args.Formats[0] == "" {
				// -f "" means only files with no extension
				if filepath.Ext(content.Name) != "" {
					continue
				}
			} else if len(args.Formats) > 0 {
				// -f image or -f [jpg,pdf]
				ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(content.Name), "."))
				if !contains(args.Formats, ext) {
					continue
				}
			}

			// Add file to structure
			structure.Files = append(structure.Files, itemPath)
			structure.FilesName = append(structure.FilesName, content.Name)
			structure.FilesSha = append(structure.FilesSha, content.Sha)
			structure.FilesHTMLURL = append(structure.FilesHTMLURL, content.HTMLURL)
			structure.FilesGitURL = append(structure.FilesGitURL, content.GitURL)
			structure.FilesURL = append(structure.FilesURL, content.URL)
			structure.FilesSize = append(structure.FilesSize, content.Size)
			if content.DownloadURL != nil {
				structure.DownloadURLs = append(structure.DownloadURLs, *content.DownloadURL)
			} else {
				structure.DownloadURLs = append(structure.DownloadURLs, "")
			}
			structure.FilesRequest = append(structure.FilesRequest, requestItemPath)
		} else if content.Type == "dir" {
			// Add directory to structure only if it contains matching files
			var folderRequestPath string
			if parentPath != "" && strings.HasPrefix(itemPath, parentPath+"/") {
				folderRequestPath = strings.TrimPrefix(itemPath, parentPath+"/")
			} else {
				folderRequestPath = itemPath
			}

			// Recursively fetch subdirectory contents
			subStructure, err := FetchGitHubStructure(owner, repo, ref, itemPath, "dir", token, args)
			if err != nil {
				return structure, err
			}

			// Only add folder if it contains files or subfolders with matching files
			if len(subStructure.Files) > 0 || len(subStructure.Folders) > 0 {
				structure.Folders = append(structure.Folders, folderRequestPath)
				structure.Files = append(structure.Files, subStructure.Files...)
				structure.FilesName = append(structure.FilesName, subStructure.FilesName...)
				structure.FilesSha = append(structure.FilesSha, subStructure.FilesSha...)
				structure.FilesHTMLURL = append(structure.FilesHTMLURL, subStructure.FilesHTMLURL...)
				structure.FilesGitURL = append(structure.FilesGitURL, subStructure.FilesGitURL...)
				structure.FilesURL = append(structure.FilesURL, subStructure.FilesURL...)
				structure.FilesSize = append(structure.FilesSize, subStructure.FilesSize...)
				structure.Folders = append(structure.Folders, subStructure.Folders...)
				structure.DownloadURLs = append(structure.DownloadURLs, subStructure.DownloadURLs...)
				structure.FilesRequest = append(structure.FilesRequest, subStructure.FilesRequest...)
			}
		}
	}

	return structure, nil
}

// fetchSingleFile fetches details for a single file from GitHub API
func fetchSingleFile(owner, repo, ref, path, token string) (types.GitHubContent, error) {
	var content types.GitHubContent
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	if ref != "" {
		api += "?ref=" + ref
	}
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return content, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return content, fmt.Errorf("failed to fetch file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return content, ErrPathNotFound
	} else if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return content, fmt.Errorf("%d %s: %s", resp.StatusCode, http.StatusText(resp.StatusCode), string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return content, fmt.Errorf("failed to decode file details: %v", err)
	}

	return content, nil
}

// contains checks if a string slice contains a specific item
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}