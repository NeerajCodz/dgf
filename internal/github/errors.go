package github

import "fmt"

// Common errors for GitHub operations
var (
	ErrPathNotFound = fmt.Errorf("path not found")
	ErrRateLimited  = fmt.Errorf("rate limit exceeded")
	ErrUnauthorized = fmt.Errorf("unauthorized - check your token")
	ErrRepoNotFound = fmt.Errorf("repository not found")
)
