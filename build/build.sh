#!/usr/bin/env bash
set -e

cd ..

PROJECT_NAME="dgf"
VERSION="1.0"
OUT_DIR="./build"
mkdir -p "${OUT_DIR}"

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
  env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "${OUT_DIR}/${OUTPUT_NAME}" .

  # Gather file size
  FILE_SIZE=$(stat -c%s "${OUT_DIR}/${OUTPUT_NAME}")

  # Append to JSON
  METADATA="${METADATA}{\"filename\":\"${OUTPUT_NAME}\",\"goos\":\"${GOOS}\",\"goarch\":\"${GOARCH}\",\"size_bytes\":${FILE_SIZE}},"
done

# Remove last comma and close JSON
METADATA=${METADATA%,}
METADATA="${METADATA}]"

# Save to file
echo "${METADATA}" > "${OUT_DIR}/metadata-${VERSION}.json"

echo "✅ All binaries built into ${OUT_DIR}"
echo "✅ Metadata saved to ${OUT_DIR}/metadata-${VERSION}.json"
