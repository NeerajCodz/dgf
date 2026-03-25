package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"

	"github.com/NeerajCodz/dgf/internal/utils"
	"github.com/NeerajCodz/dgf/pkg/types"
)

// ASCII art logo
const logo = `
    ____  ____________
   / __ \/ ____/ ____/
  / / / / / __/ /_    
 / /_/ / /_/ / __/    
/_____/\____/_/       
                      
Direct Git Fetch v2.0
`

// viewInput renders the URL input screen
func (m Model) viewInput() string {
	var b strings.Builder

	// Logo
	logoStyle := lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Bold(true).
		Align(lipgloss.Center)

	centeredLogo := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, logoStyle.Render(logo))
	b.WriteString(centeredLogo)
	b.WriteString("\n\n")

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(ColorSubtle).
		Align(lipgloss.Center)

	b.WriteString(instructionStyle.Render("Enter a GitHub repository URL to browse"))
	b.WriteString("\n\n")

	// URL input
	inputBox := InputFocusedStyle.
		Width(60).
		Align(lipgloss.Center).
		Render(m.urlInput.View())

	b.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, inputBox))
	b.WriteString("\n\n")

	// Hints
	hintStyle := lipgloss.NewStyle().Foreground(ColorSubtle)
	hints := []string{
		"Examples:",
		"  github.com/charmbracelet/bubbletea",
		"  https://github.com/golang/go/tree/master/src",
		"",
		"Press Enter to browse  •  ? for help  •  q to quit",
	}
	for _, hint := range hints {
		b.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, hintStyle.Render(hint)))
		b.WriteString("\n")
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}

// viewLoading renders the loading screen
func (m Model) viewLoading() string {
	loadingStyle := lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Bold(true)

	spinnerStr := m.spinner.View()
	message := "Fetching repository..."

	if m.state.Owner != "" && m.state.Repo != "" {
		message = fmt.Sprintf("Fetching %s/%s...", m.state.Owner, m.state.Repo)
	}

	content := fmt.Sprintf("%s %s", spinnerStr, loadingStyle.Render(message))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// viewBrowser renders the file browser
func (m Model) viewBrowser() string {
	var b strings.Builder

	// Header with centered logo and breadcrumb
	headerLogoStyle := lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Bold(true)
	b.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, headerLogoStyle.Render("Direct Git Fetch • DGF v2.0")))
	b.WriteString("\n")

	// Header with breadcrumb
	breadcrumb := m.state.GetBreadcrumb()
	if breadcrumb != "" {
		headerStyle := lipgloss.NewStyle().
			Background(ColorBorder).
			Foreground(ColorForeground).
			Padding(0, 1).
			Width(m.width)

		b.WriteString(headerStyle.Render("📁 " + breadcrumb))
		b.WriteString("\n")
	}

	// File list
	items := m.state.GetVisibleItems()
	icons := GetIcons(m.state.ASCIIMode)

	// Keep selection state in sync before rendering
	m.selection.SyncWithItems(m.state.Items)

	// Calculate visible area
	listHeight := m.height - 8 // Leave room for logo, breadcrumb, table header, status
	if listHeight < 5 {
		listHeight = 5
	}

	// Adjust scroll offset to keep cursor visible
	if m.state.Cursor < m.state.ScrollOffset {
		m.state.ScrollOffset = m.state.Cursor
	} else if m.state.Cursor >= m.state.ScrollOffset+listHeight {
		m.state.ScrollOffset = m.state.Cursor - listHeight + 1
	}

	// Render table header
	tableHeader := lipgloss.NewStyle().
		Foreground(ColorSubtle).
		Bold(true).
		Render(fmt.Sprintf("%-4s %-2s %-40s %-10s", "Sel", "T", "Name", "Size"))
	b.WriteString(tableHeader)
	b.WriteString("\n")

	// Render items
	if len(items) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(ColorSubtle).
			Italic(true).
			Padding(2, 0)
		b.WriteString(emptyStyle.Render("  (empty directory)"))
		b.WriteString("\n")
	} else {
		for i := m.state.ScrollOffset; i < len(items) && i < m.state.ScrollOffset+listHeight; i++ {
			item := items[i]
			line := m.renderItem(item, i == m.state.Cursor, icons)
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	// Search overlay
	if m.state.Mode == types.ModeSearch {
		searchBox := SearchStyle.
			Width(40).
			Render("/ " + m.searchInput.View())

		searchInfo := lipgloss.NewStyle().
			Foreground(ColorSubtle).
			Render(fmt.Sprintf(" (%d results)", len(m.state.FilteredItems)))

		b.WriteString("\n")
		b.WriteString(searchBox + searchInfo)
		b.WriteString("\n")
	}

	// Footer status bar
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar(icons))

	return BaseStyle.Render(b.String())
}

// renderItem renders a single file/folder item
func (m Model) renderItem(item types.RepoItem, isCursor bool, icons IconSet) string {
	selectionMarker := "[ ]"
	if item.Selected {
		selectionMarker = "[.]"
	}

	// Icon
	var icon, name string
	if item.IsDir() {
		icon = icons.Folder
		name = m.highlightMatches(item.Name+"/", FolderStyle, isCursor)
	} else {
		icon = icons.File
		if item.IsLFS {
			icon = icons.LFS
		}
		name = m.highlightMatches(item.Name, FileStyle, isCursor)
	}

	// Size (for files)
	var sizeStr string
	if item.IsFile() {
		sizeStr = utils.FormatBytes(item.Size)
	} else {
		sizeStr = "-"
	}

	// Build line
	itemType := "F"
	if item.IsDir() {
		itemType = "D"
	}
	displayName := stripANSI(fmt.Sprintf("%s %s", icon, name))
	if len(displayName) > 40 {
		displayName = displayName[:37] + "..."
	}
	lineContent := fmt.Sprintf("%-4s %-2s %-40s %-10s", selectionMarker, itemType, displayName, sizeStr)

	// Apply cursor style
	if isCursor {
		return CursorStyle.Width(m.width - 2).Render(lineContent)
	}

	return "  " + lineContent
}

// highlightMatches highlights search query characters in the text
func (m Model) highlightMatches(text string, baseStyle lipgloss.Style, isCursor bool) string {
	// If no search query, return styled text as-is
	if m.state.SearchQuery == "" {
		return baseStyle.Render(text)
	}

	query := strings.ToLower(m.state.SearchQuery)
	
	// Find all match positions
	var result strings.Builder
	lastPos := 0
	highlightStyle := baseStyle.Copy().Bold(true).Foreground(ColorHighlight)
	
	for i := 0; i < len(text); i++ {
		// Look for next match starting at current position in query
		found := false
		for j := i; j < len(text); j++ {
			if j-i < len(query) && strings.ToLower(string(text[j])) == string(query[j-i]) {
				if j-i == len(query)-1 {
					// Full query matched
					// Add non-matching part before match
					if lastPos < i {
						result.WriteString(baseStyle.Render(text[lastPos:i]))
					}
					// Add highlighted match
					result.WriteString(highlightStyle.Render(text[i : j+1]))
					lastPos = j + 1
					i = j
					found = true
					break
				}
			} else {
				break
			}
		}
		if found {
			break
		}
	}
	
	// Add remaining text
	if lastPos < len(text) {
		result.WriteString(baseStyle.Render(text[lastPos:]))
	}
	
	// If nothing was added, just return the styled text
	if result.Len() == 0 {
		return baseStyle.Render(text)
	}
	
	return result.String()
}

// renderStatusBar renders the bottom status bar
func (m Model) renderStatusBar(_ IconSet) string {
	var parts []string

	// Selection info
	if m.selection.HasSelection() {
		count := m.selection.Count()
		size := utils.FormatBytes(m.selection.TotalSize())
		parts = append(parts, SelectedStyle.Render(fmt.Sprintf("%d selected (%s)", count, size)))
	} else {
		parts = append(parts, StatusBarStyle.Render("0 selected"))
	}

	// Position info
	items := m.state.GetVisibleItems()
	if len(items) > 0 {
		parts = append(parts, StatusBarStyle.Render(fmt.Sprintf("%d/%d", m.state.Cursor+1, len(items))))
	}

	// Key hints
	hints := []string{
		StatusKeyStyle.Render("K/Space") + ":select",
		StatusKeyStyle.Render("L/→") + ":open",
		StatusKeyStyle.Render("J/←") + ":back",
		StatusKeyStyle.Render("Enter") + ":download",
		StatusKeyStyle.Render("I") + ":invert",
		StatusKeyStyle.Render("/") + ":search",
		StatusKeyStyle.Render("O") + ":icons",
		StatusKeyStyle.Render("?") + ":help",
	}

	hintsStr := StatusBarStyle.Render(strings.Join(hints, " │ "))

	// Combine
	left := strings.Join(parts, " │ ")
	right := hintsStr

	// Calculate spacing
	spacing := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if spacing < 0 {
		spacing = 1
	}

	line := left + strings.Repeat(" ", spacing) + right
	runes := []rune(line)
	if len(runes) > m.width {
		return string(runes[:max(0, m.width)])
	}
	return line
}

func stripANSI(s string) string {
	var b strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == 0x1b {
			inEscape = true
			continue
		}
		if inEscape {
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
				inEscape = false
			}
			continue
		}
		b.WriteByte(ch)
	}
	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// highlightCode applies syntax highlighting to code content
func highlightCode(content string, filePath string) string {
	// Get lexer based on file extension
	ext := filepath.Ext(filePath)
	lexer := lexers.Get(strings.TrimPrefix(ext, "."))
	
	// Fall back to plain text if language not recognized
	if lexer == nil {
		// Try to guess from filename
		lexer = lexers.Match(filePath)
		if lexer == nil {
			lexer = lexers.Fallback
		}
	}
	
	// Use a terminal-friendly formatter with ANSI colors
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Get("terminal")
	}
	
	// Get a style suitable for dark terminals
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}
	
	// Tokenize and format
	tokens, err := lexer.Tokenise(nil, content)
	if err != nil {
		return content // Fall back to plain text on error
	}
	
	var buf strings.Builder
	err = formatter.Format(&buf, style, tokens)
	if err != nil {
		return content // Fall back to plain text on error
	}
	
	return buf.String()
}

// viewPreview renders the file preview modal
func (m Model) viewPreview() string {
	var b strings.Builder

	// Title
	title := PreviewTitleStyle.Render("Preview: " + m.state.PreviewPath)
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.width-4))
	b.WriteString("\n")

	// Apply syntax highlighting
	highlightedContent := highlightCode(m.state.PreviewContent, m.state.PreviewPath)
	
	// Content
	lines := strings.Split(highlightedContent, "\n")
	maxLines := m.height - 8
	if maxLines < 5 {
		maxLines = 5
	}

	start := m.state.PreviewScroll
	if start > len(lines)-maxLines {
		start = len(lines) - maxLines
	}
	if start < 0 {
		start = 0
	}

	end := start + maxLines
	if end > len(lines) {
		end = len(lines)
	}

	for i := start; i < end; i++ {
		line := lines[i]
		if len(line) > m.width-4 {
			line = line[:m.width-7] + "..."
		}
		b.WriteString(PreviewContentStyle.Render(line))
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.width-4))
	b.WriteString("\n")

	scrollInfo := fmt.Sprintf("Line %d-%d of %d", start+1, end, len(lines))
	hints := "↑/↓: scroll │ Esc: close"
	footer := StatusBarStyle.Render(scrollInfo + "  │  " + hints)
	b.WriteString(footer)

	// Wrap in modal border
	return ModalStyle.
		Width(m.width - 4).
		Height(m.height - 2).
		Render(b.String())
}

// viewHelp renders the help screen
func (m Model) viewHelp() string {
	var b strings.Builder

	title := TitleStyle.Render("dgf Help - Keyboard Shortcuts")
	b.WriteString(title)
	b.WriteString("\n\n")

	sections := []struct {
		name  string
		bindings []struct{ key, desc string }
	}{
		{
			name: "Navigation",
			bindings: []struct{ key, desc string }{
				{"↑", "Move up"},
				{"↓", "Move down"},
				{"J/←", "Go back"},
				{"L/→", "Enter folder"},
				{"Home/g", "Go to top"},
				{"End/G", "Go to bottom"},
			},
		},
		{
			name: "Selection",
			bindings: []struct{ key, desc string }{
				{"K/Space", "Toggle selection"},
				{"a", "Select all"},
				{"u", "Unselect all"},
				{"I", "Inverse (confirm)"},
			},
		},
		{
			name: "Actions",
			bindings: []struct{ key, desc string }{
				{"Enter", "Download selected (confirm)"},
				{"p", "Preview file"},
				{"/", "Search"},
				{"r", "Refresh"},
				{"o", "Toggle icons"},
			},
		},
		{
			name: "General",
			bindings: []struct{ key, desc string }{
				{"?", "Toggle help"},
				{"Esc", "Close/Cancel"},
				{"q/Ctrl+C", "Quit"},
			},
		},
	}

	for _, section := range sections {
		b.WriteString(SubtitleStyle.Render(section.name))
		b.WriteString("\n")
		for _, binding := range section.bindings {
			key := HelpKeyStyle.Width(20).Render(binding.key)
			desc := HelpDescStyle.Render(binding.desc)
			b.WriteString(fmt.Sprintf("  %s %s\n", key, desc))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(StatusBarStyle.Render("Press Esc or ? to close"))

	return ModalStyle.
		Width(m.width - 8).
		Height(m.height - 4).
		Render(b.String())
}

// viewDownload renders the download progress screen
func (m Model) viewDownload() string {
	var b strings.Builder

	title := TitleStyle.Render("Downloading Files...")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Progress bar with percentage
	barWidth := 50
	filled := int(m.state.DownloadProgress * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}

	bar := "[" + strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled) + "]"
	percentage := int(m.state.DownloadProgress * 100)
	progressLine := fmt.Sprintf("%s %3d%%", bar, percentage)
	b.WriteString(ProgressBarStyle.Render(progressLine))
	b.WriteString("\n\n")

	// Progress info: completed / total files
	info := fmt.Sprintf("Files: %d / %d", m.state.DownloadDone, m.state.DownloadTotal)
	b.WriteString(ProgressTextStyle.Render(info))
	b.WriteString("\n\n")

	// Current file being downloaded
	if m.state.DownloadCurrent != "" {
		currentLabel := "Current File:"
		b.WriteString(StatusBarStyle.Render(currentLabel))
		b.WriteString("\n")
		
		current := "  " + m.state.DownloadCurrent
		// Truncate with ellipsis if too long
		maxLen := m.width - 6
		if len(current) > maxLen {
			current = current[:maxLen-3] + "..."
		}
		b.WriteString(StatusBarStyle.Render(current))
		b.WriteString("\n")
	} else {
		b.WriteString(StatusBarStyle.Render("Initializing download..."))
		b.WriteString("\n")
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}

// overlayToast adds a toast notification overlay
func (m Model) overlayToast(content string) string {
	if m.state.Toast == nil {
		return content
	}

	var style lipgloss.Style
	switch m.state.Toast.Type {
	case types.ToastSuccess:
		style = ToastSuccessStyle
	case types.ToastError:
		style = ToastErrorStyle
	case types.ToastWarning:
		style = ToastWarningStyle
	default:
		style = ToastInfoStyle
	}

	toast := style.Render(m.state.Toast.Message)

	// Position at top-right
	toastWidth := lipgloss.Width(toast)
	x := m.width - toastWidth - 2
	if x < 0 {
		x = 0
	}

	// Overlay the toast on the content
	lines := strings.Split(content, "\n")
	if len(lines) > 1 {
		// Replace part of line 1 with toast
		line := lines[1]
		if x < len(line) {
			lines[1] = line[:x] + toast
		} else {
			lines[1] = line + strings.Repeat(" ", x-len(line)) + toast
		}
	}

	return strings.Join(lines, "\n")
}
