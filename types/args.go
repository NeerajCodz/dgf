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

