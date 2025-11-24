package tui

import (
	"github.com/uralys/check-projects/internal/git"
	"github.com/uralys/check-projects/internal/scanner"
)

// ProjectWithStatus represents a project with its Git status
type ProjectWithStatus struct {
	Project scanner.Project
	Status  *git.Status
}

// scanCompleteMsg is sent when the initial scan is complete
type scanCompleteMsg struct {
	projects []ProjectWithStatus
	err      error
}

// fetchCompleteMsg is sent when a fetch operation is complete
type fetchCompleteMsg struct {
	projectIndex int
	err          error
}

// fetchingMsg is sent when a fetch operation starts
type fetchingMsg struct {
	projectIndex int
}
