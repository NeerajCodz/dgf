package cli

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/NeerajCodz/dgf/internal/github"
	"github.com/NeerajCodz/dgf/internal/utils"
	"github.com/NeerajCodz/dgf/pkg/types"
)

//go:embed git.json
var configData []byte

// Execute runs the CLI mode workflow
func Execute(args types.Args) error {
	// Parse platforms configuration
	var platforms []types.Platform
	if err := json.Unmarshal(configData, &platforms); err != nil {
		return fmt.Errorf("error parsing config: %v", err)
	}

	// Determine selected platform
	var selectedPlatform types.Platform
	if args.Site != "" {
		siteID := strings.ToLower(args.Site)
		for _, p := range platforms {
			if p.ID == siteID {
				selectedPlatform = p
				break
			}
		}
		if selectedPlatform.ID == "" {
			return fmt.Errorf("invalid site ID '%s'", args.Site)
		}
	} else if args.URL != "" {
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
			return fmt.Errorf("URL does not match any configured platform")
		}
	} else {
		return fmt.Errorf("must provide either a URL or --site, --username, and --repo")
	}

	// Process GitHub
	if selectedPlatform.ID == "github" {
		return executeGitHub(args, selectedPlatform)
	}

	return fmt.Errorf("platform not supported")
}

func executeGitHub(args types.Args, platform types.Platform) error {
	client := github.NewClient(args.Token)

	// Parse URL or construct from args
	var parsed types.ParsedURL
	var err error

	if args.URL != "" {
		parsed, err = github.ParseURL(args.URL, platform)
	} else {
		parsed, err = github.ParseFromArgs(args, platform)
	}
	if err != nil {
		return fmt.Errorf("failed to parse URL: %v", err)
	}

	// Override path if provided
	if args.Path != "" {
		parsed.Path = args.Path
		pathSegments := strings.Split(args.Path, "/")
		if len(pathSegments) > 1 {
			parsed.ParentPath = strings.Join(pathSegments[:len(pathSegments)-1], "/")
			parsed.RequestPath = pathSegments[len(pathSegments)-1]
		} else {
			parsed.ParentPath = ""
			parsed.RequestPath = args.Path
		}
	}

	// Determine reference
	var ref string
	if args.Commit != "" {
		ref = args.Commit
		parsed.Commit = args.Commit
		parsed.Branch = ""
	} else if parsed.Commit != "" {
		ref = parsed.Commit
	} else if args.Branch != "" {
		ref = args.Branch
		parsed.Branch = args.Branch
	} else if parsed.Branch != "" {
		ref = parsed.Branch
	} else {
		defaultBranch, err := client.FetchDefaultBranch(parsed.Username, parsed.Repo)
		if err != nil {
			return fmt.Errorf("failed to fetch default branch: %v", err)
		}
		ref = defaultBranch
		parsed.Branch = defaultBranch
	}

	// Update parsed URL
	if parsed.Path != "" {
		parsed.URL = fmt.Sprintf("https://github.com/%s/%s/tree/%s/%s", parsed.Username, parsed.Repo, ref, parsed.Path)
	} else {
		parsed.URL = fmt.Sprintf("https://github.com/%s/%s/tree/%s", parsed.Username, parsed.Repo, ref)
	}

	// Determine request type
	if parsed.Path != "" {
		requestType, err := github.GetRequestType(client, parsed.Username, parsed.Repo, ref, parsed.ParentPath, parsed.RequestPath)
		if err != nil {
			if args.Check {
				if !args.NoPrint {
					fmt.Println(`{"exists": false}`)
				}
				return nil
			}
			return fmt.Errorf("failed to determine request type: %v", err)
		}
		parsed.RequestType = requestType
	}

	// Fetch structure
	structure, err := github.FetchStructure(client, parsed.Username, parsed.Repo, ref, parsed.Path, parsed.RequestType, args.Formats)
	if err != nil {
		if args.Check {
			if !args.NoPrint {
				fmt.Println(`{"exists": false}`)
			}
			return nil
		}
		return err
	}

	// Handle output modes
	if args.Check {
		if !args.NoPrint {
			fmt.Println(`{"exists": true}`)
		}
		return nil
	}

	if args.PrintInfo {
		if !args.NoPrint {
			info := struct {
				Parsed    types.ParsedURL           `json:"parsed"`
				Structure types.RepositoryStructure `json:"structure"`
			}{parsed, structure}
			jsonData, _ := json.MarshalIndent(info, "", "  ")
			fmt.Println(string(jsonData))
		}
		return nil
	}

	if args.PrintTree {
		if !args.NoPrint {
			utils.TreePrint(structure)
		}
		return nil
	}

	// Download
	github.DownloadWithProgress(structure, args.Token, args.Output, args, parsed)
	return nil
}
