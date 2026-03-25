# DGF (Direct Git Fetch) v2.0

## Overview

**DGF (Direct Git Fetch)** is a powerful tool for downloading files and folders directly from GitHub repositories. Version 2.0 introduces an **interactive terminal UI (TUI)** with visual file browsing, search, and multi-select capabilities, while maintaining full backward compatibility with the CLI interface.

## What's New in v2.0

- 🖥️ **Interactive TUI** - Visual file browser with keyboard navigation
- 🔍 **Real-time Search** - Filter files instantly with `/` key
- 📂 **Visual Selection** - Multi-select with space bar, select all with `a`
- 👁️ **File Preview** - Preview text files before downloading with `p`
- ⚡ **Parallel Downloads** - Configurable worker pool for faster downloads
- 🔧 **Config Management** - Persistent settings with `dgf config` command
- 🤖 **Agent/JSON Mode** - Machine-readable output for scripts
- 📦 **Git LFS Support** - Automatic detection and download of LFS files
- 🎨 **Icon Toggle** - Switch between emoji and ASCII icons with `i`

## Quick Start

### Interactive Mode (NEW!)
```sh
# Launch the interactive TUI
dgf

# Or start with a URL
dgf https://github.com/user/repo
```

### CLI Mode (Backward Compatible)
```sh
# Download files with CLI flags
dgf https://github.com/user/repo -f code -o ./output

# Using component flags
dgf -s github -u user -r repo -p src/
```

## Features

### Interactive TUI Controls

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate up/down |
| `Enter` or `l` | Enter folder |
| `Backspace` or `h` | Go back |
| `Space` | Toggle selection |
| `a` | Select all |
| `u` | Unselect all |
| `/` | Search/filter |
| `p` | Preview file |
| `d` | Download selected |
| `i` | Toggle icons |
| `r` | Refresh |
| `?` | Help |
| `q` | Quit |

### Config Management

```sh
# Show all config
dgf config show

# Get a specific value
dgf config get token

# Set values
dgf config set token ghp_xxxxx
dgf config set download_path ./downloads
dgf config set workers 10
dgf config set ascii_mode true

# Show config file location
dgf config path
```

### Agent/JSON Mode

```sh
# Get repository tree as JSON
dgf agent tree https://github.com/user/repo

# Download specific files programmatically
dgf agent download https://github.com/user/repo README.md src/main.go --out ./output
```

## Installation

### Prerequisites

- **bash** (Linux, macOS, or WSL on Windows)
- **curl** and **jq** (required for the installer script)
- **sudo** privileges for system-wide install (optional)

### Install or Update DGF

```sh
# Download installer
curl -LO https://raw.githubusercontent.com/NeerajCodz/dgf/main/dgf-installer.sh
chmod +x dgf-installer.sh

# Install latest version
sudo ./dgf-installer.sh

# Install specific version
sudo ./dgf-installer.sh -v 2.0.0
```

### Uninstall

```sh
sudo ./dgf-installer.sh --uninstall
```

## CLI Reference

```
dgf                                    Launch interactive TUI
dgf <URL>                              Download from URL (CLI mode)
dgf -s <site> -u <user> -r <repo>      Download using components (CLI mode)
dgf config <get|set|show> [key] [val]  Manage configuration
dgf agent <tree|download> <url> ...    JSON API mode for scripts
```

### Options

| Flag | Short | Description |
|------|-------|-------------|
| `--site` | `-s` | Platform ID (e.g., github) |
| `--username` | `-u` | Repository owner |
| `--repo` | `-r` | Repository name |
| `--token` | `-t` | GitHub token |
| `--branch` | `-b` | Branch name |
| `--commit` | `-c` | Commit SHA |
| `--path` | `-p` | Path in repository |
| `--output` | `-o` | Output directory |
| `--format` | `-f` | File formats |
| `--no-print` | `-n` | Silent mode |
| `--print-tree` | | Print directory tree |
| `--check` | | Check if path exists |
| `--print-info` | `-i` | Print repo info as JSON |
| `--help` | `-h` | Show help |

## Token Usage

For private repositories:

```sh
# Using environment variable (recommended)
export GITHUB_TOKEN=your_token_here
dgf https://github.com/user/private-repo

# Using --token flag
dgf https://github.com/user/private-repo -t your_token_here

# Using config (persisted)
dgf config set token your_token_here
```

## Supported File Formats

The `--format` option accepts categories or comma-separated extensions:

- **image:** jpg, jpeg, png, gif, svg, webp, ...
- **video:** mp4, avi, mkv, mov, webm, ...
- **audio:** mp3, wav, aac, flac, ...
- **document:** pdf, doc, docx, txt, md, ...
- **code:** html, css, js, ts, py, go, rs, ...
- **archive:** zip, rar, tar, gz, ...

Example:
```sh
# Download only images
dgf https://github.com/user/repo -f image

# Download specific extensions
dgf https://github.com/user/repo -f [jpg,png,svg]
```

## Examples

### Interactive Use Cases

```sh
# Browse a repository interactively
dgf https://github.com/golang/go

# Browse and search for specific files
# Press / and type "readme" to filter
```

### CLI Use Cases

```sh
# Download all Go source files
dgf -s github -u golang -r go -p src -f go

# Download specific folder
dgf https://github.com/user/repo/tree/main/docs -o ./local-docs

# Check if a path exists
dgf -s github -u user -r repo -p src/main.go --check
```

### Scripting

```sh
# Get file list as JSON
files=$(dgf agent tree https://github.com/user/repo)

# Download files programmatically  
dgf agent download https://github.com/user/repo README.md LICENSE --out ./deps
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push: `git push origin feature/amazing`
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) file.

## Contact

- GitHub: [@NeerajCodz](https://github.com/NeerajCodz)
- Issues: [github.com/NeerajCodz/dgf/issues](https://github.com/NeerajCodz/dgf/issues)