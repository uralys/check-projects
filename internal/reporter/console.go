package reporter

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/uralys/check-projects/internal/config"
	"github.com/uralys/check-projects/internal/git"
)

var (
	green     = color.New(color.FgGreen).SprintFunc()
	red       = color.New(color.FgRed).SprintFunc()
	greenBold = color.New(color.FgGreen, color.Bold).SprintFunc()
	redBold   = color.New(color.FgRed, color.Bold).SprintFunc()
	bold      = color.New(color.Bold).SprintFunc()
	underline = color.New(color.Bold, color.Underline).SprintFunc()
)

// Reporter handles output formatting
type Reporter struct {
	config  *config.Config
	verbose bool
}

// NewReporter creates a new Reporter
func NewReporter(cfg *config.Config, verbose bool) *Reporter {
	return &Reporter{
		config:  cfg,
		verbose: verbose,
	}
}

// ProjectResult represents the result of checking a project
type ProjectResult struct {
	Name     string
	Status   *git.Status
	Category string
}

// Report generates and displays the final report
func (r *Reporter) Report(results []ProjectResult) {
	// Group results by category
	categoryResults := make(map[string][]ProjectResult)
	for _, result := range results {
		categoryResults[result.Category] = append(categoryResults[result.Category], result)
	}

	// Check if all projects are clean
	allClean := true
	for _, result := range results {
		if result.Status.Type != git.StatusSync && result.Status.Type != git.StatusIgnored {
			allClean = false
			break
		}
	}

	if allClean {
		fmt.Println(greenBold("✔ All projects are clean!"))
		return
	}

	// Display results by category
	for category, categoryProjects := range categoryResults {
		r.displayCategory(category, categoryProjects)
	}
}

func (r *Reporter) displayCategory(category string, results []ProjectResult) {
	// Check if all projects in this category are clean
	allClean := true
	for _, result := range results {
		if result.Status.Type != git.StatusSync && result.Status.Type != git.StatusIgnored {
			allClean = false
			break
		}
	}

	// Display category header
	if allClean {
		fmt.Printf("%s %s\n", greenBold("✔"), greenBold(category))
	} else {
		fmt.Printf("%s %s\n", redBold("x"), underline(category))
	}

	// Display projects
	if !allClean {
		for _, result := range results {
			// Skip ignored projects if configured
			if r.config.Display.HideIgnored && result.Status.Type == git.StatusIgnored {
				continue
			}

			// Skip clean projects unless verbose mode
			if r.config.Display.HideClean && !r.verbose && result.Status.Type == git.StatusSync {
				continue
			}

			r.displayProject(result)
		}
	} else if r.verbose {
		// In verbose mode, show all projects even if category is clean
		for _, result := range results {
			if r.config.Display.HideIgnored && result.Status.Type == git.StatusIgnored {
				continue
			}
			r.displayProject(result)
		}
	}
}

func (r *Reporter) displayProject(result ProjectResult) {
	switch result.Status.Type {
	case git.StatusSync:
		// Green tick + white project name
		fmt.Printf("  %s %s\n", green(result.Status.Symbol), result.Name)
	case git.StatusUnsync:
		// Red status
		message := fmt.Sprintf("%s %s", result.Status.Symbol, result.Name)
		fmt.Printf("  %s\n", red(message))
	case git.StatusError:
		// Red error
		message := fmt.Sprintf("%s %s", result.Status.Symbol, result.Name)
		fmt.Printf("  %s\n", red(message))
	case git.StatusNoUpstream:
		// Yellow/default for no upstream
		message := fmt.Sprintf("%s %s", result.Status.Symbol, result.Name)
		fmt.Printf("  %s\n", message)
	default:
		// Default color
		message := fmt.Sprintf("%s %s", result.Status.Symbol, result.Name)
		fmt.Printf("  %s\n", message)
	}
}
