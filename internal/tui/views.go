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
	logoStyle := lipgloss.NewStyle().
		Foreground(ColorLogo).
		Bold(true)
	
	logoText := "╔═══════════════════════════════════════╗\n"
	logoText += "║   Direct Git Fetch • DGF v2.0        ║\n"
	logoText += "╚═══════════════════════════════════════╝"
	
	b.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, logoStyle.Render(logoText)))
	b.WriteString("\n")

	// Breadcrumb bar
	breadcrumb := m.state.GetBreadcrumb()
	if breadcrumb != "" {
		breadcrumbStyle := lipgloss.NewStyle().
			Background(ColorBorder).
			Foreground(ColorBreadcrumb).
			Bold(true).
			Padding(0, 2).
			Width(m.width)

		b.WriteString(breadcrumbStyle.Render("📂 " + breadcrumb))
		b.WriteString("\n")
	} else {
		// Show helpful message when no repo loaded
		breadcrumbStyle := lipgloss.NewStyle().
			Background(ColorBorder).
			Foreground(ColorSubtle).
			Italic(true).
			Padding(0, 2).
			Width(m.width)
		b.WriteString(breadcrumbStyle.Render("⚡ Ready to browse GitHub repositories"))
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

	// Render table header with styling
	headerStyle := lipgloss.NewStyle().
		Background(ColorBorder).
		Foreground(ColorForeground).
		Bold(true).
		Width(m.width)
	
	tableHeader := fmt.Sprintf("  %-4s %-2s %-40s %-10s", "SEL", "T", "NAME", "SIZE")
	b.WriteString(headerStyle.Render(tableHeader))
	b.WriteString("\n")
	
	// Add separator line
	separator := lipgloss.NewStyle().
		Foreground(ColorBorder).
		Render(strings.Repeat("─", m.width))
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

// renderItem renders a single file/folder item with enhanced styling
func (m Model) renderItem(item types.RepoItem, isCursor bool, icons IconSet) string {
	// Selection marker with color
	selectionMarker := "[ ]"
	selectionStyle := lipgloss.NewStyle().Foreground(ColorSubtle)
	if item.Selected {
		selectionMarker = "[●]"
		selectionStyle = lipgloss.NewStyle().Foreground(ColorSelected).Bold(true)
	}

	// Icon and name with appropriate styling
	var icon, name string
	nameStyle := FileStyle
	if item.IsDir() {
		icon = icons.Folder
		nameStyle = FolderStyle
		name = item.Name + "/"
	} else {
		icon = icons.File
		if item.IsLFS {
			icon = icons.LFS
			nameStyle = lipgloss.NewStyle().Foreground(ColorLFS).Bold(true)
		}
		name = item.Name
	}

	// Apply highlighting if searching
	if m.state.SearchQuery != "" {
		name = m.highlightMatches(name, nameStyle, isCursor)
	} else {
		name = nameStyle.Render(name)
	}

	// Size display
	var sizeStr string
	sizeStyle := lipgloss.NewStyle().Foreground(ColorSubtle)
	if item.IsFile() {
		sizeStr = utils.FormatBytes(item.Size)
		if item.Size > 1024*1024 { // > 1MB
			sizeStyle = lipgloss.NewStyle().Foreground(ColorWarning)
		}
	} else {
		sizeStr = "—"
	}

	// Type indicator
	itemType := "F"
	typeStyle := lipgloss.NewStyle().Foreground(ColorInfo)
	if item.IsDir() {
		itemType = "D"
		typeStyle = lipgloss.NewStyle().Foreground(ColorFolder).Bold(true)
	}

	// Build line with proper spacing
	displayName := stripANSI(fmt.Sprintf("%s %s", icon, name))
	if len(displayName) > 40 {
		displayName = displayName[:37] + "..."
	}
	
	lineContent := fmt.Sprintf("%s %s %-40s %s",
		selectionStyle.Render(selectionMarker),
		typeStyle.Render(itemType),
		displayName,
		sizeStyle.Render(fmt.Sprintf("%10s", sizeStr)))

	// Apply cursor highlighting
	if isCursor {
		return CursorStyle.Width(m.width).Render(" "+lineContent+" ")
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

	// Right: Key hints (compact)
	hints := []string{
		StatusKeyStyle.Render("K") + ":sel",
		StatusKeyStyle.Render("L") + ":open",
		StatusKeyStyle.Render("J") + ":back",
		StatusKeyStyle.Render("⏎") + ":dl",
		StatusKeyStyle.Render("I") + ":inv",
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
	
	b.WriteString(headerStyle.Render("⌨️  DGF Keyboard Shortcuts & Help"))
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
			name: "🧭 Navigation",
			bindings: []struct{ key, desc string }{
				{"↑ / ↓", "Move cursor up/down"},
				{"Home / g", "Jump to top"},
				{"End / G", "Jump to bottom"},
				{"J / ←", "Go back to parent folder"},
				{"L / →", "Enter selected folder"},
			},
		},
		{
			name: "✓ Selection",
			bindings: []struct{ key, desc string }{
				{"K / Space", "Toggle selection of current item"},
				{"a", "Select all items"},
				{"u", "Unselect all items"},
				{"I (twice)", "Invert selection (with confirmation)"},
			},
		},
		{
			name: "⚡ Actions",
			bindings: []struct{ key, desc string }{
				{"Enter", "Download selected items"},
				{"p", "Preview file content"},
				{"/", "Search files"},
				{"r", "Refresh current view"},
				{"o", "Toggle icon mode (emoji/ASCII)"},
				{"y", "Copy current item path to clipboard"},
			},
		},
		{
			name: "ℹ️  General",
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
	b.WriteString(sectionTitleStyle.Render("💡 Tips"))
	b.WriteString("\n")
	
	tipStyle := lipgloss.NewStyle().
		Foreground(ColorInfo).
		Italic(true)
	
	tips := []string{
		"• Use [●] markers to see which items are selected",
		"• Download directory can be set via 'dgf config set download_path <path>'",
		"• Support for GitHub tokens: 'dgf config set github_token <token>'",
		"• Press Enter twice when downloading to confirm your selection",
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
