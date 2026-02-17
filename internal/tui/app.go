package tui

import (
	"fmt"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/uralys/check-projects/internal/config"
	"github.com/uralys/check-projects/internal/git"
	"github.com/uralys/check-projects/internal/scanner"
)

// Run starts the TUI application
func Run(cfg *config.Config, version string) error {
	m := NewModel(cfg, version)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}

// Init initializes the model and starts the initial scan
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		scanProjectsCmd(m.config),
	)
}

// scanProjectsCmd scans all projects and returns their status
func scanProjectsCmd(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		// Scan for projects
		s := scanner.NewScanner(cfg)
		projects, err := s.ScanAll()
		if err != nil {
			return scanCompleteMsg{err: err}
		}

		// Check git status for each project concurrently
		results := make([]ProjectWithStatus, len(projects))
		var wg sync.WaitGroup
		sem := make(chan struct{}, 10) // Limit concurrency to 10

		for i, project := range projects {
			wg.Add(1)
			go func(idx int, proj scanner.Project) {
				defer wg.Done()
				sem <- struct{}{}        // Acquire semaphore
				defer func() { <-sem }() // Release semaphore

				if proj.Repository == nil {
					results[idx] = ProjectWithStatus{
						Project: proj,
						Status:  &git.Status{Type: git.StatusBrokenSymlink, Symbol: "🔗 ✗"},
					}
					return
				}

				status, err := proj.Repository.GetStatus()
				if err != nil {
					// Handle error by marking as error status
					status = &git.Status{
						Type:    git.StatusError,
						Message: err.Error(),
						Symbol:  "❌",
					}
				}

				results[idx] = ProjectWithStatus{
					Project: proj,
					Status:  status,
				}
			}(i, project)
		}

		wg.Wait()

		return scanCompleteMsg{
			projects: results,
			err:      nil,
		}
	}
}

// fetchProjectCmd fetches a single project and refreshes its status
func fetchProjectCmd(projectWithStatus *ProjectWithStatus, projectIndex int) tea.Cmd {
	return func() tea.Msg {
		if projectWithStatus.Project.Repository == nil {
			return fetchCompleteMsg{projectIndex: projectIndex, err: nil}
		}

		// Fetch from remote
		if err := projectWithStatus.Project.Repository.Fetch(); err != nil {
			return fetchCompleteMsg{
				projectIndex: projectIndex,
				err:          err,
			}
		}

		// Get updated status after fetch
		status, err := projectWithStatus.Project.Repository.GetStatus()
		if err != nil {
			return fetchCompleteMsg{
				projectIndex: projectIndex,
				err:          err,
			}
		}

		// Update the status in the project
		projectWithStatus.Status = status

		return fetchCompleteMsg{
			projectIndex: projectIndex,
			err:          nil,
		}
	}
}
