package utils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/NeerajCodz/dgf/types"
)

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// TreePrint prints the repository structure as a tree
func TreePrint(structure types.RepositoryStructure) {
	fmt.Println("Repository structure:")
	if len(structure.Files) == 0 && len(structure.Folders) == 0 {
		fmt.Println("  (empty)")
		return
	}

	// Combine files and folders
	structurePaths := append(structure.Files, structure.Folders...)
	sort.Strings(structurePaths)

	// Build tree structure
	type node struct {
		name     string
		children map[string]*node
		isFile   bool
	}
	root := &node{children: make(map[string]*node)}

	for _, path := range structurePaths {
		parts := strings.Split(path, "/")
		current := root
		for i, part := range parts {
			if _, exists := current.children[part]; !exists {
				isFile := i == len(parts)-1 && contains(structure.Files, path)
				current.children[part] = &node{
					name:     part,
					children: make(map[string]*node),
					isFile:   isFile,
				}
			}
			current = current.children[part]
		}
	}

	// Print tree recursively
	var printNode func(*node, string, string)
	printNode = func(n *node, prefix, indent string) {
		keys := make([]string, 0, len(n.children))
		for key := range n.children {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for i, key := range keys {
			child := n.children[key]
			isLast := i == len(keys)-1
			connector := "├── "
			nextIndent := indent + "│   "
			if isLast {
				connector = "└── "
				nextIndent = indent + "    "
			}
			fmt.Printf("%s%s%s\n", prefix, connector, child.name)
			if !child.isFile {
				printNode(child, indent, nextIndent)
			}
		}
	}

	printNode(root, "", "")
}
