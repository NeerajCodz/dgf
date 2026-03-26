package types

// AppMode represents the current mode of the TUI application
type AppMode int

const (
	ModeInput    AppMode = iota // URL input screen
	ModeLoading                 // Loading/fetching data
	ModeBrowse                  // File browser
	ModePreview                 // File preview modal
	ModeHelp                    // Help overlay
	ModeDownload                // Download progress
	ModeDownloadComplete        // Download completed screen (auto-exit)
	ModeSearch                  // Search overlay active
	ModeCommitInput             // Commit ID input
	ModeBranchSelect            // Branch selection
)

// String returns the string representation of the mode
func (m AppMode) String() string {
	switch m {
	case ModeInput:
		return "input"
	case ModeLoading:
		return "loading"
	case ModeBrowse:
		return "browse"
	case ModePreview:
		return "preview"
	case ModeHelp:
		return "help"
	case ModeDownload:
		return "download"
	case ModeDownloadComplete:
		return "download_complete"
	case ModeSearch:
		return "search"
	case ModeCommitInput:
		return "commit_input"
	case ModeBranchSelect:
		return "branch_select"
	default:
		return "unknown"
	}
}

// NavigationEntry represents a navigation history entry
type NavigationEntry struct {
	URL    string
	Path   string
	Cursor int
	Scroll int
}

// DownloadProgress represents download progress for a single file
type DownloadProgress struct {
	Path       string
	Downloaded int64
	Total      int64
	Done       bool
	Error      error
}

// Toast represents a notification message
type Toast struct {
	Message string
	Type    ToastType
	TTL     int // Frames remaining
}

// ToastType represents the type of toast notification
type ToastType int

const (
	ToastInfo ToastType = iota
	ToastSuccess
	ToastWarning
	ToastError
)
