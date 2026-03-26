package selection

import (
	"sort"
	"testing"

	"github.com/NeerajCodz/dgf/pkg/types"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.selected == nil {
		t.Fatal("selected map is nil")
	}

	if manager.sizes == nil {
		t.Fatal("sizes map is nil")
	}

	if len(manager.selected) != 0 {
		t.Errorf("Expected empty selected map, got %d items", len(manager.selected))
	}

	if len(manager.sizes) != 0 {
		t.Errorf("Expected empty sizes map, got %d items", len(manager.sizes))
	}
}

func TestToggle_SelectNewItem(t *testing.T) {
	manager := NewManager()

	result := manager.Toggle("file.txt", 1024)

	if !result {
		t.Error("Expected Toggle to return true for new item")
	}

	if !manager.IsSelected("file.txt") {
		t.Error("Expected file.txt to be selected")
	}

	if manager.Count() != 1 {
		t.Errorf("Expected 1 selected item, got %d", manager.Count())
	}
}

func TestToggle_DeselectItem(t *testing.T) {
	manager := NewManager()

	manager.Toggle("file.txt", 1024)
	result := manager.Toggle("file.txt", 1024)

	if result {
		t.Error("Expected Toggle to return false when deselecting")
	}

	if manager.IsSelected("file.txt") {
		t.Error("Expected file.txt to be deselected")
	}

	if manager.Count() != 0 {
		t.Errorf("Expected 0 selected items, got %d", manager.Count())
	}
}

func TestSelect_MultipleItems(t *testing.T) {
	manager := NewManager()

	manager.Select("file1.txt", 1024)
	manager.Select("file2.txt", 2048)
	manager.Select("file3.txt", 4096)

	if manager.Count() != 3 {
		t.Errorf("Expected 3 selected items, got %d", manager.Count())
	}

	if !manager.IsSelected("file1.txt") || !manager.IsSelected("file2.txt") || !manager.IsSelected("file3.txt") {
		t.Error("Not all files are selected")
	}
}

func TestUnselect_RemoveItem(t *testing.T) {
	manager := NewManager()

	manager.Select("file1.txt", 1024)
	manager.Select("file2.txt", 2048)

	manager.Unselect("file1.txt")

	if manager.Count() != 1 {
		t.Errorf("Expected 1 selected item, got %d", manager.Count())
	}

	if manager.IsSelected("file1.txt") {
		t.Error("Expected file1.txt to be deselected")
	}

	if !manager.IsSelected("file2.txt") {
		t.Error("Expected file2.txt to remain selected")
	}
}

func TestIsSelected_Default(t *testing.T) {
	manager := NewManager()

	if manager.IsSelected("nonexistent") {
		t.Error("Expected IsSelected to return false for non-existent item")
	}
}

func TestSelectAll(t *testing.T) {
	manager := NewManager()

	items := []types.RepoItem{
		{Name: "file1.txt", Path: "file1.txt", Type: "file", Size: 1024},
		{Name: "file2.txt", Path: "file2.txt", Type: "file", Size: 2048},
		{Name: "dir1", Path: "dir1", Type: "dir", Size: 0},
	}

	manager.SelectAll(items)

	if manager.Count() != 3 {
		t.Errorf("Expected 3 selected items, got %d", manager.Count())
	}

	for _, item := range items {
		if !manager.IsSelected(item.Path) {
			t.Errorf("Expected %s to be selected", item.Path)
		}
	}
}

func TestSelectAll_Empty(t *testing.T) {
	manager := NewManager()

	items := []types.RepoItem{}
	manager.SelectAll(items)

	if manager.Count() != 0 {
		t.Errorf("Expected 0 selected items, got %d", manager.Count())
	}
}

func TestUnselectAll(t *testing.T) {
	manager := NewManager()

	manager.Select("file1.txt", 1024)
	manager.Select("file2.txt", 2048)
	manager.Select("file3.txt", 4096)

	manager.UnselectAll()

	if manager.Count() != 0 {
		t.Errorf("Expected 0 selected items, got %d", manager.Count())
	}

	if manager.IsSelected("file1.txt") || manager.IsSelected("file2.txt") || manager.IsSelected("file3.txt") {
		t.Error("Expected all items to be deselected")
	}
}

func TestGetSelected(t *testing.T) {
	manager := NewManager()

	manager.Select("file1.txt", 1024)
	manager.Select("file2.txt", 2048)
	manager.Select("file3.txt", 4096)

	selected := manager.GetSelected()

	if len(selected) != 3 {
		t.Errorf("Expected 3 selected items, got %d", len(selected))
	}

	// Check that all expected paths are in the result
	expectedPaths := map[string]bool{
		"file1.txt": true,
		"file2.txt": true,
		"file3.txt": true,
	}

	for _, path := range selected {
		if !expectedPaths[path] {
			t.Errorf("Unexpected path in selected: %s", path)
		}
	}
}

func TestCount(t *testing.T) {
	manager := NewManager()

	if manager.Count() != 0 {
		t.Errorf("Expected 0 items initially, got %d", manager.Count())
	}

	manager.Select("file1.txt", 1024)
	if manager.Count() != 1 {
		t.Errorf("Expected 1 item after select, got %d", manager.Count())
	}

	manager.Select("file2.txt", 2048)
	if manager.Count() != 2 {
		t.Errorf("Expected 2 items after second select, got %d", manager.Count())
	}

	manager.Unselect("file1.txt")
	if manager.Count() != 1 {
		t.Errorf("Expected 1 item after unselect, got %d", manager.Count())
	}
}

func TestTotalSize(t *testing.T) {
	manager := NewManager()

	if manager.TotalSize() != 0 {
		t.Errorf("Expected 0 bytes initially, got %d", manager.TotalSize())
	}

	manager.Select("file1.txt", 1024)
	if manager.TotalSize() != 1024 {
		t.Errorf("Expected 1024 bytes, got %d", manager.TotalSize())
	}

	manager.Select("file2.txt", 2048)
	if manager.TotalSize() != 3072 {
		t.Errorf("Expected 3072 bytes, got %d", manager.TotalSize())
	}

	manager.Unselect("file1.txt")
	if manager.TotalSize() != 2048 {
		t.Errorf("Expected 2048 bytes after unselect, got %d", manager.TotalSize())
	}
}

func TestTotalSize_MultipleItems(t *testing.T) {
	manager := NewManager()

	manager.Select("file1.txt", 100)
	manager.Select("file2.txt", 200)
	manager.Select("file3.txt", 300)
	manager.Select("file4.txt", 400)

	expected := int64(1000)
	if manager.TotalSize() != expected {
		t.Errorf("Expected %d bytes total, got %d", expected, manager.TotalSize())
	}
}

func TestHasSelection_Empty(t *testing.T) {
	manager := NewManager()

	if manager.HasSelection() {
		t.Error("Expected HasSelection to return false for empty manager")
	}
}

func TestHasSelection_WithItems(t *testing.T) {
	manager := NewManager()

	manager.Select("file1.txt", 1024)

	if !manager.HasSelection() {
		t.Error("Expected HasSelection to return true when items are selected")
	}

	manager.UnselectAll()

	if manager.HasSelection() {
		t.Error("Expected HasSelection to return false after UnselectAll")
	}
}

func TestSelectFolder(t *testing.T) {
	manager := NewManager()

	items := []types.RepoItem{
		{Name: "src", Path: "src", Type: "dir", Size: 0},
		{Name: "main.go", Path: "src/main.go", Type: "file", Size: 1024},
		{Name: "utils.go", Path: "src/utils.go", Type: "file", Size: 2048},
		{Name: "helper.go", Path: "src/helper.go", Type: "file", Size: 512},
		{Name: "README.md", Path: "README.md", Type: "file", Size: 256},
	}

	manager.SelectFolder("src", items)

	if manager.Count() != 4 {
		t.Errorf("Expected 4 selected items in folder, got %d", manager.Count())
	}

	if !manager.IsSelected("src") {
		t.Error("Expected src folder to be selected")
	}

	if !manager.IsSelected("src/main.go") {
		t.Error("Expected src/main.go to be selected")
	}

	if !manager.IsSelected("src/utils.go") {
		t.Error("Expected src/utils.go to be selected")
	}

	if !manager.IsSelected("src/helper.go") {
		t.Error("Expected src/helper.go to be selected")
	}

	if manager.IsSelected("README.md") {
		t.Error("Expected README.md to NOT be selected")
	}
}

func TestSelectFolder_NestedStructure(t *testing.T) {
	manager := NewManager()

	items := []types.RepoItem{
		{Name: "src", Path: "src", Type: "dir", Size: 0},
		{Name: "pkg", Path: "src/pkg", Type: "dir", Size: 0},
		{Name: "main.go", Path: "src/main.go", Type: "file", Size: 1024},
		{Name: "utils.go", Path: "src/pkg/utils.go", Type: "file", Size: 2048},
		{Name: "helper.go", Path: "src/pkg/helper.go", Type: "file", Size: 512},
		{Name: "README.md", Path: "README.md", Type: "file", Size: 256},
	}

	manager.SelectFolder("src/pkg", items)

	if manager.Count() != 3 {
		t.Errorf("Expected 3 selected items in src/pkg (src/pkg dir + 2 files), got %d", manager.Count())
	}

	if !manager.IsSelected("src/pkg") {
		t.Error("Expected src/pkg dir to be selected")
	}

	if !manager.IsSelected("src/pkg/utils.go") {
		t.Error("Expected src/pkg/utils.go to be selected")
	}

	if !manager.IsSelected("src/pkg/helper.go") {
		t.Error("Expected src/pkg/helper.go to be selected")
	}

	if manager.IsSelected("src/main.go") {
		t.Error("Expected src/main.go to NOT be selected")
	}
}

func TestUnselectFolder(t *testing.T) {
	manager := NewManager()

	items := []types.RepoItem{
		{Name: "src", Path: "src", Type: "dir", Size: 0},
		{Name: "main.go", Path: "src/main.go", Type: "file", Size: 1024},
		{Name: "utils.go", Path: "src/utils.go", Type: "file", Size: 2048},
		{Name: "README.md", Path: "README.md", Type: "file", Size: 256},
	}

	// First select everything
	manager.SelectAll(items)

	if manager.Count() != 4 {
		t.Errorf("Expected 4 items initially selected, got %d", manager.Count())
	}

	// Then unselect folder
	manager.UnselectFolder("src", items)

	if manager.Count() != 1 {
		t.Errorf("Expected 1 item remaining after unselectFolder, got %d", manager.Count())
	}

	if !manager.IsSelected("README.md") {
		t.Error("Expected README.md to remain selected")
	}

	if manager.IsSelected("src/main.go") {
		t.Error("Expected src/main.go to be deselected")
	}
}

func TestSyncWithItems(t *testing.T) {
	manager := NewManager()

	manager.Select("file1.txt", 1024)
	manager.Select("file3.txt", 4096)

	items := []types.RepoItem{
		{Name: "file1.txt", Path: "file1.txt", Type: "file", Size: 1024, Selected: false},
		{Name: "file2.txt", Path: "file2.txt", Type: "file", Size: 2048, Selected: false},
		{Name: "file3.txt", Path: "file3.txt", Type: "file", Size: 4096, Selected: false},
	}

	manager.SyncWithItems(items)

	if !items[0].Selected {
		t.Error("Expected file1.txt.Selected to be true")
	}

	if items[1].Selected {
		t.Error("Expected file2.txt.Selected to be false")
	}

	if !items[2].Selected {
		t.Error("Expected file3.txt.Selected to be true")
	}
}

func TestConcurrentOperations(t *testing.T) {
	manager := NewManager()
	done := make(chan bool, 5)

	// Concurrent selects
	go func() {
		manager.Select("file1.txt", 1024)
		done <- true
	}()

	go func() {
		manager.Select("file2.txt", 2048)
		done <- true
	}()

	go func() {
		manager.Toggle("file3.txt", 4096)
		done <- true
	}()

	go func() {
		_ = manager.GetSelected()
		done <- true
	}()

	go func() {
		_ = manager.Count()
		done <- true
	}()

	for i := 0; i < 5; i++ {
		<-done
	}

	if manager.Count() < 2 {
		t.Errorf("Expected at least 2 items selected, got %d", manager.Count())
	}
}

func TestToggleMultipleTimes(t *testing.T) {
	manager := NewManager()

	// Select
	result1 := manager.Toggle("file.txt", 1024)
	if !result1 {
		t.Error("First toggle should return true")
	}

	// Deselect
	result2 := manager.Toggle("file.txt", 1024)
	if result2 {
		t.Error("Second toggle should return false")
	}

	// Select again
	result3 := manager.Toggle("file.txt", 1024)
	if !result3 {
		t.Error("Third toggle should return true")
	}

	if manager.Count() != 1 {
		t.Errorf("Expected 1 selected item, got %d", manager.Count())
	}
}

func TestSelectAll_PreservesExistingSelections(t *testing.T) {
	manager := NewManager()

	manager.Select("file1.txt", 1024)

	items := []types.RepoItem{
		{Name: "file2.txt", Path: "file2.txt", Type: "file", Size: 2048},
		{Name: "file3.txt", Path: "file3.txt", Type: "file", Size: 4096},
	}

	manager.SelectAll(items)

	if manager.Count() != 3 {
		t.Errorf("Expected 3 selected items, got %d", manager.Count())
	}

	if !manager.IsSelected("file1.txt") {
		t.Error("Expected file1.txt to remain selected")
	}
}

func TestGetSelected_Order(t *testing.T) {
	manager := NewManager()

	manager.Select("file1.txt", 1024)
	manager.Select("file2.txt", 2048)
	manager.Select("file3.txt", 4096)

	selected := manager.GetSelected()

	// Sort both slices to compare
	sort.Strings(selected)
	expected := []string{"file1.txt", "file2.txt", "file3.txt"}
	sort.Strings(expected)

	if len(selected) != len(expected) {
		t.Errorf("Expected %d items, got %d", len(expected), len(selected))
	}

	for i, v := range selected {
		if v != expected[i] {
			t.Errorf("Expected %s at index %d, got %s", expected[i], i, v)
		}
	}
}
