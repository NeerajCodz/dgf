package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/NeerajCodz/dgf/pkg/types"
)

// Client handles GitHub API interactions
type Client struct {
	token      string
	httpClient *http.Client
}

// NewClient creates a new GitHub API client
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{},
	}
}

// SetToken updates the authentication token
func (c *Client) SetToken(token string) {
	c.token = token
}

// doWithAuthFallback executes a request and retries once without Authorization
// if GitHub returns 401 for an authenticated request.
// This helps when a stale/invalid token is configured but the target repository is public.
func (c *Client) doWithAuthFallback(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized && c.token != "" {
		resp.Body.Close()
		retryReq := req.Clone(req.Context())
		retryReq.Header = req.Header.Clone()
		retryReq.Header.Del("Authorization")
		return c.httpClient.Do(retryReq)
	}

	return resp, nil
}

// FetchContents fetches directory contents from GitHub API
func (c *Client) FetchContents(owner, repo, ref, path string) ([]types.GitHubContent, error) {
	owner = strings.ToLower(owner)
	repo = strings.ToLower(repo)
	
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", owner, repo)
	if path != "" {
		api += "/" + path
	}
	if ref != "" {
		api += "?ref=" + ref
	}

	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	c.setHeaders(req)

	resp, err := c.doWithAuthFallback(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contents: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, ErrPathNotFound
	} else if resp.StatusCode == 403 {
		return nil, ErrRateLimited
	} else if resp.StatusCode == 401 {
		return nil, ErrUnauthorized
	} else if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%d %s: %s", resp.StatusCode, http.StatusText(resp.StatusCode), string(body))
	}

	var contents []types.GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, fmt.Errorf("failed to decode contents: %v", err)
	}

	return contents, nil
}

// FetchFile fetches a single file's metadata from GitHub API
func (c *Client) FetchFile(owner, repo, ref, path string) (types.GitHubContent, error) {
	var content types.GitHubContent
	
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	if ref != "" {
		api += "?ref=" + ref
	}

	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return content, fmt.Errorf("failed to create request: %v", err)
	}

	c.setHeaders(req)

	resp, err := c.doWithAuthFallback(req)
	if err != nil {
		return content, fmt.Errorf("failed to fetch file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return content, ErrPathNotFound
	} else if resp.StatusCode == 403 {
		return content, ErrRateLimited
	} else if resp.StatusCode == 401 {
		return content, ErrUnauthorized
	} else if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return content, fmt.Errorf("%d %s: %s", resp.StatusCode, http.StatusText(resp.StatusCode), string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return content, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check if response is an array (directory) instead of object (file)
	if len(body) > 0 && body[0] == '[' {
		return content, ErrPathNotFound
	}

	if err := json.Unmarshal(body, &content); err != nil {
		return content, fmt.Errorf("failed to decode file details: %v", err)
	}

	return content, nil
}

// FetchDefaultBranch retrieves the default branch of a repository
func (c *Client) FetchDefaultBranch(owner, repo string) (string, error) {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	c.setHeaders(req)

	resp, err := c.doWithAuthFallback(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch repo info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("repository not found: %s/%s", owner, repo)
	} else if resp.StatusCode == 403 {
		return "", ErrRateLimited
	} else if resp.StatusCode == 401 {
		return "", fmt.Errorf("%w - invalid or expired token (use `dgf config set token \"\"` to clear saved token)", ErrUnauthorized)
	} else if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch repo info: status %d", resp.StatusCode)
	}

	var repoInfo struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return "", fmt.Errorf("failed to decode repo info: %v", err)
	}

	if repoInfo.DefaultBranch == "" {
		return "", fmt.Errorf("no default branch found for %s/%s", owner, repo)
	}

	return repoInfo.DefaultBranch, nil
}

// FetchRawFile fetches raw file content
func (c *Client) FetchRawFile(downloadURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if c.token != "" {
		req.Header.Add("Authorization", "token "+c.token)
	}

	resp, err := c.doWithAuthFallback(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("%w - invalid or expired token", ErrUnauthorized)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to download: status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// setHeaders sets common headers for GitHub API requests
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Add("Accept", "application/vnd.github+json")
	if c.token != "" {
		req.Header.Add("Authorization", "token "+c.token)
	}
}
