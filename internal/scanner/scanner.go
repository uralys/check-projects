package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/uralys/check-projects/internal/config"
	"github.com/uralys/check-projects/internal/git"
)

// Project represents a discovered project
type Project struct {
	Name          string
	Path          string
	Category      string
	Repository    *git.Repository
	IsSymlink     bool
	SymlinkTarget string
}

// Scanner scans for projects based on configuration
type Scanner struct {
	config *config.Config
}

// NewScanner creates a new Scanner
func NewScanner(cfg *config.Config) *Scanner {
	return &Scanner{config: cfg}
}

// ScanAll scans all categories and returns discovered projects
func (s *Scanner) ScanAll() ([]Project, error) {
	var projects []Project

	for _, category := range s.config.Categories {
		categoryProjects, err := s.scanCategory(category)
		if err != nil {
			// Log error but continue with other categories
			continue
		}
		projects = append(projects, categoryProjects...)
	}

	return projects, nil
}

func (s *Scanner) scanCategory(category config.Category) ([]Project, error) {
	var projects []Project

	// Mode 1: Explicit projects list (full paths)
	if len(category.Projects) > 0 {
		for _, projectPath := range category.Projects {
			expandedPath := config.ExpandPath(projectPath)
			if !git.IsGitRepository(expandedPath) {
				continue
			}
			// Extract project name from path
			projectName := filepath.Base(expandedPath)

			// Check if ignored in this category
			if s.isIgnored(projectName, category.Ignore) {
				continue
			}

			projects = append(projects, Project{
				Name:       projectName,
				Path:       expandedPath,
				Category:   category.Name,
				Repository: git.NewRepository(expandedPath, projectName),
			})
		}
		return projects, nil
	}

	// Mode 2: Auto-scan root directory recursively
	if category.Root != "" {
		rootPath := config.ExpandPath(category.Root)
		projects = s.scanRecursive(rootPath, category.Name, category.Ignore)
		return projects, nil
	}

	return projects, nil
}

// scanRecursive recursively scans a directory for git repositories
func (s *Scanner) scanRecursive(rootPath, categoryName string, ignored []string) []Project {
	var projects []Project
	s.scanRecursiveHelper(rootPath, rootPath, categoryName, ignored, &projects)
	return projects
}

func (s *Scanner) scanRecursiveHelper(basePath, currentPath, categoryName string, ignored []string, projects *[]Project) {
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(currentPath, name)

		isDir := entry.IsDir()
		isSymlink := entry.Type()&os.ModeSymlink != 0
		symlinkTarget := ""

		if !isDir && isSymlink {
			target, err := os.Readlink(fullPath)
			if err != nil {
				continue
			}
			if !filepath.IsAbs(target) {
				target = filepath.Join(currentPath, target)
			}
			symlinkTarget = target

			// Skip ignored before any expensive I/O on the target
			if s.shouldIgnore(name) {
				continue
			}

			// Try git repo check first (single stat on target/.git)
			if git.IsGitRepository(fullPath) {
				relPath, relErr := filepath.Rel(basePath, fullPath)
				if relErr != nil {
					relPath = name
				}
				if !s.isIgnored(relPath, ignored) {
					*projects = append(*projects, Project{
						Name:          relPath,
						Path:          fullPath,
						Category:      categoryName,
						Repository:    git.NewRepository(fullPath, relPath),
						IsSymlink:     true,
						SymlinkTarget: symlinkTarget,
					})
				}
				continue
			}

			// Not a git repo: check if it's a directory to recurse into
			info, err := os.Lstat(target)
			if err != nil {
				// Broken symlink
				relPath, relErr := filepath.Rel(basePath, fullPath)
				if relErr != nil {
					relPath = name
				}
				if !s.isIgnored(relPath, ignored) {
					*projects = append(*projects, Project{
						Name:          relPath,
						Path:          fullPath,
						Category:      categoryName,
						IsSymlink:     true,
						SymlinkTarget: symlinkTarget,
					})
				}
				continue
			}

			if !info.IsDir() {
				continue
			}

			// Symlink to a non-git directory: recurse
			s.scanRecursiveHelper(basePath, fullPath, categoryName, ignored, projects)
			continue
		} else if !isDir {
			continue
		}

		// Skip ignored directories (hardcoded)
		if s.shouldIgnore(name) {
			continue
		}

		// If this directory is a git repo, check if it should be added
		if git.IsGitRepository(fullPath) {
			relPath, err := filepath.Rel(basePath, fullPath)
			if err != nil {
				relPath = name
			}

			if !s.isIgnored(relPath, ignored) {
				*projects = append(*projects, Project{
					Name:       relPath,
					Path:       fullPath,
					Category:   categoryName,
					Repository: git.NewRepository(fullPath, relPath),
				})
			}

			continue
		}

		// Recurse into subdirectories
		s.scanRecursiveHelper(basePath, fullPath, categoryName, ignored, projects)
	}
}

// shouldIgnore checks if a directory name should be ignored
// These are common patterns that should always be skipped during scanning
func (s *Scanner) shouldIgnore(name string) bool {
	// Common directories that should never be scanned for git repos
	commonIgnored := []string{"node_modules", ".DS_Store"}
	for _, ignored := range commonIgnored {
		if name == ignored {
			return true
		}
	}
	return false
}

// isIgnored checks if a project path matches any ignored pattern from config
func (s *Scanner) isIgnored(projectPath string, ignored []string) bool {
	for _, pattern := range ignored {
		// Exact match
		if projectPath == pattern || filepath.Base(projectPath) == pattern {
			return true
		}

		// Prefix match with wildcard (e.g., "_archives/*" matches "_archives/anything")
		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(projectPath, prefix+"/") || projectPath == prefix {
				return true
			}
		}

		// Wildcard match (e.g., "*-deprecated" matches "foo-deprecated")
		if strings.Contains(pattern, "*") {
			matched, err := filepath.Match(pattern, projectPath)
			if err == nil && matched {
				return true
			}
			// Also try matching against basename
			matched, err = filepath.Match(pattern, filepath.Base(projectPath))
			if err == nil && matched {
				return true
			}
		}
	}
	return false
}
