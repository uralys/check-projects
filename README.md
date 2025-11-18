# check-projects

A fast, cross-platform CLI tool to check the git status of multiple projects organized by categories.

Run `check-projects` to see which of your projects have uncommitted changes, are ahead of remote, or have other git status indicators.

## Output Example

```sh
x mozilla
  * M firefox
  ✱ ✚ thunderbird

✔ godot

x gamedev
  ⬆ flying-ones
  * M avindi
```

## Git Status Symbols

- `✔` - Clean (synced with remote)
- `⬆` - Ahead of remote
- `⬆⬆` - Diverged from remote
- `* M` - Modified files
- `* D` - Deleted files
- `✱ ✚` - Untracked files
- `❌` - Error

## Features

- **Multi-category organization**: Group your projects by team, client, or any category
- **Nested project discovery**: Automatically scan nested folder structures
- **Concurrent scanning**: Fast parallel git status checks
- **Flexible configuration**: YAML-based config with local and global support
- **Smart filtering**: Hide clean projects by default, show only what needs attention
- **Cross-platform**: Single binary for macOS, Linux, and Windows

## Installation

### Quick Install (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/uralys/check-projects/main/install.sh | sh
```

This will download the latest release and install it to `~/.local/bin/check-projects`.

<details>
<summary><b>show manual Installation</b></summary>

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

### From Source

```bash
git clone https://github.com/uralys/check-projects.git
cd check-projects
make install
```

</details>

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
  # Mode 1: Explicit project list (using 'projects' field)
  # Use full paths to specific git repositories
  - name: core
    projects:
      - ~/fox
      - ~/cherry

  # Mode 2: Auto-scan directory (using 'root' field)
  # Recursively scans for all git repositories in the directory
  - name: godot
    root: ~/Projects/godot

  # Mode 2 with ignore patterns
  # Projects listed in 'ignore' will be skipped in this category
  - name: uralys
    root: ~/Projects/uralys
    ignore:
      - deprecated-project        # Exact match
      - _archives/*              # Wildcard: ignore all projects in _archives/
      - "*-old"                  # Pattern: ignore all projects ending with -old

# Display options
display:
  hide_clean: true      # Hide projects with ✔ status by default
  hide_ignored: true    # Hide ignored projects from output
```

### Ignore Patterns

You can ignore specific projects in a category using the `ignore` field. Supported patterns:

- **Exact match**: `project-name` - ignores the exact project name
- **Wildcard prefix**: `_archives/*` - ignores all projects in the `_archives/` directory
- **Glob patterns**: `*-deprecated` - ignores all projects ending with `-deprecated`

Common ignore patterns are automatically applied:
- `node_modules` - always skipped during scanning
- `.DS_Store` - always skipped during scanning

## Usage

```bash
# Check all projects
check-projects

# Show all projects including clean ones
check-projects --verbose
check-projects -v

# Check only specific category
check-projects --category gamedev

# Use custom config file
check-projects --config /path/to/config.yml

# Show version
check-projects --version
```

## Updates

`check-projects` automatically checks for new versions on startup. When a new version is available, you'll see:

```
⚠ New version available: 1.0.0 → 1.1.0
Install update? [Y/n]:
```

- Press **Enter** or type **Y** to automatically download and install the update
- Type **n** to skip and continue with your current version

The update check is non-blocking and will silently fail if GitHub is unreachable.

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
