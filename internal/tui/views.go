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

// viewInput renders the URL input screen with better guidance
func (m Model) viewInput() string {
	var b strings.Builder

	// Centered logo with enhanced styling
	logoStyle := lipgloss.NewStyle().
		Foreground(ColorLogo).
		Bold(true).
		Align(lipgloss.Center)

	enhancedLogo := `
    ____  ____________
   / __ \/ ____/ ____/
  / / / / / __/ /_    
 / /_/ / /_/ / __/    
/_____/\____/_/       
                      
 v2.0 • Terminal UI
`
	
	b.WriteString(logoStyle.Render(enhancedLogo))
	b.WriteString("\n\n")

	// Tagline
	taglineStyle := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Italic(true).
		Align(lipgloss.Center)

	b.WriteString(taglineStyle.Render("Download files from GitHub repos without full clones"))
	b.WriteString("\n\n")

	// URL input box with better styling
	inputBoxStyle := InputFocusedStyle.
		Width(70).
		Align(lipgloss.Center)

	b.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, inputBoxStyle.Render(m.urlInput.View())))
	b.WriteString("\n\n")

	// Enhanced hints section
	hintTitleStyle := lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)
	
	hintStyle := lipgloss.NewStyle().
		Foreground(ColorSubtle)
	
	exampleStyle := lipgloss.NewStyle().
		Foreground(ColorInfo)

	hints := []string{
		hintTitleStyle.Render("📝 Examples:"),
		exampleStyle.Render("  • github.com/charmbracelet/bubbletea"),
		exampleStyle.Render("  • https://github.com/golang/go/tree/master/src"),
		exampleStyle.Render("  • github.com/user/repo/tree/branch/path"),
		"",
		hintStyle.Render("💡 Tip: Paste any GitHub URL and press Enter"),
		"",
		hintTitleStyle.Render("⌨️  Controls:"),
		hintStyle.Render("  Enter: Browse repository  •  ?: Help  •  q: Quit"),
	}

	for _, hint := range hints {
		b.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, hint))
		b.WriteString("\n")
	}

	// Show error if present
	if m.state.Error != "" {
		b.WriteString("\n")
		errorBox := lipgloss.NewStyle().
			Background(ColorError).
			Foreground(ColorBackground).
			Bold(true).
			Padding(0, 2).
			Width(70).
			Align(lipgloss.Center).
			Render("❌ " + m.state.Error)
		
		b.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, errorBox))
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}

// viewLoading renders the enhanced loading screen
func (m Model) viewLoading() string {
	var b strings.Builder

	// Animated loading box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Padding(2, 4).
		Align(lipgloss.Center)

	spinnerStyle := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)
	
	messageStyle := lipgloss.NewStyle().
		Foreground(ColorForeground)

	spinnerStr := m.spinner.View()
	
	var message string
	if m.state.Owner != "" && m.state.Repo != "" {
		message = fmt.Sprintf("Fetching %s/%s...", m.state.Owner, m.state.Repo)
		if m.state.Path != "" {
			message += fmt.Sprintf("\n📂 %s", m.state.Path)
		}
	} else {
		message = "Connecting to GitHub..."
	}

	content := spinnerStyle.Render(spinnerStr) + " " + messageStyle.Render(message)
	
	b.WriteString(boxStyle.Render(content))
	b.WriteString("\n\n")
	
	// Helpful tip while loading
	tipStyle := lipgloss.NewStyle().
		Foreground(ColorSubtle).
		Italic(true)
	b.WriteString(tipStyle.Render("Tip: You can select multiple files before downloading"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}

// viewBrowser renders the file browser
func (m Model) viewBrowser() string {
	var b strings.Builder

	// Centered title
	titleStyle := lipgloss.NewStyle().
		Foreground(ColorLogo).
		Bold(true).
		Align(lipgloss.Center)
	
	title := titleStyle.Width(m.width).Render("Direct Git Fetch v2.0")
	b.WriteString(title)
	b.WriteString("\n")
	
	// Separator line
	separatorStyle := lipgloss.NewStyle().Foreground(ColorBorder)
	separator := separatorStyle.Render(strings.Repeat("─", m.width))
	b.WriteString(separator)
	b.WriteString("\n")

	// Breadcrumb path (centered)
	breadcrumb := m.state.GetBreadcrumb()
	if breadcrumb != "" {
		pathStyle := lipgloss.NewStyle().
			Foreground(ColorBreadcrumb).
			Align(lipgloss.Center)
		
		path := pathStyle.Width(m.width).Render(breadcrumb)
		b.WriteString(path)
		b.WriteString("\n")
		b.WriteString(separator)
		b.WriteString("\n")
	}

	// File list
	items := m.state.GetVisibleItems()
	icons := GetIcons(m.state.ASCIIMode)

	// Keep selection state in sync before rendering
	m.selection.SyncWithItems(m.state.Items)

	// Calculate visible area
	listHeight := m.height - 8 // Leave room for title, breadcrumb, table header, status
	if listHeight < 5 {
		listHeight = 5
	}

	// Adjust scroll offset to keep cursor visible
	if m.state.Cursor < m.state.ScrollOffset {
		m.state.ScrollOffset = m.state.Cursor
	} else if m.state.Cursor >= m.state.ScrollOffset+listHeight {
		m.state.ScrollOffset = m.state.Cursor - listHeight + 1
	}

	// Render table header with clean columns
	headerStyle := lipgloss.NewStyle().
		Foreground(ColorForeground).
		Bold(true)
	
	// Column layout: SELECT (6) | NAME (40) | FILE TYPE (15) | SIZE (10)
	tableHeader := fmt.Sprintf("%-6s %-40s %-15s %10s", "SELECT", "NAME", "FILE TYPE", "SIZE")
	b.WriteString(headerStyle.Render(tableHeader))
	b.WriteString("\n")
	b.WriteString(separator)
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

	// Search overlay with enhanced styling
	if m.state.Mode == types.ModeSearch {
		searchBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorInfo).
			Background(ColorBackground).
			Foreground(ColorForeground).
			Padding(0, 1).
			Width(50)

		searchContent := "🔍 " + m.searchInput.View()
		
		searchInfo := lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true).
			Render(fmt.Sprintf(" (%d results)", len(m.state.FilteredItems)))

		b.WriteString("\n")
		b.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, searchBoxStyle.Render(searchContent)+searchInfo))
		b.WriteString("\n")
	}

	// Footer status bar
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar(icons))

	return BaseStyle.Render(b.String())
}

// renderItem renders a single file/folder item with enhanced styling
func (m Model) renderItem(item types.RepoItem, isCursor bool, icons IconSet) string {
	// Selection marker
	selectionMarker := "[ ]"
	selColor := ColorSubtle
	if item.Selected {
		selectionMarker = "[●]"
		selColor = ColorSelected
	}

	// Get file type and icon
	fileType, fileIcon := GetFileTypeAndIcon(item.Name, item.IsDir())
	
	// Name styling
	nameColor := ColorFile
	if item.IsDir() {
		nameColor = ColorFolder
	}
	if item.IsLFS {
		fileIcon = icons.LFS
		fileType = "lfs"
		nameColor = ColorLFS
	}

	// Format name with icon (40 chars total including icon)
	displayName := fileIcon + " " + item.Name
	nameLimit := 38
	if len(displayName) > nameLimit {
		// Account for emoji width
		displayName = displayName[:nameLimit-3] + "..."
	}

	// Size display
	var sizeStr string
	sizeColor := ColorSubtle
	if item.IsFile() {
		sizeStr = utils.FormatBytes(item.Size)
		if item.Size > 10*1024*1024 { // > 10MB
			sizeColor = ColorWarning
		}
		if item.Size > 50*1024*1024 { // > 50MB
			sizeColor = ColorError
		}
	} else {
		sizeStr = "—"
	}

	// Truncate file type if needed
	if len(fileType) > 13 {
		fileType = fileType[:13]
	}

	// Build line with proper spacing
	// SELECT (6) | NAME (40) | FILE TYPE (15) | SIZE (10)
	selPart := lipgloss.NewStyle().Foreground(selColor).Bold(item.Selected).Render(selectionMarker)
	namePart := lipgloss.NewStyle().Foreground(nameColor).Bold(item.IsDir()).Render(displayName)
	typePart := fileType
	sizePart := lipgloss.NewStyle().Foreground(sizeColor).Render(sizeStr)
	
	// Pad to exact widths
	line := fmt.Sprintf("%-6s %-40s %-15s %10s", selPart, namePart, typePart, sizePart)

	// Apply cursor highlighting
	if isCursor {
		return CursorStyle.Width(m.width).Render(line)
	}

	return line
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

// renderStatusBar renders the enhanced bottom status bar
func (m Model) renderStatusBar(_ IconSet) string {
	// Build status bar sections
	var left, center, right []string

	// Left: Selection info
	if m.selection.HasSelection() {
		count := m.selection.Count()
		size := utils.FormatBytes(m.selection.TotalSize())
		left = append(left, StatusValueStyle.Render(fmt.Sprintf("%d", count))+" selected "+StatusValueStyle.Render(size))
	} else {
		left = append(left, StatusBarStyle.Foreground(ColorSubtle).Render("No selection"))
	}

	// Center: Position and context
	items := m.state.GetVisibleItems()
	if len(items) > 0 {
		center = append(center, StatusKeyStyle.Render(fmt.Sprintf("%d/%d", m.state.Cursor+1, len(items))))
		if m.state.Path != "" {
			center = append(center, StatusBarStyle.Foreground(ColorSubtle).Render(fmt.Sprintf("• %d items", len(items))))
		} else {
			center = append(center, StatusBarStyle.Foreground(ColorSubtle).Render(fmt.Sprintf("• %d items • root", len(items))))
		}
	}
	
	// Add download directory hint
	if m.state.DownloadPath != "" && m.state.DownloadPath != "." {
		center = append(center, StatusBarStyle.Foreground(ColorSubtle).Render("• "+m.state.DownloadPath))
	}

	// Right: Key hints (essential only - user can press ? for full help)
	hints := []string{
		StatusKeyStyle.Render("D") + ":download",
		StatusKeyStyle.Render("⏎") + ":open",
		StatusKeyStyle.Render("Esc") + ":back",
		StatusKeyStyle.Render("?") + ":help",
	}
	right = append(right, strings.Join(hints, " "))

	// Build full status bar
	leftStr := strings.Join(left, " │ ")
	centerStr := strings.Join(center, " ")
	rightStr := strings.Join(right, " ")

	// Calculate spacing
	leftWidth := lipgloss.Width(leftStr)
	centerWidth := lipgloss.Width(centerStr)
	rightWidth := lipgloss.Width(rightStr)
	
	totalContentWidth := leftWidth + centerWidth + rightWidth
	if totalContentWidth >= m.width-4 {
		// Too wide, simplify
		return StatusBarStyle.Width(m.width).Render(leftStr + " " + rightStr)
	}
	
	// Calculate center position
	centerStart := (m.width - centerWidth) / 2
	leftSpace := centerStart - leftWidth
	rightSpace := m.width - centerStart - centerWidth - rightWidth
	
	if leftSpace < 2 {
		leftSpace = 2
	}
	if rightSpace < 2 {
		rightSpace = 2
	}

	statusLine := leftStr + strings.Repeat(" ", leftSpace) + centerStr + strings.Repeat(" ", rightSpace) + rightStr
	
	return StatusBarStyle.Width(m.width).Render(statusLine)
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

// viewPreview renders the file preview modal with line numbers
func (m Model) viewPreview() string {
	var b strings.Builder

	// Title bar with file info
	titleBar := lipgloss.NewStyle().
		Background(ColorAccent).
		Foreground(ColorBackground).
		Bold(true).
		Padding(0, 2).
		Width(m.width - 4)
	
	fileInfo := fmt.Sprintf("📄 %s", m.state.PreviewPath)
	b.WriteString(titleBar.Render(fileInfo))
	b.WriteString("\n")

	// Apply syntax highlighting
	highlightedContent := highlightCode(m.state.PreviewContent, m.state.PreviewPath)
	
	// Content with line numbers
	lines := strings.Split(highlightedContent, "\n")
	maxLines := m.height - 10
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

	// Line number styling
	lineNumStyle := lipgloss.NewStyle().
		Foreground(ColorSubtle).
		Width(5).
		Align(lipgloss.Right)
	
	contentWidth := m.width - 12 // Account for line numbers and padding

	for i := start; i < end; i++ {
		lineNum := lineNumStyle.Render(fmt.Sprintf("%d", i+1))
		line := lines[i]
		
		// Truncate if too long
		if len(line) > contentWidth {
			line = line[:contentWidth-3] + "..."
		}
		
		lineContent := fmt.Sprintf("%s │ %s", lineNum, line)
		b.WriteString(lineContent)
		b.WriteString("\n")
	}

	// Footer with navigation info
	b.WriteString("\n")
	footerStyle := lipgloss.NewStyle().
		Background(ColorBorder).
		Foreground(ColorForeground).
		Padding(0, 2).
		Width(m.width - 4)
	
	scrollInfo := fmt.Sprintf("Lines %d-%d of %d", start+1, end, len(lines))
	hints := "  │  ↑↓: scroll  │  ESC: close  │  ?: help"
	footer := scrollInfo + hints
	b.WriteString(footerStyle.Render(footer))

	// Wrap in modal border
	return BorderStyle.
		BorderForeground(ColorAccent).
		Width(m.width - 2).
		Height(m.height - 2).
		Padding(1).
		Render(b.String())
}

// viewHelp renders the enhanced help screen
func (m Model) viewHelp() string {
	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Background(ColorAccent).
		Foreground(ColorBackground).
		Bold(true).
		Padding(0, 2).
		Width(m.width - 8)
	
	b.WriteString(headerStyle.Render("DGF Keyboard Shortcuts & Help"))
	b.WriteString("\n\n")

	// Sections with enhanced styling
	sectionTitleStyle := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		Underline(true)
	
	keyStyle := lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true).
		Width(18)
	
	descStyle := lipgloss.NewStyle().
		Foreground(ColorForeground)

	sections := []struct {
		name     string
		bindings []struct{ key, desc string }
	}{
		{
			name: "Navigation",
			bindings: []struct{ key, desc string }{
				{"↑ / ↓", "Move cursor up/down"},
				{"Home / g", "Jump to top"},
				{"End / G", "Jump to bottom"},
				{"J / ← / Esc", "Go back to parent folder"},
				{"L / → / Enter", "Enter selected folder"},
			},
		},
		{
			name: "Selection",
			bindings: []struct{ key, desc string }{
				{"K / Space", "Toggle selection of current item"},
				{"a", "Select all items"},
				{"u", "Unselect all items"},
				{"I (twice)", "Invert selection (with confirmation)"},
			},
		},
		{
			name: "Actions",
			bindings: []struct{ key, desc string }{
				{"D / Shift+Enter", "Download selected items"},
				{"p", "Preview file content"},
				{"/", "Search files"},
				{"r", "Refresh current view"},
				{"o", "Toggle icon mode (emoji/ASCII)"},
				{"y", "Copy current item path to clipboard"},
				{"C", "Specify commit ID"},
				{"B", "Select branch"},
			},
		},
		{
			name: "General",
			bindings: []struct{ key, desc string }{
				{"?", "Toggle this help screen"},
				{"Esc", "Cancel/Close/Back"},
				{"q / Ctrl+C", "Quit application"},
			},
		},
	}

	for _, section := range sections {
		b.WriteString(sectionTitleStyle.Render(section.name))
		b.WriteString("\n")
		for _, binding := range section.bindings {
			key := keyStyle.Render("  " + binding.key)
			desc := descStyle.Render(binding.desc)
			b.WriteString(fmt.Sprintf("%s  %s\n", key, desc))
		}
		b.WriteString("\n")
	}

	// Tips section
	b.WriteString(sectionTitleStyle.Render("Tips"))
	b.WriteString("\n")
	
	tipStyle := lipgloss.NewStyle().
		Foreground(ColorInfo).
		Italic(true)
	
	tips := []string{
		"• Use [●] markers to see which items are selected",
		"• Download directory can be set via 'dgf config set download_path <path>'",
		"• Support for GitHub tokens: 'dgf config set github_token <token>'",
		"• Directories are cached in memory - navigating back is instant",
		"• Search is case-insensitive and matches filenames",
	}
	
	for _, tip := range tips {
		b.WriteString(tipStyle.Render("  "+tip))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	
	// Footer
	footerStyle := lipgloss.NewStyle().
		Foreground(ColorSubtle).
		Italic(true)
	b.WriteString(footerStyle.Render("Press Esc or ? to close this help screen"))

	// Wrap in modal
	return BorderStyle.
		BorderForeground(ColorAccent).
		Width(m.width - 6).
		Height(m.height - 4).
		Padding(2).
		Render(b.String())
}

// viewDownload renders the enhanced download progress screen
func (m Model) viewDownload() string {
	var b strings.Builder

	// Progress box with border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSuccess).
		Padding(2, 4).
		Width(70).
		Align(lipgloss.Center)

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)
	
	b.WriteString(titleStyle.Render("Downloading Files"))
	b.WriteString("\n\n")

	// Progress bar with enhanced visuals
	barWidth := 50
	filled := int(m.state.DownloadProgress * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}

	barStyle := lipgloss.NewStyle().Foreground(ColorSuccess)
	emptyStyle := lipgloss.NewStyle().Foreground(ColorBorder)
	
	bar := barStyle.Render(strings.Repeat("█", filled)) + emptyStyle.Render(strings.Repeat("░", barWidth-filled))
	percentage := int(m.state.DownloadProgress * 100)
	progressLine := fmt.Sprintf("[%s] %3d%%", bar, percentage)
	b.WriteString(progressLine)
	b.WriteString("\n\n")

	// Progress stats
	statsStyle := lipgloss.NewStyle().
		Foreground(ColorInfo).
		Bold(true)
	
	info := fmt.Sprintf("Files: %s/%d", statsStyle.Render(fmt.Sprintf("%d", m.state.DownloadDone)), m.state.DownloadTotal)
	b.WriteString(info)
	b.WriteString("\n\n")

	// Current file
	if m.state.DownloadCurrent != "" {
		labelStyle := lipgloss.NewStyle().
			Foreground(ColorSubtle)
		
		fileStyle := lipgloss.NewStyle().
			Foreground(ColorAccent).
			Italic(true)
		
		b.WriteString(labelStyle.Render("Current: "))
		
		current := m.state.DownloadCurrent
		maxLen := 60
		if len(current) > maxLen {
			current = "..." + current[len(current)-maxLen+3:]
		}
		b.WriteString(fileStyle.Render(current))
		b.WriteString("\n")
	} else {
		statusStyle := lipgloss.NewStyle().
			Foreground(ColorSubtle).
			Italic(true)
		b.WriteString(statusStyle.Render("Preparing download..."))
		b.WriteString("\n")
	}

	content := boxStyle.Render(b.String())
	
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
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
