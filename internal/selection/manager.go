package selection

import (
	"strings"
	"sync"

	"github.com/NeerajCodz/dgf/pkg/types"
)

// Manager handles multi-selection state for repository items
type Manager struct {
	selected map[string]bool
	sizes    map[string]int64
	mu       sync.RWMutex
}

// NewManager creates a new selection manager
func NewManager() *Manager {
	return &Manager{
		selected: make(map[string]bool),
		sizes:    make(map[string]int64),
	}
}

// Toggle toggles the selection state of an item
func (m *Manager) Toggle(path string, size int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.selected[path] {
		delete(m.selected, path)
		delete(m.sizes, path)
		return false
	}
	m.selected[path] = true
	m.sizes[path] = size
	return true
}

// Select marks an item as selected
func (m *Manager) Select(path string, size int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.selected[path] = true
	m.sizes[path] = size
}

// Unselect marks an item as unselected
func (m *Manager) Unselect(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.selected, path)
	delete(m.sizes, path)
}

// IsSelected checks if an item is selected
func (m *Manager) IsSelected(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.selected[path]
}

// SelectAll selects all items in the list
func (m *Manager) SelectAll(items []types.RepoItem) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range items {
		m.selected[item.Path] = true
		m.sizes[item.Path] = item.Size
	}
}

// UnselectAll clears all selections
func (m *Manager) UnselectAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.selected = make(map[string]bool)
	m.sizes = make(map[string]int64)
}

// GetSelected returns all selected paths
func (m *Manager) GetSelected() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	paths := make([]string, 0, len(m.selected))
	for path := range m.selected {
		paths = append(paths, path)
	}
	return paths
}

// Count returns the number of selected items
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.selected)
}

// TotalSize returns the total size of selected items
func (m *Manager) TotalSize() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var total int64
	for _, size := range m.sizes {
		total += size
	}
	return total
}

// HasSelection returns true if any items are selected
func (m *Manager) HasSelection() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.selected) > 0
}

// SelectFolder recursively selects all items under a folder path
func (m *Manager) SelectFolder(folderPath string, allItems []types.RepoItem) {
	m.mu.Lock()
	defer m.mu.Unlock()

	prefix := folderPath + "/"
	for _, item := range allItems {
		if item.Path == folderPath || strings.HasPrefix(item.Path, prefix) {
			m.selected[item.Path] = true
			m.sizes[item.Path] = item.Size
		}
	}
}

// UnselectFolder recursively unselects all items under a folder path
func (m *Manager) UnselectFolder(folderPath string, allItems []types.RepoItem) {
	m.mu.Lock()
	defer m.mu.Unlock()

	prefix := folderPath + "/"
	for _, item := range allItems {
		if item.Path == folderPath || strings.HasPrefix(item.Path, prefix) {
			delete(m.selected, item.Path)
			delete(m.sizes, item.Path)
		}
	}
}

// SyncWithItems updates item.Selected based on manager state
func (m *Manager) SyncWithItems(items []types.RepoItem) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for i := range items {
		items[i].Selected = m.selected[items[i].Path]
	}
}
