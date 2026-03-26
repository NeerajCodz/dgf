package github

import (
	"strings"
	"testing"

	"github.com/NeerajCodz/dgf/pkg/types"
)

func TestParseURL_HTTPS(t *testing.T) {
	platform := types.Platform{
		Name: "github",
		ID:   "github",
		URL: types.URL{
			Site: []string{"https://github.com/"},
		},
	}

	url := "https://github.com/testuser/testrepo"
	result, err := ParseURL(url, platform)

	if err != nil {
		t.Fatalf("ParseURL failed: %v", err)
	}

	if result.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", result.Username)
	}

	if result.Repo != "testrepo" {
		t.Errorf("Expected repo testrepo, got %s", result.Repo)
	}
}

func TestParseURL_HTTPPrefix(t *testing.T) {
	platform := types.Platform{
		Name: "github",
		ID:   "github",
		URL: types.URL{
			Site: []string{"http://github.com/"},
		},
	}

	url := "http://github.com/testuser/testrepo"
	result, err := ParseURL(url, platform)

	if err != nil {
		t.Fatalf("ParseURL failed: %v", err)
	}

	if result.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", result.Username)
	}

	if result.Repo != "testrepo" {
		t.Errorf("Expected repo testrepo, got %s", result.Repo)
	}
}

func TestParseURL_WithoutProtocol(t *testing.T) {
	platform := types.Platform{
		Name: "github",
		ID:   "github",
		URL: types.URL{
			Site: []string{"https://github.com/", "github.com/"},
		},
	}

	url := "github.com/testuser/testrepo"
	result, err := ParseURL(url, platform)

	if err != nil {
		t.Fatalf("ParseURL failed: %v", err)
	}

	if result.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", result.Username)
	}

	if result.Repo != "testrepo" {
		t.Errorf("Expected repo testrepo, got %s", result.Repo)
	}
}

func TestParseURL_WithBranch(t *testing.T) {
	platform := types.Platform{
		Name: "github",
		ID:   "github",
		URL: types.URL{
			Site: []string{"https://github.com/"},
		},
	}

	url := "https://github.com/testuser/testrepo/tree/main/src/main.go"
	result, err := ParseURL(url, platform)

	if err != nil {
		t.Fatalf("ParseURL failed: %v", err)
	}

	if result.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", result.Username)
	}

	if result.Repo != "testrepo" {
		t.Errorf("Expected repo testrepo, got %s", result.Repo)
	}

	if result.Branch != "main" {
		t.Errorf("Expected branch main, got %s", result.Branch)
	}

	if result.Path != "src/main.go" {
		t.Errorf("Expected path src/main.go, got %s", result.Path)
	}
}

func TestParseURL_WithCommit(t *testing.T) {
	platform := types.Platform{
		Name: "github",
		ID:   "github",
		URL: types.URL{
			Site: []string{"https://github.com/"},
		},
	}

	url := "https://github.com/testuser/testrepo/blob/abc1234567890/README.md"
	result, err := ParseURL(url, platform)

	if err != nil {
		t.Fatalf("ParseURL failed: %v", err)
	}

	if result.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", result.Username)
	}

	if result.Commit != "abc1234567890" {
		t.Errorf("Expected commit abc1234567890, got %s", result.Commit)
	}

	if result.Path != "README.md" {
		t.Errorf("Expected path README.md, got %s", result.Path)
	}
}

func TestParseURL_InvalidURL(t *testing.T) {
	platform := types.Platform{
		Name: "github",
		ID:   "github",
		URL: types.URL{
			Site: []string{"https://github.com/"},
		},
	}

	tests := []string{
		"https://gitlab.com/testuser/testrepo", // Wrong domain
		"https://github.com/testuser",           // Missing repo
		"https://github.com/",                   // Missing username and repo
		"invalid-url",                           // Invalid format
	}

	for _, url := range tests {
		_, err := ParseURL(url, platform)
		if err != ErrPathNotFound {
			t.Errorf("Expected ErrPathNotFound for %s, got %v", url, err)
		}
	}
}

func TestParseFromArgs(t *testing.T) {
	platform := types.Platform{
		Name: "github",
		ID:   "github",
		URLStruc: types.URLStruc{
			Site: "https://github.com/<username>/<repo>",
		},
	}

	args := types.Args{
		Username: "testuser",
		Repo:     "testrepo",
	}

	result, err := ParseFromArgs(args, platform)

	if err != nil {
		t.Fatalf("ParseFromArgs failed: %v", err)
	}

	if result.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", result.Username)
	}

	if result.Repo != "testrepo" {
		t.Errorf("Expected repo testrepo, got %s", result.Repo)
	}

	if !strings.Contains(result.URL, "testuser") {
		t.Errorf("Expected URL to contain username, got %s", result.URL)
	}

	if !strings.Contains(result.URL, "testrepo") {
		t.Errorf("Expected URL to contain repo, got %s", result.URL)
	}
}

func TestIsPotentialCommitHash(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"abc1234", true},           // Valid 7-char hash
		{"abc123456789", true},      // Valid 12-char hash
		{"abcdef0123456789", true},  // Valid 16-char hash
		{"0123456789abcdef", true},  // Valid with numbers
		{"ABCDEF0123456789", true},  // Valid uppercase
		{"abc", false},              // Too short
		{"abcdefg0123456789abcdefg0123456789abcdef1", false}, // Too long
		{"abc123g", false},          // Invalid character 'g'
		{"abc 123", false},          // Contains space
		{"abc-123", false},          // Contains dash
		{"", false},                 // Empty
		{"abc12345678901234567890123456789012345678901", false}, // > 40 chars
	}

	for _, tt := range tests {
		result := isPotentialCommitHash(tt.input)
		if result != tt.expected {
			t.Errorf("isPotentialCommitHash(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		username string
		repo     string
		ref      string
		path     string
		expected string
	}{
		{
			username: "testuser",
			repo:     "testrepo",
			ref:      "",
			path:     "",
			expected: "https://github.com/testuser/testrepo",
		},
		{
			username: "testuser",
			repo:     "testrepo",
			ref:      "main",
			path:     "",
			expected: "https://github.com/testuser/testrepo/tree/main",
		},
		{
			username: "testuser",
			repo:     "testrepo",
			ref:      "main",
			path:     "src/main.go",
			expected: "https://github.com/testuser/testrepo/tree/main/src/main.go",
		},
		{
			username: "testuser",
			repo:     "testrepo",
			ref:      "abc1234567",
			path:     "README.md",
			expected: "https://github.com/testuser/testrepo/tree/abc1234567/README.md",
		},
	}

	for _, tt := range tests {
		result := BuildURL(tt.username, tt.repo, tt.ref, tt.path)
		if result != tt.expected {
			t.Errorf("BuildURL(%q, %q, %q, %q) = %q, want %q",
				tt.username, tt.repo, tt.ref, tt.path, result, tt.expected)
		}
	}
}

func TestParseURL_EdgeCases(t *testing.T) {
	platform := types.Platform{
		Name: "github",
		ID:   "github",
		URL: types.URL{
			Site: []string{"https://github.com/", "http://github.com/"},
		},
	}

	// Test with trailing slashes
	url := "https://github.com/testuser/testrepo/"
	result, err := ParseURL(url, platform)

	if err != nil {
		t.Fatalf("ParseURL failed: %v", err)
	}

	if result.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", result.Username)
	}

	if result.Repo != "testrepo" {
		t.Errorf("Expected repo testrepo, got %s", result.Repo)
	}
}

func TestParseURL_WithNestedPath(t *testing.T) {
	platform := types.Platform{
		Name: "github",
		ID:   "github",
		URL: types.URL{
			Site: []string{"https://github.com/"},
		},
	}

	url := "https://github.com/testuser/testrepo/tree/develop/src/pkg/utils/helper.go"
	result, err := ParseURL(url, platform)

	if err != nil {
		t.Fatalf("ParseURL failed: %v", err)
	}

	if result.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", result.Username)
	}

	if result.Repo != "testrepo" {
		t.Errorf("Expected repo testrepo, got %s", result.Repo)
	}

	if result.Branch != "develop" {
		t.Errorf("Expected branch develop, got %s", result.Branch)
	}

	if result.Path != "src/pkg/utils/helper.go" {
		t.Errorf("Expected path src/pkg/utils/helper.go, got %s", result.Path)
	}
}
