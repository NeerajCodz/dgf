package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/NeerajCodz/dgf/types"
	"github.com/spf13/pflag"
)

func ParseArgs() types.Args {
	var args types.Args

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: ./dgf <URL> [options] ...`)
	}

	// Custom usage
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage:
  ./dgf <URL> [options]

Options:
  --token, -t <token>         GitHub token
  --branch, -b <branch>       Branch name
  --commit, -c <commit>       Commit ID
  --path, -p <path>           Path in repository
  --output, -o <dir>          Output directory (default: .)
  --no-print, -n              Suppress all output
  --print-tree                Print directory tree
  --check                     Check if path exists
  --print-info                Print repository info as JSON
  --help, -h                  Show this help message
`)
	}

	// Define flags
	pflag.StringVarP(&args.Token, "token", "t", "", "GitHub token")
	pflag.StringVarP(&args.Branch, "branch", "b", "", "Branch name")
	pflag.StringVarP(&args.Commit, "commit", "c", "", "Commit ID")
	pflag.StringVarP(&args.Path, "path", "p", "", "Path in repository")
	pflag.StringVarP(&args.Output, "output", "o", ".", "Output directory for downloads (default: current directory)")
	pflag.BoolVarP(&args.NoPrint, "no-print", "n", false, "Suppress all output")
	pflag.BoolVar(&args.PrintTree, "print-tree", false, "Print directory tree")
	pflag.BoolVar(&args.Check, "check", false, "Check if path exists")
	pflag.BoolVar(&args.PrintInfo, "print-info", false, "Print info as JSON")

	// Help flag
	help := pflag.BoolP("help", "h", false, "Show this help message")

	pflag.Parse()

	// Show help and exit
	if *help {
		pflag.Usage()
		os.Exit(0)
	}

	// Validate positional arguments
	if pflag.NArg() != 1 {
		pflag.Usage()
		os.Exit(1)
	}

	// Parse positional argument
	args.URL = pflag.Arg(0)

	// Normalize path
	if args.Path != "" {
		args.Path = strings.Trim(args.Path, "/")
	}
	if args.Token == "" {
		args.Token = os.Getenv("GITHUB_TOKEN")
	}
	if args.Output != "" {
		args.Output = strings.TrimRight(args.Output, "/")
	}

	return args
}
