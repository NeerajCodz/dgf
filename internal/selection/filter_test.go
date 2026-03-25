package selection

import (
	"testing"

	"github.com/NeerajCodz/dgf/pkg/types"
)

func TestFilter_EmptyQuery(t *testing.T) {
	items := []types.RepoItem{
		{Name: "file1.txt", Path: "file1.txt", Type: "file", Size: 1024},
		{Name: "file2.txt", Path: "file2.txt", Type: "file", Size: 2048},
		{Name: "src", Path: "src", Type: "dir", Size: 0},
	}

	result := Filter(items, "")

	if len(result) != 3 {
		t.Errorf("Expected 3 items with empty query, got %d", len(result))
	}
}

func TestFilter_EmptyItems(t *testing.T) {
	items := []types.RepoItem{}

	result := Filter(items, "test")

	if len(result) != 0 {
		t.Errorf("Expected 0 items, got %d", len(result))
	}
}

func TestFilter_ByName(t *testing.T) {
	items := []types.RepoItem{
		{Name: "README.md", Path: "README.md", Type: "file", Size: 1024},
		{Name: "main.go", Path: "main.go", Type: "file", Size: 2048},
		{Name: "test_main.go", Path: "test_main.go", Type: "file", Size: 512},
		{Name: "utils.go", Path: "utils.go", Type: "file", Size: 1024},
	}

	result := Filter(items, "main")

	if len(result) != 2 {
		t.Errorf("Expected 2 items matching 'main', got %d", len(result))
	}

	// Verify the results contain main.go and test_main.go
	found := map[string]bool{
		"main.go":      false,
		"test_main.go": false,
	}

	for _, item := range result {
		if _, exists := found[item.Name]; exists {
			found[item.Name] = true
		}
	}

	for name, wasFound := range found {
		if !wasFound {
			t.Errorf("Expected %s to be in results", name)
		}
	}
}

func TestFilter_ByPath(t *testing.T) {
	items := []types.RepoItem{
		{Name: "main.go", Path: "src/main.go", Type: "file", Size: 2048},
		{Name: "main_test.go", Path: "src/main_test.go", Type: "file", Size: 512},
		{Name: "utils.go", Path: "pkg/utils.go", Type: "file", Size: 1024},
		{Name: "main.go", Path: "cmd/main.go", Type: "file", Size: 256},
	}

	result := Filter(items, "src")

	if len(result) != 2 {
		t.Errorf("Expected 2 items in src/ path, got %d", len(result))
	}

	for _, item := range result {
		if item.Path != "src/main.go" && item.Path != "src/main_test.go" {
			t.Errorf("Unexpected item in results: %s", item.Path)
		}
	}
}

func TestFilter_CaseInsensitive(t *testing.T) {
	items := []types.RepoItem{
		{Name: "README.MD", Path: "README.MD", Type: "file", Size: 1024},
		{Name: "src", Path: "src", Type: "dir", Size: 0},
		{Name: "Main.Go", Path: "Main.Go", Type: "file", Size: 2048},
	}

	result := Filter(items, "README")

	if len(result) != 1 {
		t.Errorf("Expected 1 item matching 'README' (case-insensitive), got %d", len(result))
	}

	if result[0].Name != "README.MD" {
		t.Errorf("Expected README.MD, got %s", result[0].Name)
	}

	result2 := Filter(items, "main")

	if len(result2) != 1 {
		t.Errorf("Expected 1 item matching 'main' (case-insensitive), got %d", len(result2))
	}

	if result2[0].Name != "Main.Go" {
		t.Errorf("Expected Main.Go, got %s", result2[0].Name)
	}
}

func TestFilter_PartialMatch(t *testing.T) {
	items := []types.RepoItem{
		{Name: "configuration.json", Path: "config/configuration.json", Type: "file", Size: 1024},
		{Name: "config.yaml", Path: "config/config.yaml", Type: "file", Size: 512},
		{Name: "app.js", Path: "app.js", Type: "file", Size: 2048},
	}

	result := Filter(items, "config")

	if len(result) != 2 {
		t.Errorf("Expected 2 items matching 'config', got %d", len(result))
	}
}

func TestFilter_NoMatches(t *testing.T) {
	items := []types.RepoItem{
		{Name: "file1.txt", Path: "file1.txt", Type: "file", Size: 1024},
		{Name: "file2.txt", Path: "file2.txt", Type: "file", Size: 2048},
	}

	result := Filter(items, "xyz")

	if len(result) != 0 {
		t.Errorf("Expected 0 items matching 'xyz', got %d", len(result))
	}
}

func TestHighlightMatches_Success(t *testing.T) {
	text := "hello world"
	query := "world"

	indices := HighlightMatches(text, query)

	if len(indices) != len(query) {
		t.Errorf("Expected %d indices, got %d", len(query), len(indices))
	}

	// "world" starts at index 6 in "hello world"
	expectedStart := 6
	for i, idx := range indices {
		if idx != expectedStart+i {
			t.Errorf("Expected index %d at position %d, got %d", expectedStart+i, i, idx)
		}
	}
}

func TestHighlightMatches_EmptyQuery(t *testing.T) {
	indices := HighlightMatches("hello", "")

	if len(indices) != 0 {
		t.Errorf("Expected 0 indices for empty query, got %d", len(indices))
	}
}

func TestHighlightMatches_NoMatch(t *testing.T) {
	indices := HighlightMatches("hello", "xyz")

	if len(indices) != 0 {
		t.Errorf("Expected 0 indices for no match, got %d", len(indices))
	}
}

func TestHighlightMatches_CaseInsensitive(t *testing.T) {
	text := "Hello World"
	query := "world"

	indices := HighlightMatches(text, query)

	if len(indices) != len(query) {
		t.Errorf("Expected %d indices, got %d", len(query), len(indices))
	}

	// "world" (case-insensitive) starts at index 6
	expectedStart := 6
	for i, idx := range indices {
		if idx != expectedStart+i {
			t.Errorf("Expected index %d at position %d, got %d", expectedStart+i, i, idx)
		}
	}
}

func TestHighlightMatches_BeginningOfString(t *testing.T) {
	text := "testing"
	query := "test"

	indices := HighlightMatches(text, query)

	if len(indices) != len(query) {
		t.Errorf("Expected %d indices, got %d", len(query), len(indices))
	}

	for i, idx := range indices {
		if idx != i {
			t.Errorf("Expected index %d at position %d, got %d", i, i, idx)
		}
	}
}

func TestFuzzyMatch_ExactMatch(t *testing.T) {
	text := "hello world"
	query := "world"

	match, score := FuzzyMatch(text, query)

	if !match {
		t.Error("Expected FuzzyMatch to return true for substring match")
	}

	if score <= 0 {
		t.Errorf("Expected positive score for exact substring match, got %d", score)
	}
}

func TestFuzzyMatch_EmptyQuery(t *testing.T) {
	match, score := FuzzyMatch("hello", "")

	if !match {
		t.Error("Expected FuzzyMatch to return true for empty query")
	}

	if score != 0 {
		t.Errorf("Expected 0 score for empty query, got %d", score)
	}
}

func TestFuzzyMatch_CharacterByCharacter(t *testing.T) {
	text := "main.go"
	query := "mn"

	match, score := FuzzyMatch(text, query)

	if !match {
		t.Error("Expected FuzzyMatch to return true for fuzzy match")
	}

	if score <= 0 {
		t.Errorf("Expected positive score for fuzzy match, got %d", score)
	}
}

func TestFuzzyMatch_NoMatch(t *testing.T) {
	text := "hello"
	query := "xyz"

	match, score := FuzzyMatch(text, query)

	if match {
		t.Error("Expected FuzzyMatch to return false for non-matching query")
	}

	if score != 0 {
		t.Errorf("Expected 0 score for non-matching query, got %d", score)
	}
}

func TestFuzzyMatch_FullCharacterMatch(t *testing.T) {
	text := "hello world"
	query := "hlo"

	match, score := FuzzyMatch(text, query)

	if !match {
		t.Error("Expected FuzzyMatch to return true for character fuzzy match")
	}

	if score <= 0 {
		t.Errorf("Expected positive score, got %d", score)
	}
}

func TestFilterByExtension_Success(t *testing.T) {
	items := []types.RepoItem{
		{Name: "main.go", Path: "main.go", Type: "file", Size: 1024},
		{Name: "utils.go", Path: "utils.go", Type: "file", Size: 2048},
		{Name: "README.md", Path: "README.md", Type: "file", Size: 512},
		{Name: "config.yaml", Path: "config.yaml", Type: "file", Size: 256},
		{Name: "src", Path: "src", Type: "dir", Size: 0},
	}

	result := FilterByExtension(items, []string{"go"})

	if len(result) != 3 {
		t.Errorf("Expected 3 items with .go extension, got %d", len(result))
	}

	// Should include the directory
	foundDir := false
	for _, item := range result {
		if item.Type == "dir" {
			foundDir = true
		}
		if item.Type == "file" && item.Name != "main.go" && item.Name != "utils.go" {
			t.Errorf("Unexpected file in results: %s", item.Name)
		}
	}

	if !foundDir {
		t.Error("Expected directory to be included in results")
	}
}

func TestFilterByExtension_MultipleExtensions(t *testing.T) {
	items := []types.RepoItem{
		{Name: "main.go", Path: "main.go", Type: "file", Size: 1024},
		{Name: "README.md", Path: "README.md", Type: "file", Size: 512},
		{Name: "config.yaml", Path: "config.yaml", Type: "file", Size: 256},
		{Name: "package.json", Path: "package.json", Type: "file", Size: 128},
	}

	result := FilterByExtension(items, []string{"go", "md", "yaml"})

	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result))
	}

	found := map[string]bool{}
	for _, item := range result {
		found[item.Name] = true
	}

	if !found["main.go"] || !found["README.md"] || !found["config.yaml"] {
		t.Error("Expected files not found in results")
	}

	if found["package.json"] {
		t.Error("package.json should not be in results")
	}
}

func TestFilterByExtension_EmptyExtensions(t *testing.T) {
	items := []types.RepoItem{
		{Name: "file.txt", Path: "file.txt", Type: "file", Size: 1024},
		{Name: "src", Path: "src", Type: "dir", Size: 0},
	}

	result := FilterByExtension(items, []string{})

	if len(result) != 2 {
		t.Errorf("Expected all items with empty extension filter, got %d", len(result))
	}
}

func TestFilterByExtension_WithDot(t *testing.T) {
	items := []types.RepoItem{
		{Name: "main.go", Path: "main.go", Type: "file", Size: 1024},
		{Name: "config.yaml", Path: "config.yaml", Type: "file", Size: 256},
	}

	// Test with dot prefix
	result := FilterByExtension(items, []string{".go"})

	if len(result) != 1 {
		t.Errorf("Expected 1 item with .go extension, got %d", len(result))
	}

	if result[0].Name != "main.go" {
		t.Errorf("Expected main.go, got %s", result[0].Name)
	}
}

func TestFilterByExtension_NoExtension(t *testing.T) {
	items := []types.RepoItem{
		{Name: "Makefile", Path: "Makefile", Type: "file", Size: 1024},
		{Name: "README", Path: "README", Type: "file", Size: 512},
		{Name: "main.go", Path: "main.go", Type: "file", Size: 2048},
	}

	result := FilterByExtension(items, []string{"go"})

	if len(result) != 1 {
		t.Errorf("Expected 1 item with .go extension, got %d", len(result))
	}

	if result[0].Name != "main.go" {
		t.Errorf("Expected main.go, got %s", result[0].Name)
	}
}

func TestFilterDirectoriesOnly(t *testing.T) {
	items := []types.RepoItem{
		{Name: "main.go", Path: "main.go", Type: "file", Size: 1024},
		{Name: "src", Path: "src", Type: "dir", Size: 0},
		{Name: "README.md", Path: "README.md", Type: "file", Size: 512},
		{Name: "pkg", Path: "pkg", Type: "dir", Size: 0},
	}

	result := FilterDirectoriesOnly(items)

	if len(result) != 2 {
		t.Errorf("Expected 2 directories, got %d", len(result))
	}

	for _, item := range result {
		if item.Type != "dir" {
			t.Errorf("Expected type dir, got %s", item.Type)
		}
	}
}

func TestFilterDirectoriesOnly_NoDirectories(t *testing.T) {
	items := []types.RepoItem{
		{Name: "main.go", Path: "main.go", Type: "file", Size: 1024},
		{Name: "README.md", Path: "README.md", Type: "file", Size: 512},
	}

	result := FilterDirectoriesOnly(items)

	if len(result) != 0 {
		t.Errorf("Expected 0 directories, got %d", len(result))
	}
}

func TestFilterFilesOnly(t *testing.T) {
	items := []types.RepoItem{
		{Name: "main.go", Path: "main.go", Type: "file", Size: 1024},
		{Name: "src", Path: "src", Type: "dir", Size: 0},
		{Name: "README.md", Path: "README.md", Type: "file", Size: 512},
		{Name: "pkg", Path: "pkg", Type: "dir", Size: 0},
	}

	result := FilterFilesOnly(items)

	if len(result) != 2 {
		t.Errorf("Expected 2 files, got %d", len(result))
	}

	for _, item := range result {
		if item.Type != "file" {
			t.Errorf("Expected type file, got %s", item.Type)
		}
	}
}

func TestFilterFilesOnly_NoFiles(t *testing.T) {
	items := []types.RepoItem{
		{Name: "src", Path: "src", Type: "dir", Size: 0},
		{Name: "pkg", Path: "pkg", Type: "dir", Size: 0},
	}

	result := FilterFilesOnly(items)

	if len(result) != 0 {
		t.Errorf("Expected 0 files, got %d", len(result))
	}
}

func TestMatchesQuery_SubstringMatch(t *testing.T) {
	if !matchesQuery("hello world", "world") {
		t.Error("Expected to match 'world' in 'hello world'")
	}

	if !matchesQuery("Hello World", "hello") {
		t.Error("Expected case-insensitive match")
	}

	if matchesQuery("hello", "xyz") {
		t.Error("Expected no match for 'xyz' in 'hello'")
	}
}

func TestFilter_MixedTypes(t *testing.T) {
	items := []types.RepoItem{
		{Name: "src", Path: "src", Type: "dir", Size: 0},
		{Name: "main.go", Path: "src/main.go", Type: "file", Size: 1024},
		{Name: "utils.go", Path: "src/utils.go", Type: "file", Size: 2048},
		{Name: "config.yaml", Path: "config.yaml", Type: "file", Size: 256},
		{Name: "tests", Path: "tests", Type: "dir", Size: 0},
	}

	result := Filter(items, "src")

	if len(result) != 3 {
		t.Errorf("Expected 3 items matching 'src', got %d", len(result))
	}

	dirCount := 0
	fileCount := 0

	for _, item := range result {
		if item.Type == "dir" {
			dirCount++
		} else if item.Type == "file" {
			fileCount++
		}
	}

	if dirCount != 1 {
		t.Errorf("Expected 1 directory, got %d", dirCount)
	}

	if fileCount != 2 {
		t.Errorf("Expected 2 files, got %d", fileCount)
	}
}
