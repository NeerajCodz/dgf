# DGF (Direct Git Fetch)

## Overview

**DGF (Direct Git Fetch)** is a command-line tool for downloading files and folders directly from Git repositories hosted on platforms like GitHub, GitLab, and HuggingFace. It supports flexible options for specifying repositories, branches, commits, file paths, and file formats, making it easy to fetch specific assets without cloning entire repositories.

## Features

- Download files or folders from Git repositories using a URL or platform-specific details (site, username, repo).
- Filter downloads by file formats (e.g., `pdf`, `jpg`, `go`) or predefined categories (e.g., `image`, `video`, `code`).
- Support for GitHub, GitLab, and HuggingFace platforms.
- Options to check paths, print repository info as JSON, or display directory trees.
- Lightweight installation with a single shell script.

## Installation

### Prerequisites

- `bash` (available on Linux, macOS, or WSL on Windows)
- `wget` or `curl` for downloading the installer
- `sudo` privileges for installation (optional for system-wide install)

### Steps

1. **Download the installer script:**
    ```sh
    wget https://raw.githubusercontent.com/NeerajCodz/dgf/main/dgf-installer.sh
    ```
2. **Make the script executable:**
    ```sh
    chmod +x dgf-installer.sh
    ```

## Usage Examples

- **Install the latest version for the current system:**
  ```sh
  sudo ./dgf-installer.sh
  ```
- **Install a specific version (e.g., 1.0) for a specific OS and architecture:**
  ```sh
  sudo ./dgf-installer.sh -v 1.0 -os linux -arch amd64
  ```
- **Download only without installing:**
  ```sh
  ./dgf-installer.sh --download-only
  ```
- **Keep the original filename:**
  ```sh
  sudo ./dgf-installer.sh --no-rename
  ```

## Usage

Run DGF to download files or folders from a Git repository:

```sh
./dgf [<URL> | -s <site> -u <username> -r <repo>] [options]
```

### Options

- `--site, -s <site>`: Platform ID (e.g., `github`, `gitlab`, `huggingface`)
- `--username, -u <username>`: Repository username
- `--repo, -r <repo>`: Repository name
- `--token, -t <token>`: GitHub token (for private repositories)
- `--branch, -b <branch>`: Branch name
- `--commit, -c <commit>`: Commit ID
- `--path, -p <path>`: Path in the repository
- `--output, -o <dir>`: Output directory (default: current directory)
- `--format, -f <format>`: File formats to include (e.g., `[pdf,jpg,go]`, `image`, or `""` for no-extension files)
- `--no-print, -n`: Suppress all output
- `--print-tree`: Print directory tree
- `--check`: Check if path exists
- `--print-info, -i`: Print repository info as JSON
- `--help, -h`: Show help message

> **Note:** Only one of `--no-print`, `--print-tree`, `--check`, or `--print-info` can be used at a time.

## Supported File Formats

The `--format` option accepts either a comma-separated list (e.g., `[pdf,jpg,go]`) or a predefined category. Supported categories and their extensions:

- **image:** jpg, jpeg, png, gif, bmp, webp, tiff, svg, heic, raw, ico, psd, ai, eps, svgz
- **video:** mp4, avi, mkv, mov, wmv, flv, webm, 3gp, m4v, mpeg, mpg, ogv
- **audio:** mp3, wav, aac, flac, ogg, m4a, wma, amr, aiff, opus
- **document:** pdf, doc, docx, xls, xlsx, ppt, pptx, txt, rtf, odt, csv, md, epub
- **archive:** zip, rar, 7z, tar, gz, bz2, iso, xz, lz
- **code:** html, css, js, ts, jsx, tsx, py, java, c, cpp, go, rs, json, xml, yaml, yml, sh, bat, ps1, rb, php, pl, kt, dart
- **e-books:** epub, mobi, azw3, fb2, lit
- **fonts:** ttf, otf, woff, woff2, eot, fon
- **3d-models:** obj, stl, fbx, gltf, glb, dae, 3ds, blend
- **spreadsheets:** xls, xlsx, ods, csv
- **presentations:** ppt, pptx, odp, key
- **databases:** sql, sqlite, db, mdb
- **executables:** exe, apk, dmg, bin
- **log:** log, env, ini, toml

## Examples

- **Download all .pdf and .jpg files from a GitHub repository:**
  ```sh
  ./dgf -s github -u NeerajCodz -r dgf -f [pdf,jpg] -o ./downloads
  ```
- **Download all images from a specific path:**
  ```sh
  ./dgf -s github -u NeerajCodz -r dgf -p assets/images -f image
  ```
- **Check if a path exists in a repository:**
  ```sh
  ./dgf -s github -u NeerajCodz -r dgf -p src --check
  ```
- **Download from a direct repository URL:**
  ```sh
  ./dgf https://github.com/NeerajCodz/dgf -f code
  ```

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository at [github.com/NeerajCodz/dgf](https://github.com/NeerajCodz/dgf).
2. Create a new branch:
    ```sh
    git checkout -b feature/your-feature
    ```
3. Commit your changes:
    ```sh
    git commit -m 'Add your feature'
    ```
4. Push to the branch:
    ```sh
    git push origin feature/your-feature
    ```
5. Open a pull request.

Please read our **Contributing Guidelines** for more details.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contact

For questions or feedback, reach out to **Neeraj SathishKumar** via [GitHub](https://github.com/NeerajCodz) or open an issue in the repository.