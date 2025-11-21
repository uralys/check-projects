# Installation

## Quick Install (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/uralys/check-projects/main/install.sh | sh
```

This will download the latest release and install it to `~/.local/bin/check-projects`.

## Manual Installation

### macOS/Linux

1. Download the latest release for your platform from [GitHub Releases](https://github.com/uralys/check-projects/releases)

2. Extract and install:

```bash
tar -xzf check-projects-*.tar.gz
chmod +x check-projects
sudo mv check-projects /usr/local/bin/

# Or install to user directory (no sudo required)
mkdir -p ~/.local/bin
mv check-projects ~/.local/bin/
# Add ~/.local/bin to your PATH if not already done
```

### Windows

Download the Windows executable from [GitHub Releases](https://github.com/uralys/check-projects/releases)

```powershell
# PowerShell as Administrator
Move-Item check-projects.exe C:\Windows\System32\
```

## From Source

```bash
git clone https://github.com/uralys/check-projects.git
cd check-projects
make install
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
