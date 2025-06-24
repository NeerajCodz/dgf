package types

// Platform represents a platform configuration
type Platform struct {
	Name        string      `json:"name"`
	ID          string      `json:"id"`
	PublicToken string      `json:"public_token"`
	URL         URL         `json:"URL"`
	URLStruc    URLStruc    `json:"URLStruc"`
	RawURLStruc RawURLStruc `json:"rawURLStruc"`
}

// URL represents the base URLs for a platform
type URL struct {
	Site []string `json:"site"`
	Raw  []string `json:"raw"`
}

// URLStruc represents the URL structure templates for a platform
type URLStruc struct {
	Site         string `json:"site"`
	CommitFolder string `json:"commit_folder"`
	CommitFile   string `json:"commit_file"`
	BranchFolder string `json:"branch_folder"`
	BranchFile   string `json:"branch_file"`
}

// RawURLStruc represents the raw URL structure templates for a platform
type RawURLStruc struct {
	Site   string `json:"site"`
	Commit string `json:"commit"`
	Branch string `json:"branch"`
}

// ParsedURL represents parsed components of a platform URL
type ParsedURL struct {
	URL         string `json:"url"`
	Name        string `json:"name"`
	ID          string `json:"id"`
	Username    string `json:"username"`
	Repo        string `json:"repo"`
	Branch      string `json:"branch"`
	Commit      string `json:"commit"`
	Path        string `json:"path"`
	ParentPath  string `json:"parent_path"`
	RequestPath string `json:"request_path"`
	RequestType string `json:"request_type"` // file or dir
}

// GitHubContent represents an item in a GitHub repository's contents
type GitHubContent struct {
	Name        string  `json:"name"`
	Path        string  `json:"path"`
	Type        string  `json:"type"` // file or dir
	Size        int     `json:"size"`
	DownloadURL *string `json:"download_url"` // Pointer to handle null
	Sha         string  `json:"sha"`
	URL         string  `json:"url"`
	HTMLURL     string  `json:"html_url"`
	GitURL      string  `json:"git_url"`
}

// RepositoryStructure represents the categorized structure of a repository
type RepositoryStructure struct {
	Files        []string `json:"files"`
	FilesName    []string `json:"files_name"`
	FilesSha     []string `json:"files_sha"`
	FilesHTMLURL []string `json:"files_html_url"`
	FilesGitURL  []string `json:"files_git_url"`
	FilesURL     []string `json:"files_url"`
	FilesSize    []int    `json:"files_size"`
	Folders      []string `json:"folders"`
	DownloadURLs []string `json:"download_urls"`
	FilesRequest []string `json:"files_request"`
}