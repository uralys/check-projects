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

// GetCurrentBranch returns the name of the current branch
func (r *Repository) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = r.Path

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get current branch: %v", err)
	}

	return string(bytes.TrimSpace(stdout.Bytes())), nil
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

// Pull updates a local branch from its remote.
// If branch is the currently checked-out branch, it pulls in place.
// Otherwise it updates the branch through a temporary worktree, leaving the
// current working tree untouched.
func (r *Repository) Pull(branch string) error {
	current, err := r.GetCurrentBranch()
	if err != nil {
		return err
	}

	if branch == current {
		return r.pullInPlace(branch)
	}

	return r.pullViaWorktree(branch)
}

// pullInPlace runs git pull origin <branch> in the repository directory.
func (r *Repository) pullInPlace(branch string) error {
	cmd := exec.Command("git", "pull", "origin", branch)
	cmd.Dir = r.Path

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pull failed: %s", stderr.String())
	}

	return nil
}

// pullViaWorktree updates a branch that is not currently checked out by
// checking it out in a temporary worktree, pulling, then removing the worktree.
// This avoids merging the target branch into the current working branch.
func (r *Repository) pullViaWorktree(branch string) error {
	tmpDir, err := os.MkdirTemp("", "check-projects-worktree-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	worktreePath := filepath.Join(tmpDir, "wt")

	addCmd := exec.Command("git", "worktree", "add", worktreePath, branch)
	addCmd.Dir = r.Path
	var addStderr bytes.Buffer
	addCmd.Stderr = &addStderr
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("worktree add failed: %s", addStderr.String())
	}

	// Always remove the worktree, even if the pull fails.
	defer func() {
		removeCmd := exec.Command("git", "worktree", "remove", "--force", worktreePath)
		removeCmd.Dir = r.Path
		_ = removeCmd.Run()
	}()

	pullCmd := exec.Command("git", "pull", "origin", branch)
	pullCmd.Dir = worktreePath
	var pullStderr bytes.Buffer
	pullCmd.Stderr = &pullStderr
	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("pull failed: %s", pullStderr.String())
	}

	return nil
}
