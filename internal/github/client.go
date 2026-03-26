package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/NeerajCodz/dgf/pkg/types"
)

// CommitInfo stores minimal commit metadata for selector UI.
type CommitInfo struct {
	SHA     string
	Message string
	Author  string
	Date    time.Time
}

// BranchInfo stores branch details and commit counts.
type BranchInfo struct {
	Name        string
	CommitCount int
}

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

// FetchBranches fetches the list of branches for a repository
func (c *Client) FetchBranches(owner, repo string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches", owner, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.doWithAuthFallback(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch branches: status %d", resp.StatusCode)
	}

	var branchData []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&branchData); err != nil {
		return nil, err
	}

	branches := make([]string, len(branchData))
	for i, b := range branchData {
		branches[i] = b.Name
	}

	return branches, nil
}

// FetchBranchesWithCounts fetches branches and estimated commit counts for each branch.
func (c *Client) FetchBranchesWithCounts(owner, repo string) ([]BranchInfo, error) {
	names, err := c.FetchBranches(owner, repo)
	if err != nil {
		return nil, err
	}

	result := make([]BranchInfo, 0, len(names))
	for _, name := range names {
		count, countErr := c.FetchBranchCommitCount(owner, repo, name)
		if countErr != nil {
			count = -1
		}
		result = append(result, BranchInfo{
			Name:        name,
			CommitCount: count,
		})
	}
	return result, nil
}

// FetchBranchCommitCount returns approximate total commit count for a branch.
func (c *Client) FetchBranchCommitCount(owner, repo, branch string) (int, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?sha=%s&per_page=1", owner, repo, branch)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	c.setHeaders(req)

	resp, err := c.doWithAuthFallback(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to fetch branch commit count: status %d", resp.StatusCode)
	}

	link := resp.Header.Get("Link")
	if link != "" {
		if last := parseLastPageFromLink(link); last > 0 {
			return last, nil
		}
	}

	var rows []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return 0, err
	}
	return len(rows), nil
}

// FetchCommits fetches recent commits for the selected repository ref.
func (c *Client) FetchCommits(owner, repo, ref string, perPage int) ([]CommitInfo, error) {
	if perPage <= 0 {
		perPage = 50
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?per_page=%d", owner, repo, perPage)
	if ref != "" {
		url += "&sha=" + ref
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.doWithAuthFallback(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrRateLimited
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch commits: status %d", resp.StatusCode)
	}

	var payload []struct {
		SHA    string `json:"sha"`
		Commit struct {
			Message string `json:"message"`
			Author  struct {
				Name string    `json:"name"`
				Date time.Time `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	out := make([]CommitInfo, 0, len(payload))
	for _, row := range payload {
		msg := row.Commit.Message
		if idx := strings.Index(msg, "\n"); idx >= 0 {
			msg = msg[:idx]
		}
		out = append(out, CommitInfo{
			SHA:     row.SHA,
			Message: msg,
			Author:  row.Commit.Author.Name,
			Date:    row.Commit.Author.Date,
		})
	}
	return out, nil
}

func parseLastPageFromLink(linkHeader string) int {
	re := regexp.MustCompile(`[\?&]page=(\d+)>;\s*rel="last"`)
	m := re.FindStringSubmatch(linkHeader)
	if len(m) != 2 {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	return n
}
