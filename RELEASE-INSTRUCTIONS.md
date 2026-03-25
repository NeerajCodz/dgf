# DGF v2.0.0 Release Instructions

**This document provides step-by-step instructions for completing the GitHub release of v2.0.0**

---

## ✅ Pre-Release Verification Checklist

The following items have been verified:

- [x] **CHANGELOG.md** updated with v2.0.0 section (date: 2026-03-26)
- [x] **Release notes** prepared in `v2.0.0-RELEASE-NOTES.md`
- [x] **Build artifacts** ready (see "Build Information" section)
- [x] **Build scripts** updated:
  - `build/build.sh` - Bash build script with metadata generation
  - `build/build.ps1` - PowerShell build script with metadata generation
- [x] **Installer script** updated (`dgf-installer.sh`) - Compatible with v2.0.0 artifacts
- [x] **Documentation** updated:
  - `README.md` - v2.0 features and usage
  - `CONTRIBUTING.md` - Contributor guidelines
  - Migration guide for v1 → v2 transition
- [x] **Integration tests** passing (`test: validate core v2.0 integration workflows`)
- [x] **All commits** on beta-2.0 branch include release-ready code

---

## 📦 Build Information

### To Build Binaries Locally (Optional)

Before creating the GitHub release, binaries should be built and checksummed.

#### On Linux/macOS:
```bash
cd E:/codz/Projects/dgf/dgf
chmod +x build/build.sh
./build/build.sh
```

#### On Windows (with Go installed):
```powershell
cd E:\codz\Projects\dgf\dgf
powershell -ExecutionPolicy Bypass -File .\build\build.ps1
```

This generates:
- 8 cross-platform binaries in `build/` directory
- `build/metadata-2.0.0.json` with checksums and metadata

### Binary Targets
- `dgf-2.0.0-linux-amd64`
- `dgf-2.0.0-linux-arm64`
- `dgf-2.0.0-linux-arm`
- `dgf-2.0.0-darwin-amd64`
- `dgf-2.0.0-darwin-arm64`
- `dgf-2.0.0-windows-amd64.exe`
- `dgf-2.0.0-windows-arm64.exe`
- `dgf-2.0.0-android-arm64`

---

## 🚀 Creating the GitHub Release

### Option 1: Using GitHub CLI (gh) - RECOMMENDED

Requires: `gh` CLI installed and authenticated to GitHub

```bash
# Navigate to repository
cd E:\codz\Projects\dgf\dgf

# Create release using the release notes
gh release create v2.0.0 \
  --title "DGF v2.0.0 - Interactive TUI Release" \
  --notes-file v2.0.0-RELEASE-NOTES.md \
  --draft false \
  build/dgf-2.0.0-*
```

**If metadata.json should also be included:**
```bash
gh release create v2.0.0 \
  --title "DGF v2.0.0 - Interactive TUI Release" \
  --notes-file v2.0.0-RELEASE-NOTES.md \
  --draft false \
  build/dgf-2.0.0-* \
  build/metadata-2.0.0.json
```

### Option 2: Using GitHub Web UI

1. Go to: https://github.com/NeerajCodz/dgf/releases
2. Click **"Draft a new release"**
3. Fill in release details:
   - **Tag version:** `v2.0.0`
   - **Release title:** `DGF v2.0.0 - Interactive TUI Release`
   - **Description:** Copy content from `v2.0.0-RELEASE-NOTES.md`
4. Upload binary artifacts:
   - Click **"Attach binaries by dropping them here or selecting them"**
   - Upload all files from `build/` directory:
     - All 8 `dgf-2.0.0-*` binaries
     - `metadata-2.0.0.json`
5. Click **"Publish release"**

### Option 3: Using GitHub API

```bash
# Requires: jq (JSON processor) and curl

TOKEN="your_github_token"  # Set your GitHub personal access token
REPO="NeerajCodz/dgf"
TAG="v2.0.0"
TITLE="DGF v2.0.0 - Interactive TUI Release"
RELEASE_NOTES_FILE="v2.0.0-RELEASE-NOTES.md"

# Read release notes
BODY=$(cat "$RELEASE_NOTES_FILE")

# Create release
RESPONSE=$(curl -s -X POST \
  -H "Authorization: token $TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  "https://api.github.com/repos/$REPO/releases" \
  -d @- <<EOF
{
  "tag_name": "$TAG",
  "target_commitish": "beta-2.0",
  "name": "$TITLE",
  "body": "$(echo "$BODY" | jq -Rs .)",
  "draft": false,
  "prerelease": false
}
EOF
)

# Extract upload URL
UPLOAD_URL=$(echo $RESPONSE | jq -r '.upload_url' | sed 's/{?name,label}//')

# Upload binaries
for FILE in build/dgf-2.0.0-* build/metadata-2.0.0.json; do
    [ -f "$FILE" ] && curl -s -X POST \
      -H "Authorization: token $TOKEN" \
      -H "Content-Type: application/octet-stream" \
      "${UPLOAD_URL}?name=$(basename $FILE)" \
      --data-binary @"$FILE"
done

echo "Release created successfully!"
```

---

## 📋 Post-Release Tasks

After creating the GitHub release:

1. **Verify Release** - Check at https://github.com/NeerajCodz/dgf/releases/tag/v2.0.0
2. **Update Main Branch** (when ready for production):
   ```bash
   git checkout main
   git merge beta-2.0 --ff-only
   git push origin main
   ```
3. **Announce** - Share release on social media, forums, or community channels
4. **Update Installer** - Ensure `dgf-installer.sh` can fetch v2.0.0 correctly
5. **Monitor Issues** - Watch for bug reports and user feedback

---

## 🔍 Verification Checklist

After release is created, verify:

- [ ] Release appears on GitHub releases page
- [ ] All 8 binaries are downloadable
- [ ] `metadata-2.0.0.json` contains correct checksums
- [ ] Release notes are properly formatted and visible
- [ ] Release is marked as "Latest release"
- [ ] Tag `v2.0.0` exists in repository
- [ ] Installer script can fetch and install the release

---

## 🚨 If Release Creation Fails

### Common Issues & Solutions

**Issue: `gh` CLI not installed**
```bash
# Install gh CLI
# macOS: brew install gh
# Linux: See https://github.com/cli/cli/blob/trunk/docs/install.md
# Windows: choco install gh
```

**Issue: Not authenticated with GitHub**
```bash
gh auth login
# Follow prompts to authenticate
```

**Issue: Binary files not found**
```bash
# Ensure binaries are built first
cd build/
./build.sh  # or build.ps1 on Windows
```

**Issue: Release already exists**
```bash
# Delete existing release (via GitHub UI) or
gh release delete v2.0.0 -y  # then recreate
```

### Manual Verification

If automated release fails, verify files exist:
```bash
ls -la build/
# Should see: dgf-2.0.0-linux-amd64, dgf-2.0.0-windows-amd64.exe, etc.
# Should see: metadata-2.0.0.json

# Verify checksums in metadata
cat build/metadata-2.0.0.json | jq '.'
```

---

## 📦 What's Included in v2.0.0 Release

### Binaries (8 platforms)
- Linux (amd64, arm64, arm)
- macOS (amd64, arm64)
- Windows (amd64, arm64)
- Android (arm64)

### Documentation
- Complete release notes with features and improvements
- Installation and update instructions
- Token support documentation
- Migration guide from v1.x to v2.0

### Metadata
- `metadata-2.0.0.json` with SHA256 checksums for integrity verification

---

## 📝 Related Files

- **Release Notes:** `v2.0.0-RELEASE-NOTES.md`
- **Changelog:** `CHANGELOG.md`
- **Build Scripts:** `build/build.sh`, `build/build.ps1`
- **Installer:** `dgf-installer.sh`
- **README:** `README.md`

---

## 🔗 Resources

- GitHub Releases API: https://docs.github.com/en/rest/releases
- gh CLI Documentation: https://cli.github.com/manual/
- Repository: https://github.com/NeerajCodz/dgf
- Issues: https://github.com/NeerajCodz/dgf/issues

---

**Status:** ✅ Ready for Release  
**Version:** 2.0.0  
**Date:** March 26, 2026  
**Branch:** beta-2.0  
**Commit:** 009d4a3 (or latest on beta-2.0)

