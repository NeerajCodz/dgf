package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/NeerajCodz/dgf/types"
	"github.com/spf13/pflag"
)

// ParseArgs parses command-line arguments into a types.Args struct
func ParseArgs() types.Args {
	var args types.Args
	var format string // Temporary variable for --format flag

	// Define custom usage message
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage:
  ./dgf [ <URL> | -s <site> -u <username> -r <repo> ] [options]

Options:
  --site, -s <site>           Platform ID (e.g., github, gitlab, huggingface)
  --username, -u <username>   Repository username
  --repo, -r <repo>           Repository name
  --token, -t <token>         GitHub token
  --branch, -b <branch>       Branch name
  --commit, -c <commit>       Commit ID
  --path, -p <path>           Path in repository
  --output, -o <dir>          Output directory (default: .)
  --format, -f <format>       File formats to include (e.g., image, [jpg,pdf,png], or "" for no-extension files)
  --no-print, -n              Suppress all output
  --print-tree                Print directory tree
  --check                     Check if path exists
  --print-info, -i            Print repository info as JSON
  --help, -h                  Show this help message

Note: Only one of --no-print, --print-tree, --check, or --print-info can be provided.
`)
	}

	// Define command-line flags
	pflag.StringVarP(&args.Site, "site", "s", "", "Platform ID (e.g., github, gitlab, huggingface)")
	pflag.StringVarP(&args.Username, "username", "u", "", "Repository username")
	pflag.StringVarP(&args.Repo, "repo", "r", "", "Repository name")
	pflag.StringVarP(&args.Token, "token", "t", "", "GitHub token")
	pflag.StringVarP(&args.Branch, "branch", "b", "", "Branch name")
	pflag.StringVarP(&args.Commit, "commit", "c", "", "Commit ID")
	pflag.StringVarP(&args.Path, "path", "p", "", "Path in repository")
	pflag.StringVarP(&args.Output, "output", "o", ".", "Output directory for downloads (default: current directory)")
	pflag.StringVarP(&format, "format", "f", "", "File formats to include (e.g., image, [jpg,pdf,png])")
	pflag.BoolVarP(&args.NoPrint, "no-print", "n", false, "Suppress all output")
	pflag.BoolVar(&args.PrintTree, "print-tree", false, "Print directory tree")
	pflag.BoolVar(&args.Check, "check", false, "Check if path exists")
	pflag.BoolVarP(&args.PrintInfo, "print-info", "i", false, "Print info as JSON")

	// Help flag
	help := pflag.BoolP("help", "h", false, "Show this help message")
	pflag.Parse()

	// Show help and exit if --help is provided
	if *help {
		pflag.Usage()
		os.Exit(0)
	}

	// Validate that only one of NoPrint, PrintTree, Check, or PrintInfo is set
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

	// Validate input: either URL or site args, but not both
	hasSiteArgs := args.Site != "" || args.Username != "" || args.Repo != ""
	hasURL := pflag.NArg() == 1
	if (hasSiteArgs && hasURL) || (!hasSiteArgs && !hasURL) {
		fmt.Fprintf(os.Stderr, "Error: Must provide either a URL or all of --site, --username, and --repo\n")
		pflag.Usage()
		os.Exit(1)
	}

	// If site args are provided, ensure all are present
	if hasSiteArgs {
		if args.Site == "" || args.Username == "" || args.Repo == "" {
			fmt.Fprintf(os.Stderr, "Error: Must provide all of --site, --username, and --repo\n")
			pflag.Usage()
			os.Exit(1)
		}
	}

	// Set URL from positional argument if provided
	if hasURL {
		args.URL = pflag.Arg(0)
	}

	// Process --format flag using config/format.json
	if format != "" {
		if format == `""` || format == "" {
			// Handle -f "" or -f=""
			args.Formats = []string{""}
		} else {
			// Read formats configuration
			formatsData, err := os.ReadFile("config/format.json")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading format config file: %v\n", err)
				os.Exit(1)
			}

			var formatsMap map[string]map[string][]string
			if err := json.Unmarshal(formatsData, &formatsMap); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing format config file: %v\n", err)
				os.Exit(1)
			}

			// Check if format is a category (e.g., "image")
			if formats, exists := formatsMap["formats"][format]; exists {
				// Ensure all extensions are lowercase
				for i, ext := range formats {
					formats[i] = strings.ToLower(ext)
				}
				args.Formats = formats
			} else {
				// Parse as a list (e.g., "[jpg,pdf,png]")
				cleanFormat := strings.Trim(format, "[]")
				if cleanFormat != "" {
					extensions := strings.Split(cleanFormat, ",")
					for i, ext := range extensions {
						extensions[i] = strings.TrimSpace(strings.ToLower(ext))
					}
					args.Formats = extensions
				} else {
					fmt.Fprintf(os.Stderr, "Error: Invalid format '%s'\n", format)
					os.Exit(1)
				}
			}
		}
	}

	// Normalize path by trimming slashes
	if args.Path != "" {
		args.Path = strings.Trim(args.Path, "/")
	}

	// Set token from environment if not provided
	if args.Token == "" {
		args.Token = os.Getenv("GITHUB_TOKEN")
	}

	// Normalize output directory
	if args.Output != "" {
		args.Output = strings.TrimRight(args.Output, "/")
	}
	// fmt.Printf("DEBUG: Args = %+v\n", args)
	return args
}
