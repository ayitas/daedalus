package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Color Palette
	ColorPrimary   = lipgloss.Color("#7D56F4") // Indigo/Violet
	ColorSecondary = lipgloss.Color("#06B6D4") // Cyan
	ColorAccent    = lipgloss.Color("#F43F5E") // Rose
	ColorSuccess   = lipgloss.Color("#10B981") // Emerald Green
	ColorWarning   = lipgloss.Color("#F59E0B") // Amber
	ColorError     = lipgloss.Color("#EF4444") // Coral Red
	ColorMuted     = lipgloss.Color("#6B7280") // Slate
	ColorBg        = lipgloss.Color("#181825") // Dark Background
	ColorFg        = lipgloss.Color("#CDD6F4") // Bright Text

	// Typography & Badges
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	BadgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorPrimary).
			Padding(0, 1)

	SuccessBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorSuccess).
			Padding(0, 1)

	WarningBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(ColorWarning).
			Padding(0, 1)

	// Containers
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	MagicPromptBox = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorSecondary).
			Background(lipgloss.Color("#11111B")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	PromptTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1")). // Soft green
			Bold(true)

	StepNumberStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary)

	StepNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorFg)

	StepDescStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)
)

// RenderBanner displays the Daedalus stylized header with author attribution.
func RenderBanner(version, author string) string {
	asciiLogo := `
  ____             _       _           
 |  _ \  __ _  ___| | __ _| |_   _ ___ 
 | | | |/ _` + "`" + ` |/ _ \ |/ _` + "`" + ` | | | | / __|
 | |_| | (_| |  __/ | (_| | | |_| \__ \
 |____/ \__,_|\___|_|\__,_|_|\__,_|___/`

	banner := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Render(asciiLogo)

	tagline := SubtitleStyle.Render(fmt.Sprintf(" The Autonomous Agent Harness CLI • v%s", version))
	authorCredit := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorSecondary).
		Render(fmt.Sprintf(" ⚡ Created by @%s (github.com/%s/daedalus)", author, author))
	divider := lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("─", 56))

	return fmt.Sprintf("%s\n%s\n%s\n%s\n", banner, tagline, authorCredit, divider)
}

// RenderMagicPrompt renders the stylized handoff prompt for coding agents.
func RenderMagicPrompt(prompt string, projectDir string) string {
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorSecondary).
		Render("✨ THE MAGIC PROMPT (Copy & Paste to Claude Code, Codex, or Google Antigravity)")

	subHeader := lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render(fmt.Sprintf("Target Directory: %s", projectDir))

	content := PromptTextStyle.Render(prompt)

	inner := fmt.Sprintf("%s\n%s\n\n%s", header, subHeader, content)
	return MagicPromptBox.Render(inner)
}

// RenderSuccessCard renders a summary box after scaffolding.
func RenderSuccessCard(projectName, projectDir string, phases, tasks int, files []string) string {
	title := SuccessBadge.Render(" READY TO BUILD ") + " " +
		lipgloss.NewStyle().Bold(true).Foreground(ColorFg).Render(fmt.Sprintf("Project '%s' successfully harnessed!", projectName))

	stats := fmt.Sprintf(
		"• %s: %s\n• %s: %d\n• %s: %d\n• %s: %s",
		lipgloss.NewStyle().Bold(true).Render("Directory"), projectDir,
		lipgloss.NewStyle().Bold(true).Render("Phases"), phases,
		lipgloss.NewStyle().Bold(true).Render("Atomic Tasks"), tasks,
		lipgloss.NewStyle().Bold(true).Render("Generated Files"), strings.Join(files, ", "),
	)

	return CardStyle.Render(fmt.Sprintf("%s\n\n%s", title, stats))
}
