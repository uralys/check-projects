package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/uralys/check-projects/internal/config"
	"github.com/uralys/check-projects/internal/git"
)

// Model represents the application state for the TUI
type Model struct {
	// Configuration
	config *config.Config

	// Projects and results
	projects []ProjectWithStatus

	// UI state
	loading         bool
	hideClean       bool
	errorMsg        string
	fetchingProject int // Index of project being fetched (-1 means none)

	// Selection
	selectedCategory int
	selectedProject  int
	detailsScroll    int  // Scroll offset for git status details
	focusedPanel     bool // true = details panel, false = projects panel

	// Bubble components
	spinner  spinner.Model
	viewport viewport.Model

	// Window size
	width  int
	height int

	// Categories for navigation
	categories []string

	// Version info
	version string
}

// NewModel creates a new TUI model
func NewModel(cfg *config.Config, version string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	// Extract category names
	categories := make([]string, len(cfg.Categories))
	for i, cat := range cfg.Categories {
		categories[i] = cat.Name
	}

	return Model{
		config:           cfg,
		loading:          true,
		hideClean:        true, // Hide clean projects by default in TUI
		spinner:          s,
		categories:       categories,
		selectedCategory: 0,
		selectedProject:  0,
		version:          version,
		fetchingProject:  -1, // No project being fetched initially
	}
}

// getFilteredProjects returns projects filtered by current settings
func (m Model) getFilteredProjects() []ProjectWithStatus {
	if len(m.projects) == 0 {
		return nil
	}

	filtered := make([]ProjectWithStatus, 0)
	currentCategory := ""
	if m.selectedCategory < len(m.categories) {
		currentCategory = m.categories[m.selectedCategory]
	}

	for _, p := range m.projects {
		// Filter by category
		if currentCategory != "" && p.Project.Category != currentCategory {
			continue
		}

		// Filter by clean status - skip if clean AND no behind branches
		if m.hideClean && p.Status != nil && p.Status.Type == git.StatusSync && len(p.Status.BehindBranches) == 0 {
			continue
		}

		filtered = append(filtered, p)
	}

	return filtered
}

// categoryHasChanges checks if a category has any projects with changes or behind branches
func (m Model) categoryHasChanges(categoryName string) bool {
	for _, p := range m.projects {
		if p.Project.Category == categoryName {
			if p.Status != nil {
				// Check if status is not clean
				if p.Status.Type != git.StatusSync {
					return true
				}
				// Check if there are branches behind remote
				if len(p.Status.BehindBranches) > 0 {
					return true
				}
			}
		}
	}
	return false
}

// getVisibleCategories returns categories filtered by hideClean setting
func (m Model) getVisibleCategories() []string {
	if !m.hideClean {
		return m.categories
	}

	// Filter out clean categories when hideClean is true
	var visible []string
	for _, cat := range m.categories {
		if m.categoryHasChanges(cat) {
			visible = append(visible, cat)
		}
	}
	return visible
}

// hasAnyChanges checks if there are any projects with changes or behind branches across all categories
func (m Model) hasAnyChanges() bool {
	for _, p := range m.projects {
		if p.Status != nil {
			// Check if status is not clean
			if p.Status.Type != git.StatusSync {
				return true
			}
			// Check if there are branches behind remote
			if len(p.Status.BehindBranches) > 0 {
				return true
			}
		}
	}
	return false
}
