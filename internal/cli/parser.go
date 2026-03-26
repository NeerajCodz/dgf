package cli

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/NeerajCodz/dgf/pkg/types"
	"github.com/spf13/pflag"
)

//go:embed format.json
var formatsData []byte

// ParseArgs parses command-line arguments into a types.Args struct
func ParseArgs() types.Args {
	var args types.Args
	var format string

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, `dgf - Direct Git Fetch v2.0.0

Usage:
  dgf                                    Launch interactive TUI
  dgf <URL>                              Download from URL (CLI mode)
  dgf -s <site> -u <user> -r <repo>      Download using components (CLI mode)
  dgf config <get|set|show> [key] [val]  Manage configuration
  dgf agent <tree|download> <url> ...    JSON API mode for scripts

Options:
  --site, -s <site>           Platform ID (e.g., github)
  --username, -u <username>   Repository owner
  --repo, -r <repo>           Repository name
  --token, -t <token>         GitHub token (or use GITHUB_TOKEN env)
  --branch, -b <branch>       Branch name
  --commit, -c <commit>       Commit SHA
  --path, -p <path>           Path in repository
  --output, -o <dir>          Output directory (default: .)
  --format, -f <format>       File formats (e.g., image, [jpg,pdf], or "" for no-extension)
  --no-print, -n              Suppress all output
  --print-tree                Print directory tree
  --check                     Check if path exists (returns JSON)
  --print-info, -i            Print repository info as JSON
  --help, -h                  Show this help message

Note: Only one of --no-print, --print-tree, --check, or --print-info can be used.
`)
	}

	// Define flags
	pflag.StringVarP(&args.Site, "site", "s", "", "Platform ID")
	pflag.StringVarP(&args.Username, "username", "u", "", "Repository owner")
	pflag.StringVarP(&args.Repo, "repo", "r", "", "Repository name")
	pflag.StringVarP(&args.Token, "token", "t", "", "GitHub token")
	pflag.StringVarP(&args.Branch, "branch", "b", "", "Branch name")
	pflag.StringVarP(&args.Commit, "commit", "c", "", "Commit SHA")
	pflag.StringVarP(&args.Path, "path", "p", "", "Path in repository")
	pflag.StringVarP(&args.Output, "output", "o", ".", "Output directory")
	pflag.StringVarP(&format, "format", "f", "", "File formats")
	pflag.BoolVarP(&args.NoPrint, "no-print", "n", false, "Suppress output")
	pflag.BoolVar(&args.PrintTree, "print-tree", false, "Print directory tree")
	pflag.BoolVar(&args.Check, "check", false, "Check path exists")
	pflag.BoolVarP(&args.PrintInfo, "print-info", "i", false, "Print info as JSON")
	help := pflag.BoolP("help", "h", false, "Show help")

	pflag.Parse()

	if *help {
		pflag.Usage()
		os.Exit(0)
	}

	// Validate output mode exclusivity
	count := 0
	if args.NoPrint {
		count++
	}
	if args.PrintTree {
		count++
	}
	if args.Check {
		count++
	}
	if args.PrintInfo {
		count++
	}
	if count > 1 {
		fmt.Fprintf(os.Stderr, "Error: Only one of --no-print, --print-tree, --check, or --print-info can be provided\n")
		pflag.Usage()
		os.Exit(1)
	}

	// Check for positional URL argument
	if pflag.NArg() > 0 {
		args.URL = pflag.Arg(0)
	}

	// Process format flag
	if format != "" {
		args.Formats = parseFormats(format)
	}

	// Normalize path
	if args.Path != "" {
		args.Path = strings.Trim(args.Path, "/")
	}

	// Token from environment if not provided
	if args.Token == "" {
		args.Token = os.Getenv("GITHUB_TOKEN")
	}

	// Normalize output directory
	if args.Output != "" {
		args.Output = strings.TrimRight(args.Output, "/")
	}

	return args
}

// parseFormats parses the format string into a slice of extensions
func parseFormats(format string) []string {
	if format == `""` || format == "" {
		return []string{""}
	}

	var formatsMap map[string]map[string][]string
	if err := json.Unmarshal(formatsData, &formatsMap); err == nil {
		if formats, exists := formatsMap["formats"][format]; exists {
			for i, ext := range formats {
				formats[i] = strings.ToLower(ext)
			}
			return formats
		}
	}

	cleanFormat := strings.Trim(format, "[]")
	if cleanFormat != "" {
		extensions := strings.Split(cleanFormat, ",")
		for i, ext := range extensions {
			extensions[i] = strings.TrimSpace(strings.ToLower(ext))
		}
		return extensions
	}

	return nil
}

// HasCLIArgs returns true if arguments indicate CLI mode should be used
func HasCLIArgs(args types.Args) bool {
	return args.URL != "" || args.Site != "" || args.Username != "" || args.Repo != "" ||
		args.PrintTree || args.Check || args.PrintInfo || len(args.Formats) > 0
}
