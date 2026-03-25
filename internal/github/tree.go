package github

import (
	"path/filepath"
	"strings"

	"github.com/NeerajCodz/dgf/pkg/types"
)

// FetchStructure fetches the complete repository structure
func FetchStructure(client *Client, owner, repo, ref, path, requestType string, formats []string) (types.RepositoryStructure, error) {
	owner = strings.ToLower(owner)
	repo = strings.ToLower(repo)

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

	// Single file request
	if requestType == "file" && path != "" {
		content, err := client.FetchFile(owner, repo, ref, path)
		if err != nil {
			return structure, err
		}

		if !matchesFormat(content.Name, formats) {
			return structure, nil
		}

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
		structure.FilesRequest = []string{content.Name}
		return structure, nil
	}

	// Directory request
	contents, err := client.FetchContents(owner, repo, ref, path)
	if err != nil {
		return structure, err
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
			if !matchesFormat(content.Name, formats) {
				continue
			}

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
			folderRequestPath := requestItemPath

			subStructure, err := FetchStructure(client, owner, repo, ref, itemPath, "dir", formats)
			if err != nil {
				return structure, err
			}

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

// FetchItems fetches repository items for the TUI browser
func FetchItems(client *Client, owner, repo, ref, path string) ([]types.RepoItem, error) {
	contents, err := client.FetchContents(owner, repo, ref, path)
	if err != nil {
		return nil, err
	}

	items := make([]types.RepoItem, 0, len(contents))
	for _, c := range contents {
		downloadURL := ""
		if c.DownloadURL != nil {
			downloadURL = *c.DownloadURL
		}
		
		items = append(items, types.RepoItem{
			Name:        c.Name,
			Path:        c.Path,
			Type:        c.Type,
			Size:        int64(c.Size),
			Sha:         c.Sha,
			DownloadURL: downloadURL,
			Selected:    false,
		})
	}

	return items, nil
}

// GetRequestType determines whether a path is a file or directory
func GetRequestType(client *Client, owner, repo, ref, parentPath, requestPath string) (string, error) {
	if requestPath == "" {
		return "", nil
	}

	owner = strings.ToLower(owner)
	repo = strings.ToLower(repo)

	fullPath := requestPath
	if parentPath != "" {
		fullPath = parentPath + "/" + requestPath
	}

	// Check parent path contents if available
	if parentPath != "" {
		contents, err := client.FetchContents(owner, repo, ref, parentPath)
		if err != nil {
			if err == ErrPathNotFound {
				return "", err
			}
			return "", err
		}
		for _, content := range contents {
			if content.Name == requestPath {
				return content.Type, nil
			}
		}
		return "", ErrPathNotFound
	}

	// Try as directory first
	contents, err := client.FetchContents(owner, repo, ref, fullPath)
	if err == nil && len(contents) > 0 {
		return "dir", nil
	} else if err != nil && err != ErrPathNotFound {
		return "", err
	}

	// Try as file
	content, err := client.FetchFile(owner, repo, ref, fullPath)
	if err == nil && content.Type == "file" {
		return "file", nil
	} else if err != ErrPathNotFound {
		return "", err
	}

	return "", ErrPathNotFound
}

// matchesFormat checks if a filename matches the requested formats
func matchesFormat(fileName string, formats []string) bool {
	if len(formats) == 0 {
		return true
	}
	if len(formats) == 1 && formats[0] == "" {
		return filepath.Ext(fileName) == ""
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(fileName), "."))
	for _, f := range formats {
		if strings.ToLower(f) == ext {
			return true
		}
	}
	return false
}
