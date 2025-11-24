package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles incoming messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 6 // Reserve space for header and footer

	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "r":
			// Refresh
			m.loading = true
			return m, scanProjectsCmd(m.config)

		case "f":
			// Fetch selected project
			filtered := m.getFilteredProjects()
			if len(filtered) > 0 && m.selectedProject < len(filtered) {
				// Find the actual index in m.projects
				selectedProj := filtered[m.selectedProject]
				actualIndex := -1
				for i, p := range m.projects {
					if p.Project.Path == selectedProj.Project.Path {
						actualIndex = i
						break
					}
				}

				if actualIndex != -1 {
					m.fetchingProject = actualIndex
					return m, fetchProjectCmd(&m.projects[actualIndex], actualIndex)
				}
			}

		case "h":
			// Toggle hide clean
			m.hideClean = !m.hideClean
			m.selectedProject = 0

			// If current category is now hidden, move to first visible category
			if m.hideClean {
				currentCat := ""
				if m.selectedCategory < len(m.categories) {
					currentCat = m.categories[m.selectedCategory]
				}

				// Check if current category has changes (is visible)
				if currentCat != "" && !m.categoryHasChanges(currentCat) {
					// Current category is now hidden, move to first visible one
					visibleCategories := m.getVisibleCategories()
					if len(visibleCategories) > 0 {
						// Find first visible category in full list
						for i, cat := range m.categories {
							if cat == visibleCategories[0] {
								m.selectedCategory = i
								break
							}
						}
					}
				}
			}

		case "enter":
			// Toggle focus between panels
			m.focusedPanel = !m.focusedPanel
		}

		// Navigation keys - behavior depends on focused panel
		switch msg.String() {
		case "up", "k":
			if m.focusedPanel {
				// Focused on details - scroll up
				if m.detailsScroll > 0 {
					m.detailsScroll--
				}
			} else {
				// Focused on projects - navigate projects
				if m.selectedProject > 0 {
					m.selectedProject--
					m.detailsScroll = 0 // Reset details scroll when changing project
				}
			}

		case "down", "j":
			if m.focusedPanel {
				// Focused on details - scroll down
				m.detailsScroll++
			} else {
				// Focused on projects - navigate projects
				filtered := m.getFilteredProjects()
				if m.selectedProject < len(filtered)-1 {
					m.selectedProject++
					m.detailsScroll = 0 // Reset details scroll when changing project
				}
			}

		case "left":
			// Navigate to previous visible category
			visibleCategories := m.getVisibleCategories()
			if len(visibleCategories) > 0 && m.selectedCategory > 0 {
				// Find current category in visible list
				currentCat := m.categories[m.selectedCategory]
				currentIndex := -1
				for i, cat := range visibleCategories {
					if cat == currentCat {
						currentIndex = i
						break
					}
				}

				// Move to previous visible category
				if currentIndex > 0 {
					prevCat := visibleCategories[currentIndex-1]
					// Find index in full categories list
					for i, cat := range m.categories {
						if cat == prevCat {
							m.selectedCategory = i
							break
						}
					}
					m.selectedProject = 0
					m.detailsScroll = 0
					m.focusedPanel = false
				}
			}

		case "right":
			// Navigate to next visible category
			visibleCategories := m.getVisibleCategories()
			if len(visibleCategories) > 0 && m.selectedCategory < len(m.categories)-1 {
				// Find current category in visible list
				currentCat := m.categories[m.selectedCategory]
				currentIndex := -1
				for i, cat := range visibleCategories {
					if cat == currentCat {
						currentIndex = i
						break
					}
				}

				// Move to next visible category
				if currentIndex >= 0 && currentIndex < len(visibleCategories)-1 {
					nextCat := visibleCategories[currentIndex+1]
					// Find index in full categories list
					for i, cat := range m.categories {
						if cat == nextCat {
							m.selectedCategory = i
							break
						}
					}
					m.selectedProject = 0
					m.detailsScroll = 0
					m.focusedPanel = false
				}
			}

		case "pgup":
			// Page up in details
			m.detailsScroll -= 10
			if m.detailsScroll < 0 {
				m.detailsScroll = 0
			}

		case "pgdown":
			// Page down in details
			m.detailsScroll += 10
		}

	case scanCompleteMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
		} else {
			m.projects = msg.projects
			m.errorMsg = ""

			// Ensure selected category is visible when hideClean is enabled
			if m.hideClean && len(m.categories) > 0 {
				// Check if current selected category has changes
				currentCat := ""
				if m.selectedCategory < len(m.categories) {
					currentCat = m.categories[m.selectedCategory]
				}

				// If current category has no changes, select first visible category
				if currentCat != "" && !m.categoryHasChanges(currentCat) {
					visibleCategories := m.getVisibleCategories()
					if len(visibleCategories) > 0 {
						// Find first visible category in full list
						for i, cat := range m.categories {
							if cat == visibleCategories[0] {
								m.selectedCategory = i
								m.selectedProject = 0
								break
							}
						}
					}
				}
			}
		}

	case fetchingMsg:
		// Mark project as being fetched
		m.fetchingProject = msg.projectIndex

	case fetchCompleteMsg:
		// Clear fetching state
		m.fetchingProject = -1

		if msg.err != nil {
			// Show error briefly (could be improved with a status bar)
			m.errorMsg = fmt.Sprintf("Fetch failed: %v", msg.err)
		} else {
			// Clear any error
			m.errorMsg = ""
		}

	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}
