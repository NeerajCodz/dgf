package types

type Args struct {
	URL       string
	Token     string
	Branch    string
	Commit    string
	Path      string
	NoPrint   bool
	PrintTree bool
	Check     bool
	PrintInfo bool
	Output    string
}