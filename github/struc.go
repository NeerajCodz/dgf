package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/NeerajCodz/dgf/types"
)

// FetchGitHubStructure fetches the repository structure for the specified path
func FetchGitHubStructure(owner, repo, ref, path, requestType, token string) (types.RepositoryStructure, error) {
	var parentPath string
	if path != "" {
		pathSegments := strings.Split(path, "/")
		if len(pathSegments) > 1 {
			parentPath = strings.Join(pathSegments[:len(pathSegments)-1], "/")
		}
	}

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

	// If RequestType is file, set Files to the path and fetch file details
	if requestType == "file" && path != "" {
		content, err := fetchSingleFile(owner, repo, ref, path, token)
		if err != nil {
			if err == ErrPathNotFound {
				return structure, ErrPathNotFound
			}
			return structure, fmt.Errorf("failed to fetch file details for %s: %v", path, err)
		}
		requestPath := content.Name // Use Name as RequestPath for single file
		structure.Files = []string{path}
		structure.FilesName = []string{content.Name}
		structure.FilesSha = []string{content.Sha}
		structure.FilesHTMLURL = []string{content.HTMLURL}
		structure.FilesGitURL = []string{content.GitURL}
		structure.FilesURL = []string{content.URL}
		structure.FilesSize = []int{content.Size}
		if content.DownloadURL != "" {
			structure.DownloadURLs = []string{content.DownloadURL}
		} else {
			structure.DownloadURLs = []string{""}
		}
		structure.FilesRequest = []string{requestPath}
		return structure, nil
	}

	// If path is empty or not a file, proceed with directory fetching
	if path == "" {
		return structure, nil
	}

	contents, err := FetchGitHubContents(owner, repo, ref, path, token)
	if err != nil {
		if err == ErrPathNotFound {
			return structure, ErrPathNotFound
		}
		return structure, fmt.Errorf("failed to fetch contents for path %s: %v", path, err)
	}

	for _, content := range contents {
		itemPath := content.Path
		var requestItemPath string
		if parentPath != "" && strings.HasPrefix(itemPath, parentPath+"/") {
			requestItemPath = strings.TrimPrefix(itemPath, parentPath+"/")
		} else {
			requestItemPath = itemPath
		}

		if content.Type == "file" {
			structure.Files = append(structure.Files, itemPath)
			structure.FilesName = append(structure.FilesName, content.Name)
			structure.FilesSha = append(structure.FilesSha, content.Sha)
			structure.FilesHTMLURL = append(structure.FilesHTMLURL, content.HTMLURL)
			structure.FilesGitURL = append(structure.FilesGitURL, content.GitURL)
			structure.FilesURL = append(structure.FilesURL, content.URL)
			structure.FilesSize = append(structure.FilesSize, content.Size)
			if content.DownloadURL != "" {
				structure.DownloadURLs = append(structure.DownloadURLs, content.DownloadURL)
			} else {
				structure.DownloadURLs = append(structure.DownloadURLs, "")
			}
			structure.FilesRequest = append(structure.FilesRequest, requestItemPath)
		} else if content.Type == "dir" {
			// Only append RequestPath for directories
			var folderRequestPath string
			if parentPath != "" && strings.HasPrefix(itemPath, parentPath+"/") {
				folderRequestPath = strings.TrimPrefix(itemPath, parentPath+"/")
			} else {
				folderRequestPath = itemPath
			}
			structure.Folders = append(structure.Folders, folderRequestPath)
			// Recursively fetch contents of directories
			subStructure, err := FetchGitHubStructure(owner, repo, ref, itemPath, "dir", token)
			if err != nil {
				return structure, err
			}
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

	return structure, nil
}

// fetchSingleFile fetches details for a single file
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

	err = json.NewDecoder(resp.Body).Decode(&content)
	if err != nil {
		return content, fmt.Errorf("failed to decode file details: %v", err)
	}

	return content, nil
}
