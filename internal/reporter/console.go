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
	blue      = color.New(color.FgCyan, color.Bold).SprintFunc()
	greenBold = color.New(color.FgGreen, color.Bold).SprintFunc()
	redBold   = color.New(color.FgRed, color.Bold).SprintFunc()
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
	Name          string
	Status        *git.Status
	Category      string
	IsSymlink     bool
	SymlinkTarget string
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
		// Also check if there are behind branches
		if len(result.Status.BehindBranches) > 0 {
			allClean = false
			break
		}
	}

	if allClean && !r.verbose {
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
		// Also check if there are behind branches
		if len(result.Status.BehindBranches) > 0 {
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

			// Skip clean projects unless verbose mode or they have behind branches
			if r.config.Display.HideClean && !r.verbose && result.Status.Type == git.StatusSync && len(result.Status.BehindBranches) == 0 {
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
	displayName := result.Name
	if result.IsSymlink && result.SymlinkTarget != "" {
		displayName = fmt.Sprintf("%s -> %s", result.Name, result.SymlinkTarget)
	}

	switch result.Status.Type {
	case git.StatusSync:
		fmt.Printf("  %s %s\n", green(result.Status.Symbol), displayName)
		r.displayBehindBranches(result)
	case git.StatusUnsync:
		if len(result.Status.Symbol) >= 3 && result.Status.Symbol[0:3] == "✱ " {
			letter := result.Status.Symbol[len("✱ "):]
			if result.Status.Branch != "" {
				fmt.Printf("  %s %s %s - %s\n", red("✱"), green(letter), displayName, blue(result.Status.Branch))
			} else {
				fmt.Printf("  %s %s %s\n", red("✱"), green(letter), displayName)
			}
		} else if result.Status.Symbol == "⬆" && result.Status.Branch != "" {
			fmt.Printf("  %s %s - %s\n", green(result.Status.Symbol), displayName, blue(result.Status.Branch))
		} else if result.Status.Branch != "" {
			message := fmt.Sprintf("%s %s", result.Status.Symbol, displayName)
			fmt.Printf("  %s - %s\n", red(message), blue(result.Status.Branch))
		} else {
			message := fmt.Sprintf("%s %s", result.Status.Symbol, displayName)
			fmt.Printf("  %s\n", red(message))
		}
		r.displayBehindBranches(result)
	case git.StatusError:
		message := fmt.Sprintf("%s %s", result.Status.Symbol, displayName)
		fmt.Printf("  %s\n", red(message))
		r.displayBehindBranches(result)
	case git.StatusBrokenSymlink:
		message := fmt.Sprintf("🔗 ✗ %s (broken symlink)", displayName)
		fmt.Printf("  %s\n", red(message))
	case git.StatusNoUpstream:
		message := fmt.Sprintf("%s %s", result.Status.Symbol, displayName)
		fmt.Printf("  %s\n", message)
		r.displayBehindBranches(result)
	default:
		message := fmt.Sprintf("%s %s", result.Status.Symbol, displayName)
		fmt.Printf("  %s\n", message)
		r.displayBehindBranches(result)
	}
}

func (r *Reporter) displayBehindBranches(result ProjectResult) {
	if len(result.Status.BehindBranches) > 0 {
		for _, branch := range result.Status.BehindBranches {
			fmt.Printf("    %s %s: %s\n", red("↓"), branch.Branch, branch.Message)
		}
	}
}
