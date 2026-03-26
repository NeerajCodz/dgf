
#!/usr/bin/env bash
set -e

# Navigate to project root
cd "$(dirname "$0")/.."

PROJECT_NAME="dgf"
VERSION="2.0.0"
OUT_DIR="./build/v2.0"
mkdir -p "${OUT_DIR}"

# Replace previous artifacts for this version
rm -f "${OUT_DIR}"/*

# Pure Go binaries
export CGO_ENABLED=0

TARGETS=(
  "linux amd64"
  "linux arm64"
  "linux arm"
  "darwin amd64"
  "darwin arm64"
  "windows amd64"
  "windows arm64"
  "android arm64"
)

# Prepare metadata JSON array
METADATA="["

for target in "${TARGETS[@]}"; do
  set -- $target
  GOOS=$1
  GOARCH=$2
  OUTPUT_NAME="${PROJECT_NAME}-${VERSION}-${GOOS}-${GOARCH}"
  if [ "$GOOS" = "windows" ]; then
    OUTPUT_NAME="${OUTPUT_NAME}.exe"
  fi

  echo "→ Building ${OUTPUT_NAME}"
  env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "${OUT_DIR}/${OUTPUT_NAME}" ./cmd/dgf/

  # Gather file size (compatible with both Linux and macOS)
  if [[ "$OSTYPE" == "darwin"* ]]; then
    FILE_SIZE=$(stat -f%z "${OUT_DIR}/${OUTPUT_NAME}")
  else
    FILE_SIZE=$(stat -c%s "${OUT_DIR}/${OUTPUT_NAME}")
  fi

  # Compute SHA256
  if command -v sha256sum &> /dev/null; then
    SHA256=$(sha256sum "${OUT_DIR}/${OUTPUT_NAME}" | awk '{print $1}')
  elif command -v shasum &> /dev/null; then
    SHA256=$(shasum -a 256 "${OUT_DIR}/${OUTPUT_NAME}" | awk '{print $1}')
  else
    SHA256=""
  fi

  # Append to JSON
  METADATA="${METADATA}{\"filename\":\"${OUTPUT_NAME}\",\"goos\":\"${GOOS}\",\"goarch\":\"${GOARCH}\",\"size_bytes\":${FILE_SIZE},\"sha256\":\"${SHA256}\"},"
done

# Remove last comma and close JSON
METADATA=${METADATA%,}
METADATA="${METADATA}]"

# Save to file
echo "${METADATA}" > "${OUT_DIR}/metadata-${VERSION}.json"

echo "✅ All binaries built into ${OUT_DIR}"
echo "✅ Metadata saved to ${OUT_DIR}/metadata-${VERSION}.json"
