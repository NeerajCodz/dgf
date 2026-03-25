# Changelog

All notable changes to the DGF (Direct Git Fetch) project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [2.0.0] - 2026-03-26

### Added

#### Interactive TUI
- **File Browser**: Visual terminal UI with keyboard navigation for browsing repository contents
- **Real-time Search**: Press `/` to filter files instantly as you type
- **Multi-select**: Use `Space` to toggle selection, `a` for select all, `u` for unselect all
- **File Preview**: Preview text files before downloading with `p` key
- **Icon Toggle**: Switch between emoji and ASCII icons with `i` key
- **Breadcrumb Navigation**: Visual path indicator with back button support

#### Config Management
- New `dgf config` subcommand for persistent configuration
- `dgf config show` - Display all settings
- `dgf config get <key>` - Get specific value
- `dgf config set <key> <value>` - Set configuration
- `dgf config path` - Show config file location
- Configurable options: `token`, `download_path`, `ascii_mode`, `workers`

#### Agent/JSON Mode
- New `dgf agent tree <url>` - Get repository tree as JSON
- New `dgf agent download <url> <paths...>` - Download files programmatically
- Machine-readable JSON output for scripting

#### Git LFS Support
- Automatic detection of Git LFS pointer files
- Seamless download of LFS-tracked files
- Falls back to pointer file if LFS download fails

#### Download Improvements
- Parallel download with configurable worker pool
- Recursive folder selection and download
- Progress tracking for batch downloads
- Better error handling and retry logic

### Changed
- Complete codebase restructure for modularity
- Migrated from flat structure to `cmd/`, `internal/`, `pkg/` layout
- Improved argument parsing with better error messages
- Enhanced help text with more examples

### Technical
- Added bubbletea for TUI framework
- Added lipgloss for terminal styling
- Added bubbles for UI components
- New internal packages: `tui`, `config`, `selection`, `agent`
- Refactored GitHub client with cleaner API

---

## [1.0.1] - 2025-08-15

### Added

- **Installer enhancements:**
  - Seamless update support: installer checks for the latest version and updates only if needed.
  - Robust uninstallation: `--uninstall`/`-u` flag to remove DGF from default or custom locations.
  - Improved error handling, debug output, and file verification (size and SHA256 checksum).
  - More flexible installation for Windows, Linux, macOS, and Android.
- **Token support:**
  - Provide a GitHub token via the `--token` flag or the `GITHUB_TOKEN` environment variable for private repositories. The flag takes precedence if both are set.
- **Documentation:**
  - Updated README with detailed installation, update, uninstall, and argument usage instructions, including token handling and all command-line options.
- **New builds:**
  - Binaries for: linux-amd64, linux-arm64, linux-arm, darwin-amd64, darwin-arm64, windows-amd64.exe, windows-arm64.exe, android-arm64.

### Changed

- Improved installer script logic and user feedback.
- Expanded documentation for all command-line arguments and installer options.

### Fixed

- Patch fixes for installer reliability and platform compatibility.

---

## [1.0.0] - 2025-06-24

### Added

- **Initial release of DGF**: A command-line tool for downloading files and folders directly from Git repositories.
- **Supported platforms**: GitHub, GitLab, and HuggingFace.
- **Flexible filtering**: Download specific file formats by passing a list (e.g., `[pdf,jpg,go]`) or using predefined categories like `image`, `video`, `document`, `code`, etc.
- **Comprehensive format categories**:  
    - `image`: jpg, jpeg, png, gif, bmp, webp, tiff, svg, heic, raw, ico, psd, ai, eps, svgz  
    - `video`: mp4, avi, mkv, mov, wmv, flv, webm, 3gp, m4v, mpeg, mpg, ogv  
    - `audio`: mp3, wav, aac, flac, ogg, m4a, wma, amr, aiff, opus  
    - `document`: pdf, doc, docx, xls, xlsx, ppt, pptx, txt, rtf, odt, csv, md, epub  
    - `archive`: zip, rar, 7z, tar, gz, bz2, iso, xz, lz  
    - `code`: html, css, js, ts, jsx, tsx, py, java, c, cpp, go, rs, json, xml, yaml, yml, sh, bat, ps1, rb, php, pl, kt, dart  
    - `e-books`, `fonts`, `3d-models`, `spreadsheets`, `presentations`, `databases`, `executables`, `log`
- **Installer script**: `dgf-installer.sh` with options for version, OS, architecture, download-only, and no-rename.
- **Command-line options**:
    - `--site, -s <site>`: Platform ID (github, gitlab, huggingface)
    - `--username, -u <username>`: Repository username
    - `--repo, -r <repo>`: Repository name
    - `--token, -t <token>`: GitHub token
    - `--branch, -b <branch>`: Branch name
    - `--commit, -c <commit>`: Commit ID
    - `--path, -p <path>`: Path in repository
    - `--output, -o <dir>`: Output directory (default: .)
    - `--format, -f <format>`: File formats to include (e.g., `image`, `[jpg,pdf,png]`, or `""` for no-extension files)
    - `--no-print, -n`: Suppress all output
    - `--print-tree`: Print directory tree
    - `--check`: Check if path exists
    - `--print-info, -i`: Print repository info as JSON
    - `--help, -h`: Show help message
- **Usage**:
    ```
    ./dgf [ <URL> | -s <site> -u <username> -r <repo> ] [options]
    ```
    > Note: Only one of `--no-print`, `--print-tree`, `--check`, or `--print-info` can be provided at a time.
- **Features**:
    - Download files via direct repository URLs or detailed parameters.
    - Check repository paths, print directory trees, and output repository info as JSON.

---

**Author**: [Neeraj SathishKumar](https://github.com/NeerajCodz)
