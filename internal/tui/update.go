package tui

import (
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

type tickMsg time.Time

// Commands

func tick() tea.Cmd {
	return tea.Tick(time.Second/30, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
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
			return m.refreshView()
		}
		return nil
	}

	m.state.Path = entry.Path
	m.state.Cursor = entry.Cursor
	m.state.ScrollOffset = entry.Scroll
	return m.refreshView()
}

// refreshView reloads the current directory
func (m *Model) refreshView() tea.Cmd {
	m.state.SetMode(types.ModeLoading)

	return func() tea.Msg {
		items, err := github.FetchItems(
			m.state.Client,
			m.state.Owner,
			m.state.Repo,
			m.state.GetRef(),
			m.state.Path,
		)
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
	m.state.DownloadTotal = m.selection.Count()

	selected := m.selection.GetSelected()
	client := m.state.Client
	owner := m.state.Owner
	repo := m.state.Repo
	ref := m.state.GetRef()
	downloadPath := m.state.DownloadPath
	token := m.state.Token
	workers := m.state.Config.Workers

	return func() tea.Msg {
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
						if err == nil {
							structure.Folders = append(structure.Folders, folderStructure.Folders...)
							structure.Files = append(structure.Files, folderStructure.Files...)
							structure.DownloadURLs = append(structure.DownloadURLs, folderStructure.DownloadURLs...)
							structure.FilesRequest = append(structure.FilesRequest, folderStructure.FilesRequest...)
							structure.FilesSize = append(structure.FilesSize, folderStructure.FilesSize...)
						}
					}
				}
			}
		}

		err := github.Download(structure, github.DownloadOptions{
			OutputDir:    downloadPath,
			Token:        token,
			Workers:      workers,
			CheckLFS:     true,
			Owner:        owner,
			Repo:         repo,
			GitHubClient: client,
		})

		return downloadDoneMsg{count: len(structure.Files), err: err}
	}
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
