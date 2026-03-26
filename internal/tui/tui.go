package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.design/x/clipboard"

	"github.com/NeerajCodz/dgf/internal/config"
	"github.com/NeerajCodz/dgf/internal/github"
	"github.com/NeerajCodz/dgf/internal/selection"
	"github.com/NeerajCodz/dgf/pkg/types"
)

// Model is the main Bubble Tea model for the TUI
type Model struct {
	state      *AppState
	keys       KeyMap
	selection  *selection.Manager

	// UI components
	urlInput    textinput.Model
	searchInput textinput.Model
	commitInput textinput.Model
	spinner     spinner.Model
	viewport    viewport.Model

	// Dimensions
	width  int
	height int
	ready  bool
	
	// Download progress monitoring
	downloadProgressChan <-chan downloadProgressMsg
}

// NewModel creates a new TUI model
func NewModel(initialURL, token string) Model {
	// Load config
	cfg, _ := config.Load()

	// Create text inputs
	urlInput := textinput.New()
	urlInput.Placeholder = "Enter GitHub URL (e.g., github.com/user/repo)"
	urlInput.Focus()
	urlInput.CharLimit = 256
	urlInput.Width = 60

	searchInput := textinput.New()
	searchInput.Placeholder = "Search files..."
	searchInput.CharLimit = 64
	searchInput.Width = 40

	commitInput := textinput.New()
	commitInput.Placeholder = "Enter commit SHA (or leave empty for latest)"
	commitInput.CharLimit = 40
	commitInput.Width = 50

	// Create enhanced spinner with smooth animation
	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)

	// Create state
	state := NewAppState()
	state.Config = cfg
	state.ASCIIMode = cfg.ASCIIMode
	state.DownloadPath = cfg.DownloadPath
	if state.DownloadPath == "" {
		state.DownloadPath = "."
	}

	// Set token from config or parameter
	if token != "" {
		state.Token = token
	} else if cfg.GithubToken != "" {
		state.Token = cfg.GithubToken
	}

	// Create GitHub client
	state.Client = github.NewClient(state.Token)

	// Pre-fill URL if provided and mark for auto-fetch
	autoFetch := false
	if initialURL != "" {
		urlInput.SetValue(initialURL)
		autoFetch = true
	}

	m := Model{
		state:       state,
		keys:        DefaultKeyMap(),
		selection:   selection.NewManager(),
		urlInput:    urlInput,
		searchInput: searchInput,
		commitInput: commitInput,
		spinner:     sp,
	}

	// Mark for auto-fetch after init
	if autoFetch {
		m.state.AutoFetch = true
	}

	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		textinput.Blink,
		m.spinner.Tick,
	}
	
	// Auto-fetch if URL was provided
	if m.state.AutoFetch {
		m.state.AutoFetch = false
		cmds = append(cmds, m.fetchRepository(m.urlInput.Value()))
	}
	
	return tea.Batch(cmds...)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd := m.handleKeyPress(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.state.Width = msg.Width
		m.state.Height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-4)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 4
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case fetchDoneMsg:
		m.state.SetMode(types.ModeBrowse)
		m.state.Items = msg.items
		m.selection.SyncWithItems(m.state.Items)
		if msg.err != nil {
			m.state.SetError(msg.err.Error())
		} else {
			// Keep layout stable by not showing transient loaded-toasts in browse view
			m.state.Toast = nil
		}

	case branchesDoneMsg:
		if msg.err != nil {
			m.state.SetError(msg.err.Error())
			m.state.SetMode(types.ModeBrowse)
		} else {
			m.state.AvailableBranches = msg.branches
			m.state.BranchCursor = 0
			m.state.SetMode(types.ModeBranchSelect)
		}

	case previewDoneMsg:
		m.state.PreviewLoading = false
		if msg.err != nil {
			m.state.SetError(msg.err.Error())
			m.state.SetMode(types.ModeBrowse)
		} else {
			m.state.PreviewContent = msg.content
			m.state.PreviewPath = msg.path
			m.state.SetMode(types.ModePreview)
		}

	case progressMonitorMsg:
		// Forward progress updates to the state
		m.state.DownloadProgress = msg.progress
		m.state.DownloadCurrent = msg.current
		m.state.DownloadDone = msg.done
		m.state.DownloadTotal = msg.total
		// Schedule another progress read
		if m.downloadProgressChan != nil {
			cmds = append(cmds, monitorDownloadProgress(m.downloadProgressChan))
		}

	case downloadProgressMsg:
		m.state.DownloadProgress = msg.progress
		m.state.DownloadCurrent = msg.current
		m.state.DownloadDone = msg.done
		m.state.DownloadTotal = msg.total

	case downloadDoneMsg:
		m.state.IsDownloading = false
		m.downloadProgressChan = nil // Clear the channel
		m.state.SetMode(types.ModeBrowse)
		if msg.err != nil {
			m.state.SetError(msg.err.Error())
		} else {
			m.state.ShowToast(fmt.Sprintf("Downloaded %d files successfully", msg.count), types.ToastSuccess)
		}

	case tickMsg:
		m.state.FrameCount++
		m.state.TickToast()
		cmds = append(cmds, tick())
	}

	// Update inputs based on mode
	switch m.state.Mode {
	case types.ModeInput:
		var cmd tea.Cmd
		m.urlInput, cmd = m.urlInput.Update(msg)
		cmds = append(cmds, cmd)

	case types.ModeSearch:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)
		// Update filtered items
		m.state.FilteredItems = selection.Filter(m.state.Items, m.searchInput.Value())
		m.state.SearchQuery = m.searchInput.Value()
	case types.ModeCommitInput:
		var cmd tea.Cmd
		m.commitInput, cmd = m.commitInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var content string

	switch m.state.Mode {
	case types.ModeInput:
		content = m.viewInput()
	case types.ModeLoading:
		content = m.viewLoading()
	case types.ModeBrowse:
		content = m.viewBrowser()
	case types.ModePreview:
		content = m.viewPreview()
	case types.ModeHelp:
		content = m.viewHelp()
	case types.ModeDownload:
		content = m.viewDownload()
	case types.ModeCommitInput:
		content = m.viewCommitInput()
	case types.ModeBranchSelect:
		content = m.viewBranchSelect()
	default:
		content = m.viewInput()
	}

	// Add toast overlay if present
	if m.state.Toast != nil {
		content = m.overlayToast(content)
	}

	return content
}

// handleKeyPress handles keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch m.state.Mode {
	case types.ModeInput:
		return m.handleInputKeys(key, msg)
	case types.ModeBrowse:
		return m.handleBrowseKeys(key)
	case types.ModeSearch:
		return m.handleSearchKeys(key)
	case types.ModePreview:
		return m.handlePreviewKeys(key)
	case types.ModeHelp:
		return m.handleHelpKeys(key)
	case types.ModeCommitInput:
		return m.handleCommitInputKeys(key)
	case types.ModeBranchSelect:
		return m.handleBranchSelectKeys(key)
	}

	return nil
}

// handleInputKeys handles keys in input mode
func (m *Model) handleInputKeys(key string, msg tea.KeyMsg) tea.Cmd {
	switch key {
	case "enter":
		val := strings.TrimSpace(m.urlInput.Value())
		if val == "" {
			return nil
		}
		if m.state.TokenEntry {
			m.state.Token = val
			m.state.Client.SetToken(val)
			m.state.Config.GithubToken = val
			if err := config.Save(m.state.Config); err != nil {
				m.state.SetError(err.Error())
			} else {
				m.state.ShowToast("Token updated", types.ToastSuccess)
			}
			m.state.TokenEntry = false
			m.urlInput.SetValue("")
			m.urlInput.Placeholder = "Enter GitHub URL (e.g., github.com/user/repo)"
			return nil
		}
		if val != "" {
			return m.fetchRepository(val)
		}
	case "ctrl+c", "q":
		return tea.Quit
	case "?":
		m.state.SetMode(types.ModeHelp)
	case "esc":
		if m.state.TokenEntry {
			m.state.TokenEntry = false
			m.urlInput.SetValue("")
			m.urlInput.Placeholder = "Enter GitHub URL (e.g., github.com/user/repo)"
		}
	}
	return nil
}

// handleBrowseKeys handles keys in browse mode
func (m *Model) handleBrowseKeys(key string) tea.Cmd {
	items := m.state.GetVisibleItems()
	itemCount := len(items)
	errLower := strings.ToLower(m.state.Error)
	rateLimited := strings.Contains(errLower, "rate limit") || strings.Contains(errLower, "status 403")

	if rateLimited {
		switch key {
		case "enter", "t", "T":
			m.state.TokenEntry = true
			m.urlInput.SetValue("")
			m.urlInput.Placeholder = "Enter GitHub token and press Enter"
			m.state.SetMode(types.ModeInput)
			return nil
		case "esc":
			m.state.ClearError()
			m.state.SetMode(types.ModeInput)
			return nil
		case "ctrl+c", "q":
			return tea.Quit
		}
	}

	switch key {
	case "up":
		if m.state.Cursor > 0 {
			m.state.Cursor--
		}
		m.state.ConfirmDownload = false
		m.state.ConfirmInverseSelection = false
	case "down":
		if m.state.Cursor < itemCount-1 {
			m.state.Cursor++
		}
		m.state.ConfirmDownload = false
		m.state.ConfirmInverseSelection = false
	case "home", "g":
		m.state.Cursor = 0
		m.state.ConfirmDownload = false
		m.state.ConfirmInverseSelection = false
	case "end", "G":
		if itemCount > 0 {
			m.state.Cursor = itemCount - 1
		}
		m.state.ConfirmDownload = false
		m.state.ConfirmInverseSelection = false
	case " ", "k", "K":
		if item := m.state.CurrentItem(); item != nil {
			m.selection.Toggle(item.Path, item.Size)
			item.Selected = m.selection.IsSelected(item.Path)
		}
		m.state.ConfirmDownload = false
		m.state.ConfirmInverseSelection = false
	case "a":
		m.selection.SelectAll(items)
		m.selection.SyncWithItems(m.state.Items)
		m.state.ShowToast(fmt.Sprintf("Selected %d items", m.selection.Count()), types.ToastInfo)
		m.state.ConfirmDownload = false
		m.state.ConfirmInverseSelection = false
	case "u":
		m.selection.UnselectAll()
		m.selection.SyncWithItems(m.state.Items)
		m.state.ShowToast("Selection cleared", types.ToastInfo)
		m.state.ConfirmDownload = false
		m.state.ConfirmInverseSelection = false
	case "l", "L", "right", "enter":
		// Enter or L/right arrow: navigate into folder
		if item := m.state.CurrentItem(); item != nil && item.IsDir() {
			m.state.PushNavigation()
			return m.navigateToFolder(item.Path)
		}
		m.state.ConfirmDownload = false
		m.state.ConfirmInverseSelection = false
	case "j", "J", "left", "esc":
		// Esc or J/left arrow: navigate back to parent
		m.state.ConfirmDownload = false
		m.state.ConfirmInverseSelection = false
		return m.navigateBack()
	case "d", "D", "shift+enter":
		// D or Shift+Enter: download selected items
		if !m.selection.HasSelection() {
			m.state.ShowToast("No items selected - use K/space to select", types.ToastWarning)
			return nil
		}
		return m.startDownload()
	case "o", "O":
		m.state.ASCIIMode = !m.state.ASCIIMode
		mode := "emoji"
		if m.state.ASCIIMode {
			mode = "ASCII"
		}
		m.state.ShowToast("Icons: "+mode, types.ToastInfo)
	case "y":
		if item := m.state.CurrentItem(); item != nil {
			err := clipboard.Write(clipboard.FmtText, []byte(item.Path))
			if err != nil {
				m.state.ShowToast("Failed to copy to clipboard", types.ToastError)
			} else {
				m.state.ShowToast(fmt.Sprintf("Copied to clipboard: %s", item.Path), types.ToastSuccess)
			}
		}
	case "p":
		if item := m.state.CurrentItem(); item != nil && item.IsFile() {
			return m.previewFile(item)
		}
	case "/":
		m.searchInput.SetValue("")
		m.searchInput.Focus()
		m.state.SetMode(types.ModeSearch)
		m.state.IsSearching = true
	case "i", "I":
		if m.state.ConfirmInverseSelection {
			m.state.ConfirmInverseSelection = false
			selectionSet := make(map[string]struct{}, len(m.selection.GetSelected()))
			for _, path := range m.selection.GetSelected() {
				selectionSet[path] = struct{}{}
			}
			m.selection.UnselectAll()
			for _, item := range items {
				if _, exists := selectionSet[item.Path]; !exists {
					m.selection.Select(item.Path, item.Size)
				}
			}
			m.selection.SyncWithItems(m.state.Items)
			m.state.ShowToast(fmt.Sprintf("Inverted selection (%d selected)", m.selection.Count()), types.ToastSuccess)
			break
		}
		m.state.ConfirmInverseSelection = true
		m.state.ShowToast("Press I again to confirm inverse selection", types.ToastWarning)
	case "r":
		m.state.ConfirmDownload = false
		m.state.ConfirmInverseSelection = false
		return m.refreshView()
	case "c", "C":
		// Enter commit ID
		m.commitInput.SetValue(m.state.Commit)
		m.commitInput.Focus()
		m.state.SetMode(types.ModeCommitInput)
	case "b", "B":
		// Fetch and show branches
		m.state.Error = ""
		m.state.SetMode(types.ModeLoading)
		return m.fetchBranches()
	case "?":
		m.state.SetMode(types.ModeHelp)
	case "q", "ctrl+c":
		return tea.Quit
	}

	// Ensure cursor stays within filtered results bounds
	if itemCount == 0 {
		m.state.Cursor = 0
		m.state.ScrollOffset = 0
	} else if m.state.Cursor >= itemCount {
		m.state.Cursor = itemCount - 1
	}

	return nil
}

// handleSearchKeys handles keys in search mode
func (m *Model) handleSearchKeys(key string) tea.Cmd {
	switch key {
	case "enter":
		m.state.SetMode(types.ModeBrowse)
		m.state.Cursor = 0
	case "esc":
		m.searchInput.SetValue("")
		m.state.SearchQuery = ""
		m.state.IsSearching = false
		m.state.FilteredItems = nil
		m.state.SetMode(types.ModeBrowse)
	}
	return nil
}

func (m *Model) handleCommitInputKeys(key string) tea.Cmd {
	switch key {
	case "enter":
		commit := strings.TrimSpace(m.commitInput.Value())
		m.state.Commit = commit
		m.state.SetMode(types.ModeBrowse)
		return m.refreshView()
	case "esc":
		m.state.SetMode(types.ModeBrowse)
	}
	return nil
}

func (m *Model) handleBranchSelectKeys(key string) tea.Cmd {
	switch key {
	case "up", "k":
		if m.state.BranchCursor > 0 {
			m.state.BranchCursor--
		}
	case "down", "j":
		if m.state.BranchCursor < len(m.state.AvailableBranches)-1 {
			m.state.BranchCursor++
		}
	case "enter":
		if len(m.state.AvailableBranches) > 0 && m.state.BranchCursor >= 0 && m.state.BranchCursor < len(m.state.AvailableBranches) {
			m.state.Branch = m.state.AvailableBranches[m.state.BranchCursor]
			m.state.Commit = ""
			m.state.SetMode(types.ModeBrowse)
			return m.refreshView()
		}
	case "esc":
		m.state.SetMode(types.ModeBrowse)
	}
	return nil
}

// handlePreviewKeys handles keys in preview mode
func (m *Model) handlePreviewKeys(key string) tea.Cmd {
	switch key {
	case "esc", "q", "p":
		m.state.SetMode(types.ModeBrowse)
	case "up", "k":
		m.state.PreviewScroll--
		if m.state.PreviewScroll < 0 {
			m.state.PreviewScroll = 0
		}
	case "down", "j":
		m.state.PreviewScroll++
	}
	return nil
}

// handleHelpKeys handles keys in help mode
func (m *Model) handleHelpKeys(key string) tea.Cmd {
	switch key {
	case "esc", "q", "?":
		m.state.GoBack()
	}
	return nil
}

// Run starts the TUI application
func Run(initialURL, token string) error {
	m := NewModel(initialURL, token)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
