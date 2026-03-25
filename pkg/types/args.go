package types

// Args represents command-line arguments
type Args struct {
	URL       string
	Site      string
	Username  string
	Repo      string
	Token     string
	Branch    string
	Commit    string
	Path      string
	NoPrint   bool
	PrintTree bool
	Check     bool
	PrintInfo bool
	Output    string
	Formats   []string
}

// HasCLIArgs returns true if any CLI arguments were provided that indicate CLI mode
func (a *Args) HasCLIArgs() bool {
	return a.URL != "" || a.Site != "" || a.Username != "" || a.Repo != "" ||
		a.PrintTree || a.Check || a.PrintInfo || len(a.Formats) > 0
}
