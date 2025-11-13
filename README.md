# check-projects

A fast, cross-platform CLI tool to check the git status of multiple projects organized by categories.

## Features

- **Multi-category organization**: Group your projects by team, client, or any category
- **Nested project discovery**: Automatically scan nested folder structures
- **Concurrent scanning**: Fast parallel git status checks
- **Flexible configuration**: YAML-based config with local and global support
- **Smart filtering**: Hide clean projects by default, show only what needs attention
- **Cross-platform**: Single binary for macOS, Linux, and Windows

## Installation

### Using Homebrew (macOS/Linux)

```bash
brew tap uralys/tap
brew install check-projects
```

### From Source

```bash
git clone https://github.com/uralys/check-projects.git
cd check-projects
make install
```

### Download Binary

1. Download the latest release for your platform from [GitHub Releases](https://github.com/uralys/check-projects/releases)

2. Extract and install:

```bash
# macOS/Linux
tar -xzf check-projects-*.tar.gz
chmod +x check-projects
sudo mv check-projects /usr/local/bin/

# Or install to user directory (no sudo required)
mkdir -p ~/.local/bin
mv check-projects ~/.local/bin/
# Add ~/.local/bin to your PATH if not already done
```

```powershell
# Windows (PowerShell as Administrator)
Move-Item check-projects.exe C:\Windows\System32\
```

## Quick Start

1. Create a configuration file:

```bash
cp check-projects.example.yml ~/check-projects.yml
```

2. Edit `~/check-projects.yml` to match your project structure

3. Run the tool:

```bash
check-projects
```

## Configuration

Configuration files are searched in this order:

1. Path specified with `--config` flag
2. `./check-projects.yml` (current directory)
3. `~/check-projects.yml` (home directory)

### Example Configuration

```yaml
categories:
  # Simple category with explicit project list
  - name: core
    root: ~/
    projects:
      - dev-tools
      - my-project

  # Nested category - auto-scans folder/subfolder/project
  - name: clients
    root: ~/Projects/clients
    nested: true

  # Category with scan patterns
  - name: web
    root: ~/Projects
    scan:
      - frontend/*
      - backend/*

ignored:
  - godot-google-play-billing
  - "**/.DS_Store"

display:
  hide_clean: true      # Hide projects with ✔ status
  hide_ignored: true    # Hide ignored projects
```

## Usage

```bash
# Check all projects
check-projects

# Show all projects including clean ones
check-projects --verbose
check-projects -v

# Check only specific category
check-projects --category core

# Use custom config file
check-projects --config /path/to/config.yml
```

## Git Status Symbols

- `✔` - Clean (synced with remote)
- `⬆` - Ahead of remote
- `⬆⬆` - Diverged from remote
- `* M` - Modified files
- `* D` - Deleted files
- `✱ ✚` - Untracked files
- `❌` - Error

## Output Example

```sh
x alterego
  * M project-a
  ✱ ✚ project-b

✔ core

x nomodata
  ⬆ nomosense/backend
  * M nomosense/frontend
```

## Development

```bash
# Install dependencies
make deps

# Run without building
make dev

# Build binary
make build

# Run tests
make test

# Build for all platforms
make release
```

## Releasing

Releases are automated using GitHub Actions and [GoReleaser](https://goreleaser.com/).

### Creating a New Release

1. **Tag the release:**

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

2. **GitHub Actions will automatically:**
   - Build binaries for all platforms (macOS, Linux, Windows)
   - Create a GitHub Release with changelog
   - Attach binaries and checksums
   - Update the Homebrew tap (if configured)

### First-time Setup

To enable Homebrew tap updates, create a GitHub token and add it as a repository secret:

1. Create a [Personal Access Token](https://github.com/settings/tokens) with `repo` scope
2. Add it as `HOMEBREW_TAP_GITHUB_TOKEN` in repository secrets
3. Create the `uralys/homebrew-tap` repository

## License

MIT
