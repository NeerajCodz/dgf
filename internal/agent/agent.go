package agent

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/NeerajCodz/dgf/internal/github"
	"github.com/NeerajCodz/dgf/pkg/types"
)

// Envelope is the standard JSON response wrapper
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// TreeItem represents a file/folder in the tree output
type TreeItem struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	Size        int64  `json:"size"`
	Sha         string `json:"sha"`
	DownloadURL string `json:"download_url,omitempty"`
}

// DownloadResult represents the result of downloading a file
type DownloadResult struct {
	Path    string `json:"path"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// outputJSON outputs a JSON response and exits
func outputJSON(envelope Envelope) {
	data, _ := json.MarshalIndent(envelope, "", "  ")
	fmt.Println(string(data))
	if envelope.Success {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

// Tree fetches and outputs the repository structure as JSON
func Tree(url, token string) {
	platform := types.Platform{
		Name: "GitHub",
		ID:   "github",
		URL: types.URL{
			Site: []string{"https://github.com"},
		},
	}

	parsed, err := github.ParseURL(url, platform)
	if err != nil {
		outputJSON(Envelope{Success: false, Error: fmt.Sprintf("failed to parse URL: %v", err)})
		return
	}

	client := github.NewClient(token)

	// Get default branch if needed
	ref := parsed.Branch
	if ref == "" && parsed.Commit == "" {
		defaultBranch, err := client.FetchDefaultBranch(parsed.Username, parsed.Repo)
		if err != nil {
			outputJSON(Envelope{Success: false, Error: fmt.Sprintf("failed to fetch default branch: %v", err)})
			return
		}
		ref = defaultBranch
	} else if parsed.Commit != "" {
		ref = parsed.Commit
	}

	// Fetch items recursively
	items, err := fetchTreeRecursive(client, parsed.Username, parsed.Repo, ref, parsed.Path)
	if err != nil {
		outputJSON(Envelope{Success: false, Error: fmt.Sprintf("failed to fetch tree: %v", err)})
		return
	}

	outputJSON(Envelope{Success: true, Data: items})
}

// fetchTreeRecursive recursively fetches all items in a directory
func fetchTreeRecursive(client *github.Client, owner, repo, ref, path string) ([]TreeItem, error) {
	items, err := github.FetchItems(client, owner, repo, ref, path)
	if err != nil {
		return nil, err
	}

	result := make([]TreeItem, 0)
	for _, item := range items {
		treeItem := TreeItem{
			Name:        item.Name,
			Path:        item.Path,
			Type:        item.Type,
			Size:        item.Size,
			Sha:         item.Sha,
			DownloadURL: item.DownloadURL,
		}
		result = append(result, treeItem)

		// Recurse into directories
		if item.IsDir() {
			subItems, err := fetchTreeRecursive(client, owner, repo, ref, item.Path)
			if err != nil {
				return nil, err
			}
			result = append(result, subItems...)
		}
	}

	return result, nil
}

// Download downloads specified paths and outputs results as JSON
func Download(url, token, outputDir string, paths []string) {
	platform := types.Platform{
		Name: "GitHub",
		ID:   "github",
		URL: types.URL{
			Site: []string{"https://github.com"},
		},
	}

	parsed, err := github.ParseURL(url, platform)
	if err != nil {
		outputJSON(Envelope{Success: false, Error: fmt.Sprintf("failed to parse URL: %v", err)})
		return
	}

	client := github.NewClient(token)

	// Get default branch if needed
	ref := parsed.Branch
	if ref == "" && parsed.Commit == "" {
		defaultBranch, err := client.FetchDefaultBranch(parsed.Username, parsed.Repo)
		if err != nil {
			outputJSON(Envelope{Success: false, Error: fmt.Sprintf("failed to fetch default branch: %v", err)})
			return
		}
		ref = defaultBranch
	} else if parsed.Commit != "" {
		ref = parsed.Commit
	}

	// Fetch all items to find download URLs
	allItems, err := fetchTreeRecursive(client, parsed.Username, parsed.Repo, ref, "")
	if err != nil {
		outputJSON(Envelope{Success: false, Error: fmt.Sprintf("failed to fetch tree: %v", err)})
		return
	}

	// Build map of path to item
	itemMap := make(map[string]TreeItem)
	for _, item := range allItems {
		itemMap[item.Path] = item
	}

	// Build download structure
	structure := types.RepositoryStructure{
		Files:        make([]string, 0),
		DownloadURLs: make([]string, 0),
		FilesRequest: make([]string, 0),
		FilesSize:    make([]int, 0),
	}

	results := make([]DownloadResult, 0)
	for _, path := range paths {
		item, exists := itemMap[path]
		if !exists {
			results = append(results, DownloadResult{Path: path, Success: false, Error: "path not found"})
			continue
		}
		if item.Type == "dir" {
			// Add all files under this directory
			for _, subItem := range allItems {
				if subItem.Type == "file" && (subItem.Path == path || len(subItem.Path) > len(path) && subItem.Path[:len(path)+1] == path+"/") {
					structure.Files = append(structure.Files, subItem.Path)
					structure.DownloadURLs = append(structure.DownloadURLs, subItem.DownloadURL)
					structure.FilesRequest = append(structure.FilesRequest, subItem.Path)
					structure.FilesSize = append(structure.FilesSize, int(subItem.Size))
				}
			}
		} else {
			structure.Files = append(structure.Files, item.Path)
			structure.DownloadURLs = append(structure.DownloadURLs, item.DownloadURL)
			structure.FilesRequest = append(structure.FilesRequest, item.Path)
			structure.FilesSize = append(structure.FilesSize, int(item.Size))
		}
	}

	// Download
	if len(structure.Files) > 0 {
		err = github.Download(structure, github.DownloadOptions{
			OutputDir: outputDir,
			Token:     token,
			Workers:   5,
			Silent:    true,
		})
		if err != nil {
			for _, f := range structure.Files {
				results = append(results, DownloadResult{Path: f, Success: false, Error: err.Error()})
			}
		} else {
			for _, f := range structure.Files {
				results = append(results, DownloadResult{Path: f, Success: true})
			}
		}
	}

	outputJSON(Envelope{Success: true, Data: results})
}
