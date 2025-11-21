# TUI Mode

The Terminal User Interface (TUI) provides an interactive way to navigate and manage your projects.

![TUI Interface](images/tui-screenshot.png)

## Launch

```bash
check-projects --tui
```

## Keybindings

### Navigation
- `↑`/`↓` - Navigate through projects (git status updates automatically)
- `←`/`→` - Switch between categories

### Actions
- `h` - Toggle hide/show clean projects
- `r` - Refresh all projects
- `q`, `ESC` or `Ctrl+C` - Quit

## Features

- **Automatic split-screen**: Git status always visible on the right panel
- **Category navigation**: Switch between categories with arrow keys
- **Visual feedback**: Color-coded status symbols
- **Responsive layout**: Adapts to terminal size (minimum 60x10)
- **Fast scanning**: Concurrent git status checks

## Split-Screen Layout

The TUI automatically displays two panels:
- **Left**: Project list with navigation
- **Right**: Git status details for the currently selected project

As you navigate through projects with `↑↓`, the right panel automatically updates to show the git status for the selected project.

For projects with changes, you'll see the output of `git status --short` with colors:
- **Green** - Staged changes (A, M, D in index)
- **Red** - Unstaged/untracked changes (??, M, D in worktree)

For clean projects, you'll see a "✔ Project is clean" message.
