package scanner

import (
	"os"
	"path/filepath"

	"github.com/uralys/check-projects/internal/config"
	"github.com/uralys/check-projects/internal/git"
)

// Project represents a discovered project
type Project struct {
	Name       string
	Path       string
	Category   string
	Repository *git.Repository
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

			// Check if ignored
			if s.isIgnored(projectName, s.config.Ignored) {
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
		projects = s.scanRecursive(rootPath, category.Name, s.config.Ignored)
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
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		fullPath := filepath.Join(currentPath, name)

		// Skip ignored directories (hardcoded)
		if s.shouldIgnore(name) {
			continue
		}

		// If this directory is a git repo, check if it should be added
		if git.IsGitRepository(fullPath) {
			// Calculate relative path from basePath
			relPath, err := filepath.Rel(basePath, fullPath)
			if err != nil {
				relPath = name
			}

			// Check if this project matches any ignored pattern
			if !s.isIgnored(relPath, ignored) {
				*projects = append(*projects, Project{
					Name:       relPath,
					Path:       fullPath,
					Category:   categoryName,
					Repository: git.NewRepository(fullPath, relPath),
				})
			}

			// Don't scan inside git repositories (nested repos are not scanned)
			continue
		}

		// Recurse into subdirectories
		s.scanRecursiveHelper(basePath, fullPath, categoryName, ignored, projects)
	}
}

// shouldIgnore checks if a directory name should be ignored (hardcoded patterns)
func (s *Scanner) shouldIgnore(name string) bool {
	ignoredNames := []string{".DS_Store", "node_modules", "_archives", "_assets", "_tools"}
	for _, ignored := range ignoredNames {
		if name == ignored {
			return true
		}
	}
	return false
}

// isIgnored checks if a project path matches any ignored pattern from config
func (s *Scanner) isIgnored(projectPath string, ignored []string) bool {
	for _, pattern := range ignored {
		// Simple pattern matching: exact match or suffix match
		if projectPath == pattern || filepath.Base(projectPath) == pattern {
			return true
		}
	}
	return false
}
