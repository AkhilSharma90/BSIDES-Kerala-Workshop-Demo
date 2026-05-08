package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color palette — kept narrow so the TUI looks consistent across themes.
var (
	colorAccent    = lipgloss.Color("#7d56f4") // purple
	colorMuted     = lipgloss.Color("#7c7c7c")
	colorAttacker  = lipgloss.Color("#5fafff") // blue
	colorTarget    = lipgloss.Color("#d787ff") // magenta
	colorScoreLow  = lipgloss.Color("#ff6b6b")
	colorScoreMid  = lipgloss.Color("#ffd166")
	colorScoreHigh = lipgloss.Color("#06d6a0")
	colorBorder    = lipgloss.Color("#444444")
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(colorAccent).
			Padding(0, 1)

	subHeaderStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1)

	paneTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			Underline(true)

	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	attackerStyle = lipgloss.NewStyle().Foreground(colorAttacker)
	targetStyle   = lipgloss.NewStyle().Foreground(colorTarget)

	paneBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Background(lipgloss.Color("#1e1e1e")).
			Padding(0, 1)

	statusRunning = lipgloss.NewStyle().Foreground(colorScoreHigh).Bold(true).Render("● running")
	statusPaused  = lipgloss.NewStyle().Foreground(colorScoreMid).Bold(true).Render("⏸ paused ")
	statusDone    = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("✓ done   ")
)

// scoreColor picks a color for the 0-100 score badge.
func scoreColor(score int) lipgloss.Color {
	switch {
	case score >= 80:
		return colorScoreHigh
	case score >= 50:
		return colorScoreMid
	default:
		return colorScoreLow
	}
}

// scoreBadge renders "score N" right-aligned in a color matching the score.
func scoreBadge(score int) string {
	return lipgloss.NewStyle().
		Foreground(scoreColor(score)).
		Bold(true).
		Render(fmt.Sprintf("score %3d", score))
}
