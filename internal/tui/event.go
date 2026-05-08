// Package tui is the Bubble Tea TUI for redteam-box. The same patterns
// that drive the CLI run unmodified — the TUI just plugs in a different
// agents.Reporter implementation that pumps events into a tea.Program.
package tui

import (
	"github.com/akhilsharma/redteam-box/internal/agents"
)

// Events flow from the pattern goroutine (via the reporter) into the
// Bubble Tea program. Each event is one logical update to the model.

type bannerMsg struct{ title string }
type sectionMsg struct{ name string }
type fieldMsg struct{ label, value string }
type planMsg struct{ plan string }
type turnMsg struct{ turn agents.Turn }
type resultMsg struct{ campaign *agents.Campaign }

// doneMsg signals that the pattern goroutine has finished (success or
// failure). The TUI uses it to switch to a final review state.
type doneMsg struct {
	campaign *agents.Campaign
	err      error
}
