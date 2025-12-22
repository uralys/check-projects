# Configuration

Configuration files are searched in this order:

1. Path specified with `--config` flag
2. `./check-projects.yml` (current directory)
3. `~/check-projects.yml` (home directory)

## Example Configuration

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
  hide_clean: true      # Hide projects with ✔ status by default (CLI mode)
  hide_ignored: true    # Hide ignored projects from output
```

## Category Modes

### Mode 1: Explicit Project List

Use the `projects` field to specify exact paths to git repositories:

```yaml
- name: my-category
  projects:
    - ~/path/to/project1
    - ~/path/to/project2
```

### Mode 2: Auto-Scan Directory

Use the `root` field to automatically scan a directory for all git repositories:

```yaml
- name: my-category
  root: ~/Projects/my-projects
```

This will recursively find all git repositories under the specified directory.

## Ignore Patterns

You can ignore specific projects in a category using the `ignore` field. Supported patterns:

- **Exact match**: `project-name` - ignores the exact project name
- **Wildcard prefix**: `_archives/*` - ignores all projects in the `_archives/` directory
- **Glob patterns**: `*-deprecated` - ignores all projects ending with `-deprecated`

Common ignore patterns are automatically applied:
- `node_modules` - always skipped during scanning
- `.DS_Store` - always skipped during scanning

## Display Options

### hide_clean

When set to `true`, only shows projects with changes in CLI mode (default: `true`).

**Note**: TUI mode always shows all projects by default, regardless of this setting. You can toggle visibility with the `h` key.

### hide_ignored

When set to `true`, hides ignored projects from the output (default: `true`).

## Fetch Options

### fetch

When set to `true`, always fetch from remote before checking status (default: `false`).

### fetch_concurrency

Number of parallel fetches when using `-f` or `fetch: true` (default: `10`).

```yaml
fetch: true
fetch_concurrency: 30  # Run up to 30 fetches in parallel
```
