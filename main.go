package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/NeerajCodz/dgf/github"
	"github.com/NeerajCodz/dgf/types"
	"github.com/NeerajCodz/dgf/utils"
)

func main() {
	args := ParseArgs()

	// Read the git.json config file
	configData, err := os.ReadFile("config/git.json")
	if err != nil {
		if !args.NoPrint {
			fmt.Printf("Error reading config file: %v\n", err)
		}
		os.Exit(1)
	}

	var platforms []types.Platform
	err = json.Unmarshal(configData, &platforms)
	if err != nil {
		if !args.NoPrint {
			fmt.Printf("Error parsing config file: %v\n", err)
		}
		os.Exit(1)
	}

	// Check if the URL matches any platform's site
	for _, platform := range platforms {
		for _, site := range platform.URL.Site {
			if strings.HasPrefix(args.URL, site) {
				if platform.ID == "github" {
					// Process GitHub URL with arguments
					parsed, structure, err := github.ProcessGitHubURL(args.URL, args.Token, args.Branch, args.Commit, args.Path, platform)
					if args.Check {
						if !args.NoPrint {
							if err == github.ErrPathNotFound {
								fmt.Println(`{"exists": false}`)
							} else if err == nil {
								fmt.Println(`{"exists": true}`)
							} else {
								fmt.Printf("Error: %v\n", err)
							}
						}
					} else {
						if err != nil {
							if !args.NoPrint {
								fmt.Printf("Error: %v\n", err)
							}
							os.Exit(1)
						}
						if !args.NoPrint {
							if args.PrintInfo {
								info := struct {
									Parsed    types.ParsedURL           `json:"parsed"`
									Structure types.RepositoryStructure `json:"structure"`
								}{parsed, structure}
								jsonData, _ := json.MarshalIndent(info, "", "  ")
								fmt.Println(string(jsonData))
							}
							if args.PrintTree {
								utils.TreePrint(structure)
							}
							// Call Download if neither -t nor -i is set
							if !args.PrintTree && !args.PrintInfo {
								github.Download(structure, args.Token, args.Output, args, parsed)
							}
						}
					}
					return
				}
			}
		}
	}

	if !args.NoPrint {
		fmt.Println("Error: URL does not match any configured platform")
	}
	os.Exit(1)
}
