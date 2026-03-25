# DGF v1.x → v2.0 Migration Guide

Welcome to DGF v2.0! This guide will help you upgrade from v1.x and take advantage of powerful new features while maintaining backward compatibility with your existing workflows.

---

## 🎯 Overview: What's New in v2.0

DGF v2.0 is a major release introducing an interactive Terminal UI (TUI), persistent configuration management, and a new JSON API for scripting. The codebase has been refactored into modular components while maintaining full backward compatibility with v1.x CLI flags.

### Key Highlights
- **Interactive TUI** - Visual file browser with search, preview, and multi-select
- **Config Management** - Persistent settings without environment variables
- **Agent Mode** - JSON API for programmatic access and automation
- **Git LFS Support** - Automatic detection and download of LFS-tracked files
- **Parallel Downloads** - Faster batch downloads with configurable worker pool
- **Modular Architecture** - Cleaner codebase with better maintainability

---

## ✅ Backward Compatibility: What Still Works

All v1.x CLI flags and options remain fully functional. Your existing scripts and workflows will continue to work:

```bash
# All of these still work exactly as in v1.x
dgf https://github.com/golang/go -f go -o ./output
dgf -s github -u golang -r go -p src/ -f go -o ./output
dgf https://github.com/user/repo/blob/main/README.md -o ./output

# Environment variable authentication still works
export GITHUB_TOKEN=ghp_xxxxx
dgf https://github.com/user/private-repo

# All format filtering options unchanged
dgf -s github -u user -r repo -f [jpg,png,svg]
dgf https://github.com/user/repo -f image -o ./images

# All flags supported: --token, --branch, --commit, --path, --format, etc.
```

---

## ⚠️ Breaking Changes

### 1. Default Behavior Without Arguments

**v1.x:** Running `dgf` with no arguments displayed the help message  
**v2.0:** Running `dgf` with no arguments **launches the interactive TUI**

#### Migration Required For:
- **Shell scripts expecting help output** - Add explicit `--help` flag
- **Automated systems without terminal** - Provide full CLI arguments or use `dgf agent` mode

#### Examples of Breaking Behavior:

```bash
# v1.x behavior - shows help
$ dgf
# Help output...

# v2.0 behavior - launches TUI
$ dgf
# [Interactive Terminal UI appears]

# To get help in v2.0:
$ dgf --help
# Help output...

# To run CLI mode, provide URL/flags:
$ dgf https://github.com/golang/go -f go -o ./output
$ dgf -s github -u golang -r go -f go -o ./output
```

**Fix Your Scripts:**
```bash
# If you have scripts calling 'dgf' for help, add --help explicitly
# Before: dgf > help.txt
# After: dgf --help > help.txt

# If you have background processes or cron jobs, explicitly provide arguments
# Before: dgf -s github -u user -r repo  (would show help if config missing)
# After: dgf -s github -u user -r repo -t $GITHUB_TOKEN -o ./downloads -n
```

### 2. Environments Without Terminal Support

If running in non-interactive environments (CI/CD, cron, headless containers):
- **TUI cannot launch** - These environments don't have terminal capabilities
- **Solution:** Use full CLI arguments or use `dgf agent` mode for JSON API

```bash
# In non-interactive environments, always use:
dgf https://github.com/user/repo -f code -o ./src -n  # -n suppresses prompts

# Or use agent mode for JSON responses:
dgf agent tree https://github.com/user/repo
dgf agent download https://github.com/user/repo src/main.go --out ./output
```

---

## 🚀 New Features Guide

### 1. Interactive Terminal UI

The TUI launches when you run `dgf` with a URL or without arguments.

```bash
# Launch TUI with a repository URL
dgf https://github.com/golang/go

# Launch TUI without URL (enter URL in the UI)
dgf
```

**Keyboard Controls:**
- `↑` / `↓` or `j` / `k` - Navigate files
- `Enter` or `l` - Open folder / expand
- `Backspace` or `h` - Go back
- `Space` - Toggle file/folder selection
- `a` - Select all files
- `u` - Unselect all
- `/` - Live search/filter
- `p` - Preview file (text files only)
- `d` - Download selected files
- `i` - Toggle between emoji and ASCII icons
- `r` - Refresh
- `?` - Show help
- `q` - Quit

**Use Cases:**
- Exploring large repositories visually
- Filtering files with real-time search
- Previewing file contents before download
- Selecting multiple files at once
- Downloading entire folders recursively

### 2. Persistent Configuration

Instead of relying solely on environment variables, v2.0 includes a config file system:

```bash
# Show all configuration
dgf config show

# Get specific setting (shows masked token)
dgf config get token

# Set configuration
dgf config set token ghp_xxxxx
dgf config set download_path ./my-downloads
dgf config set workers 10
dgf config set ascii_mode true

# Show where config is stored
dgf config path
```

**Config File Locations:**
- **Windows:** `%APPDATA%\dgf\config.json`
- **macOS:** `~/Library/Application Support/dgf/config.json`
- **Linux:** `~/.config/dgf/config.json`

**Available Options:**
```json
{
  "github_token": "ghp_xxxxxxxxxxxxxxxxxxxxx",
  "download_path": "./downloads",
  "ascii_mode": false,
  "workers": 5
}
```

**Priority/Precedence** (highest to lowest):
1. CLI flags (`--token`, `--output`, etc.)
2. Environment variables (`GITHUB_TOKEN`)
3. Config file settings
4. Built-in defaults

**Migration from Environment Variables:**

```bash
# v1.x approach - set environment variable
export GITHUB_TOKEN=ghp_xxxxx
export DOWNLOAD_PATH=./downloads

# v2.0 approach - use config file (or keep using env vars)
dgf config set token ghp_xxxxx
dgf config set download_path ./downloads

# v2.0 also supports env vars for backward compatibility
# You can use either approach - both work!
```

### 3. Agent Mode (JSON API for Scripting)

New `dgf agent` subcommand provides JSON responses for automation and integration:

```bash
# Get repository structure as JSON
dgf agent tree https://github.com/golang/go | jq .

# Download files programmatically
dgf agent download https://github.com/golang/go \
  src/main.go \
  src/utils.go \
  --out ./output

# Works with private repositories
GITHUB_TOKEN=ghp_xxxxx dgf agent tree https://github.com/user/private-repo
```

**Response Format:**
```json
{
  "success": true,
  "data": {
    "files": [...],
    "folders": [...],
    "structure": {...}
  },
  "error": null
}
```

**Use Cases:**
- CI/CD pipelines with structured output
- Integration with other tools via JSON parsing
- Programmatic file selection and batch downloads
- Error handling with clear success/error flags

---

## 🔧 Configuration Migration

### From Environment Variables to Config File

If you've been using environment variables in v1.x, you have several options in v2.0:

#### Option 1: Keep Using Environment Variables (No Migration Needed)
```bash
# v1.x - still works in v2.0
export GITHUB_TOKEN=ghp_xxxxx
dgf https://github.com/golang/go -f go
```

#### Option 2: Migrate to Config File (Recommended)
```bash
# Set once, use everywhere
dgf config set token ghp_xxxxx
dgf config set download_path ~/downloads

# Now you don't need to export GITHUB_TOKEN every time
dgf https://github.com/golang/go -f go -o ~/downloads
```

#### Option 3: Mix and Match
```bash
# Config file for defaults
dgf config set download_path ~/downloads

# CLI flags for one-off settings
dgf https://github.com/golang/go -f go -o ./tmp-folder -t ghp_different_token
```

### Platform-Specific Setup

**Windows:**
```powershell
# View config location
dgf config path
# Output: C:\Users\YourName\AppData\Roaming\dgf\config.json

# Set persistent token
dgf config set token ghp_xxxxx

# View config file
cat $env:APPDATA\dgf\config.json
```

**macOS/Linux:**
```bash
# View config location
dgf config path
# Output: ~/.config/dgf/config.json

# Set persistent token
dgf config set token ghp_xxxxx

# View config file
cat ~/.config/dgf/config.json
```

---

## 💡 Code Examples for Common Tasks

### Example 1: Download Specific File Types from a Repository

**v1.x approach (still works in v2.0):**
```bash
dgf -s github -u golang -r go -p src/ -f go -o ./go-source-files
```

**v2.0 approach - with config:**
```bash
# Setup once
dgf config set download_path ~/my-downloads

# Then use simple URL format
dgf https://github.com/golang/go/tree/master/src -f go
```

### Example 2: Download from Private Repository

**v1.x approach:**
```bash
export GITHUB_TOKEN=ghp_xxxxx
dgf https://github.com/myorg/private-repo -f code
```

**v2.0 approach - options:**

```bash
# Option A: Keep using environment variable (works)
export GITHUB_TOKEN=ghp_xxxxx
dgf https://github.com/myorg/private-repo -f code

# Option B: Store in config file (recommended)
dgf config set token ghp_xxxxx
dgf https://github.com/myorg/private-repo -f code

# Option C: CLI flag for one-time override
dgf https://github.com/myorg/private-repo -f code --token ghp_xxxxx
```

### Example 3: Batch Download Multiple File Types

**v1.x approach:**
```bash
dgf -s github -u user -r repo -f [py,js,ts] -o ./source-files
```

**v2.0 TUI approach (new):**
```bash
# Interactive visual selection
dgf https://github.com/user/repo
# Then use TUI: search, select files, press 'd' to download
```

**v2.0 CLI approach (same as v1.x):**
```bash
dgf https://github.com/user/repo -f [py,js,ts] -o ./source-files
```

### Example 4: Recursive Folder Download

**v1.x approach:**
```bash
# Limited support - download everything in path
dgf https://github.com/user/repo/tree/main/src -o ./src-folder
```

**v2.0 TUI approach (new):**
```bash
# Visual selection with multi-select
dgf https://github.com/user/repo
# Navigate to desired folder, press 'Space' to select, then 'd' to download recursively
```

**v2.0 agent mode approach (new):**
```bash
dgf agent download https://github.com/user/repo src/ --out ./output
```

### Example 5: Check If Path Exists

**v1.x approach:**
```bash
dgf -s github -u user -r repo -p src/main.go --check
```

**v2.0 approach (unchanged):**
```bash
dgf -s github -u user -r repo -p src/main.go --check
# Returns JSON indicating if path exists
```

### Example 6: Get Repository Structure as JSON

**v1.x approach:**
```bash
dgf -s github -u user -r repo --print-info | jq .
```

**v2.0 approach - options:**

```bash
# Option A: Use existing flag (works exactly same)
dgf -s github -u user -r repo --print-info | jq .

# Option B: Use new agent mode (preferred)
dgf agent tree https://github.com/user/repo | jq .
```

### Example 7: Silent/Non-Interactive Download for Scripts

**v1.x approach:**
```bash
dgf -s github -u golang -r go -p src/ -f go -o ./go-src -n
```

**v2.0 approach (unchanged):**
```bash
dgf -s github -u golang -r go -p src/ -f go -o ./go-src -n
# -n flag still works for silent mode
```

### Example 8: Integration with Build Process

**Bash script - v1.x:**
```bash
#!/bin/bash
export GITHUB_TOKEN=$MY_GITHUB_TOKEN
dgf -s github -u golang -r go -p src/ -f go -o ./vendor/go-src -n
if [ $? -eq 0 ]; then
    echo "Downloaded Go source files"
else
    echo "Failed to download"
    exit 1
fi
```

**Bash script - v2.0 (recommended approach):**
```bash
#!/bin/bash
# Setup config once (or use environment variable)
export GITHUB_TOKEN=$MY_GITHUB_TOKEN

# Use agent mode for reliable JSON parsing
RESULT=$(dgf agent download https://github.com/golang/go \
    src/main.go \
    src/utils.go \
    --out ./vendor/go-src)

if echo "$RESULT" | jq -e '.success' > /dev/null; then
    echo "Downloaded successfully"
else
    echo "Download failed: $(echo $RESULT | jq -r '.error')"
    exit 1
fi
```

### Example 9: Download All Images from Documentation Folder

**CLI approach:**
```bash
dgf https://github.com/user/repo/tree/main/docs -f image -o ./images
```

**TUI approach (new):**
```bash
# Interactive visual browsing
dgf https://github.com/user/repo
# Search: /image
# Select desired images with Space, then press 'd'
```

### Example 10: Parallel Download with Custom Worker Count

**v2.0 feature - set worker pool:**
```bash
# Set workers for faster downloads (default: 5)
dgf config set workers 10

# Now all downloads will use 10 parallel workers
dgf https://github.com/large-repo/with-many-files -f code -o ./source
```

---

## 📋 Upgrade Checklist

### Before Upgrading
- [ ] Review breaking changes section above
- [ ] Check if you have any scripts calling `dgf` without arguments
- [ ] Note your current GitHub token setup (environment variable or CLI flag)

### After Upgrading
- [ ] Test your existing scripts still work
- [ ] Update any automation expecting help output (add `--help`)
- [ ] Fix non-interactive environments (add explicit arguments)
- [ ] (Optional) Migrate GitHub token to config file with `dgf config set token`
- [ ] (Optional) Try the interactive TUI: `dgf https://github.com/golang/go`
- [ ] (Optional) Try agent mode for scripting: `dgf agent tree https://github.com/golang/go`

### Testing Your Migration

```bash
# Test 1: Verify existing CLI commands still work
dgf https://github.com/golang/go -f go -o ./test-output

# Test 2: Test with environment variable
export GITHUB_TOKEN=ghp_xxxxx
dgf https://github.com/user/private-repo -f code -o ./test-private

# Test 3: Test TUI (if running interactively)
dgf

# Test 4: Test agent mode
dgf agent tree https://github.com/golang/go | head -20

# Test 5: Verify config management works
dgf config show
dgf config get token
```

---

## 🆘 Troubleshooting

### Issue: TUI Launches Instead of Showing Help

**Problem:** You're running `dgf` without arguments and expecting help

**Solution:** Use explicit flag
```bash
dgf --help
```

### Issue: Scripts Hang in Non-Interactive Environments

**Problem:** TUI tries to launch but terminal isn't available

**Solution:** Provide full CLI arguments or use agent mode
```bash
# Option 1: CLI with all arguments
dgf https://github.com/user/repo -f code -o ./output -n

# Option 2: Agent mode
dgf agent download https://github.com/user/repo file.go --out ./output
```

### Issue: "Permission denied" on Config File

**Problem:** Config file has wrong permissions or location doesn't exist

**Solution:** Check config location
```bash
dgf config path
dgf config show  # Creates config directory if needed
```

### Issue: Token Not Being Read from Config

**Problem:** Token not working even though `dgf config set token` succeeded

**Solution:** Verify precedence - CLI flags and environment variables override config
```bash
# Check what's being used
dgf config get token
echo $GITHUB_TOKEN  # Check if env var is overriding

# Test with explicit flag to confirm it's a config issue
dgf https://github.com/user/repo --token ghp_xxxxx
```

---

## 📞 Getting Help

- **Help text:** `dgf --help`
- **Version info:** `dgf --version`
- **Config help:** `dgf config --help`
- **Agent help:** `dgf agent --help`
- **GitHub Issues:** Report bugs or request features

---

## Summary

**Key Takeaways:**
1. ✅ All v1.x CLI flags work - backward compatible
2. ⚠️ No-args now launches TUI instead of showing help
3. 🚀 New features: TUI, config file, agent mode
4. 🔧 Config file provides persistent settings
5. 📝 Environment variables still fully supported
6. 🤖 Agent mode for scripting and integration

**Recommended Next Steps:**
1. Upgrade and test your existing scripts
2. Fix any scripts affected by the no-args behavior change
3. (Optional) Set up config file: `dgf config set token <token>`
4. (Optional) Try the TUI: `dgf https://github.com/golang/go`
5. (Optional) Explore agent mode for new automation possibilities

Welcome to DGF v2.0! 🎉
