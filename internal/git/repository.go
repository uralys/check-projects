package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Repository represents a git repository
type Repository struct {
	Path string
	Name string
}

// IsGitRepository checks if a path is a git repository
func IsGitRepository(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// NewRepository creates a new Repository instance
func NewRepository(path, name string) *Repository {
	return &Repository{
		Path: path,
		Name: name,
	}
}

// SetUpstream configures upstream tracking locally without pushing
func (r *Repository) SetUpstream() error {
	// Get current branch name
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchCmd.Dir = r.Path

	var branchOut bytes.Buffer
	branchCmd.Stdout = &branchOut

	if err := branchCmd.Run(); err != nil {
		return fmt.Errorf("failed to get current branch: %v", err)
	}

	branch := bytes.TrimSpace(branchOut.Bytes())
	branchName := string(branch)

	// Set remote tracking locally (without pushing)
	remoteCmd := exec.Command("git", "config", fmt.Sprintf("branch.%s.remote", branchName), "origin")
	remoteCmd.Dir = r.Path
	if err := remoteCmd.Run(); err != nil {
		return fmt.Errorf("failed to set branch remote: %v", err)
	}

	mergeCmd := exec.Command("git", "config", fmt.Sprintf("branch.%s.merge", branchName), fmt.Sprintf("refs/heads/%s", branchName))
	mergeCmd.Dir = r.Path
	if err := mergeCmd.Run(); err != nil {
		return fmt.Errorf("failed to set branch merge: %v", err)
	}

	return nil
}
