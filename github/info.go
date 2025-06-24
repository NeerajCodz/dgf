package github

import (
	"fmt"

	"github.com/NeerajCodz/dgf/types"
)

// PrintGitHubInfo prints all fields of a ParsedURL struct for debugging
func PrintGitHubInfo(info types.ParsedURL) {
	fmt.Printf("url: %s\n", info.URL)
	fmt.Printf("name: %s\n", info.Name)
	fmt.Printf("id: %s\n", info.ID)
	fmt.Printf("username: %s\n", info.Username)
	fmt.Printf("repo: %s\n", info.Repo)
	fmt.Printf("branch: %s\n", info.Branch)
	fmt.Printf("path: %s\n", info.Path)
	fmt.Printf("parent_path: %s\n", info.ParentPath)
	fmt.Printf("request_path: %s\n", info.RequestPath)
	fmt.Printf("request_type: %s\n", info.RequestType)
}