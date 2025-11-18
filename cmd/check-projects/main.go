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
	"github.com/uralys/check-projects/internal/updater"
)

var (
	configPath string
	verbose    bool
	category   string

	// Version information (set by ldflags during build)
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "check-projects",
		Short: "Check git status of multiple projects",
		Long:  `A tool to quickly check the git status of all your projects organized by categories.`,
		RunE:  run,
	}

	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "Config file path (default: ./check-projects.yml or ~/check-projects.yml)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show all projects including clean ones")
	rootCmd.Flags().StringVar(&category, "category", "", "Only check projects in this category")
	rootCmd.Version = fmt.Sprintf("%s (built: %s)", Version, BuildTime)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Check for updates (non-blocking)
	updater.CheckForUpdates(Version)

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

	// Scan for projects
	fmt.Println("Processing projects...")
	s := scanner.NewScanner(cfg)
	projects, err := s.ScanAll()
	if err != nil {
		return fmt.Errorf("failed to scan projects: %w", err)
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
				Name:     proj.Name,
				Status:   status,
				Category: proj.Category,
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

	return nil
}

func handleNoUpstream(cfg *config.Config, projects []scanner.Project, results []reporter.ProjectResult) error {
	for i, result := range results {
		if result.Status.Type == git.StatusNoUpstream {
			fmt.Printf("\n⚠ Repository '%s' has no upstream configured. Setting upstream...\n", result.Name)

			// Automatically try to set upstream
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
				fmt.Printf("✅ Upstream configured successfully\n")
			}
		}
	}
	return nil
}
