package github

import (
	"context"
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
	Context      context.Context
	OutputDir    string
	Token        string
	Workers      int
	Silent       bool
	CheckLFS     bool
	Owner        string
	Repo         string
	GitHubClient *Client
	OnProgress   func(current, total int)
	OnFileStart  func(path string)
	OnFileDone   func(path string, err error)
}

// Download downloads files from a repository structure
func Download(structure types.RepositoryStructure, opts DownloadOptions) error {
	if opts.OutputDir == "" {
		opts.OutputDir = "."
	}
	if opts.Workers <= 0 {
		opts.Workers = 5
	}
	if opts.Context == nil {
		opts.Context = context.Background()
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

	// Download files with worker pool and retry logic
	var wg sync.WaitGroup
	var completed int32
	errors := make([]error, 0)
	var errMu sync.Mutex

	// Create job channel with file indices
	jobs := make(chan int, totalFiles)
	for i := 0; i < totalFiles; i++ {
		jobs <- i
	}
	close(jobs)

	const maxRetries = 3

	// Start workers
	for w := 0; w < opts.Workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{}

			for i := range jobs {
				if opts.Context.Err() != nil {
					return
				}
				downloadURL := structure.DownloadURLs[i]
				filePath := filepath.Join(opts.OutputDir, structure.FilesRequest[i])

				if opts.OnFileStart != nil {
					opts.OnFileStart(structure.FilesRequest[i])
				}

				// Retry logic: attempt up to maxRetries times
				var err error
				for attempt := 0; attempt < maxRetries; attempt++ {
					if downloadURL != "" {
						err = downloadFileWithOptions(client, DownloadFileOptions{
							Context:      opts.Context,
							URL:          downloadURL,
							DestPath:     filePath,
							Token:        opts.Token,
							Owner:        opts.Owner,
							Repo:         opts.Repo,
							CheckLFS:     opts.CheckLFS,
							GitHubClient: opts.GitHubClient,
						})
						if err == nil {
							// Success, break out of retry loop
							break
						}
						// On failure, will retry unless it's the last attempt
					} else {
						err = fmt.Errorf("no download URL")
						break // Don't retry if there's no URL
					}
				}

				if err != nil {
					errMu.Lock()
					errors = append(errors, fmt.Errorf("%s (after %d retries): %v", structure.FilesRequest[i], maxRetries, err))
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
	if opts.Context.Err() != nil {
		return opts.Context.Err()
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
// DownloadFileOptions configures individual file download
type DownloadFileOptions struct {
	Context      context.Context
	URL          string
	DestPath     string
	Token        string
	Owner        string
	Repo         string
	CheckLFS     bool
	GitHubClient *Client
}

func downloadFile(client *http.Client, url, destPath, token string) error {
	return downloadFileWithOptions(client, DownloadFileOptions{
		Context:  context.Background(),
		URL:      url,
		DestPath: destPath,
		Token:    token,
	})
}

func downloadFileWithOptions(client *http.Client, opts DownloadFileOptions) error {
	if opts.Context == nil {
		opts.Context = context.Background()
	}
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(opts.DestPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	req, err := http.NewRequestWithContext(opts.Context, "GET", opts.URL, nil)
	if err != nil {
		return err
	}

	if opts.Token != "" {
		req.Header.Add("Authorization", "token "+opts.Token)
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

	// Read content to check for LFS
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Check if this is an LFS pointer file
	if opts.CheckLFS && opts.GitHubClient != nil && IsLFSPointer(content) {
		pointer, err := ParseLFSPointer(content)
		if err != nil {
			// Not a valid LFS pointer, save as-is
			return writeFile(opts.DestPath, content)
		}

		// Fetch actual content from LFS
		lfsContent, err := opts.GitHubClient.FetchLFSFile(opts.Owner, opts.Repo, pointer)
		if err != nil {
			// LFS fetch failed, save pointer file
			return writeFile(opts.DestPath, content)
		}
		return writeFile(opts.DestPath, lfsContent)
	}

	return writeFile(opts.DestPath, content)
}

func writeFile(path string, content []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(content)
	return err
}
