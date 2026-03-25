package github

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NeerajCodz/dgf/pkg/types"
)

func TestNewClient(t *testing.T) {
	token := "test-token"
	client := NewClient(token)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.token != token {
		t.Errorf("Expected token %q, got %q", token, client.token)
	}
	if client.httpClient == nil {
		t.Fatal("httpClient is nil")
	}
}

func TestSetToken(t *testing.T) {
	client := NewClient("old-token")
	newToken := "new-token"
	client.SetToken(newToken)

	if client.token != newToken {
		t.Errorf("Expected token %q, got %q", newToken, client.token)
	}
}

func TestFetchContents_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/repos/testuser/testrepo/contents") {
			w.Header().Set("Content-Type", "application/json")
			contents := []types.GitHubContent{
				{
					Name:     "README.md",
					Path:     "README.md",
					Type:     "file",
					Size:     100,
					Sha:      "abc123",
					URL:      "https://api.github.com/repos/testuser/testrepo/contents/README.md",
					HTMLURL:  "https://github.com/testuser/testrepo/blob/main/README.md",
					GitURL:   "https://api.github.com/repos/testuser/testrepo/git/blobs/abc123",
				},
				{
					Name:     "src",
					Path:     "src",
					Type:     "dir",
					Size:     0,
					Sha:      "def456",
					URL:      "https://api.github.com/repos/testuser/testrepo/contents/src",
					HTMLURL:  "https://github.com/testuser/testrepo/tree/main/src",
					GitURL:   "https://api.github.com/repos/testuser/testrepo/git/trees/def456",
				},
			}
			json.NewEncoder(w).Encode(contents)
		}
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.httpClient = server.Client()

	// Adjust client to use test server
	oldBaseURL := "https://api.github.com"
	newBaseURL := server.URL

	contents, err := client.FetchContents("testuser", "testrepo", "", "")
	if err != nil {
		t.Fatalf("FetchContents failed: %v", err)
	}

	if len(contents) != 2 {
		t.Errorf("Expected 2 contents, got %d", len(contents))
	}

	if contents[0].Name != "README.md" {
		t.Errorf("Expected first item name to be README.md, got %s", contents[0].Name)
	}

	if contents[1].Type != "dir" {
		t.Errorf("Expected second item type to be dir, got %s", contents[1].Type)
	}

	_ = oldBaseURL
	_ = newBaseURL
}

func TestFetchContents_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.httpClient = server.Client()

	contents, err := client.FetchContents("testuser", "testrepo", "", "nonexistent")
	if err != ErrPathNotFound {
		t.Errorf("Expected ErrPathNotFound, got %v", err)
	}

	if len(contents) != 0 {
		t.Errorf("Expected 0 contents, got %d", len(contents))
	}
}

func TestFetchContents_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.httpClient = server.Client()

	contents, err := client.FetchContents("testuser", "testrepo", "", "")
	if err != ErrRateLimited {
		t.Errorf("Expected ErrRateLimited, got %v", err)
	}

	if len(contents) != 0 {
		t.Errorf("Expected 0 contents, got %d", len(contents))
	}
}

func TestFetchContents_WithRef(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("ref") != "main" {
			t.Errorf("Expected ref=main, got %s", query.Get("ref"))
		}

		w.Header().Set("Content-Type", "application/json")
		contents := []types.GitHubContent{}
		json.NewEncoder(w).Encode(contents)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.httpClient = server.Client()

	_, err := client.FetchContents("testuser", "testrepo", "main", "")
	if err != nil {
		t.Fatalf("FetchContents failed: %v", err)
	}
}

func TestFetchContents_OwnerRepoNormalization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/repos/testuser/testrepo/contents"
		if !strings.Contains(r.URL.Path, expectedPath) {
			t.Errorf("Expected path to contain %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		contents := []types.GitHubContent{}
		json.NewEncoder(w).Encode(contents)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.httpClient = server.Client()

	// Test with uppercase
	_, err := client.FetchContents("TestUser", "TestRepo", "", "")
	if err != nil {
		t.Fatalf("FetchContents failed: %v", err)
	}
}

func TestFetchFile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		content := types.GitHubContent{
			Name:        "README.md",
			Path:        "README.md",
			Type:        "file",
			Size:        100,
			Sha:         "abc123",
			URL:         "https://api.github.com/repos/testuser/testrepo/contents/README.md",
			HTMLURL:     "https://github.com/testuser/testrepo/blob/main/README.md",
			GitURL:      "https://api.github.com/repos/testuser/testrepo/git/blobs/abc123",
			DownloadURL: stringPtr("https://raw.githubusercontent.com/testuser/testrepo/main/README.md"),
		}
		json.NewEncoder(w).Encode(content)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.httpClient = server.Client()

	content, err := client.FetchFile("testuser", "testrepo", "", "README.md")
	if err != nil {
		t.Fatalf("FetchFile failed: %v", err)
	}

	if content.Name != "README.md" {
		t.Errorf("Expected name README.md, got %s", content.Name)
	}

	if content.Type != "file" {
		t.Errorf("Expected type file, got %s", content.Type)
	}
}

func TestFetchFile_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.httpClient = server.Client()

	_, err := client.FetchFile("testuser", "testrepo", "", "nonexistent.md")
	if err != ErrPathNotFound {
		t.Errorf("Expected ErrPathNotFound, got %v", err)
	}
}

func TestFetchFile_DirectoryInsteadOfFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return array (directory response)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.httpClient = server.Client()

	_, err := client.FetchFile("testuser", "testrepo", "", "somedir")
	if err != ErrPathNotFound {
		t.Errorf("Expected ErrPathNotFound for directory, got %v", err)
	}
}

func TestFetchDefaultBranch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/repos/testuser/testrepo") && !strings.Contains(r.URL.Path, "contents") {
			w.Header().Set("Content-Type", "application/json")
			repoInfo := map[string]interface{}{
				"default_branch": "main",
			}
			json.NewEncoder(w).Encode(repoInfo)
		}
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.httpClient = server.Client()

	branch, err := client.FetchDefaultBranch("testuser", "testrepo")
	if err != nil {
		t.Fatalf("FetchDefaultBranch failed: %v", err)
	}

	if branch != "main" {
		t.Errorf("Expected branch main, got %s", branch)
	}
}

func TestFetchDefaultBranch_RepositoryNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.httpClient = server.Client()

	_, err := client.FetchDefaultBranch("testuser", "nonexistent")
	if err == nil {
		t.Errorf("Expected error for non-existent repo, got nil")
	}
	if !strings.Contains(err.Error(), "repository not found") {
		t.Errorf("Expected 'repository not found' error, got %v", err)
	}
}

func TestFetchDefaultBranch_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.httpClient = server.Client()

	_, err := client.FetchDefaultBranch("testuser", "testrepo")
	if err != ErrRateLimited {
		t.Errorf("Expected ErrRateLimited, got %v", err)
	}
}

func TestFetchRawFile_Success(t *testing.T) {
	expectedContent := []byte("# Hello World\nThis is a test file")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(expectedContent)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.httpClient = server.Client()

	content, err := client.FetchRawFile(server.URL + "/raw/file.md")
	if err != nil {
		t.Fatalf("FetchRawFile failed: %v", err)
	}

	if !strings.EqualFold(string(content), string(expectedContent)) {
		t.Errorf("Expected content %s, got %s", string(expectedContent), string(content))
	}
}

func TestFetchRawFile_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.httpClient = server.Client()

	_, err := client.FetchRawFile(server.URL + "/nonexistent")
	if err == nil {
		t.Errorf("Expected error for non-existent file, got nil")
	}
}

func TestFetchRawFile_WithToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "token test-token" {
			t.Errorf("Expected Authorization header with token, got %s", auth)
		}
		w.Write([]byte("content"))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.httpClient = server.Client()

	_, err := client.FetchRawFile(server.URL + "/file")
	if err != nil {
		t.Fatalf("FetchRawFile failed: %v", err)
	}
}

func TestSetHeaders(t *testing.T) {
	client := NewClient("test-token")

	req, _ := http.NewRequest("GET", "https://api.github.com/repos/test/test", nil)
	client.setHeaders(req)

	accept := req.Header.Get("Accept")
	if accept != "application/vnd.github+json" {
		t.Errorf("Expected Accept header application/vnd.github+json, got %s", accept)
	}

	auth := req.Header.Get("Authorization")
	if auth != "token test-token" {
		t.Errorf("Expected Authorization header, got %s", auth)
	}
}

func TestSetHeaders_WithoutToken(t *testing.T) {
	client := NewClient("")

	req, _ := http.NewRequest("GET", "https://api.github.com/repos/test/test", nil)
	client.setHeaders(req)

	accept := req.Header.Get("Accept")
	if accept != "application/vnd.github+json" {
		t.Errorf("Expected Accept header application/vnd.github+json, got %s", accept)
	}

	auth := req.Header.Get("Authorization")
	if auth != "" {
		t.Errorf("Expected no Authorization header, got %s", auth)
	}
}

// Helper function to return a string pointer
func stringPtr(s string) *string {
	return &s
}
