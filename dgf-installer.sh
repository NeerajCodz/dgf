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

# Function to clean up temporary files
cleanup() {
    debug "Cleaning up temporary files"
    for file in "${TEMP_FILES[@]}"; do
        [ -f "$file" ] && rm -f "$file" && debug "Removed $file"
    done
}

# Trap to ensure cleanup on script exit
trap cleanup EXIT

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

# Function to fetch with retries
fetch_with_retries() {
    local url=$1
    local output=$2
    local attempt=1
    TEMP_FILES+=("$output")
    while [ $attempt -le $MAX_RETRIES ]; do
        debug "Attempt $attempt: curl -sL -w \"%{http_code}\" \"$url\" -o \"$output\""
        HTTP_STATUS=$(curl -sL -w "%{http_code}" "$url" -o "$output" 2> "/tmp/curl_error_$$.log" || echo "000")
        if [ "$HTTP_STATUS" = "200" ]; then
            TEMP_FILES+=("/tmp/curl_error_$$.log")
            return 0
        fi
        warning "Attempt $attempt failed with HTTP $HTTP_STATUS: $(cat "/tmp/curl_error_$$.log" 2>/dev/null || echo 'Unknown error')"
        TEMP_FILES+=("/tmp/curl_error_$$.log")
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
        *)
            echo "Usage: $0 [-v <version>] [-os <linux|darwin|windows|android>] [-arch <amd64|arm64|arm>] [--download-only] [--no-rename] [--debug] [-u|--uninstall [path]]"
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
    FOUND=false
    if [ -n "$UNINSTALL_DIR" ]; then
        if [ -f "$UNINSTALL_DIR/dgf" ]; then
            info "Removing dgf from $UNINSTALL_DIR"
            debug "Executing rm $UNINSTALL_DIR/dgf"
            rm "$UNINSTALL_DIR/dgf" 2> "/tmp/rm_error_$$.log" || {
                error_exit "Failed to remove $UNINSTALL_DIR/dgf: $(cat "/tmp/rm_error_$$.log" 2>/dev/null || echo 'Unknown error')"
            }
            TEMP_FILES+=("/tmp/rm_error_$$.log")
            info "dgf removed from $UNINSTALL_DIR"
            FOUND=true
        else
            warning "dgf not found in $UNINSTALL_DIR"
        fi
    else
        if [ -f "/usr/local/bin/dgf" ]; then
            info "Removing dgf from /usr/local/bin"
            debug "Executing sudo rm /usr/local/bin/dgf"
            sudo rm "/usr/local/bin/dgf" 2> "/tmp/rm_error_$$.log" || {
                error_exit "Failed to remove /usr/local/bin/dgf: $(cat "/tmp/rm_error_$$.log" 2>/dev/null || echo 'Unknown error')"
            }
            TEMP_FILES+=("/tmp/rm_error_$$.log")
            info "dgf removed from /usr/local/bin"
            FOUND=true
        fi
        if [ -f "$HOME/bin/dgf" ]; then
            info "Removing dgf from $HOME/bin"
            debug "Executing rm $HOME/bin/dgf"
            rm "$HOME/bin/dgf" 2> "/tmp/rm_error_$$.log" || {
                error_exit "Failed to remove $HOME/bin/dgf: $(cat "/tmp/rm_error_$$.log" 2>/dev/null || echo 'Unknown error')"
            }
            TEMP_FILES+=("/tmp/rm_error_$$.log")
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
    [ "$(id -u)" -ne 0 ] && error_exit "Root privileges required for installation to /usr/local/bin. Use sudo."
    if [ ! -w "$INSTALL_DIR" ] || [ ! -d "$INSTALL_DIR" ]; then
        INSTALL_DIR="$HOME/bin"
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
if fetch_with_retries "$API_LATEST_URL" "/tmp/latest_release_$$.json"; then
    LATEST_VERSION=$(cat "/tmp/latest_release_$$.json" | jq -r '.tag_name' 2>/dev/null || warning "Failed to extract latest release tag")
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
if ! fetch_with_retries "$API_URL" "/tmp/releases_$$.json"; then
    warning "Failed to fetch GitHub API releases after $MAX_RETRIES attempts, proceeding with metadata.json only"
    RELEASES=""
else
    RELEASES=$(cat "/tmp/releases_$$.json")
    [ -z "$RELEASES" ] && warning "Releases data is empty, proceeding with metadata.json only"
fi

# Determine metadata URL
METADATA_URL="${BASE_URL}/download/${VERSION}/metadata.json"
debug "Metadata URL: $METADATA_URL"

# Fetch metadata with HTTP status check
info "Fetching metadata from $METADATA_URL"
if ! fetch_with_retries "$METADATA_URL" "/tmp/metadata_$$.json"; then
    error_exit "Failed to fetch metadata after $MAX_RETRIES attempts: $(cat "/tmp/curl_error_$$.log" 2>/dev/null || echo 'Unknown error')"
fi
METADATA=$(cat "/tmp/metadata_$$.json")
[ -z "$METADATA" ] && error_exit "Metadata is empty or invalid"
debug "Metadata content: $METADATA"

# Extract filename and size from metadata.json
FILENAME=$(echo "$METADATA" | jq -r --arg os "$OS" --arg arch "$ARCH" \
    '.[] | select(.goos == $os and .goarch == $arch) | .filename' 2>/dev/null || error_exit "Failed to extract filename from metadata")
[ -z "$FILENAME" ] && error_exit "No matching binary for OS=$OS and ARCH=$ARCH"
FILESIZE=$(echo "$METADATA" | jq -r --arg os "$OS" --arg arch "$ARCH" \
    '.[] | select(.goos == $os and .goarch == $arch) | .size_bytes' 2>/dev/null || echo "null")
debug "Extracted filename: $FILENAME, metadata size: $FILESIZE bytes"

# Fetch size and SHA256 from GitHub API
API_SIZE="null"
API_SHA256="null"
if [ -n "$RELEASES" ]; then
    API_SIZE=$(echo "$RELEASES" | jq -r --arg fname "$FILENAME" \
        '.[0].assets[] | select(.name == $fname) | .size' 2>/dev/null || echo "null")
    API_SHA256=$(echo "$RELEASES" | jq -r --arg fname "$FILENAME" \
        '.[0].assets[] | select(.name == $fname) | .digest | ltrimstr("sha256:")' 2>/dev/null || echo "null")
fi
debug "GitHub API size: $API_SIZE bytes, SHA256: $API_SHA256"

# Determine download URL
DOWNLOAD_URL="${BASE_URL}/download/${VERSION}/${FILENAME}"
debug "Download URL: $DOWNLOAD_URL"

# Download the binary
info "Downloading $FILENAME from $DOWNLOAD_URL"
debug "Executing fetch_with_retries \"$DOWNLOAD_URL\" \"$FILENAME\""
if ! fetch_with_retries "$DOWNLOAD_URL" "$FILENAME"; then
    debug "Download failed. Contents of /tmp/curl_error_$$.log:"
    debug "$(cat "/tmp/curl_error_$$.log" 2>/dev/null || echo 'No error log available')"
    error_exit "Failed to download binary after $MAX_RETRIES attempts: $(cat "/tmp/curl_error_$$.log" 2>/dev/null || echo 'Unknown error')"
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
    if command -v sha256sum >/dev/null; then
        echo "$API_SHA256 $FILENAME" | sha256sum -c 2> "/tmp/sha_error_$$.log" || {
            error_exit "Checksum validation failed: $(cat "/tmp/sha_error_$$.log" 2>/dev/null || echo 'Unknown error')"
        }
        TEMP_FILES+=("/tmp/sha_error_$$.log")
    elif command -v shasum >/dev/null; then
        echo "$API_SHA256 $FILENAME" | shasum -a 256 -c 2> "/tmp/sha_error_$$.log" || {
            error_exit "Checksum validation failed: $(cat "/tmp/sha_error_$$.log" 2>/dev/null || echo 'Unknown error')"
        }
        TEMP_FILES+=("/tmp/sha_error_$$.log")
    else
        warning "No sha256sum or shasum found, skipping checksum validation"
    fi
else
    warning "No SHA256 checksum provided in GitHub API, skipping validation"
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
    mv "$FILENAME" "$TARGET_NAME" 2> "/tmp/mv_error_$$.log" || {
        error_exit "Failed to rename file: $(cat "/tmp/mv_error_$$.log" 2>/dev/null || echo 'Unknown error')"
    }
    TEMP_FILES+=("/tmp/mv_error_$$.log")
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
        mv "$TARGET_NAME" "$INSTALL_DIR/" 2> "/tmp/mv_error_$$.log" || {
            error_exit "Failed to move binary to $INSTALL_DIR: $(cat "/tmp/mv_error_$$.log" 2>/dev/null || echo 'Unknown error')"
        }
        TEMP_FILES+=("/tmp/mv_error_$$.log")
        debug "Setting executable permissions on $INSTALL_DIR/$TARGET_NAME"
        chmod +x "$INSTALL_DIR/$TARGET_NAME" 2> "/tmp/chmod_error_$$.log" || {
            error_exit "Failed to set executable permissions: $(cat "/tmp/chmod_error_$$.log" 2>/dev/null || echo 'Unknown error')"
        }
        TEMP_FILES+=("/tmp/chmod_error_$$.log")
        ;;
    windows)
        info "Installing $TARGET_NAME to C:\\Program Files\\dgf"
        debug "Creating directory C:\\Program Files\\dgf"
        mkdir -p "C:\\Program Files\\dgf" || error_exit "Failed to create directory (ensure you have admin privileges)"
        debug "Executing mv \"$TARGET_NAME\" \"C:\\Program Files\\dgf\\$TARGET_NAME\""
        mv "$TARGET_NAME" "C:\\Program Files\\dgf\\$TARGET_NAME" 2> "/tmp/mv_error_$$.log" || {
            error_exit "Failed to move binary: $(cat "/tmp/mv_error_$$.log" 2>/dev/null || echo 'Unknown error')"
        }
        TEMP_FILES+=("/tmp/mv_error_$$.log")
        info "Please add 'C:\\Program Files\\dgf' to your system PATH manually."
        ;;
    darwin)
        info "Installing $TARGET_NAME to $INSTALL_DIR"
        debug "Executing mv \"$TARGET_NAME\" $INSTALL_DIR/"
        mv "$TARGET_NAME" "$INSTALL_DIR/" 2> "/tmp/mv_error_$$.log" || {
            error_exit "Failed to move binary to $INSTALL_DIR: $(cat "/tmp/mv_error_$$.log" 2>/dev/null || echo 'Unknown error')"
        }
        TEMP_FILES+=("/tmp/mv_error_$$.log")
        debug "Setting executable permissions on $INSTALL_DIR/$TARGET_NAME"
        chmod +x "$INSTALL_DIR/$TARGET_NAME" 2> "/tmp/chmod_error_$$.log" || {
            error_exit "Failed to set executable permissions: $(cat "/tmp/chmod_error_$$.log" 2>/dev/null || echo 'Unknown error')"
        }
        TEMP_FILES+=("/tmp/chmod_error_$$.log")
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