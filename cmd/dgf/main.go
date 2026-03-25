package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/NeerajCodz/dgf/internal/agent"
	"github.com/NeerajCodz/dgf/internal/cli"
	"github.com/NeerajCodz/dgf/internal/config"
	"github.com/NeerajCodz/dgf/internal/tui"
	"github.com/spf13/pflag"
)

const version = "2.0.0"

func main() {
	// Check for special commands first (before pflag parsing)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "config":
			handleConfigCommand(os.Args[2:])
			return
		case "agent":
			handleAgentCommand(os.Args[2:])
			return
		case "version", "--version", "-v":
			fmt.Printf("dgf version %s\n", version)
			return
		}
	}

	// Parse CLI arguments
	args := cli.ParseArgs()

	// Determine mode: if any CLI args provided, use CLI mode
	// Otherwise, launch interactive TUI
	if cli.HasCLIArgs(args) {
		// CLI mode
		if err := cli.Execute(args); err != nil {
			if !args.NoPrint {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			os.Exit(1)
		}
	} else {
		// TUI mode
		if err := tui.Run("", args.Token); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

// handleConfigCommand handles the config subcommand
func handleConfigCommand(args []string) {
	if len(args) == 0 {
		// Show all config
		output, err := config.Show()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(output)
		return
	}

	switch args[0] {
	case "show":
		output, err := config.Show()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(output)

	case "get":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: dgf config get <key>\n")
			fmt.Fprintf(os.Stderr, "Keys: token, download_path, ascii_mode, workers\n")
			os.Exit(1)
		}
		value, err := config.Get(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(value)

	case "set":
		if len(args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: dgf config set <key> <value>\n")
			fmt.Fprintf(os.Stderr, "Keys: token, download_path, ascii_mode, workers\n")
			os.Exit(1)
		}
		if err := config.Set(args[1], strings.Join(args[2:], " ")); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Set %s successfully\n", args[1])

	case "path":
		path, err := config.GetConfigPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(path)

	default:
		fmt.Fprintf(os.Stderr, "Unknown config command: %s\n", args[0])
		fmt.Fprintf(os.Stderr, "Usage: dgf config <show|get|set|path> [args...]\n")
		os.Exit(1)
	}
}

// handleAgentCommand handles the agent subcommand for JSON API mode
func handleAgentCommand(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: dgf agent <tree|download> <url> [options]\n")
		os.Exit(1)
	}

	// Get token from env or config
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		cfg, _ := config.Load()
		token = cfg.GithubToken
	}

	switch args[0] {
	case "tree":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: dgf agent tree <url>\n")
			os.Exit(1)
		}
		agent.Tree(args[1], token)

	case "download":
		if len(args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: dgf agent download <url> <paths...> [--out <dir>]\n")
			os.Exit(1)
		}
		url := args[1]
		paths := make([]string, 0)
		outputDir := "."
		
		for i := 2; i < len(args); i++ {
			if args[i] == "--out" && i+1 < len(args) {
				outputDir = args[i+1]
				i++
			} else {
				paths = append(paths, args[i])
			}
		}
		agent.Download(url, token, outputDir, paths)

	default:
		fmt.Fprintf(os.Stderr, "Unknown agent command: %s\n", args[0])
		fmt.Fprintf(os.Stderr, "Usage: dgf agent <tree|download> <url> [options]\n")
		os.Exit(1)
	}
}

func init() {
	// Disable pflag's default help handling for special commands
	pflag.ErrHelp = nil
}
