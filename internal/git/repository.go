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

// SetUpstream configures and pushes to set upstream tracking
func (r *Repository) SetUpstream() error {
	cmd := exec.Command("git", "push", "--set-upstream", "origin", "HEAD")
	cmd.Dir = r.Path

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set upstream: %s", stderr.String())
	}

	return nil
}
