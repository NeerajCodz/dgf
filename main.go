package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/NeerajCodz/dgf/github"
	"github.com/NeerajCodz/dgf/types"
	"github.com/NeerajCodz/dgf/utils"
)

//go:embed config/git.json
var configData []byte

// main is the entry point of the dgf CLI tool
func main() {
	// Parse command-line arguments
	args := ParseArgs()

	// Parse the embedded platforms configuration
	var platforms []types.Platform
	if err := json.Unmarshal(configData, &platforms); err != nil {
		if !args.NoPrint {
			fmt.Fprintf(os.Stderr, "Error parsing embedded config file: %v\n", err)
		}
		os.Exit(1)
	}

	// Determine the selected platform based on args.Site or args.URL
	var selectedPlatform types.Platform
	if args.Site != "" {
		// Use --site to select platform (case-insensitive)
		siteID := strings.ToLower(args.Site)
		for _, p := range platforms {
			if p.ID == siteID {
				selectedPlatform = p
				break
			}
		}
		if selectedPlatform.ID == "" {
			if !args.NoPrint {
				fmt.Fprintf(os.Stderr, "Error: Invalid site ID '%s'\n", args.Site)
			}
			os.Exit(1)
		}
	} else if args.URL != "" {
		// Use URL to select platform
		for _, p := range platforms {
			for _, site := range p.URL.Site {
				if strings.HasPrefix(args.URL, site) {
					selectedPlatform = p
					break
				}
			}
			if selectedPlatform.ID != "" {
				break
			}
		}
		if selectedPlatform.ID == "" {
			if !args.NoPrint {
				fmt.Fprintf(os.Stderr, "Error: URL does not match any configured platform\n")
			}
			os.Exit(1)
		}
	} else {
		if !args.NoPrint {
			fmt.Fprintf(os.Stderr, "Error: Must provide either a URL or --site, --username, and --repo\n")
		}
		os.Exit(1)
	}

	// Process GitHub-specific logic
	if selectedPlatform.ID == "github" {
		// Use args.URL if provided; otherwise, pass empty string to construct URL from site args
		urlToUse := args.URL
		if args.Site != "" {
			urlToUse = ""
		}
		parsed, structure, err := github.ProcessGitHubURL(urlToUse, args.Token, args.Branch, args.Commit, args.Path, selectedPlatform, args)
		if args.Check {
			// Handle --check flag
			if !args.NoPrint {
				if err == github.ErrPathNotFound {
					fmt.Println(`{"exists": false}`)
				} else if err == nil {
					fmt.Println(`{"exists": true}`)
				} else {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
			}
		} else {
			// Handle normal operation
			if err != nil {
				if !args.NoPrint {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
				os.Exit(1)
			}
			if !args.NoPrint {
				if args.PrintInfo {
					// Print parsed info and structure as JSON
					info := struct {
						Parsed    types.ParsedURL           `json:"parsed"`
						Structure types.RepositoryStructure `json:"structure"`
					}{parsed, structure}
					jsonData, _ := json.MarshalIndent(info, "", "  ")
					fmt.Println(string(jsonData))
				}
				if args.PrintTree {
					// Print directory tree
					utils.TreePrint(structure)
				}
			}
			// Download files if no print flags are set
			if !args.PrintTree && !args.PrintInfo && !args.Check {
				github.Download(structure, args.Token, args.Output, args, parsed)
			}
		}
		return
	}

	// Handle unsupported platforms
	if !args.NoPrint {
		fmt.Fprintf(os.Stderr, "Error: Platform not supported\n")
	}
	os.Exit(1)
}