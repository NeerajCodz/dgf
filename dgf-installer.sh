#!/usr/bin/env bash
set -e

# Default values
VERSION="latest"
OS=""
ARCH=""
DOWNLOAD_ONLY="false"
NO_RENAME="false"
DEBUG="false"
UNINSTALL="false"
UNINSTALL_DIR=""
BASE_URL="https://github.com/NeerajCodz/dgf/releases"
API_URL="https://api.github.com/repos/NeerajCodz/dgf/releases"
API_LATEST_URL="https://api.github.com/repos/NeerajCodz/dgf/releases/latest"
CURL=$(command -v curl)
JQ=$(command -v jq)
MAX_RETRIES=3
RETRY_DELAY=5
TEMP_FILES=()
LAST_CURL_ERROR=""
TMP_DIR="${TMPDIR:-/tmp}"

# Function to clean up temporary files
cleanup() {
    debug "Cleaning up temporary files"
    for file in "${TEMP_FILES[@]}"; do
        [ -f "$file" ] && rm -f "$file" && debug "Removed $file"
    done
}

# Trap to ensure cleanup on script exit and interruption
trap cleanup EXIT INT TERM

# Function to print error and exit
error_exit() {
    echo "[X] $1" >&2
    exit 1
}

# Function to print info
info() {
    echo "[-] $1"
}

# Function to print debug messages
debug() {
    if [ "$DEBUG" = "true" ]; then
        echo "[-] DEBUG: $1"
    fi
}

# Function to print warning
warning() {
    echo "[!] $1"
}

# Function to create temporary files in a portable location
make_temp_file() {
    mktemp "${TMP_DIR}/dgf-installer.XXXXXX" 2>/dev/null || return 1
}

# Function to fetch with retries
fetch_with_retries() {
    local url=$1
    local output=$2
    local attempt=1
    local err_file
    err_file=$(make_temp_file) || error_exit "Failed to create temporary file for curl errors"
    TEMP_FILES+=("$err_file")
    LAST_CURL_ERROR="$err_file"
    TEMP_FILES+=("$output")
    while [ $attempt -le $MAX_RETRIES ]; do
        debug "Attempt $attempt: curl -sL -w \"%{http_code}\" \"$url\" -o \"$output\""
        HTTP_STATUS=$("$CURL" -sL -w "%{http_code}" "$url" -o "$output" 2> "$err_file" || echo "000")
        if [ "$HTTP_STATUS" = "200" ]; then
            return 0
        fi
        warning "Attempt $attempt failed with HTTP $HTTP_STATUS: $(cat "$err_file" 2>/dev/null || echo 'Unknown error')"
        [ $attempt -lt $MAX_RETRIES ] && sleep $RETRY_DELAY
        ((attempt++))
    done
    return 1
}

# Function to compare semantic versions (e.g., v1.0.0 vs v1.0.1)
compare_versions() {
    local ver1="$1" ver2="$2"
    # Remove 'v' prefix if present
    ver1=${ver1#v}
    ver2=${ver2#v}
    # Split into arrays
    IFS='.' read -ra v1_parts <<< "$ver1"
    IFS='.' read -ra v2_parts <<< "$ver2"
    # Compare each part
    for i in 0 1 2; do
        local v1_part=${v1_parts[$i]:-0}
        local v2_part=${v2_parts[$i]:-0}
        if [ "$v1_part" -lt "$v2_part" ]; then
            return 1  # ver1 < ver2
        elif [ "$v1_part" -gt "$v2_part" ]; then
            return 2  # ver1 > ver2
        fi
    done
    return 0  # ver1 == ver2
}

# Parse command-line arguments
while [ $# -gt 0 ]; do
    case "$1" in
        -v|--version)
            [ -z "$2" ] && error_exit "Version argument requires a value"
            VERSION="$2"
            shift 2
            ;;
        -os|--os)
            [ -z "$2" ] && error_exit "OS argument requires a value"
            OS="$2"
            shift 2
            ;;
        -arch|--arch)
            [ -z "$2" ] && error_exit "Architecture argument requires a value"
            ARCH="$2"
            shift 2
            ;;
        --download-only)
            DOWNLOAD_ONLY="true"
            shift
            ;;
        --no-rename)
            NO_RENAME="true"
            shift
            ;;
        --debug)
            DEBUG="true"
            shift
            ;;
        -u|--uninstall)
            UNINSTALL="true"
            if [ -n "$2" ] && [ "${2#-}" = "$2" ]; then
                UNINSTALL_DIR="$2"
                shift 2
            else
                shift
            fi
            ;;
        -h|--help)
            echo "DGF Installer v2.0"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "OPTIONS:"
            echo "  -v, --version <version>        Version to install (default: latest, e.g., 2.0.0)"
            echo "  -os, --os <os>                 Operating system: linux, darwin, windows, android"
            echo "  -arch, --arch <arch>           Architecture: amd64, arm64, arm"
            echo "  --download-only                Download only, do not install"
            echo "  --no-rename                    Keep original filename instead of renaming to 'dgf' or 'dgf.exe'"
            echo "  --debug                        Enable debug output"
            echo "  -u, --uninstall [path]         Uninstall dgf from specified path or default locations"
            echo "  -h, --help                     Show this help message"
            echo ""
            echo "EXAMPLES:"
            echo "  # Install latest version with sudo"
            echo "  sudo ./dgf-installer.sh"
            echo ""
            echo "  # Install specific version"
            echo "  sudo ./dgf-installer.sh -v 2.0.0"
            echo ""
            echo "  # Download only (no installation)"
            echo "  ./dgf-installer.sh --download-only"
            echo ""
            echo "  # Uninstall"
            echo "  sudo ./dgf-installer.sh --uninstall"
            exit 0
            ;;
        *)
            echo "Usage: $0 [-v <version>] [-os <linux|darwin|windows|android>] [-arch <amd64|arm64|arm>] [--download-only] [--no-rename] [--debug] [-u|--uninstall [path]] [-h|--help]"
            exit 1
            ;;
    esac
done

# Check for required tools
[ -z "$CURL" ] && error_exit "curl is required but not installed"
[ -z "$JQ" ] && error_exit "jq is required but not installed"

# Uninstall logic
if [ "$UNINSTALL" = "true" ]; then
    info "Uninstalling dgf..."
    RM_ERR_FILE=$(make_temp_file) || error_exit "Failed to create temp file for uninstall errors"
    TEMP_FILES+=("$RM_ERR_FILE")
    FOUND=false
    if [ -n "$UNINSTALL_DIR" ]; then
        if [ -f "$UNINSTALL_DIR/dgf" ]; then
            info "Removing dgf from $UNINSTALL_DIR"
            debug "Executing rm $UNINSTALL_DIR/dgf"
            rm "$UNINSTALL_DIR/dgf" 2> "$RM_ERR_FILE" || {
                error_exit "Failed to remove $UNINSTALL_DIR/dgf: $(cat "$RM_ERR_FILE" 2>/dev/null || echo 'Unknown error')"
            }
            info "dgf removed from $UNINSTALL_DIR"
            FOUND=true
        else
            warning "dgf not found in $UNINSTALL_DIR"
        fi
    else
        if [ -f "/usr/local/bin/dgf" ]; then
            info "Removing dgf from /usr/local/bin"
            debug "Executing sudo rm /usr/local/bin/dgf"
            sudo rm "/usr/local/bin/dgf" 2> "$RM_ERR_FILE" || {
                error_exit "Failed to remove /usr/local/bin/dgf: $(cat "$RM_ERR_FILE" 2>/dev/null || echo 'Unknown error')"
            }
            info "dgf removed from /usr/local/bin"
            FOUND=true
        fi
        if [ -f "$HOME/bin/dgf" ]; then
            info "Removing dgf from $HOME/bin"
            debug "Executing rm $HOME/bin/dgf"
            rm "$HOME/bin/dgf" 2> "$RM_ERR_FILE" || {
                error_exit "Failed to remove $HOME/bin/dgf: $(cat "$RM_ERR_FILE" 2>/dev/null || echo 'Unknown error')"
            }
            info "dgf removed from $HOME/bin"
            FOUND=true
        fi
    fi
    if [ "$FOUND" = "false" ]; then
        warning "dgf not found in specified or default locations"
    else
        info "dgf uninstalled successfully!"
    fi
    exit 0
fi

# Check write permissions in current directory
[ -w . ] || error_exit "No write permission in the current directory"

# Detect OS if not specified
if [ -z "$OS" ]; then
    UNAME_S=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$UNAME_S" in
        linux*)   OS="linux" ;;
        darwin*)  OS="darwin" ;;
        msys*|mingw*|cygwin*) OS="windows" ;;
        android*) OS="android" ;;
        *) error_exit "Unsupported OS: $UNAME_S"
    esac
fi

# Detect architecture if not specified
if [ -z "$ARCH" ]; then
    UNAME_M=$(uname -m)
    case "$UNAME_M" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv7l|armhf) ARCH="arm" ;;
        *) error_exit "Unsupported architecture: $UNAME_M"
    esac
fi

# Log detected OS and architecture
info "Detected OS: $OS, Architecture: $ARCH"
debug "Raw uname -s: $(uname -s), uname -m: $(uname -m)"

# Determine installation directory
INSTALL_DIR="/usr/local/bin"
if [ "$DOWNLOAD_ONLY" = "false" ] && [ "$OS" != "windows" ]; then
    if [ "$(id -u)" -ne 0 ]; then
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR" || error_exit "Failed to create $INSTALL_DIR"
        warning "Installing to $INSTALL_DIR instead of /usr/local/bin due to permissions"
        export PATH="$PATH:$INSTALL_DIR"
        info "Added $INSTALL_DIR to PATH for this session. Add 'export PATH=\$PATH:$INSTALL_DIR' to ~/.bashrc for persistence."
    fi
fi

# Check if dgf is installed and get its version
INSTALLED_VERSION=""
INSTALLED_PATH=""
if command -v dgf >/dev/null; then
    INSTALLED_PATH=$(command -v dgf)
    INSTALLED_VERSION=$(dgf --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "")
    debug "Found dgf at $INSTALLED_PATH with version: $INSTALLED_VERSION"
elif [ -f "/usr/local/bin/dgf" ]; then
    INSTALLED_PATH="/usr/local/bin/dgf"
    INSTALLED_VERSION=$(/usr/local/bin/dgf --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "")
    debug "Found dgf at $INSTALLED_PATH with version: $INSTALLED_VERSION"
elif [ -f "$HOME/bin/dgf" ]; then
    INSTALLED_PATH="$HOME/bin/dgf"
    INSTALLED_VERSION=$("$HOME/bin/dgf" --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "")
    debug "Found dgf at $INSTALLED_PATH with version: $INSTALLED_VERSION"
fi

# Fetch latest release tag
LATEST_VERSION=""
LATEST_RELEASE_FILE=$(make_temp_file) || error_exit "Failed to create temp file for latest release metadata"
TEMP_FILES+=("$LATEST_RELEASE_FILE")
if fetch_with_retries "$API_LATEST_URL" "$LATEST_RELEASE_FILE"; then
    LATEST_VERSION=$(jq -r '.tag_name' "$LATEST_RELEASE_FILE" 2>/dev/null || warning "Failed to extract latest release tag")
    [ -z "$LATEST_VERSION" ] && error_exit "Failed to fetch latest release tag"
    debug "Latest release tag from GitHub: $LATEST_VERSION"
else
    error_exit "Failed to fetch latest release after $MAX_RETRIES attempts"
fi

# Check for version update if dgf is installed and no specific version is requested
if [ -n "$INSTALLED_VERSION" ] && [ "$VERSION" = "latest" ]; then
    info "Installed dgf version: $INSTALLED_VERSION, Latest version: $LATEST_VERSION"
    compare_versions "$INSTALLED_VERSION" "$LATEST_VERSION"
    VERSION_COMPARE_RESULT=$?
    if [ "$VERSION_COMPARE_RESULT" -eq 0 ]; then
        info "dgf is up to date (version $INSTALLED_VERSION)"
        [ "$DOWNLOAD_ONLY" = "false" ] && exit 0
    elif [ "$VERSION_COMPARE_RESULT" -eq 2 ]; then
        warning "Installed version ($INSTALLED_VERSION) is newer than latest release ($LATEST_VERSION)"
        [ "$DOWNLOAD_ONLY" = "false" ] && exit 0
    else
        info "Updating dgf from $INSTALLED_VERSION to $LATEST_VERSION"
        VERSION="$LATEST_VERSION"
    fi
elif [ -n "$INSTALLED_VERSION" ]; then
    info "dgf is installed (version $INSTALLED_VERSION) at $INSTALLED_PATH"
    if [ "$VERSION" != "latest" ]; then
        compare_versions "$INSTALLED_VERSION" "$VERSION"
        VERSION_COMPARE_RESULT=$?
        if [ "$VERSION_COMPARE_RESULT" -eq 0 ]; then
            info "Requested version ($VERSION) matches installed version"
            [ "$DOWNLOAD_ONLY" = "false" ] && exit 0
        elif [ "$VERSION_COMPARE_RESULT" -eq 2 ]; then
            warning "Installed version ($INSTALLED_VERSION) is newer than requested version ($VERSION)"
            [ "$DOWNLOAD_ONLY" = "false" ] && exit 0
        else
            info "Updating dgf from $INSTALLED_VERSION to $VERSION"
        fi
    fi
fi

# If version is still "latest", set it to the fetched latest version
[ "$VERSION" = "latest" ] && VERSION="$LATEST_VERSION"

# Fetch GitHub API releases data for asset details
info "Fetching releases from $API_URL"
RELEASES_FILE=$(make_temp_file) || error_exit "Failed to create temp file for releases payload"
TEMP_FILES+=("$RELEASES_FILE")
if ! fetch_with_retries "$API_URL" "$RELEASES_FILE"; then
    warning "Failed to fetch GitHub API releases after $MAX_RETRIES attempts, proceeding with metadata.json only"
    RELEASES=""
else
    RELEASES=$(cat "$RELEASES_FILE")
    [ -z "$RELEASES" ] && warning "Releases data is empty, proceeding with metadata.json only"
fi

# Determine metadata URL (v2.0 uses versioned metadata file)
METADATA_URL="${BASE_URL}/download/${VERSION}/metadata-${VERSION}.json"
debug "Metadata URL: $METADATA_URL"

# Fetch metadata with HTTP status check
info "Fetching metadata from $METADATA_URL"
METADATA_FILE=$(make_temp_file) || error_exit "Failed to create temp file for metadata payload"
TEMP_FILES+=("$METADATA_FILE")
if ! fetch_with_retries "$METADATA_URL" "$METADATA_FILE"; then
    error_exit "Failed to fetch metadata after $MAX_RETRIES attempts: $(cat "$LAST_CURL_ERROR" 2>/dev/null || echo 'Unknown error')"
fi
METADATA=$(cat "$METADATA_FILE")
[ -z "$METADATA" ] && error_exit "Metadata is empty or invalid"
debug "Metadata content: $METADATA"

# Extract filename and size from metadata.json
FILENAME=$(echo "$METADATA" | jq -r --arg os "$OS" --arg arch "$ARCH" \
    '.[] | select(.goos == $os and .goarch == $arch) | .filename' 2>/dev/null || error_exit "Failed to extract filename from metadata")
[ -z "$FILENAME" ] && error_exit "No matching binary for OS=$OS and ARCH=$ARCH"
FILESIZE=$(echo "$METADATA" | jq -r --arg os "$OS" --arg arch "$ARCH" \
    '.[] | select(.goos == $os and .goarch == $arch) | .size_bytes' 2>/dev/null || echo "null")
debug "Extracted filename: $FILENAME, metadata size: $FILESIZE bytes"

# Extract SHA256 from metadata (v2.0+)
METADATA_SHA256=$(echo "$METADATA" | jq -r --arg os "$OS" --arg arch "$ARCH" \
    '.[] | select(.goos == $os and .goarch == $arch) | .sha256' 2>/dev/null || echo "null")

# Fetch size from GitHub API as backup
API_SIZE="null"
API_SHA256="null"
if [ -n "$RELEASES" ]; then
    API_SIZE=$(echo "$RELEASES" | jq -r --arg fname "$FILENAME" \
        '.[0].assets[] | select(.name == $fname) | .size' 2>/dev/null || echo "null")
fi

# Prefer metadata SHA256, fallback to API if available
if [ "$METADATA_SHA256" != "null" ] && [ -n "$METADATA_SHA256" ]; then
    API_SHA256="$METADATA_SHA256"
    debug "Using SHA256 from metadata: $API_SHA256"
elif [ -n "$RELEASES" ]; then
    API_SHA256=$(echo "$RELEASES" | jq -r --arg fname "$FILENAME" \
        '.[0].assets[] | select(.name == $fname) | .digest | ltrimstr("sha256:")' 2>/dev/null || echo "null")
    debug "Falling back to GitHub API digest: $API_SHA256"
fi
debug "GitHub API size: $API_SIZE bytes, SHA256: $API_SHA256"

# Determine download URL
DOWNLOAD_URL="${BASE_URL}/download/${VERSION}/${FILENAME}"
debug "Download URL: $DOWNLOAD_URL"

# Download the binary
info "Downloading $FILENAME from $DOWNLOAD_URL"
debug "Executing fetch_with_retries \"$DOWNLOAD_URL\" \"$FILENAME\""
if ! fetch_with_retries "$DOWNLOAD_URL" "$FILENAME"; then
    debug "Download failed. Curl error output:"
    debug "$(cat "$LAST_CURL_ERROR" 2>/dev/null || echo 'No error log available')"
    error_exit "Failed to download binary after $MAX_RETRIES attempts: $(cat "$LAST_CURL_ERROR" 2>/dev/null || echo 'Unknown error')"
fi
[ -f "$FILENAME" ] || error_exit "Downloaded file $FILENAME does not exist"

# Verify file size (warn on mismatch)
ACTUAL_SIZE=$(stat -c %s "$FILENAME" 2>/dev/null || stat -f %z "$FILENAME" 2>/dev/null || error_exit "Failed to get file size")
debug "Actual file size: $ACTUAL_SIZE bytes"
if [ -n "$API_SIZE" ] && [ "$API_SIZE" != "null" ]; then
    [ "$ACTUAL_SIZE" -eq "$API_SIZE" ] || warning "File size mismatch: expected $API_SIZE bytes (GitHub API), got $ACTUAL_SIZE bytes"
elif [ -n "$FILESIZE" ] && [ "$FILESIZE" != "null" ]; then
    [ "$ACTUAL_SIZE" -eq "$FILESIZE" ] || warning "File size mismatch: expected $FILESIZE bytes (metadata.json), got $ACTUAL_SIZE bytes"
else
    warning "No size information available for validation"
fi

# Verify SHA256 checksum if available
if [ -n "$API_SHA256" ] && [ "$API_SHA256" != "null" ]; then
    debug "Verifying SHA256 checksum: $API_SHA256"
    SHA_ERR_FILE=$(make_temp_file) || error_exit "Failed to create temp file for checksum validation"
    TEMP_FILES+=("$SHA_ERR_FILE")
    if command -v sha256sum >/dev/null; then
        echo "$API_SHA256 $FILENAME" | sha256sum -c 2> "$SHA_ERR_FILE" || {
            error_exit "Checksum validation failed: $(cat "$SHA_ERR_FILE" 2>/dev/null || echo 'Unknown error')"
        }
    elif command -v shasum >/dev/null; then
        echo "$API_SHA256 $FILENAME" | shasum -a 256 -c 2> "$SHA_ERR_FILE" || {
            error_exit "Checksum validation failed: $(cat "$SHA_ERR_FILE" 2>/dev/null || echo 'Unknown error')"
        }
    else
        error_exit "No sha256sum or shasum found; cannot validate binary integrity"
    fi
else
    error_exit "No SHA256 checksum available for this asset; aborting for safety"
fi

# Handle renaming
TARGET_NAME="$FILENAME"
if [ "$NO_RENAME" = "false" ]; then
    if [ "$OS" = "windows" ]; then
        TARGET_NAME="dgf.exe"
    else
        TARGET_NAME="dgf"
    fi
    [ -f "$TARGET_NAME" ] && error_exit "Target file $TARGET_NAME already exists"
    info "Renaming $FILENAME to $TARGET_NAME"
    debug "Executing mv \"$FILENAME\" \"$TARGET_NAME\""
    MV_ERR_FILE=$(make_temp_file) || error_exit "Failed to create temp file for rename errors"
    TEMP_FILES+=("$MV_ERR_FILE")
    mv "$FILENAME" "$TARGET_NAME" 2> "$MV_ERR_FILE" || {
        error_exit "Failed to rename file: $(cat "$MV_ERR_FILE" 2>/dev/null || echo 'Unknown error')"
    }
fi

# If download-only, exit here
if [ "$DOWNLOAD_ONLY" = "true" ]; then
    info "Binary downloaded as $TARGET_NAME in current directory"
    exit 0
fi

# Install the binary
case "$OS" in
    linux|android)
        info "Installing $TARGET_NAME to $INSTALL_DIR"
        debug "Executing mv \"$TARGET_NAME\" $INSTALL_DIR/"
        INSTALL_ERR_FILE=$(make_temp_file) || error_exit "Failed to create temp file for install errors"
        TEMP_FILES+=("$INSTALL_ERR_FILE")
        mv "$TARGET_NAME" "$INSTALL_DIR/" 2> "$INSTALL_ERR_FILE" || {
            error_exit "Failed to move binary to $INSTALL_DIR: $(cat "$INSTALL_ERR_FILE" 2>/dev/null || echo 'Unknown error')"
        }
        debug "Setting executable permissions on $INSTALL_DIR/$TARGET_NAME"
        chmod +x "$INSTALL_DIR/$TARGET_NAME" 2> "$INSTALL_ERR_FILE" || {
            error_exit "Failed to set executable permissions: $(cat "$INSTALL_ERR_FILE" 2>/dev/null || echo 'Unknown error')"
        }
        ;;
    windows)
        info "Installing $TARGET_NAME to C:\\Program Files\\dgf"
        debug "Creating directory C:\\Program Files\\dgf"
        mkdir -p "C:\\Program Files\\dgf" || error_exit "Failed to create directory (ensure you have admin privileges)"
        debug "Executing mv \"$TARGET_NAME\" \"C:\\Program Files\\dgf\\$TARGET_NAME\""
        INSTALL_ERR_FILE=$(make_temp_file) || error_exit "Failed to create temp file for install errors"
        TEMP_FILES+=("$INSTALL_ERR_FILE")
        mv "$TARGET_NAME" "C:\\Program Files\\dgf\\$TARGET_NAME" 2> "$INSTALL_ERR_FILE" || {
            error_exit "Failed to move binary: $(cat "$INSTALL_ERR_FILE" 2>/dev/null || echo 'Unknown error')"
        }
        info "Please add 'C:\\Program Files\\dgf' to your system PATH manually."
        ;;
    darwin)
        info "Installing $TARGET_NAME to $INSTALL_DIR"
        debug "Executing mv \"$TARGET_NAME\" $INSTALL_DIR/"
        INSTALL_ERR_FILE=$(make_temp_file) || error_exit "Failed to create temp file for install errors"
        TEMP_FILES+=("$INSTALL_ERR_FILE")
        mv "$TARGET_NAME" "$INSTALL_DIR/" 2> "$INSTALL_ERR_FILE" || {
            error_exit "Failed to move binary to $INSTALL_DIR: $(cat "$INSTALL_ERR_FILE" 2>/dev/null || echo 'Unknown error')"
        }
        debug "Setting executable permissions on $INSTALL_DIR/$TARGET_NAME"
        chmod +x "$INSTALL_DIR/$TARGET_NAME" 2> "$INSTALL_ERR_FILE" || {
            error_exit "Failed to set executable permissions: $(cat "$INSTALL_ERR_FILE" 2>/dev/null || echo 'Unknown error')"
        }
        ;;
    *)
        error_exit "Unsupported OS for installation: $OS"
        ;;
esac

# Verify binary is in PATH
if ! command -v "$TARGET_NAME" >/dev/null; then
    warning "$TARGET_NAME installed to $INSTALL_DIR but not found in PATH. Add 'export PATH=\$PATH:$INSTALL_DIR' to ~/.bashrc and run 'source ~/.bashrc'."
fi

info "$TARGET_NAME installed successfully!"
