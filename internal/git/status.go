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

// BranchTracking represents the tracking status of a branch
type BranchTracking struct {
	Branch  string
	Message string
}

// Status represents the git status of a repository
type Status struct {
	Type            StatusType
	Message         string
	Symbol          string
	Branch          string           // Current branch name
	BehindBranches  []BranchTracking // Branches that are behind their remote
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

// GetBranchesTrackingStatus checks all local branches and returns those that are behind their remote
func (r *Repository) GetBranchesTrackingStatus() ([]BranchTracking, error) {
	// Get all local branches
	cmd := exec.Command("git", "branch", "--format=%(refname:short)")
	cmd.Dir = r.Path

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get branches: %s", stderr.String())
	}

	branches := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	var behindBranches []BranchTracking

	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if branch == "" {
			continue
		}

		// Check if this branch has a remote tracking branch
		trackingCmd := exec.Command("git", "rev-parse", "--abbrev-ref", branch+"@{u}")
		trackingCmd.Dir = r.Path

		var trackingStderr bytes.Buffer
		trackingCmd.Stderr = &trackingStderr

		if err := trackingCmd.Run(); err != nil {
			// No upstream for this branch, skip it
			continue
		}

		// Check if branch is behind its remote
		statusCmd := exec.Command("git", "status", "-b", "--porcelain")
		statusCmd.Dir = r.Path
		statusCmd.Env = append(statusCmd.Env, "GIT_OPTIONAL_LOCKS=0")

		// Temporarily checkout this branch to get its status
		// Actually, we can use a different approach - check commits behind
		behindCmd := exec.Command("git", "rev-list", "--count", branch+".."+branch+"@{u}")
		behindCmd.Dir = r.Path

		var behindOut bytes.Buffer
		behindCmd.Stdout = &behindOut

		if err := behindCmd.Run(); err != nil {
			// Error checking behind status, skip
			continue
		}

		behindCount := strings.TrimSpace(behindOut.String())
		if behindCount != "0" && behindCount != "" {
			// Get ahead count as well
			aheadCmd := exec.Command("git", "rev-list", "--count", branch+"@{u}.."+branch)
			aheadCmd.Dir = r.Path

			var aheadOut bytes.Buffer
			aheadCmd.Stdout = &aheadOut

			aheadCount := "0"
			if err := aheadCmd.Run(); err == nil {
				aheadCount = strings.TrimSpace(aheadOut.String())
			}

			message := fmt.Sprintf("behind by %s commit(s)", behindCount)
			if aheadCount != "0" {
				message = fmt.Sprintf("behind by %s, ahead by %s commit(s)", behindCount, aheadCount)
			}

			behindBranches = append(behindBranches, BranchTracking{
				Branch:  branch,
				Message: message,
			})
		}
	}

	return behindBranches, nil
}

// GetStatus retrieves the git status of a repository
func (r *Repository) GetStatus() (*Status, error) {
	// Get current branch name
	branch, _ := r.GetCurrentBranch()

	// Check all branches for tracking status
	behindBranches, err := r.GetBranchesTrackingStatus()
	if err != nil {
		// Log error but continue with regular status check
		behindBranches = []BranchTracking{}
	}

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
				Type:           StatusNoUpstream,
				Message:        "No upstream configured",
				Symbol:         "⚠ No upstream",
				Branch:         branch,
				BehindBranches: behindBranches,
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
			Type:           StatusError,
			Message:        fmt.Sprintf("Error: %s", stderr.String()),
			Symbol:         "❌",
			Branch:         branch,
			BehindBranches: behindBranches,
		}, nil
	}

	output := stdout.String()

	// Check for various git states
	// Check for staged changes first (Changes to be committed)
	if strings.Contains(output, "Changes to be committed:") {
		// Determine the type of staged change
		if strings.Contains(output, "renamed:") {
			return &Status{
				Type:           StatusUnsync,
				Message:        "Staged renames",
				Symbol:         "✱ R",
				Branch:         branch,
				BehindBranches: behindBranches,
			}, nil
		}
		if strings.Contains(output, "new file:") {
			return &Status{
				Type:           StatusUnsync,
				Message:        "Staged files",
				Symbol:         "✱ +",
				Branch:         branch,
				BehindBranches: behindBranches,
			}, nil
		}
		// Generic staged changes
		return &Status{
			Type:           StatusUnsync,
			Message:        "Staged changes",
			Symbol:         "✱",
			Branch:         branch,
			BehindBranches: behindBranches,
		}, nil
	}

	if strings.Contains(output, "modified:") {
		return &Status{
			Type:           StatusUnsync,
			Message:        "Modified files",
			Symbol:         "* M",
			Branch:         branch,
			BehindBranches: behindBranches,
		}, nil
	}

	if strings.Contains(output, "deleted:") {
		return &Status{
			Type:           StatusUnsync,
			Message:        "Deleted files",
			Symbol:         "* D",
			Branch:         branch,
			BehindBranches: behindBranches,
		}, nil
	}

	if strings.Contains(output, "Untracked files:") {
		return &Status{
			Type:           StatusUnsync,
			Message:        "Untracked files",
			Symbol:         "✱ ✚",
			Branch:         branch,
			BehindBranches: behindBranches,
		}, nil
	}

	if strings.Contains(output, "is ahead of") {
		return &Status{
			Type:           StatusUnsync,
			Message:        "Ahead of remote",
			Symbol:         "⬆",
			Branch:         branch,
			BehindBranches: behindBranches,
		}, nil
	}

	if strings.Contains(output, "is behind") {
		return &Status{
			Type:           StatusUnsync,
			Message:        "Behind remote",
			Symbol:         "↓",
			Branch:         branch,
			BehindBranches: behindBranches,
		}, nil
	}

	if strings.Contains(output, "diverged") {
		return &Status{
			Type:           StatusUnsync,
			Message:        "Diverged from remote",
			Symbol:         "⬆⬆",
			Branch:         branch,
			BehindBranches: behindBranches,
		}, nil
	}

	if strings.Contains(output, "nothing to commit, working tree clean") {
		return &Status{
			Type:           StatusSync,
			Message:        "Clean",
			Symbol:         "✔",
			Branch:         branch,
			BehindBranches: behindBranches,
		}, nil
	}

	// Default to unsync if we can't determine
	return &Status{
		Type:           StatusUnsync,
		Message:        "Unknown state",
		Symbol:         "*",
		Branch:         branch,
		BehindBranches: behindBranches,
	}, nil
}
