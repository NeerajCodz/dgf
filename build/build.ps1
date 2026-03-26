# PowerShell build script for cross-platform release binaries
$ErrorActionPreference = "Stop"

# Navigate to project root
Push-Location (Split-Path -Parent $PSScriptRoot)

$PROJECT_NAME = "dgf"
$VERSION = "2.0.0"
$OUT_DIR = "./build/v2.0"

# Create output directory if it doesn't exist
if (-not (Test-Path $OUT_DIR)) {
    New-Item -ItemType Directory -Path $OUT_DIR -Force | Out-Null
}

# Replace previous artifacts for this version
Get-ChildItem -Path $OUT_DIR -File -ErrorAction SilentlyContinue | Remove-Item -Force -ErrorAction SilentlyContinue

# Pure Go binaries - set CGO_ENABLED=0
$env:CGO_ENABLED = 0

$TARGETS = @(
    @("linux", "amd64"),
    @("linux", "arm64"),
    @("linux", "arm"),
    @("darwin", "amd64"),
    @("darwin", "arm64"),
    @("windows", "amd64"),
    @("windows", "arm64"),
    @("android", "arm64")
)

# Prepare metadata array
$metadata = @()

foreach ($target in $TARGETS) {
    $GOOS = $target[0]
    $GOARCH = $target[1]
    $OUTPUT_NAME = "$PROJECT_NAME-$VERSION-$GOOS-$GOARCH"
    
    if ($GOOS -eq "windows") {
        $OUTPUT_NAME += ".exe"
    }
    
    $OUTPUT_PATH = Join-Path $OUT_DIR $OUTPUT_NAME
    
    Write-Host "→ Building $OUTPUT_NAME"
    
    # Build the binary
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    
    & go build -ldflags="-s -w" -o $OUTPUT_PATH ./cmd/dgf/
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build $OUTPUT_NAME"
        exit 1
    }
    
    # Get file size
    $file_info = Get-Item $OUTPUT_PATH
    $FILE_SIZE = $file_info.Length
    
    # Compute SHA256
    $SHA256 = (Get-FileHash -Path $OUTPUT_PATH -Algorithm SHA256).Hash.ToLower()
    
    # Add to metadata array
    $metadata += @{
        filename = $OUTPUT_NAME
        goos = $GOOS
        goarch = $GOARCH
        size_bytes = $FILE_SIZE
        sha256 = $SHA256
    }
}

# Convert to JSON and save
$metadata_json = $metadata | ConvertTo-Json
$metadata_file = Join-Path $OUT_DIR "metadata-$VERSION.json"
Set-Content -Path $metadata_file -Value $metadata_json

Write-Host "✅ All binaries built into $OUT_DIR"
Write-Host "✅ Metadata saved to $metadata_file"

Pop-Location
