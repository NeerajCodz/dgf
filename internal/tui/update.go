package tui

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/bubbletea"

	"github.com/NeerajCodz/dgf/internal/github"
	"github.com/NeerajCodz/dgf/pkg/types"
)

// Messages

type fetchDoneMsg struct {
	items []types.RepoItem
	err   error
}

type previewDoneMsg struct {
	content string
	path    string
	err     error
}

type downloadProgressMsg struct {
	progress float64
	current  string
	done     int
	total    int
}

type downloadDoneMsg struct {
	count int
	err   error
}

type branchesDoneMsg struct {
	branches []string
	details  []github.BranchInfo
	err      error
}

type commitsDoneMsg struct {
	commits []github.CommitInfo
	err     error
}

type autoExitMsg struct{}
type selectorSearchDoneMsg struct {
	commits []github.CommitInfo
	err     error
}

func autoExit() tea.Cmd {
	return tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
		return autoExitMsg{}
	})
}

type tickMsg time.Time

// progressMonitorMsg wraps a download progress message from the monitor goroutine
type progressMonitorMsg downloadProgressMsg

// Commands

func tick() tea.Cmd {
	return tea.Tick(time.Second/30, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// monitorDownloadProgress monitors a progress channel and sends messages to the TUI
// This runs as a separate command that continuously forwards progress updates
func monitorDownloadProgress(progressChan <-chan downloadProgressMsg) tea.Cmd {
	return func() tea.Msg {
		// Read from the progress channel and forward the first message
		// This will be called repeatedly to drain the channel
		msg, ok := <-progressChan
		if !ok {
			// Channel closed, download is complete
			return nil
		}
		return progressMonitorMsg(msg)
	}
}

// fetchRepository fetches repository contents from GitHub
func (m *Model) fetchRepository(url string) tea.Cmd {
	m.state.SetMode(types.ModeLoading)

	return func() tea.Msg {
		// Parse URL
		platform := types.Platform{
			Name: "GitHub",
			ID:   "github",
			URL: types.URL{
				Site: []string{"https://github.com"},
			},
		}

		parsed, err := github.ParseURL(url, platform)
		if err != nil {
			return fetchDoneMsg{err: err}
		}

		m.state.Owner = parsed.Username
		m.state.Repo = parsed.Repo
		m.state.Branch = parsed.Branch
		m.state.Commit = parsed.Commit
		m.state.Path = parsed.Path

		// Get default branch if needed
		ref := parsed.Branch
		if ref == "" && parsed.Commit == "" {
			defaultBranch, err := m.state.Client.FetchDefaultBranch(parsed.Username, parsed.Repo)
			if err != nil {
				return fetchDoneMsg{err: err}
			}
			ref = defaultBranch
			m.state.Branch = defaultBranch
		} else if parsed.Commit != "" {
			ref = parsed.Commit
		}

		// Fetch items
		items, err := github.FetchItems(m.state.Client, parsed.Username, parsed.Repo, ref, parsed.Path)
		return fetchDoneMsg{items: items, err: err}
	}
}

// navigateToFolder navigates into a folder
func (m *Model) navigateToFolder(path string) tea.Cmd {
	// Check if we have this directory cached
	cacheKey := m.state.DirCacheKey(path)
	if cachedItems, exists := m.state.DirCache[cacheKey]; exists {
		// Use cached items immediately
		m.state.Items = cachedItems
		m.state.Path = path
		m.state.Cursor = 0
		m.state.ScrollOffset = 0
		m.state.SetMode(types.ModeBrowse)
		m.selection.SyncWithItems(m.state.Items)
		return nil
	}

	// Not cached, fetch from API
	m.state.SetMode(types.ModeLoading)
	m.state.Path = path
	m.state.Cursor = 0

	return func() tea.Msg {
		items, err := github.FetchItems(
			m.state.Client,
			m.state.Owner,
			m.state.Repo,
			m.state.GetRef(),
			path,
		)

		// Cache the result if successful
		if err == nil && items != nil {
			m.state.DirCache[cacheKey] = items
		}

		return fetchDoneMsg{items: items, err: err}
	}
}

// navigateBack goes back to the previous directory
func (m *Model) navigateBack() tea.Cmd {
	entry := m.state.PopNavigation()
	if entry == nil {
		// Go to parent directory
		if m.state.Path != "" {
			parts := splitPath(m.state.Path)
			if len(parts) > 1 {
				m.state.Path = joinPath(parts[:len(parts)-1])
			} else {
				m.state.Path = ""
			}
			m.state.Cursor = 0
			cacheKey := m.state.DirCacheKey(m.state.Path)
			if cached, ok := m.state.DirCache[cacheKey]; ok {
				m.state.Items = cached
				m.selection.SyncWithItems(m.state.Items)
				m.state.SetMode(types.ModeBrowse)
				return nil
			}
			return m.refreshView()
		}
		return nil
	}

	m.state.Path = entry.Path
	m.state.Cursor = entry.Cursor
	m.state.ScrollOffset = entry.Scroll
	cacheKey := m.state.DirCacheKey(m.state.Path)
	if cached, ok := m.state.DirCache[cacheKey]; ok {
		m.state.Items = cached
		m.selection.SyncWithItems(m.state.Items)
		m.state.SetMode(types.ModeBrowse)
		return nil
	}
	return m.refreshView()
}

// refreshView reloads the current directory
func (m *Model) refreshView() tea.Cmd {
	cacheKey := m.state.DirCacheKey(m.state.Path)
	if cached, ok := m.state.DirCache[cacheKey]; ok {
		m.state.Items = cached
		m.selection.SyncWithItems(m.state.Items)
		m.state.SetMode(types.ModeBrowse)
		return nil
	}

	m.state.SetMode(types.ModeLoading)

	return func() tea.Msg {
		items, err := github.FetchItems(
			m.state.Client,
			m.state.Owner,
			m.state.Repo,
			m.state.GetRef(),
			m.state.Path,
		)
		if err == nil {
			m.state.DirCache[cacheKey] = items
		}
		return fetchDoneMsg{items: items, err: err}
	}
}

// previewFile fetches file content for preview
func (m *Model) previewFile(item *types.RepoItem) tea.Cmd {
	m.state.PreviewLoading = true

	return func() tea.Msg {
		if item.DownloadURL == "" {
			return previewDoneMsg{err: github.ErrPathNotFound, path: item.Path}
		}

		// Check file size (limit to 1MB)
		if item.Size > 1024*1024 {
			return previewDoneMsg{
				content: "(File too large to preview - over 1MB)",
				path:    item.Path,
			}
		}

		content, err := m.state.Client.FetchRawFile(item.DownloadURL)
		if err != nil {
			return previewDoneMsg{err: err, path: item.Path}
		}

		// Check if binary
		if isBinary(content) {
			return previewDoneMsg{
				content: "(Binary file - cannot preview)",
				path:    item.Path,
			}
		}

		return previewDoneMsg{content: string(content), path: item.Path}
	}
}

// startDownload begins downloading selected items
func (m *Model) startDownload() tea.Cmd {
	m.state.SetMode(types.ModeDownload)
	m.state.IsDownloading = true
	m.state.DownloadDone = 0
	m.state.DownloadProgress = 0
	m.state.DownloadCurrent = ""
	m.state.DownloadTotal = m.selection.Count()

	// Validate: check for zero-file selection
	if m.state.DownloadTotal == 0 {
		m.state.IsDownloading = false
		m.state.SetMode(types.ModeBrowse)
		return func() tea.Msg {
			return downloadDoneMsg{count: 0, err: fmt.Errorf("no files selected for download")}
		}
	}

	selected := m.selection.GetSelected()
	client := m.state.Client
	owner := m.state.Owner
	repo := m.state.Repo
	ref := m.state.GetRef()
	downloadPath := m.state.DownloadPath
	token := m.state.Token
	workers := m.state.Config.Workers

	// Create the download command
	downloadCmd := func() tea.Msg {
		// Build structure from selected items
		structure := types.RepositoryStructure{
			Files:        make([]string, 0),
			Folders:      make([]string, 0),
			DownloadURLs: make([]string, 0),
			FilesRequest: make([]string, 0),
			FilesSize:    make([]int, 0),
		}

		// Process selected items - fetch folder contents recursively
		for _, path := range selected {
			for _, item := range m.state.Items {
				if item.Path == path {
					if item.IsFile() {
						structure.Files = append(structure.Files, item.Path)
						structure.DownloadURLs = append(structure.DownloadURLs, item.DownloadURL)
						structure.FilesRequest = append(structure.FilesRequest, item.Name)
						structure.FilesSize = append(structure.FilesSize, int(item.Size))
					} else if item.IsDir() {
						// Fetch folder contents recursively
						folderStructure, err := github.FetchFolderRecursive(client, owner, repo, ref, item.Path)
						if err != nil {
							// Log warning but continue - folder may be empty or inaccessible
							continue
						}
						structure.Folders = append(structure.Folders, folderStructure.Folders...)
						structure.Files = append(structure.Files, folderStructure.Files...)
						structure.DownloadURLs = append(structure.DownloadURLs, folderStructure.DownloadURLs...)
						structure.FilesRequest = append(structure.FilesRequest, folderStructure.FilesRequest...)
						structure.FilesSize = append(structure.FilesSize, folderStructure.FilesSize...)
					}
				}
			}
		}

		// Validate: check if any files were found
		if len(structure.Files) == 0 {
			return downloadDoneMsg{count: 0, err: fmt.Errorf("no files found in selection (selected folders may be empty)")}
		}

		// Create a channel to capture progress updates
		progressChan := make(chan downloadProgressMsg, 10)

		// Track current file being downloaded
		var currentFile string
		var currentMutex sync.Mutex

		// Start download in background goroutine
		go func() {
			err := github.Download(structure, github.DownloadOptions{
				OutputDir:    downloadPath,
				Token:        token,
				Workers:      workers,
				CheckLFS:     true,
				Owner:        owner,
				Repo:         repo,
				GitHubClient: client,
				OnFileStart: func(path string) {
					currentMutex.Lock()
					currentFile = path
					currentMutex.Unlock()
					progressChan <- downloadProgressMsg{
						current:  path,
						total:    len(structure.Files),
						progress: 0,
						done:     0,
					}
				},
				OnProgress: func(current, total int) {
					currentMutex.Lock()
					filename := currentFile
					currentMutex.Unlock()
					progressChan <- downloadProgressMsg{
						current:  filename,
						total:    total,
						progress: float64(current) / float64(total),
						done:     current,
					}
				},
			})

			close(progressChan)

			// Send final download completion message after progress channel is exhausted
			// Note: We'll drain the progress channel in Update() via monitorDownloadProgress
			if err != nil {
				// Wait a bit to let remaining progress messages get through
				time.Sleep(100 * time.Millisecond)
			}
		}()

		// Wait for all progress to complete
		for range progressChan {
			// Drain the channel, messages are handled by monitorDownloadProgress
		}

		return downloadDoneMsg{count: len(structure.Files), err: nil}
	}

	// Return batch command: one to monitor progress, one to do the download
	return tea.Batch(downloadCmd)
}

// Helper functions

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	result := make([]string, 0)
	current := ""
	for _, c := range path {
		if c == '/' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func joinPath(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += "/" + parts[i]
	}
	return result
}

func isBinary(data []byte) bool {
	// Check first 8000 bytes for null bytes
	checkLen := 8000
	if len(data) < checkLen {
		checkLen = len(data)
	}
	for i := 0; i < checkLen; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}

// fetchBranches fetches the list of branches for the current repository
func (m *Model) fetchBranches() tea.Cmd {
	cacheKey := m.state.BranchCacheKey()
	if cached, ok := m.state.BranchCache[cacheKey]; ok && len(cached) > 0 {
		return func() tea.Msg {
			branches := make([]string, 0, len(cached))
			for _, b := range cached {
				branches = append(branches, b.Name)
			}
			return branchesDoneMsg{branches: branches, details: cached}
		}
	}

	return func() tea.Msg {
		branches, err := m.state.Client.FetchBranches(m.state.Owner, m.state.Repo)
		if err != nil {
			return branchesDoneMsg{err: err}
		}
		details, detailErr := m.state.Client.FetchBranchesWithCounts(m.state.Owner, m.state.Repo)
		if detailErr != nil {
			details = nil
		} else {
			m.state.BranchCache[cacheKey] = details
		}
		return branchesDoneMsg{branches: branches, details: details}
	}
}

func (m *Model) fetchCommits() tea.Cmd {
	cacheKey := m.state.CommitCacheKey("")
	if cached, ok := m.state.CommitCache[cacheKey]; ok && len(cached) > 0 {
		return func() tea.Msg {
			return commitsDoneMsg{commits: cached, err: nil}
		}
	}

	return func() tea.Msg {
		ref := m.state.GetRef()
		commits, err := m.state.Client.FetchCommits(m.state.Owner, m.state.Repo, ref, 10)
		if err == nil {
			m.state.CommitCache[cacheKey] = commits
		}
		return commitsDoneMsg{commits: commits, err: err}
	}
}

func (m *Model) searchCommits(query string) tea.Cmd {
	cacheKey := m.state.CommitCacheKey(query)
	if cached, ok := m.state.CommitCache[cacheKey]; ok {
		return func() tea.Msg {
			return selectorSearchDoneMsg{commits: cached}
		}
	}

	return func() tea.Msg {
		commits, err := m.state.Client.SearchCommits(m.state.Owner, m.state.Repo, query, 25)
		if err == nil {
			m.state.CommitCache[cacheKey] = commits
		}
		return selectorSearchDoneMsg{commits: commits, err: err}
	}
}
