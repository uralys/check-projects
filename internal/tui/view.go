package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme colors - centralized color definitions
var (
	// Status colors
	colorStatusClean  = lipgloss.Color("2")  // Green for clean/success
	colorStatusError  = lipgloss.Color("1")  // Dark red for errors/modifications
	colorStatusUnsync = lipgloss.Color("1")  // Dark red for unsync

	// UI colors
	colorTitle       = lipgloss.Color("86")  // Cyan for titles
	colorVersion     = lipgloss.Color("11")  // Yellow for version
	colorLink        = lipgloss.Color("5")   // Purple for links
	colorCategory    = lipgloss.Color("12")  // Blue for categories
	colorLabel       = lipgloss.Color("5")   // Purple for labels
	colorHelp        = lipgloss.Color("255") // White for help text
	colorBorder      = lipgloss.Color("241") // Gray for borders
	colorScrollbar   = lipgloss.Color("240") // Gray for scrollbars
	colorScrollThumb = lipgloss.Color("12")  // Blue for scroll thumb
	colorArrow       = lipgloss.Color("10")  // Green for navigation arrows

	// Background colors
	colorBackground = lipgloss.Color("237") // Dark gray for backgrounds
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorTitle).
			MarginBottom(1)

	categoryStyle = lipgloss.NewStyle().
			Foreground(colorCategory).
			PaddingLeft(2)

	selectedCategoryStyle = lipgloss.NewStyle().
				Foreground(colorCategory).
				Bold(true).
				PaddingLeft(1)

	projectStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedProjectStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Bold(true)

	statusCleanStyle = lipgloss.NewStyle().
				Foreground(colorStatusClean)

	statusUnsyncStyle = lipgloss.NewStyle().
				Foreground(colorStatusUnsync)

	statusErrorStyle = lipgloss.NewStyle().
				Foreground(colorStatusError)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorHelp).
			MarginTop(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(colorLabel)
)

// View renders the current state of the model
func (m Model) View() string {
	// Check minimum terminal size
	if m.width < 60 || m.height < 10 {
		return "Terminal too small. Minimum size: 60x10\nPress q to quit."
	}

	// Loading state
	if m.loading {
		return fmt.Sprintf("%s Loading projects...\n%s", m.spinner.View(), helpStyle.Render("\nPress q to quit"))
	}

	// Error state
	if m.errorMsg != "" {
		return fmt.Sprintf("%s\n%s", statusErrorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg)), helpStyle.Render("\nPress q to quit | r to retry"))
	}

	// Check if we have any projects at all
	hasProjects := len(m.projects) > 0

	// If no projects at all, show error/empty state
	if !hasProjects {
		return "No projects found.\n" + renderHelpBar(m)
	}

	// Get filtered projects
	filtered := m.getFilteredProjects()

	// If hiding clean and no projects with changes, show "all clean" view
	if m.hideClean && len(filtered) == 0 {
		return renderNormalView(m)
	}

	// Otherwise show split view (even if all projects are clean when not hiding)
	return renderSplitView(m)
}

func renderNormalView(m Model) string {
	var b strings.Builder

	// Top margin
	b.WriteString("\n")

	// Big centered box with "All projects are clean!"
	message := statusCleanStyle.Render("✔") + " All projects are clean!"
	messageBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(2, 4).
		Width(m.width - 10).
		Align(lipgloss.Center).
		Render(message)

	b.WriteString(messageBox)
	b.WriteString("\n")

	// Footer at the bottom
	b.WriteString("\n\n")
	b.WriteString(renderFooter(m))

	return b.String()
}

func renderSplitView(m Model) string {
	var b strings.Builder

	// Add top margin (space above header)
	b.WriteString("\n")

	// Render categories without scrollbar
	categoryTabsOnly := renderCategoryTabsOnly(m)

	// Render horizontal scrollbar for categories
	categoryScrollbar := renderCategoryHorizontalScrollbar(m, lipgloss.Width(categoryTabsOnly))

	// Build inner content with just categories (no title)
	var innerContent string
	if categoryScrollbar != "" {
		innerContent = categoryTabsOnly + "\n" + categoryScrollbar
	} else {
		innerContent = categoryTabsOnly
	}

	// Wrap categories in border - full width
	headerContent := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		PaddingLeft(2).
		PaddingRight(2).
		PaddingTop(0).
		PaddingBottom(0).
		Width(m.width - 4).
		Render(innerContent)

	b.WriteString(headerContent)
	b.WriteString("\n\n")

	// Calculate dimensions for split panels
	// Total available width for both panels
	totalPanelsWidth := m.width - 4
	leftWidth := totalPanelsWidth * 40 / 100
	rightWidth := totalPanelsWidth - leftWidth

	// Height calculation - use fixed reserved space like width
	// Reserve: top margin (1) + header box (~4-5) + blank lines (2) + footer (2) = ~10 lines
	reservedHeight := 10

	// Remaining height for panels (including their borders)
	panelTotalHeight := m.height - reservedHeight

	// Panel content height (subtract 2 for top/bottom border)
	contentHeight := panelTotalHeight - 2

	// Safety check
	if contentHeight < 3 {
		contentHeight = 3
		panelTotalHeight = 5
	}

	// Left panel - projects list (scrollable)
	// Subtract 4 from width to account for border (2) + padding (2)
	leftPanel := renderProjectsList(m, leftWidth-4, contentHeight)

	// Right panel - git status details
	// Subtract 4 from width to account for border (2) + padding (2)
	rightPanel := renderDetailsPanel(m, rightWidth-4, contentHeight)

	// Use lipgloss to properly handle ANSI codes and join horizontally
	// Set explicit height to prevent overflow
	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(panelTotalHeight).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	rightStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(panelTotalHeight).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	if !m.focusedPanel {
		// Projects panel is focused
		leftStyle = leftStyle.BorderForeground(colorScrollThumb)
		rightStyle = rightStyle.BorderForeground(colorScrollbar) // Dim border
	} else {
		// Details panel is focused
		leftStyle = leftStyle.BorderForeground(colorScrollbar) // Dim border
		rightStyle = rightStyle.BorderForeground(colorScrollThumb)
	}

	// Calculate spacing needed to align right panel with header
	leftPanelRendered := leftStyle.Render(leftPanel)
	rightPanelRendered := rightStyle.Render(rightPanel)

	// Gap between left panel and right panel to align with header
	gap := m.width - lipgloss.Width(leftPanelRendered) - lipgloss.Width(rightPanelRendered)
	if gap < 0 {
		gap = 0
	}

	splitContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanelRendered,
		strings.Repeat(" ", gap),
		rightPanelRendered,
	)

	b.WriteString(splitContent)

	// Footer - title with version, then help bar
	b.WriteString("\n\n")
	b.WriteString(renderFooter(m))

	return b.String()
}

func renderProjectsList(m Model, width, height int) string {
	filtered := m.getFilteredProjects()

	// Use the provided height directly for available space
	availableHeight := height
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Calculate scroll window
	startIdx := 0
	endIdx := len(filtered)
	needsScroll := len(filtered) > availableHeight

	if needsScroll {
		// Need to scroll - center the selected project
		halfHeight := availableHeight / 2
		startIdx = m.selectedProject - halfHeight
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + availableHeight
		if endIdx > len(filtered) {
			endIdx = len(filtered)
			startIdx = endIdx - availableHeight
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	// Build project lines
	var lines []string
	for i := startIdx; i < endIdx && i < len(filtered); i++ {
		p := filtered[i]
		style := projectStyle
		prefix := "  "
		if i == m.selectedProject {
			style = selectedProjectStyle
			prefix = "> "
		}

		statusSymbol := "?"
		statusStyle := lipgloss.NewStyle()
		if p.Status != nil {
			statusSymbol = p.Status.Symbol
			switch p.Status.Type {
			case "sync":
				statusStyle = statusCleanStyle
			case "unsync":
				// Special case: if symbol is ⬆ (ahead of remote), use green
				if statusSymbol == "⬆" {
					statusStyle = statusCleanStyle
				} else {
					statusStyle = statusUnsyncStyle
				}
			case "error":
				statusStyle = statusErrorStyle
			}
		}

		line := fmt.Sprintf("%s%s %s", prefix, statusStyle.Render(statusSymbol), style.Render(p.Project.Name))
		lines = append(lines, line)
	}

	// Truncate if somehow we have too many lines
	if len(lines) > availableHeight {
		lines = lines[:availableHeight]
	}

	// Pad with empty lines if needed to fill available height
	for len(lines) < availableHeight {
		lines = append(lines, "")
	}

	// Add scroll indicator on the left (always show for consistency)
	scrollbarStyle := lipgloss.NewStyle().Foreground(colorBorder)
	for lineIdx := range lines {
		scrollChar := " "
		if needsScroll {
			// Calculate scrollbar position - show thumb at selected position
			scrollChar = "│"
			if len(filtered) > 0 {
				// Calculate where the thumb should be based on selected project position
				selectedPercentage := float64(m.selectedProject) / float64(len(filtered)-1)
				thumbPosition := int(selectedPercentage * float64(availableHeight-1))

				if lineIdx == thumbPosition {
					scrollChar = "█"
				}
			}
		}
		lines[lineIdx] = scrollbarStyle.Render(scrollChar) + " " + lines[lineIdx]
	}

	return strings.Join(lines, "\n")
}

func renderDetailsPanel(m Model, width, height int) string {
	filtered := m.getFilteredProjects()
	if len(filtered) == 0 || m.selectedProject >= len(filtered) {
		return ""
	}

	selectedProj := filtered[m.selectedProject]

	// Build content lines
	var contentLines []string

	// Path
	contentLines = append(contentLines, labelStyle.Render(selectedProj.Project.Path))
	contentLines = append(contentLines, "") // Empty line

	// Always check remote status first
	remoteStatus := getRemoteStatus(selectedProj.Project.Path)

	// Show git status --short output for non-clean projects
	if selectedProj.Status != nil && selectedProj.Status.Type != "sync" {
		// Get branch name
		branchName := getGitBranch(selectedProj.Project.Path)
		if branchName != "" {
			contentLines = append(contentLines, labelStyle.Render(fmt.Sprintf("[%s]", branchName)))
		}

		gitOutput := getGitStatusShort(selectedProj.Project.Path)
		if gitOutput != "" {
			// Split git output into lines
			gitLines := strings.Split(colorizeGitStatus(gitOutput), "\n")
			contentLines = append(contentLines, gitLines...)
		} else {
			// No local changes but status is unsync - likely ahead of remote
			contentLines = append(contentLines, statusCleanStyle.Render("✔")+" No local changes")

			// Check remote status
			if remoteStatus.HasRemote {
				if remoteStatus.AheadCount > 0 {
					contentLines = append(contentLines, statusCleanStyle.Render("⬆")+" Ready to be pushed")
				} else if remoteStatus.RemoteAhead {
					contentLines = append(contentLines, statusErrorStyle.Render("↓")+" Remote is ahead: Requires pull")
				}
			}
		}

		// Always show remote status check for modified files too
		if remoteStatus.HasRemote && gitOutput != "" {
			contentLines = append(contentLines, "") // Empty line
			if remoteStatus.RemoteAhead {
				contentLines = append(contentLines, statusErrorStyle.Render("↓")+" Remote is ahead: Requires pull")
			} else if remoteStatus.AheadCount > 0 {
				contentLines = append(contentLines, statusCleanStyle.Render("⬆")+" Also ahead of remote")
			}
		}
	} else {
		// Project is clean - show remote status
		// Show local status
		contentLines = append(contentLines, statusCleanStyle.Render("✔")+" No local changes")

		// Show remote status if available
		if remoteStatus.HasRemote {
			if remoteStatus.IsUpToDate {
				contentLines = append(contentLines, statusCleanStyle.Render("✔")+" Up to date with remote")
			} else if remoteStatus.RemoteAhead {
				contentLines = append(contentLines, statusErrorStyle.Render("↓")+" Remote is ahead: Requires pull")
			} else if remoteStatus.AheadCount > 0 {
				contentLines = append(contentLines, statusCleanStyle.Render("⬆")+" Ready to be pushed")
			}
		}
	}

	// Calculate scroll window
	availableHeight := height
	if availableHeight < 1 {
		availableHeight = 1
	}

	startIdx := m.detailsScroll
	endIdx := startIdx + availableHeight
	needsScroll := len(contentLines) > availableHeight

	// Adjust scroll bounds
	if endIdx > len(contentLines) {
		endIdx = len(contentLines)
		startIdx = endIdx - availableHeight
		if startIdx < 0 {
			startIdx = 0
		}
	}

	// Extract visible lines - ensure we never exceed availableHeight
	var visibleLines []string
	if len(contentLines) > 0 {
		visibleLines = contentLines[startIdx:endIdx]
	}

	// Truncate if somehow we have too many lines
	if len(visibleLines) > availableHeight {
		visibleLines = visibleLines[:availableHeight]
	}

	// Pad with empty lines if needed to fill available height
	for len(visibleLines) < availableHeight {
		visibleLines = append(visibleLines, "")
	}

	// Add scroll indicator on the right (always show for consistency)
	scrollbarStyle := lipgloss.NewStyle().Foreground(colorBorder)

	// Ensure lines don't exceed width
	maxLineWidth := width - 2 // Reserve space for scrollbar

	for lineIdx := range visibleLines {
		scrollChar := " "
		if needsScroll {
			scrollChar = "│"
			if len(contentLines) > 1 {
				// Calculate where the thumb should be based on current scroll position
				scrollPercentage := float64(startIdx) / float64(len(contentLines)-availableHeight)
				thumbPosition := int(scrollPercentage * float64(availableHeight-1))

				if lineIdx == thumbPosition {
					scrollChar = "█"
				}
			}
		}

		// Truncate line if too long (accounting for ANSI codes)
		lineWidth := lipgloss.Width(visibleLines[lineIdx])
		if lineWidth > maxLineWidth {
			// Truncate the visible line to fit
			visibleLines[lineIdx] = truncateLine(visibleLines[lineIdx], maxLineWidth)
		}

		// Pad line to consistent width and add scrollbar on the right
		padding := maxLineWidth - lipgloss.Width(visibleLines[lineIdx])
		if padding > 0 {
			visibleLines[lineIdx] = visibleLines[lineIdx] + strings.Repeat(" ", padding)
		}
		visibleLines[lineIdx] = visibleLines[lineIdx] + " " + scrollbarStyle.Render(scrollChar)
	}

	return strings.Join(visibleLines, "\n")
}

func renderCategoryTabsOnly(m Model) string {
	// Get visible categories (filtered by hideClean)
	visibleCategories := m.getVisibleCategories()

	// Build all category tabs
	var allTabs []string
	for i, cat := range visibleCategories {
		// Check if category has changes
		hasChanges := m.categoryHasChanges(cat)

		// Determine status symbol
		var symbol string
		if hasChanges {
			symbol = "*"
		} else {
			symbol = "✔"
		}

		// Apply style
		style := categoryStyle
		if i == m.selectedCategory {
			style = selectedCategoryStyle
		}

		// Build the tab with symbol inside brackets for selected, outside for others
		var tab string
		if i == m.selectedCategory {
			// Selected: [*core] or [✔core]
			if hasChanges {
				tab = "[" + statusErrorStyle.Render(symbol) + style.Render(cat) + "]"
			} else {
				tab = "[" + statusCleanStyle.Render(symbol) + style.Render(cat) + "]"
			}
		} else {
			// Not selected: *core or ✔core
			if hasChanges {
				tab = statusErrorStyle.Render(symbol) + style.Render(cat)
			} else {
				tab = statusCleanStyle.Render(symbol) + style.Render(cat)
			}
		}

		allTabs = append(allTabs, tab)
	}

	// Add scroll indicators if selected category is not at edges
	// Check if we can navigate left/right to visible categories
	visibleCategories = m.getVisibleCategories()

	leftArrow := ""
	rightArrow := ""
	arrowStyleGreen := lipgloss.NewStyle().Foreground(colorArrow) // Green

	// Find current position in visible categories
	currentIndex := -1
	if m.selectedCategory < len(m.categories) {
		currentCategory := m.categories[m.selectedCategory]
		for i, category := range visibleCategories {
			if category == currentCategory {
				currentIndex = i
				break
			}
		}
	}

	// Left arrow: green if we can go left in visible categories
	if currentIndex > 0 {
		leftArrow = arrowStyleGreen.Render("◀ ")
	} else {
		leftArrow = "  "
	}

	// Right arrow: green if we can go right in visible categories
	if currentIndex >= 0 && currentIndex < len(visibleCategories)-1 {
		rightArrow = arrowStyleGreen.Render(" ▶")
	} else {
		rightArrow = "  "
	}

	// Join tabs and add scroll indicators
	tabsLine := leftArrow + strings.Join(allTabs, "  ") + rightArrow

	return tabsLine
}

func renderCategoryHorizontalScrollbar(m Model, width int) string {
	visibleCategories := m.getVisibleCategories()
	if len(visibleCategories) <= 1 {
		return ""
	}

	totalCategories := len(visibleCategories)

	// Find index of selected category in visible categories
	selectedIndex := 0
	if m.selectedCategory < len(m.categories) {
		selectedCat := m.categories[m.selectedCategory]
		for i, cat := range visibleCategories {
			if cat == selectedCat {
				selectedIndex = i
				break
			}
		}
	}

	thumbWidth := width / totalCategories
	if thumbWidth < 2 {
		thumbWidth = 2
	}

	thumbPosition := (selectedIndex * width) / totalCategories

	scrollbarStyle := lipgloss.NewStyle().Foreground(colorScrollbar)
	thumbStyle := lipgloss.NewStyle().Foreground(colorScrollThumb)

	var scrollbar strings.Builder
	for i := 0; i < width; i++ {
		if i >= thumbPosition && i < thumbPosition+thumbWidth {
			scrollbar.WriteString(thumbStyle.Render("━"))
		} else {
			scrollbar.WriteString(scrollbarStyle.Render("─"))
		}
	}

	return scrollbar.String()
}

func renderCategoryTabsNoBorder(m Model) string {
	tabsLine := renderCategoryTabsOnly(m)
	scrollbarLine := renderCategoryHorizontalScrollbar(m, lipgloss.Width(tabsLine))

	// Combine tabs and scrollbar vertically
	var tabsContent string
	if scrollbarLine != "" {
		tabsContent = tabsLine + "\n" + scrollbarLine
	} else {
		tabsContent = tabsLine
	}

	return tabsContent
}

func renderCategoryTabs(m Model) string {
	content := renderCategoryTabsNoBorder(m)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1)

	return borderStyle.Render(content)
}

func renderProjectList(m Model, projects []ProjectWithStatus) string {
	var b strings.Builder

	for i, p := range projects {
		style := projectStyle
		if i == m.selectedProject {
			style = selectedProjectStyle
		}

		// Render status symbol
		statusSymbol := "?"
		statusStyle := lipgloss.NewStyle()
		if p.Status != nil {
			statusSymbol = p.Status.Symbol
			switch p.Status.Type {
			case "sync":
				statusStyle = statusCleanStyle
			case "unsync":
				// Special case: if symbol is ⬆ (ahead of remote), use green
				if statusSymbol == "⬆" {
					statusStyle = statusCleanStyle
				} else {
					statusStyle = statusUnsyncStyle
				}
			case "error":
				statusStyle = statusErrorStyle
			}
		}

		line := fmt.Sprintf("%s %s", statusStyle.Render(statusSymbol), p.Project.Name)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	return b.String()
}

func renderFooter(m Model) string {
	var footer strings.Builder

	// Title and version line
	if m.version != "" {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(colorTitle)
		versionStyle := lipgloss.NewStyle().Foreground(colorVersion) // Yellow
		linkStyle := lipgloss.NewStyle().Foreground(colorLink)     // Purple

		titleLine := titleStyle.Render("check-projects") + " | " + versionStyle.Render(m.version) + " | " + linkStyle.Render("https://github.com/uralys/check-projects")
		footer.WriteString(titleLine)
	}

	// Help bar on same line
	footer.WriteString("  ")
	footer.WriteString(renderHelpBar(m))

	return footer.String()
}

func renderHelpBar(m Model) string {
	help := "q/esc: quit | ↑↓: scroll | ←→: categories | enter: switch panel | h: toggle clean | r: refresh"
	if m.hideClean {
		help = strings.Replace(help, "toggle clean", "show clean", 1)
	} else {
		help = strings.Replace(help, "toggle clean", "hide clean", 1)
	}

	return helpStyle.Render(help)
}

// getGitBranch returns the current git branch name
func getGitBranch(projectPath string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// RemoteStatus represents the status of the local branch relative to remote
type RemoteStatus struct {
	HasRemote     bool
	IsUpToDate    bool
	RemoteAhead   bool
	AheadCount    int
	BehindCount   int
	HasLocalDiffs bool
}

// getRemoteStatus checks if local branch is ahead/behind remote
func getRemoteStatus(projectPath string) RemoteStatus {
	status := RemoteStatus{
		HasRemote:     false,
		IsUpToDate:    false,
		RemoteAhead:   false,
		HasLocalDiffs: false,
	}

	// Check if there are local uncommitted changes
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		status.HasLocalDiffs = true
	}

	// Check if branch has an upstream
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "@{u}")
	cmd.Dir = projectPath
	_, err = cmd.CombinedOutput()
	if err != nil {
		// No upstream configured
		return status
	}
	status.HasRemote = true

	// Get ahead/behind counts
	// Commits ahead: local commits not in remote
	cmd = exec.Command("git", "rev-list", "--count", "@{u}..HEAD")
	cmd.Dir = projectPath
	output, err = cmd.CombinedOutput()
	if err == nil {
		fmt.Sscanf(string(output), "%d", &status.AheadCount)
	}

	// Commits behind: remote commits not in local
	cmd = exec.Command("git", "rev-list", "--count", "HEAD..@{u}")
	cmd.Dir = projectPath
	output, err = cmd.CombinedOutput()
	if err == nil {
		fmt.Sscanf(string(output), "%d", &status.BehindCount)
	}

	// Determine overall status
	if status.BehindCount > 0 {
		status.RemoteAhead = true
	}
	if status.AheadCount == 0 && status.BehindCount == 0 {
		status.IsUpToDate = true
	}

	return status
}

// getGitStatusShort returns the output of git status --short
func getGitStatusShort(projectPath string) string {
	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// truncateLine truncates a line to maxWidth, preserving ANSI codes
func truncateLine(line string, maxWidth int) string {
	if lipgloss.Width(line) <= maxWidth {
		return line
	}

	// Simple truncation - just cut the string and add ellipsis
	// This is a simplified approach; for perfect ANSI handling we'd need more complex logic
	result := ""
	currentWidth := 0
	inAnsi := false

	for _, r := range line {
		if r == '\x1b' {
			inAnsi = true
		}

		if inAnsi {
			result += string(r)
			if r == 'm' {
				inAnsi = false
			}
			continue
		}

		if currentWidth >= maxWidth-1 {
			result += "…"
			break
		}

		result += string(r)
		currentWidth++
	}

	return result
}

// colorizeGitStatus adds colors to git status output similar to bash
func colorizeGitStatus(gitOutput string) string {
	lines := strings.Split(gitOutput, "\n")
	var coloredLines []string

	for _, line := range lines {
		if len(line) < 4 {
			coloredLines = append(coloredLines, line)
			continue
		}

		// Git status --short format: XY filename
		// X = index status, Y = worktree status
		// Position 0: index status
		// Position 1: worktree status
		// Position 2: space
		// Position 3+: filename
		indexStatus := string(line[0])
		workTreeStatus := string(line[1])
		fileName := strings.TrimLeft(line[2:], " ")

		var coloredLine string
		statusCodes := line[0:2]

		// Color based on git status codes
		// Green for staged/added
		if indexStatus == "A" || indexStatus == "M" || indexStatus == "D" {
			statusPart := lipgloss.NewStyle().Foreground(colorStatusClean).Render(statusCodes) // Green
			coloredLine = statusPart + " " + fileName
		} else if indexStatus == "?" && workTreeStatus == "?" {
			// Red for untracked
			statusPart := lipgloss.NewStyle().Foreground(colorStatusError).Render("??") // Red
			coloredLine = statusPart + " " + fileName
		} else if workTreeStatus == "M" || workTreeStatus == "D" {
			// Red for modified/deleted in worktree
			statusPart := lipgloss.NewStyle().Foreground(colorStatusError).Render(statusCodes) // Red
			coloredLine = statusPart + " " + fileName
		} else {
			coloredLine = line
		}

		coloredLines = append(coloredLines, coloredLine)
	}

	return strings.Join(coloredLines, "\n")
}
