package github

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/NeerajCodz/dgf/internal/utils"
	"github.com/NeerajCodz/dgf/pkg/types"
)

// DownloadOptions configures download behavior
type DownloadOptions struct {
	OutputDir   string
	Token       string
	Workers     int
	Silent      bool
	OnProgress  func(current, total int)
	OnFileStart func(path string)
	OnFileDone  func(path string, err error)
}

// Download downloads files from a repository structure
func Download(structure types.RepositoryStructure, opts DownloadOptions) error {
	if opts.OutputDir == "" {
		opts.OutputDir = "."
	}
	if opts.Workers <= 0 {
		opts.Workers = 5
	}

	// Ensure output directory exists
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	totalFiles := len(structure.FilesRequest)
	if totalFiles == 0 {
		return nil
	}

	// Create directories first
	for _, folder := range structure.Folders {
		dirPath := filepath.Join(opts.OutputDir, folder)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			if !opts.Silent {
				fmt.Fprintf(os.Stderr, "Warning: failed to create directory %s: %v\n", dirPath, err)
			}
		}
	}

	// Download files with worker pool
	var wg sync.WaitGroup
	var completed int32
	errors := make([]error, 0)
	var errMu sync.Mutex

	// Create job channel
	jobs := make(chan int, totalFiles)
	for i := 0; i < totalFiles; i++ {
		jobs <- i
	}
	close(jobs)

	// Start workers
	for w := 0; w < opts.Workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{}

			for i := range jobs {
				downloadURL := structure.DownloadURLs[i]
				filePath := filepath.Join(opts.OutputDir, structure.FilesRequest[i])

				if opts.OnFileStart != nil {
					opts.OnFileStart(structure.FilesRequest[i])
				}

				var err error
				if downloadURL != "" {
					err = downloadFile(client, downloadURL, filePath, opts.Token)
				} else {
					err = fmt.Errorf("no download URL")
				}

				if err != nil {
					errMu.Lock()
					errors = append(errors, fmt.Errorf("%s: %v", structure.FilesRequest[i], err))
					errMu.Unlock()
				}

				if opts.OnFileDone != nil {
					opts.OnFileDone(structure.FilesRequest[i], err)
				}

				current := int(atomic.AddInt32(&completed, 1))
				if opts.OnProgress != nil {
					opts.OnProgress(current, totalFiles)
				}
			}
		}()
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("%d files failed to download", len(errors))
	}
	return nil
}

// DownloadWithProgress downloads files with CLI progress bar
func DownloadWithProgress(structure types.RepositoryStructure, token, outputDir string, args types.Args, parsed types.ParsedURL) {
	totalFiles := len(structure.FilesRequest)
	totalFolders := len(structure.Folders)

	if !args.NoPrint {
		fmt.Println()
		fmt.Println("Downloading GitHub files and folders")
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

	// Use Download with progress callback
	Download(structure, DownloadOptions{
		OutputDir: outputDir,
		Token:     token,
		Workers:   1, // Sequential for CLI progress bar
		Silent:    args.NoPrint,
		OnProgress: func(current, total int) {
			if !args.NoPrint && total > 0 {
				barWidth := 20
				filled := int(float64(current) / float64(total) * float64(barWidth))
				bar := strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled)
				fmt.Printf("\r[%s] %d/%d", bar, current, total)
			}
		},
	})

	if !args.NoPrint {
		if totalFiles > 0 {
			fmt.Println()
			fmt.Println()
		}
		fmt.Println("DONE")
	}
}

// downloadFile downloads a single file
func downloadFile(client *http.Client, url, destPath, token string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}
	req.Header.Add("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
