package github

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/NeerajCodz/dgf/types"
	"github.com/NeerajCodz/dgf/utils"
)

// Download downloads files and creates directories in the specified output directory
func Download(structure types.RepositoryStructure, token, outputDir string, args types.Args, parsed types.ParsedURL) {
	// Validate output directory
	if outputDir == "" {
		outputDir = "."
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		if !args.NoPrint {
			fmt.Fprintf(os.Stderr, "Error ensuring output directory %s exists: %v\n", outputDir, err)
		}
		return
	}
	if info, err := os.Stat(outputDir); err != nil || !info.IsDir() {
		if !args.NoPrint {
			fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", outputDir)
		}
		return
	}

	// Calculate total counts
	totalFiles := len(structure.FilesRequest)
	totalFolders := len(structure.Folders)

	// Print header if NoPrint is false
	if !args.NoPrint {
		fmt.Println()
		fmt.Println("Downloading github Folders and files")
		fmt.Println()
		fmt.Printf("REPO: %s/%s\n", parsed.Username, parsed.Repo)
		fmt.Printf("PATH: %s\n", parsed.Path)
		if args.Commit != "" {
			fmt.Printf("COMMIT: %s\n", args.Commit)
		} else if args.Branch != "" {
			fmt.Printf("BRANCH: %s\n", args.Branch)
		}
		fmt.Printf("SIZE: %s\n", utils.FormatSize(structure.FilesSize))
		fmt.Printf("OBJECTS: (%d files, %d folders)\n", totalFiles, totalFolders)
		if len(args.Formats) > 0 {
			fmt.Printf("FORMATS: %v\n", args.Formats)
		}
		fmt.Printf("SAVED IN: %s\n", outputDir)
		fmt.Println()
	}

	// Create directories (only RequestPath)
	var createdDirs []string
	for _, folder := range structure.Folders {
		dirPath := filepath.Join(outputDir, folder)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			if !args.NoPrint {
				createdDirs = append(createdDirs, fmt.Sprintf("Error creating directory %s: %v", dirPath, err))
			}
			continue
		}
		if !args.NoPrint {
			createdDirs = append(createdDirs, fmt.Sprintf("Created directory: %s", dirPath))
		}
	}

	// Download files with progress bar
	const barWidth = 20 // Width of the progress bar
	var downloadMessages []string
	for i, downloadURL := range structure.DownloadURLs {
		// Update progress bar if NoPrint is false
		if !args.NoPrint && totalFiles > 0 {
			progress := i + 1
			filled := int(float64(progress) / float64(totalFiles) * float64(barWidth))
			bar := strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled)
			fmt.Printf("\r[%s] %d/%d", bar, progress, totalFiles)
		}

		if downloadURL == "" {
			if !args.NoPrint {
				downloadMessages = append(downloadMessages, fmt.Sprintf("No download URL for file %s", structure.FilesRequest[i]))
			}
			continue
		}

		// Construct output file path using FilesRequest
		filePath := filepath.Join(outputDir, structure.FilesRequest[i])
		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			if !args.NoPrint {
				downloadMessages = append(downloadMessages, fmt.Sprintf("Error creating parent directory for %s: %v", filePath, err))
			}
			continue
		}

		// Download file
		req, err := http.NewRequest("GET", downloadURL, nil)
		if err != nil {
			if !args.NoPrint {
				downloadMessages = append(downloadMessages, fmt.Sprintf("Error creating request for %s: %v", downloadURL, err))
			}
			continue
		}

		if token != "" {
			req.Header.Add("Authorization", "token "+token)
		}
		req.Header.Add("Accept", "application/vnd.github+json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			if !args.NoPrint {
				downloadMessages = append(downloadMessages, fmt.Sprintf("Error downloading %s: %v", downloadURL, err))
			}
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			if !args.NoPrint {
				downloadMessages = append(downloadMessages, fmt.Sprintf("Failed to download %s: status %d", downloadURL, resp.StatusCode))
			}
			continue
		}

		// Save file
		file, err := os.Create(filePath)
		if err != nil {
			if !args.NoPrint {
				downloadMessages = append(downloadMessages, fmt.Sprintf("Error creating file %s: %v", filePath, err))
			}
			continue
		}
		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			if !args.NoPrint {
				downloadMessages = append(downloadMessages, fmt.Sprintf("Error saving file %s: %v", filePath, err))
			}
			continue
		}
	}

	// Print final progress bar and messages if NoPrint is false
	if !args.NoPrint {
		if totalFiles > 0 {
			bar := strings.Repeat("=", barWidth)
			fmt.Printf("\r[%s] %d/%d\n", bar, totalFiles, totalFiles)
			fmt.Println()
		}
		for _, msg := range createdDirs {
			fmt.Println(msg)
		}
		for _, msg := range downloadMessages {
			fmt.Println(msg)
		}
		fmt.Println("DONE")
	}
}