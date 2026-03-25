package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme colors for the TUI - Tokyo Night inspired
var (
	// Base colors
	ColorBackground = lipgloss.Color("#1a1b26")
	ColorForeground = lipgloss.Color("#c0caf5")
	ColorSubtle     = lipgloss.Color("#565f89")
	ColorHighlight  = lipgloss.Color("#7aa2f7")
	ColorBorder     = lipgloss.Color("#3b4261")
	ColorAccent     = lipgloss.Color("#bb9af7")

	// Semantic colors
	ColorSuccess = lipgloss.Color("#9ece6a")
	ColorWarning = lipgloss.Color("#e0af68")
	ColorError   = lipgloss.Color("#f7768e")
	ColorInfo    = lipgloss.Color("#7dcfff")

	// Item colors
	ColorFolder     = lipgloss.Color("#7aa2f7")
	ColorFile       = lipgloss.Color("#c0caf5")
	ColorSelected   = lipgloss.Color("#9ece6a")
	ColorCursor     = lipgloss.Color("#292e42")
	ColorCursorText = lipgloss.Color("#bb9af7")
	ColorLFS        = lipgloss.Color("#ff9e64")
	
	// Header colors
	ColorLogo       = lipgloss.Color("#bb9af7")
	ColorBreadcrumb = lipgloss.Color("#7dcfff")
)

// Styles
var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Background(ColorBackground).
			Foreground(ColorForeground)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle).
			Italic(true)

	// Border styles
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)

	ActiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorHighlight)

	// Item styles
	FolderStyle = lipgloss.NewStyle().
			Foreground(ColorFolder).
			Bold(true)

	FileStyle = lipgloss.NewStyle().
			Foreground(ColorFile)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorSelected).
			Bold(true)

	CursorStyle = lipgloss.NewStyle().
			Background(ColorCursor).
			Foreground(ColorCursorText).
			Bold(true)

	// Status bar styles
	StatusBarStyle = lipgloss.NewStyle().
			Background(ColorBorder).
			Foreground(ColorForeground).
			Padding(0, 1)

	StatusKeyStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)
			
	StatusValueStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	// Toast styles
	ToastBaseStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Margin(1, 2)

	ToastSuccessStyle = ToastBaseStyle.
				Background(ColorSuccess).
				Foreground(ColorBackground)

	ToastErrorStyle = ToastBaseStyle.
			Background(ColorError).
			Foreground(ColorBackground)

	ToastWarningStyle = ToastBaseStyle.
				Background(ColorWarning).
				Foreground(ColorBackground)

	ToastInfoStyle = ToastBaseStyle.
			Background(ColorInfo).
			Foreground(ColorBackground)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	InputFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorHighlight).
				Padding(0, 1)

	// Help styles
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle)

	// Progress bar styles
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(ColorHighlight)

	ProgressTextStyle = lipgloss.NewStyle().
				Foreground(ColorSubtle)

	// Breadcrumb styles
	BreadcrumbStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight)

	BreadcrumbSepStyle = lipgloss.NewStyle().
				Foreground(ColorSubtle)

	// Preview styles
	PreviewTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorHighlight).
				Padding(0, 1)

	PreviewContentStyle = lipgloss.NewStyle().
				Foreground(ColorForeground)

	// Size display
	SizeStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle).
			Align(lipgloss.Right)

	// Modal styles
	ModalStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorHighlight).
			Padding(1, 2)

	// Search styles
	SearchStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorInfo).
			Padding(0, 1)

	MatchHighlightStyle = lipgloss.NewStyle().
				Background(ColorHighlight).
				Foreground(ColorBackground)
)

// Icons (emoji and ASCII versions)
type IconSet struct {
	Folder      string
	FolderOpen  string
	File        string
	Selected    string
	Unselected  string
	ArrowRight  string
	ArrowLeft   string
	ArrowUp     string
	ArrowDown   string
	Search      string
	Download    string
	Loading     string
	Success     string
	Error       string
	Warning     string
	Info        string
	LFS         string
}

var EmojiIcons = IconSet{
	Folder:      "📁",
	FolderOpen:  "📂",
	File:        "📄",
	Selected:    "✓",
	Unselected:  "○",
	ArrowRight:  "→",
	ArrowLeft:   "←",
	ArrowUp:     "↑",
	ArrowDown:   "↓",
	Search:      "🔍",
	Download:    "⬇️",
	Loading:     "⏳",
	Success:     "✅",
	Error:       "❌",
	Warning:     "⚠️",
	Info:        "ℹ️",
	LFS:         "📦",
}

var ASCIIIcons = IconSet{
	Folder:      "[DIR]",
	FolderOpen:  "[DIR]",
	File:        "[FILE]",
	Selected:    "[x]",
	Unselected:  "[ ]",
	ArrowRight:  "->",
	ArrowLeft:   "<-",
	ArrowUp:     "^",
	ArrowDown:   "v",
	Search:      "[?]",
	Download:    "[DL]",
	Loading:     "[...]",
	Success:     "[OK]",
	Error:       "[ERR]",
	Warning:     "[!]",
	Info:        "[i]",
	LFS:         "[LFS]",
}

// GetIcons returns the appropriate icon set
func GetIcons(ascii bool) IconSet {
	if ascii {
		return ASCIIIcons
	}
	return EmojiIcons
}
