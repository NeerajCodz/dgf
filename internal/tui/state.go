package tui

import (
	"context"
	"strings"

	"github.com/NeerajCodz/dgf/internal/github"
	"github.com/NeerajCodz/dgf/pkg/types"
)

// AppState holds all state for the TUI application
type AppState struct {
	// Mode
	Mode     types.AppMode
	PrevMode types.AppMode

	// Repository info
	Owner  string
	Repo   string
	Branch string
	Commit string
	Path   string
	Token  string

	// Branch selection
	AvailableBranches []string
	BranchCursor      int
	BranchQuery       string
	FilteredBranches  []github.BranchInfo
	BranchItems       []github.BranchInfo

	// Commit selection
	CommitCursor      int
	CommitQuery       string
	FilteredCommits   []github.CommitInfo
	CommitItems       []github.CommitInfo
	SelectedCommitMsg string

	// Navigation
	NavigationStack []types.NavigationEntry
	Cursor          int
	ScrollOffset    int

	// Items
	Items         []types.RepoItem
	FilteredItems []types.RepoItem
	FullTree      []types.RepoItem

	// Directory cache to avoid refetching
	DirCache map[string][]types.RepoItem
	// Selector caches (session-scoped)
	CommitCache map[string][]github.CommitInfo
	BranchCache map[string][]github.BranchInfo

	// Selection
	SelectedPaths map[string]bool
	SelectedSize  int64
	SelectedCount int

	// Search
	IsSearching bool
	SearchQuery string

	// Preview
	PreviewContent string
	PreviewPath    string
	PreviewLoading bool
	PreviewScroll  int

	// Download
	DownloadPath            string
	DownloadProgress        float64
	DownloadCurrent         string
	DownloadTotal           int
	DownloadDone            int
	DownloadResultCount     int
	IsDownloading           bool
	ConfirmDownload         bool
	ConfirmDownloadSize     string
	ConfirmInverseSelection bool

	// UI
	ASCIIMode  bool
	Toast      *types.Toast
	FrameCount uint64
	Width      int
	Height     int
	Error      string
	TokenEntry bool

	// Config
	Config types.Config

	// GitHub client
	Client *github.Client

	// AutoFetch triggers automatic repository fetch on startup
	AutoFetch bool

	// Cancellable in-flight operation context (loading/download/search)
	opCtx    context.Context
	opCancel context.CancelFunc
}

// NewAppState creates a new application state with defaults
func NewAppState() *AppState {
	return &AppState{
		Mode:            types.ModeInput,
		SelectedPaths:   make(map[string]bool),
		NavigationStack: make([]types.NavigationEntry, 0),
		Items:           make([]types.RepoItem, 0),
		FilteredItems:   make([]types.RepoItem, 0),
		DirCache:        make(map[string][]types.RepoItem),
		CommitCache:     make(map[string][]github.CommitInfo),
		BranchCache:     make(map[string][]github.BranchInfo),
		Config:          types.DefaultConfig(),
	}
}

// SetMode changes the current mode and stores the previous one
func (s *AppState) SetMode(mode types.AppMode) {
	s.PrevMode = s.Mode
	s.Mode = mode
}

// GoBack returns to the previous mode
func (s *AppState) GoBack() {
	s.Mode = s.PrevMode
}

// IsSelected checks if a path is selected
func (s *AppState) IsSelected(path string) bool {
	return s.SelectedPaths[path]
}

// ToggleSelected toggles the selection of an item
func (s *AppState) ToggleSelected(item *types.RepoItem) {
	if s.SelectedPaths[item.Path] {
		delete(s.SelectedPaths, item.Path)
		item.Selected = false
		s.SelectedCount--
		s.SelectedSize -= item.Size
	} else {
		s.SelectedPaths[item.Path] = true
		item.Selected = true
		s.SelectedCount++
		s.SelectedSize += item.Size
	}
}

// SelectAll selects all items in the current view
func (s *AppState) SelectAll() {
	items := s.GetVisibleItems()
	for i := range items {
		if !items[i].Selected {
			s.SelectedPaths[items[i].Path] = true
			items[i].Selected = true
			s.SelectedCount++
			s.SelectedSize += items[i].Size
		}
	}
}

// UnselectAll unselects all items
func (s *AppState) UnselectAll() {
	for i := range s.Items {
		s.Items[i].Selected = false
	}
	s.SelectedPaths = make(map[string]bool)
	s.SelectedCount = 0
	s.SelectedSize = 0
}

// GetVisibleItems returns items based on search filter
func (s *AppState) GetVisibleItems() []types.RepoItem {
	if s.IsSearching && s.SearchQuery != "" {
		return s.FilteredItems
	}
	return s.Items
}

// CurrentItem returns the item under the cursor
func (s *AppState) CurrentItem() *types.RepoItem {
	items := s.GetVisibleItems()
	if s.Cursor >= 0 && s.Cursor < len(items) {
		return &items[s.Cursor]
	}
	return nil
}

// CanGoBack returns true if there's navigation history
func (s *AppState) CanGoBack() bool {
	return len(s.NavigationStack) > 0
}

// PushNavigation saves current state to navigation stack
func (s *AppState) PushNavigation() {
	entry := types.NavigationEntry{
		URL:    github.BuildURL(s.Owner, s.Repo, s.GetRef(), s.Path),
		Path:   s.Path,
		Cursor: s.Cursor,
		Scroll: s.ScrollOffset,
	}
	s.NavigationStack = append(s.NavigationStack, entry)
}

// PopNavigation restores previous navigation state
func (s *AppState) PopNavigation() *types.NavigationEntry {
	if len(s.NavigationStack) == 0 {
		return nil
	}
	last := s.NavigationStack[len(s.NavigationStack)-1]
	s.NavigationStack = s.NavigationStack[:len(s.NavigationStack)-1]
	return &last
}

// GetRef returns the current reference (commit or branch)
func (s *AppState) GetRef() string {
	if s.Commit != "" {
		return s.Commit
	}
	return s.Branch
}

// ShowToast displays a toast notification
func (s *AppState) ShowToast(message string, toastType types.ToastType) {
	s.Toast = &types.Toast{
		Message: message,
		Type:    toastType,
		TTL:     90, // ~3 seconds at 30fps
	}
}

// TickToast decrements the toast TTL and clears it when expired
func (s *AppState) TickToast() {
	if s.Toast != nil {
		s.Toast.TTL--
		if s.Toast.TTL <= 0 {
			s.Toast = nil
		}
	}
}

// ClearError clears any error message
func (s *AppState) ClearError() {
	s.Error = ""
}

// SetError sets an error message and shows a toast
func (s *AppState) SetError(err string) {
	s.Error = err
	s.ShowToast(err, types.ToastError)
}

// GetBreadcrumb returns the breadcrumb navigation string
func (s *AppState) GetBreadcrumb() string {
	if s.Owner == "" || s.Repo == "" {
		return ""
	}
	breadcrumb := "github.com/" + s.Owner + "/" + s.Repo
	if s.Path != "" {
		breadcrumb += "/" + s.Path
	}
	return breadcrumb
}

// GetBranchLabel returns branch label for header.
func (s *AppState) GetBranchLabel() string {
	if s.Branch == "" {
		return "latest"
	}
	return s.Branch
}

// GetCommitLabel returns commit label for header.
func (s *AppState) GetCommitLabel() string {
	if s.Commit == "" {
		return "latest"
	}
	if len(s.Commit) > 7 {
		return s.Commit[:7]
	}
	return s.Commit
}

func (s *AppState) FilterBranches(query string) {
	s.BranchQuery = query
	if query == "" {
		s.FilteredBranches = append([]github.BranchInfo(nil), s.BranchItems...)
		return
	}
	q := strings.ToLower(query)
	filtered := make([]github.BranchInfo, 0, len(s.BranchItems))
	for _, item := range s.BranchItems {
		if strings.Contains(strings.ToLower(item.Name), q) {
			filtered = append(filtered, item)
		}
	}
	s.FilteredBranches = filtered
	if s.BranchCursor >= len(s.FilteredBranches) {
		s.BranchCursor = max(0, len(s.FilteredBranches)-1)
	}
}

func (s *AppState) FilterCommits(query string) {
	s.CommitQuery = query
	if query == "" {
		s.FilteredCommits = append([]github.CommitInfo(nil), s.CommitItems...)
		return
	}
	q := strings.ToLower(query)
	filtered := make([]github.CommitInfo, 0, len(s.CommitItems))
	for _, item := range s.CommitItems {
		if strings.Contains(strings.ToLower(item.SHA), q) || strings.Contains(strings.ToLower(item.Message), q) {
			filtered = append(filtered, item)
		}
	}
	s.FilteredCommits = filtered
	if s.CommitCursor >= len(s.FilteredCommits) {
		s.CommitCursor = max(0, len(s.FilteredCommits)-1)
	}
}

func (s *AppState) DirCacheKey(path string) string {
	return s.Owner + "/" + s.Repo + ":" + s.GetRef() + ":" + path
}

func (s *AppState) BranchCacheKey() string {
	return s.Owner + "/" + s.Repo
}

func (s *AppState) CommitCacheKey(query string) string {
	return s.Owner + "/" + s.Repo + ":" + s.GetRef() + ":" + strings.ToLower(strings.TrimSpace(query))
}

func (s *AppState) BeginOperation() context.Context {
	s.CancelOperation()
	ctx, cancel := context.WithCancel(context.Background())
	s.opCtx = ctx
	s.opCancel = cancel
	return ctx
}

func (s *AppState) OperationContext() context.Context {
	return s.opCtx
}

func (s *AppState) CancelOperation() {
	if s.opCancel != nil {
		s.opCancel()
	}
	s.opCtx = nil
	s.opCancel = nil
}
