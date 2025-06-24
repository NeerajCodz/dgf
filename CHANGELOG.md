# Changelog

All notable changes to the DGF (Direct Git Fetch) project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
    - `e-books`, `fonts`, `3d-models`, `spreadsheets`, `presentations`, `databases`, `executables`, `logs-config`
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

### Fixed

- N/A (initial release).

### Changed

- N/A (initial release).

### Deprecated

- N/A (initial release).

### Removed

- N/A (initial release).

### Security

- N/A (initial release).

---

**Author**: [Neeraj SathishKumar](https://github.com/NeerajCodz)
