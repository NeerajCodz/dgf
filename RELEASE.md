
# DGF v1.0.1 Release Notes

## 🛠 Patch Release – v1.0.1 (2025-08-15)

We are pleased to announce **DGF v1.0.1**, a patch release focused on improving installation, update, and uninstallation workflows, as well as enhanced token support and expanded platform builds.

---

## 🆕 What's New & Improved

- **Installer Enhancements:**
  - The `dgf-installer.sh` script now supports seamless updating of existing installations. It checks for the latest version and only updates if needed.
  - Added a robust uninstallation option (`--uninstall`), allowing you to remove DGF from default or custom locations.
  - Improved error handling, debug output, and file verification (size and SHA256 checksum).
  - More flexible installation for Windows, Linux, macOS, and Android (see builds below).

- **Token Support:**
  - You can now provide a GitHub token via the `--token` flag or the `GITHUB_TOKEN` environment variable for private repositories. The CLI prioritizes the flag if both are set.

- **Documentation:**
  - The README has been updated with detailed installation, update, uninstall, and argument usage instructions, including token handling and all command-line options.

---

## 📦 New Builds Available

The following binaries are available for v1.0.1:

- dgf-1.0.1-linux-amd64
- dgf-1.0.1-linux-arm64
- dgf-1.0.1-linux-arm
- dgf-1.0.1-darwin-amd64
- dgf-1.0.1-darwin-arm64
- dgf-1.0.1-windows-amd64.exe
- dgf-1.0.1-windows-arm64.exe
- dgf-1.0.1-android-arm64

---

## 🚀 Features Recap

- Download files or folders from GitHub, GitLab, and HuggingFace using a URL or detailed parameters.
- Filter downloads by file formats (e.g., `[pdf,jpg,go]`) or categories (`image`, `video`, `code`, etc.).
- Check repository paths, print directory trees, and output repository info as JSON.
- Lightweight installation and easy updates via the installer script.

---

## 📝 Command-Line Options

| Option | Description |
|--------|-------------|
| `--site, -s <site>` | Platform ID (`github`, `gitlab`, `huggingface`) |
| `--username, -u <username>` | Repository username |
| `--repo, -r <repo>` | Repository name |
| `--token, -t <token>` | GitHub token (or use `GITHUB_TOKEN` env) |
| `--branch, -b <branch>` | Branch name |
| `--commit, -c <commit>` | Commit ID |
| `--path, -p <path>` | Path in repository |
| `--output, -o <dir>` | Output directory (default: .) |
| `--format, -f <format>` | File formats to include (e.g., `image`, `[jpg,pdf,png]`, or `""` for no-extension files) |
| `--no-print, -n` | Suppress all output |
| `--print-tree` | Print directory tree |
| `--check` | Check if path exists |
| `--print-info, -i` | Print repository info as JSON |
| `--help, -h` | Show help message |

> Only one of `--no-print`, `--print-tree`, `--check`, or `--print-info` can be provided at a time.

---

## 🛠 Installation & Update

Download the installer:

```sh
curl -LO https://raw.githubusercontent.com/NeerajCodz/dgf/main/dgf-installer.sh
chmod +x dgf-installer.sh
```

Install or update (system-wide, optional):

```sh
sudo ./dgf-installer.sh
```

Uninstall:

```sh
sudo ./dgf-installer.sh --uninstall
# or from a specific directory
sudo ./dgf-installer.sh --uninstall /path/to/dir
```

For more options, see the installer help or README.

---

## 📦 Examples

- Download all `.pdf` and `.jpg` files:
  ```sh
  dgf -s github -u user -r repo -f [pdf,jpg]
  ```
- Download all images from a path:
  ```sh
  dgf -s github -u user -r repo -p path/to/dir -f image
  ```
- Check if a path exists:
  ```sh
  dgf -s github -u user -r repo -p path/to/file --check
  ```
- Download from a direct URL:
  ```sh
  dgf https://github.com/user/repo/path/to/file.pdf
  ```

---

## 🤝 Contributing

We welcome contributions! Please see the `CONTRIBUTING.md` for guidelines on reporting issues, suggesting features, and submitting code changes.

1. Fork the repo, create a branch, make your changes, and open a pull request.
2. Follow the code style and update documentation as needed.

---

## 📄 License

MIT License. See the LICENSE file for details.

---

**Author:** Neeraj SathishKumar
