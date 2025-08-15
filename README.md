# DGF (Direct Git Fetch)

## Overview

**DGF (Direct Git Fetch)** is a command-line tool for downloading files and folders directly from Git repositories hosted on platforms like GitHub, GitLab, and HuggingFace. It supports flexible options for specifying repositories, branches, commits, file paths, and file formats, making it easy to fetch specific assets without cloning entire repositories.

## Features

- Download files or folders from Git repositories using a URL or platform-specific details (site, username, repo).
- Filter downloads by file formats (e.g., `pdf`, `jpg`, `go`) or predefined categories (e.g., `image`, `video`, `code`).
- Support for GitHub, GitLab, and HuggingFace platforms.
- Options to check paths, print repository info as JSON, or display directory trees.
- Lightweight installation with a single shell script.


## Installation, Update, and Uninstallation

### Prerequisites

- **bash** (Linux, macOS, or WSL on Windows)
- **curl** and **jq** (required for the installer script)
- **sudo** privileges for system-wide install (optional)

### Install or Update DGF

You can install or update DGF using the provided installer script. The script automatically detects your OS and architecture, fetches the latest (or specified) release, verifies the download, and installs the binary.

#### 1. Download the installer script
```sh
curl -LO https://raw.githubusercontent.com/NeerajCodz/dgf/main/dgf-installer.sh
# or
wget https://raw.githubusercontent.com/NeerajCodz/dgf/main/dgf-installer.sh
```

#### 2. Make the script executable
```sh
chmod +x dgf-installer.sh
```

#### 3. Install the latest version (auto-detects OS/arch)
```sh
sudo ./dgf-installer.sh
```

#### 4. Install a specific version, OS, or architecture
```sh
sudo ./dgf-installer.sh -v 1.0.0 -os linux -arch amd64
```

#### 5. Download only (do not install)
```sh
./dgf-installer.sh --download-only
```

#### 6. Keep the original filename (do not rename to dgf/dgf.exe)
```sh
sudo ./dgf-installer.sh --no-rename
```

#### 7. Enable debug output
```sh
./dgf-installer.sh --debug
```


#### 8. Uninstall DGF
```sh
# Uninstall from default locations
sudo ./dgf-installer.sh --uninstall
# or using the short flag
sudo ./dgf-installer.sh -u
# Uninstall from a specific directory
sudo ./dgf-installer.sh --uninstall /path/to/dir
sudo ./dgf-installer.sh -u /path/to/dir
```

> **Note:** The `--uninstall` or `-u` flag removes DGF from system or user locations. You can optionally provide a directory to target a specific uninstall path.

**Notes:**
- On Windows, the binary is installed to `C:\Program Files\dgf` (add this to your PATH manually).
- On Linux/macOS, the default install location is `/usr/local/bin` (falls back to `$HOME/bin` if permissions are insufficient).
- The installer checks for updates and will not overwrite a newer version.
- The script verifies file size and SHA256 checksum if available.

---


## Usage

Run DGF to download files or folders from a Git repository:

```sh
./dgf [<URL> | -s <site> -u <username> -r <repo>] [options]
```

### Command-Line Arguments

You can specify either a direct repository URL or the combination of `--site`, `--username`, and `--repo`. Only one method is allowed per command.

#### Main Arguments

- `--site, -s <site>`: Platform ID (e.g., `github`, `gitlab`, `huggingface`).(Only `github` is avaliable for now)
- `--username, -u <username>`: Repository username/owner.
- `--repo, -r <repo>`: Repository name.
- `<URL>`: Direct repository URL (alternative to the above three).

#### Optional Arguments

- `--token, -t <token>`: Access token for private repositories (e.g., GitHub token). If not provided, DGF will use the `GITHUB_TOKEN` environment variable if set.
- `--branch, -b <branch>`: Branch name to fetch from (default: default branch).
- `--commit, -c <commit>`: Commit SHA to fetch from (overrides branch).
- `--path, -p <path>`: Path in the repository to download (default: root).
- `--output, -o <dir>`: Output directory for downloads (default: current directory).
- `--format, -f <format>`: File formats to include. Accepts:
  - A category (e.g., `image`, `video`, `code`, see below for categories)
  - A list of extensions (e.g., `[jpg,pdf,go]`)
  - `""` (empty string) to match files with no extension
- `--no-print, -n`: Suppress all output (silent mode).
- `--print-tree`: Print the directory tree of the downloaded files.
- `--check`: Check if the specified path exists in the repository (no download).
- `--print-info, -i`: Print repository info as JSON.
- `--help, -h`: Show help message and exit.

> **Note:** Only one of `--no-print`, `--print-tree`, `--check`, or `--print-info` can be used at a time.

#### Token Usage

- For private repositories, provide a token using `--token` or set the `GITHUB_TOKEN` environment variable:
  ```sh
  export GITHUB_TOKEN=your_token_here
  ./dgf ...
  ```
- The `--token` flag overrides the environment variable if both are set.

#### Examples

- Download all `.pdf` and `.jpg` files from a GitHub repository:
  ```sh
  ./dgf -s github -u NeerajCodz -r dgf -f [pdf,jpg] -o ./downloads
  ```
- Download all images from a specific path:
  ```sh
  ./dgf -s github -u NeerajCodz -r dgf -p assets/images -f image
  ```
- Check if a path exists in a repository:
  ```sh
  ./dgf -s github -u NeerajCodz -r dgf -p src --check
  ```
- Download from a direct repository URL:
  ```sh
  ./dgf https://github.com/NeerajCodz/dgf -f code
  ```

---

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