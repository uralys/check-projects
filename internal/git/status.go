package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// StatusType represents the type of git status
type StatusType string

const (
	StatusSync       StatusType = "sync"
	StatusUnsync     StatusType = "unsync"
	StatusError      StatusType = "error"
	StatusIgnored    StatusType = "ignored"
	StatusNoUpstream StatusType = "no_upstream"
)

// Status represents the git status of a repository
type Status struct {
	Type    StatusType
	Message string
	Symbol  string
}

// Fetch runs git fetch to update remote tracking branches
func (r *Repository) Fetch() error {
	cmd := exec.Command("git", "fetch")
	cmd.Dir = r.Path

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fetch failed: %s", stderr.String())
	}

	return nil
}

// GetStatus retrieves the git status of a repository
func (r *Repository) GetStatus() (*Status, error) {
	// First check if upstream is configured
	upstreamCmd := exec.Command("git", "rev-list", "@{u}..HEAD", "--count")
	upstreamCmd.Dir = r.Path

	var upstreamStderr bytes.Buffer
	upstreamCmd.Stderr = &upstreamStderr

	if err := upstreamCmd.Run(); err != nil {
		// Check if error is due to missing upstream
		stderrStr := upstreamStderr.String()
		if strings.Contains(stderrStr, "no upstream configured") ||
		   strings.Contains(stderrStr, "upstream branch") ||
		   strings.Contains(stderrStr, "no such branch") {
			return &Status{
				Type:    StatusNoUpstream,
				Message: "No upstream configured",
				Symbol:  "⚠ No upstream",
			}, nil
		}
	}

	// Run git status
	cmd := exec.Command("git", "status")
	cmd.Dir = r.Path

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return &Status{
			Type:    StatusError,
			Message: fmt.Sprintf("Error: %s", stderr.String()),
			Symbol:  "❌",
		}, nil
	}

	output := stdout.String()

	// Check for various git states
	// Check for staged changes first (Changes to be committed)
	if strings.Contains(output, "Changes to be committed:") {
		// Determine the type of staged change
		if strings.Contains(output, "renamed:") {
			return &Status{
				Type:    StatusUnsync,
				Message: "Staged renames",
				Symbol:  "✱ R",
			}, nil
		}
		if strings.Contains(output, "new file:") {
			return &Status{
				Type:    StatusUnsync,
				Message: "Staged files",
				Symbol:  "✱ +",
			}, nil
		}
		// Generic staged changes
		return &Status{
			Type:    StatusUnsync,
			Message: "Staged changes",
			Symbol:  "✱",
		}, nil
	}

	if strings.Contains(output, "modified:") {
		return &Status{
			Type:    StatusUnsync,
			Message: "Modified files",
			Symbol:  "* M",
		}, nil
	}

	if strings.Contains(output, "deleted:") {
		return &Status{
			Type:    StatusUnsync,
			Message: "Deleted files",
			Symbol:  "* D",
		}, nil
	}

	if strings.Contains(output, "Untracked files:") {
		return &Status{
			Type:    StatusUnsync,
			Message: "Untracked files",
			Symbol:  "✱ ✚",
		}, nil
	}

	if strings.Contains(output, "is ahead of") {
		return &Status{
			Type:    StatusUnsync,
			Message: "Ahead of remote",
			Symbol:  "⬆",
		}, nil
	}

	if strings.Contains(output, "is behind") {
		return &Status{
			Type:    StatusUnsync,
			Message: "Behind remote",
			Symbol:  "↓",
		}, nil
	}

	if strings.Contains(output, "diverged") {
		return &Status{
			Type:    StatusUnsync,
			Message: "Diverged from remote",
			Symbol:  "⬆⬆",
		}, nil
	}

	if strings.Contains(output, "nothing to commit, working tree clean") {
		return &Status{
			Type:    StatusSync,
			Message: "Clean",
			Symbol:  "✔",
		}, nil
	}

	// Default to unsync if we can't determine
	return &Status{
		Type:    StatusUnsync,
		Message: "Unknown state",
		Symbol:  "*",
	}, nil
}
