package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/akhilsharma/redteam-box/internal/agents"
)

// Sidebar widths are constants — the transcript pane gets the rest.
const (
	leftPaneWidth  = 28
	rightPaneWidth = 26
)

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "starting…"
	}

	// Header line
	title := headerStyle.Render(fmt.Sprintf(" redteam-box · %s · %s ",
		m.patternName, m.targetName))
	subtitle := subHeaderStyle.Render("goal: " + truncate(m.goal, m.width-2))
	header := lipgloss.JoinVertical(lipgloss.Left, title, subtitle)

	// Three panes (left/middle/right)
	left := m.renderStatePane()
	middle := m.renderTranscriptPane()
	right := m.renderScoreboardPane()

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, middle, right)

	// Footer with keybindings
	footer := footerStyle.Width(m.width).Render(m.renderFooter())

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m model) renderFooter() string {
	status := statusRunning
	if m.paused {
		status = statusPaused
	}
	if m.finished {
		status = statusDone
	}
	keys := mutedStyle.Render("q quit · space pause · s step · e expand · j/k scroll · g/G top/bottom")
	return status + "  " + keys
}

func (m model) renderStatePane() string {
	turnLine := fmt.Sprintf("turn %d", m.currentTurn)
	if !m.finished && m.awaitingLLM {
		turnLine += " " + m.spin.View() + mutedStyle.Render(" thinking…")
	}

	lines := []string{
		paneTitleStyle.Render("STATE"),
		"",
		"pattern:  " + m.patternName,
		"target:   " + truncate(m.targetName, leftPaneWidth-12),
		turnLine,
		"best:     " + scoreBadge(m.bestScore),
	}
	if m.panel != "" {
		lines = append(lines, "", paneTitleStyle.Render("PANEL"))
		for _, p := range splitWrap(m.panel, leftPaneWidth-2) {
			lines = append(lines, p)
		}
	}
	if m.plan != "" {
		lines = append(lines, "", paneTitleStyle.Render("PLAN"))
		for _, p := range splitWrap(m.plan, leftPaneWidth-2) {
			lines = append(lines, p)
		}
	}

	content := strings.Join(lines, "\n")
	return paneBorderStyle.
		Width(leftPaneWidth).
		Height(m.bodyHeight()).
		Render(content)
}

func (m model) renderTranscriptPane() string {
	w := m.width - leftPaneWidth - rightPaneWidth - 8
	if w < 30 {
		w = m.width - 4
	}
	return paneBorderStyle.
		Width(w).
		Height(m.bodyHeight()).
		Render(m.body.View())
}

func (m model) renderScoreboardPane() string {
	lines := []string{paneTitleStyle.Render("SCOREBOARD"), ""}

	// Per-attacker average over the campaign so far.
	stats := map[string][]int{}
	for _, t := range m.turns {
		stats[t.Attacker] = append(stats[t.Attacker], t.Score)
	}
	if len(stats) == 0 {
		lines = append(lines, mutedStyle.Render("(no turns yet)"))
	} else {
		for name, scores := range stats {
			avg := 0
			for _, s := range scores {
				avg += s
			}
			avg /= len(scores)
			label := truncate(personaShort(name), rightPaneWidth-12)
			lines = append(lines, fmt.Sprintf("%-12s %s", label, sparkline(scores)))
			lines = append(lines, mutedStyle.Render(fmt.Sprintf("             avg %d  n=%d", avg, len(scores))))
		}
	}

	lines = append(lines, "", paneTitleStyle.Render("CAMPAIGN"),
		fmt.Sprintf("turns:    %d", len(m.turns)),
		fmt.Sprintf("best:     %d", m.bestScore))

	if m.finished && m.final != nil {
		verdict := mutedStyle.Render("× did not reach 80")
		if m.final.Succeeded {
			verdict = lipgloss.NewStyle().Foreground(colorScoreHigh).Bold(true).Render("★ succeeded")
		}
		lines = append(lines, "", verdict)
	}

	content := strings.Join(lines, "\n")
	return paneBorderStyle.
		Width(rightPaneWidth).
		Height(m.bodyHeight()).
		Render(content)
}

func (m model) bodyHeight() int {
	h := m.height - 4
	if h < 8 {
		h = 8
	}
	return h
}

// refreshTranscript rebuilds the viewport content from the turns slice.
// Each turn is rendered as a small block: header, attacker line, target
// line, score badge. The expanded turn shows full prompt + reply text;
// other turns show the truncated single-line form.
func (m *model) refreshTranscript() string {
	var b strings.Builder
	for _, t := range m.turns {
		fullText := m.expandedTurn == t.Index
		b.WriteString(renderTurn(t, fullText, m.body.Width-2))
		b.WriteString("\n")
	}
	if m.finished && m.final != nil {
		b.WriteString(renderResult(m.final, m.body.Width-2))
	}
	m.body.SetContent(b.String())
	return ""
}

func renderTurn(t agents.Turn, fullText bool, width int) string {
	header := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("turn %d", t.Index)) +
		"  " + mutedStyle.Render(t.Attacker)

	prompt := t.Prompt
	reply := t.Reply
	if !fullText {
		prompt = truncate(prompt, width-12)
		reply = truncate(reply, width-12)
	} else {
		prompt = wrap(prompt, width-12)
		reply = wrap(reply, width-12)
	}

	lines := []string{
		header,
		attackerStyle.Render("→ attacker:") + " " + prompt,
		targetStyle.Render("← target:  ") + " " + reply,
		scoreBadge(t.Score) + "  " + mutedStyle.Render(t.Reason),
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderResult(c *agents.Campaign, width int) string {
	verdict := lipgloss.NewStyle().Foreground(colorScoreMid).Bold(true).Render("× campaign did not reach success threshold")
	if c.Succeeded {
		verdict = lipgloss.NewStyle().Foreground(colorScoreHigh).Bold(true).Render("★ campaign succeeded")
	}
	return strings.Join([]string{
		"",
		verdict + "  " + mutedStyle.Render(fmt.Sprintf("best score: %d", c.BestScore)),
		mutedStyle.Render("best prompt: ") + truncate(c.BestPrompt, width-14),
		mutedStyle.Render("best reply:  ") + truncate(c.BestReply, width-14),
	}, "\n")
}

// truncate cuts a string to max length and appends an ellipsis if needed,
// also collapsing newlines so the single-line view stays single-line.
func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ⏎ ")
	s = strings.TrimSpace(s)
	if max <= 1 {
		return ""
	}
	if len([]rune(s)) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max-1]) + "…"
}

// wrap word-wraps at the given width preserving paragraphs.
func wrap(s string, width int) string {
	if width < 10 {
		width = 10
	}
	var b strings.Builder
	for i, para := range strings.Split(s, "\n") {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(wrapLine(para, width))
	}
	return b.String()
}

func wrapLine(s string, width int) string {
	words := strings.Fields(s)
	var b strings.Builder
	col := 0
	for i, w := range words {
		wl := len([]rune(w))
		if i == 0 {
			b.WriteString(w)
			col = wl
			continue
		}
		if col+1+wl > width {
			b.WriteString("\n")
			b.WriteString(w)
			col = wl
		} else {
			b.WriteString(" ")
			b.WriteString(w)
			col += 1 + wl
		}
	}
	return b.String()
}

func splitWrap(s string, width int) []string {
	wrapped := wrap(s, width)
	return strings.Split(wrapped, "\n")
}

// personaShort strips the "mixture:" / "evolutionary:gen1" / etc prefix
// and returns just the persona/generation tag for the scoreboard.
func personaShort(name string) string {
	if i := strings.Index(name, ":"); i >= 0 {
		return name[i+1:]
	}
	return name
}

// sparkline renders a list of 0-100 scores as a tiny block-character bar.
func sparkline(scores []int) string {
	if len(scores) == 0 {
		return ""
	}
	const blocks = "▁▂▃▄▅▆▇█"
	var b strings.Builder
	for _, s := range scores {
		idx := s * (len(blocks) - 1) / 100
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		c := []rune(blocks)[idx]
		col := lipgloss.NewStyle().Foreground(scoreColor(s))
		b.WriteString(col.Render(string(c)))
	}
	return b.String()
}
