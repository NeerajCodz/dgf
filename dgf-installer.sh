#!/usr/bin/env bash
set -e

# Default values
VERSION="latest"
OS=""
ARCH=""
DOWNLOAD_ONLY="false"
NO_RENAME="false"
BASE_URL="https://github.com/NeerajCodz/dgf/releases"
CURL=$(command -v curl)
JQ=$(command -v jq)

# Function to print error and exit
error_exit() {
    echo "Error: $1" >&2
    exit 1
}

# Function to print info
info() {
    echo "$1"
}

# Parse command-line arguments
while [ $# -gt 0 ]; do
    case "$1" in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -os|--os)
            OS="$2"
            shift 2
            ;;
        -arch|--arch)
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
        *)
            echo "Usage: $0 [-v <version>] [-os <linux|darwin|windows|android>] [-arch <amd64|arm64|arm>] [--download-only] [--no-rename]"
            exit 1
            ;;
    esac
done

# Check for required tools
[ -z "$CURL" ] && error_exit "curl is required but not installed"
[ -z "$JQ" ] && error_exit "jq is required but not installed"

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

# Determine metadata URL
if [ "$VERSION" = "latest" ]; then
    METADATA_URL="${BASE_URL}/latest/download/metadata.json"
else
    METADATA_URL="${BASE_URL}/v${VERSION}/download/metadata.json"
fi

# Fetch metadata
info "Fetching metadata from $METADATA_URL"
METADATA=$(curl -sL "$METADATA_URL") || error_exit "Failed to fetch metadata"

# Extract filename
FILENAME=$(echo "$METADATA" | jq -r --arg os "$OS" --arg arch "$ARCH" \
    '.[] | select(.goos == $os and .goarch == $arch) | .filename')

[ -z "$FILENAME" ] && error_exit "No matching binary for OS=$OS and ARCH=$ARCH"

# Determine download URL
if [ "$VERSION" = "latest" ]; then
    DOWNLOAD_URL="${BASE_URL}/latest/download/${FILENAME}"
else
    DOWNLOAD_URL="${BASE_URL}/v${VERSION}/download/${FILENAME}"
fi

# Download the binary
info "Downloading $FILENAME from $DOWNLOAD_URL"
curl -sL -o "$FILENAME" "$DOWNLOAD_URL" || error_exit "Failed to download binary"

# Handle renaming
TARGET_NAME="$FILENAME"
if [ "$NO_RENAME" = "false" ]; then
    if [ "$OS" = "windows" ]; then
        TARGET_NAME="dgf.exe"
    else
        TARGET_NAME="dgf"
    fi
    mv "$FILENAME" "$TARGET_NAME" || error_exit "Failed to rename file"

System: rename file

# If download-only, exit here
if [ "$DOWNLOAD_ONLY" = "true" ]; then
    info "Binary downloaded as $TARGET_NAME in current directory"
    exit 0
fi

# Install the binary
case "$OS" in
    linux|android)
        info "Installing $TARGET_NAME to /usr/local/bin"
        [ "$(id -u)" -ne 0 ] && error_exit "Root privileges required for installation. Use sudo."
        mv "$TARGET_NAME" /usr/local/bin/ || error_exit "Failed to move binary to /usr/local/bin"
        chmod +x "/usr/local/bin/$TARGET_NAME" || error_exit "Failed to set executable permissions"
        ;;
    windows)
        info "Installing $TARGET_NAME to C:\\Program Files\\dgf"
        mkdir -p "C:\\Program Files\\dgf" || error_exit "Failed to create directory"
        mv "$TARGET_NAME" "C:\\Program Files\\dgf\\$TARGET_NAME" || error_exit "Failed to move binary"
        echo "Please add 'C:\\Program Files\\dgf' to your system PATH manually."
        ;;
    darwin)
        info "Installing $TARGET_NAME to /usr/local/bin"
        [ "$(id -u)" -ne 0 ] && error_exit "Root privileges required for installation. Use sudo."
        mv "$TARGET_NAME" /usr/local/bin/ || error_exit "Failed to move binary to /usr/local/bin"
        chmod +x "/usr/local/bin/$TARGET_NAME" || error_exit "Failed to set executable permissions"
        ;;
    *)
        error_exit "Unsupported OS for installation: $OS"
        ;;
esac

info "$TARGET_NAME installed successfully!"