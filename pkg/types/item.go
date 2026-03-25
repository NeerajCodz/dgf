package types

// RepoItem represents a file or folder in the repository browser
type RepoItem struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" or "dir"
	Size        int64  `json:"size"`
	Sha         string `json:"sha"`
	DownloadURL string `json:"download_url"`
	Selected    bool   `json:"selected"`
	IsLFS       bool   `json:"is_lfs"`
}

// IsDir returns true if the item is a directory
func (r *RepoItem) IsDir() bool {
	return r.Type == "dir"
}

// IsFile returns true if the item is a file
func (r *RepoItem) IsFile() bool {
	return r.Type == "file"
}

// Extension returns the file extension (empty for directories)
func (r *RepoItem) Extension() string {
	if r.IsDir() {
		return ""
	}
	for i := len(r.Name) - 1; i >= 0; i-- {
		if r.Name[i] == '.' {
			return r.Name[i+1:]
		}
	}
	return ""
}
