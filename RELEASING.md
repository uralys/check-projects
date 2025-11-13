# Release Process

This document describes the automated release process for check-projects.

## Overview

Releases are fully automated using:
- **GoReleaser**: Builds multi-platform binaries
- **GitHub Actions**: Orchestrates the release workflow
- **GitHub Releases**: Hosts downloadable binaries

## Creating a Release

### 1. Update Version

Ensure your code is ready and tests pass:

```bash
make test
make build
```

### 2. Create and Push Tag

```bash
# Create an annotated tag
git tag -a v1.0.0 -m "Release v1.0.0: Description of changes"

# Push the tag to trigger the release
git push origin v1.0.0
```

### 3. Automatic Process

GitHub Actions will automatically:

1. **Build** binaries for:
   - macOS (amd64, arm64)
   - Linux (amd64, arm64)
   - Windows (amd64)

2. **Package** each binary with:
   - README.md
   - LICENSE
   - check-projects.example.yml

3. **Create** GitHub Release with:
   - Auto-generated changelog
   - Downloadable archives (.tar.gz, .zip)
   - Checksums file

4. **Update** Homebrew tap (if configured)

### 4. Verify Release

1. Go to [GitHub Releases](https://github.com/uralys/check-projects/releases)
2. Verify the new release is published
3. Test download and installation

## Homebrew Tap Setup (Optional)

To enable automatic Homebrew formula updates:

### 1. Create Homebrew Tap Repository

```bash
# Create a new repository: uralys/homebrew-tap
gh repo create uralys/homebrew-tap --public
```

### 2. Create GitHub Token

1. Go to [GitHub Settings > Tokens](https://github.com/settings/tokens)
2. Create a new token with `repo` scope
3. Copy the token

### 3. Add Repository Secret

1. Go to repository Settings > Secrets and variables > Actions
2. Add new secret: `HOMEBREW_TAP_GITHUB_TOKEN`
3. Paste the token

### 4. Test Installation

```bash
brew tap uralys/tap
brew install check-projects
```

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (v2.0.0): Breaking changes
- **MINOR** (v1.1.0): New features, backwards compatible
- **PATCH** (v1.0.1): Bug fixes, backwards compatible

## Troubleshooting

### Release Failed

Check the [Actions tab](https://github.com/uralys/check-projects/actions) for error details.

Common issues:
- Tests failing: Fix tests before releasing
- Tag already exists: Delete and recreate tag
- Token expired: Update `HOMEBREW_TAP_GITHUB_TOKEN`

### Testing Release Locally

```bash
# Install GoReleaser
brew install goreleaser

# Test release process (without publishing)
goreleaser release --snapshot --clean

# Check generated files in dist/
ls -la dist/
```

## CI/CD Workflows

### CI Workflow (ci.yml)

Runs on every push and PR:
- Runs tests
- Checks build
- Runs linter

### Release Workflow (release.yml)

Runs only on version tags (v*):
- Builds multi-platform binaries
- Creates GitHub Release
- Updates Homebrew tap
