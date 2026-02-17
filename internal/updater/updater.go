package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
)

const (
	githubAPIURL = "https://api.github.com/repos/uralys/check-projects/releases/latest"
	installURL   = "https://raw.githubusercontent.com/uralys/check-projects/main/install.sh"
)

var (
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// UpdateResult holds the result of an async update check
type UpdateResult struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
}

// CheckForUpdatesAsync checks for updates in the background and returns a channel
func CheckForUpdatesAsync(currentVersion string) <-chan *UpdateResult {
	ch := make(chan *UpdateResult, 1)

	go func() {
		defer close(ch)

		// Skip check if version is "dev" or empty
		if currentVersion == "" || currentVersion == "dev" || strings.Contains(currentVersion, "dirty") {
			ch <- nil
			return
		}

		latestVersion, err := getLatestVersion()
		if err != nil {
			ch <- nil
			return
		}

		current := strings.TrimPrefix(currentVersion, "v")
		latest := strings.TrimPrefix(latestVersion, "v")

		if current != latest {
			ch <- &UpdateResult{
				Available:      true,
				CurrentVersion: current,
				LatestVersion:  latest,
			}
		} else {
			ch <- &UpdateResult{Available: false}
		}
	}()

	return ch
}

// PrintUpdateNotice prints an update notice if available
func PrintUpdateNotice(result *UpdateResult) {
	if result == nil || !result.Available {
		return
	}

	fmt.Printf("\n%s %s → %s\n",
		yellow("⚠ New version available:"),
		cyan(result.CurrentVersion),
		green(result.LatestVersion))
	fmt.Printf("Run %s to update\n", cyan("check-projects --update"))
}

// CheckForUpdates checks if a new version is available (blocking, with prompt)
func CheckForUpdates(currentVersion string) error {
	// Skip check if version is "dev" or empty
	if currentVersion == "" || currentVersion == "dev" || strings.Contains(currentVersion, "dirty") {
		return nil
	}

	latestVersion, err := getLatestVersion()
	if err != nil {
		// Silently fail - don't block the user
		return nil
	}

	// Normalize versions (remove 'v' prefix if present)
	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(latestVersion, "v")

	if current == latest {
		fmt.Printf("%s %s\n", green("✔"), fmt.Sprintf("Already up to date (%s)", cyan(current)))
		return nil
	}

	fmt.Printf("\n%s %s → %s\n",
		yellow("⚠ New version available:"),
		cyan(current),
		green(latest))

	if err := promptAndInstall(); err != nil {
		fmt.Printf("Update cancelled or failed: %v\n", err)
	}

	return nil
}

// getLatestVersion fetches the latest version from GitHub
func getLatestVersion() (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(githubAPIURL)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

// promptAndInstall prompts the user and runs the install script
func promptAndInstall() error {
	fmt.Printf("Install update? [Y/n]: ")

	// Read single character or line
	var response string
	_, _ = fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))

	// Default to "yes" if user just pressed Enter
	if response == "" || response == "y" || response == "yes" {
		return installUpdate()
	}

	return fmt.Errorf("user declined update")
}

// installUpdate downloads and runs the install script
func installUpdate() error {
	fmt.Println("\n" + cyan("→ Downloading and running installer..."))

	// Only support Unix-like systems for now
	if runtime.GOOS == "windows" {
		fmt.Println(yellow("⚠ Auto-update not supported on Windows"))
		fmt.Printf("Please download the latest version from: %s\n", githubAPIURL)
		return fmt.Errorf("unsupported platform")
	}

	// Download install script
	resp, err := http.Get(installURL)
	if err != nil {
		return fmt.Errorf("failed to download installer: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download installer: HTTP %d", resp.StatusCode)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "check-projects-install-*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// Write script to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write installer: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Make executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to make installer executable: %w", err)
	}

	// Run installer
	cmd := exec.Command("/bin/sh", tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installer failed: %w", err)
	}

	fmt.Println(green("✔ Update completed successfully!"))
	fmt.Println(cyan("→ Restart check-projects to use the new version"))

	return nil
}
