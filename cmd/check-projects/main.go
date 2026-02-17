package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/uralys/check-projects/internal/config"
	"github.com/uralys/check-projects/internal/git"
	"github.com/uralys/check-projects/internal/reporter"
	"github.com/uralys/check-projects/internal/scanner"
	"github.com/uralys/check-projects/internal/tui"
	"github.com/uralys/check-projects/internal/updater"
)

var (
	configPath  string
	verbose     bool
	category    string
	useTUI      bool
	fetchFlag   bool
	updateFlag  bool

	// Version information (set by ldflags during build)
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "check-projects",
		Short: "Check git status of multiple projects",
		Long:  buildLongDescription(),
		RunE:  run,
	}

	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "Config file path (default: ./check-projects.yml or ~/check-projects.yml)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show all projects including clean ones")
	rootCmd.Flags().StringVar(&category, "category", "", "Only check projects in this category")
	rootCmd.Flags().BoolVar(&useTUI, "tui", false, "Use interactive TUI mode")
	rootCmd.Flags().BoolVarP(&fetchFlag, "fetch", "f", false, "Fetch from remote before checking status")
	rootCmd.Flags().BoolVar(&updateFlag, "update", false, "Check for updates and install if available")
	rootCmd.Version = fmt.Sprintf("%s (built: %s)", Version, BuildTime)

	// Customize help template with colors
	rootCmd.SetUsageTemplate(getColoredUsageTemplate())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getColoredUsageTemplate() string {
	purple := "\033[95m"
	reset := "\033[0m"

	return purple + "Usage:" + reset + `{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

` + purple + "Flags:" + reset + `
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}

func buildLongDescription() string {
	purple := "\033[95m" // Purple (same as colorLabel in TUI)
	blue := "\033[94m"   // Blue (same as colorCategory in TUI)
	reset := "\033[0m"

	description := purple + "check-projects:" + reset
	description += "\n  A tool to quickly check the git status of all your projects organized by categories."
	description += "\n  " + blue + "Created by @uralys https://github.com/uralys/check-projects" + reset

	// Add version info
	description += fmt.Sprintf("\n  Version: %s (built: %s)", Version, BuildTime)

	// Note: Update check moved to run() to avoid blocking startup

	return description
}

func run(cmd *cobra.Command, args []string) error {
	// Handle --update flag: blocking check + install prompt
	if updateFlag {
		return updater.CheckForUpdates(Version)
	}

	// Check for updates in background (truly non-blocking)
	updateCh := updater.CheckForUpdatesAsync(Version)

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Filter by category if specified
	if category != "" {
		var filteredCategories []config.Category
		for _, cat := range cfg.Categories {
			if cat.Name == category {
				filteredCategories = append(filteredCategories, cat)
			}
		}
		if len(filteredCategories) == 0 {
			return fmt.Errorf("category '%s' not found in config", category)
		}
		cfg.Categories = filteredCategories
		cfg.IsFiltered = true // Mark as filtered to prevent saving
	}

	// Determine if we should use TUI mode
	// Command line flag overrides config
	shouldUseTUI := useTUI || cfg.UseTUIByDefault

	// Determine if we should fetch
	// Command line flag overrides config
	shouldFetch := fetchFlag || cfg.Fetch

	// Use TUI mode if enabled
	if shouldUseTUI {
		return tui.Run(cfg, Version)
	}

	// Scan for projects
	fmt.Println("Processing projects...")
	s := scanner.NewScanner(cfg)
	projects, err := s.ScanAll()
	if err != nil {
		return fmt.Errorf("failed to scan projects: %w", err)
	}

	// Fetch from remote if enabled
	if shouldFetch {
		fetchProjects(projects, cfg.FetchConcurrency)
	}

	// Check git status for each project concurrently
	results := make([]reporter.ProjectResult, len(projects))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit concurrency to 10

	for i, project := range projects {
		wg.Add(1)
		go func(idx int, proj scanner.Project) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			if proj.Repository == nil {
				results[idx] = reporter.ProjectResult{
					Name:          proj.Name,
					Status:        &git.Status{Type: git.StatusBrokenSymlink, Symbol: "🔗 ✗"},
					Category:      proj.Category,
					IsSymlink:     proj.IsSymlink,
					SymlinkTarget: proj.SymlinkTarget,
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

			results[idx] = reporter.ProjectResult{
				Name:          proj.Name,
				Status:        status,
				Category:      proj.Category,
				IsSymlink:     proj.IsSymlink,
				SymlinkTarget: proj.SymlinkTarget,
			}
		}(i, project)
	}

	wg.Wait()

	// Generate report first (show all categories)
	rep := reporter.NewReporter(cfg, verbose)
	rep.Report(results)

	// Handle repositories without upstream after the report
	if err := handleNoUpstream(cfg, projects, results); err != nil {
		return err
	}

	// Check if update is available (non-blocking read)
	select {
	case result := <-updateCh:
		updater.PrintUpdateNotice(result)
	default:
		// Update check still in progress, skip notification
	}

	return nil
}

func fetchProjects(projects []scanner.Project, concurrency int) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, concurrency)

	total := len(projects)
	completed := 0

	printProgress := func() {
		barWidth := 20
		progress := float64(completed) / float64(total)
		filled := int(progress * float64(barWidth))

		bar := ""
		for i := 0; i < barWidth; i++ {
			if i < filled {
				bar += "█"
			} else {
				bar += "░"
			}
		}

		fmt.Printf("\rFetching [%s] %d/%d projects", bar, completed, total)
	}

	printProgress()

	for _, project := range projects {
		wg.Add(1)
		go func(proj scanner.Project) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			if proj.Repository != nil {
				_ = proj.Repository.Fetch()
			}

			mu.Lock()
			completed++
			printProgress()
			mu.Unlock()
		}(project)
	}

	wg.Wait()
	fmt.Println() // New line after progress bar completes
}

func handleNoUpstream(cfg *config.Config, projects []scanner.Project, results []reporter.ProjectResult) error {
	for i, result := range results {
		if result.Status.Type == git.StatusNoUpstream {
			branchName := "unknown"
			if branch, err := projects[i].Repository.GetCurrentBranch(); err == nil {
				branchName = branch
			}
			fmt.Printf("\n🧚🏻‍♀️ Repository '%s' has no upstream configured for branch '\033[95m%s\033[0m'.\n", result.Name, branchName)
			fmt.Printf("\033[38;5;208mSet upstream tracking locally?\033[0m \033[92m(Y/n):\033[0m ")

			var response string
			if _, err := fmt.Scanln(&response); err != nil {
				// Enter pressed without input - default to yes
				response = "y"
			}

			if response == "n" || response == "N" {
				continue
			}

			// Try to set upstream locally
			if err := projects[i].Repository.SetUpstream(); err != nil {
				// Failed - prompt user to ignore
				fmt.Printf("❌ Failed to set upstream: %v\n", err)
				fmt.Printf("Ignore this project? (y/n): ")

				var response string
				if _, err := fmt.Scanln(&response); err != nil {
					continue
				}

				if response == "y" || response == "Y" {
					// Check if config is filtered (--category used)
					if cfg.IsFiltered {
						fmt.Printf("⚠ Cannot ignore project when using --category flag.\n")
						fmt.Printf("   Run without --category to ignore projects.\n")
						continue
					}

					// Add to ignored list of the project's category
					projectName := result.Name
					categoryName := result.Category

					// Find the category and add to its ignore list
					for j := range cfg.Categories {
						if cfg.Categories[j].Name == categoryName {
							cfg.Categories[j].Ignore = append(cfg.Categories[j].Ignore, projectName)
							break
						}
					}

					// Save config
					if err := config.SaveConfig(cfg); err != nil {
						fmt.Printf("❌ Failed to save config: %v\n", err)
						continue
					}

					fmt.Printf("✅ Project '%s' added to ignore list in category '%s' in %s\n", projectName, categoryName, cfg.ConfigPath)
					results[i].Status.Type = git.StatusIgnored
				} else {
					fmt.Printf("Skipped.\n")
				}
			} else {
				// Success - re-check status
				newStatus, err := projects[i].Repository.GetStatus()
				if err != nil {
					return fmt.Errorf("failed to get updated status: %w", err)
				}
				results[i].Status = newStatus
				fmt.Printf("✅ Upstream configured \033[92msuccessfully\033[0m\n")
			}
		}
	}
	return nil
}
