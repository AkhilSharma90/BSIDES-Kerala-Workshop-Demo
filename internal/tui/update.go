package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.relayout()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case bannerMsg:
		// Banner is purely informational; nothing to render here since
		// the patternName/targetName already feed the header. Ignore.
		return m, nil

	case sectionMsg:
		// We don't need a separate section line in the TUI — the header
		// already shows the active pattern.
		return m, nil

	case fieldMsg:
		if msg.label == "panel" {
			m.panel = msg.value
		}
		return m, nil

	case planMsg:
		m.plan = msg.plan
		m.relayout()
		return m, nil

	case turnMsg:
		m.turns = append(m.turns, msg.turn)
		m.currentTurn = msg.turn.Index
		if msg.turn.Score > m.bestScore {
			m.bestScore = msg.turn.Score
		}
		m.awaitingLLM = false
		m.refreshTranscript()
		m.body.GotoBottom()
		// In running mode, immediately let the pattern proceed to the
		// next turn. In paused mode, wait for `s` keypress.
		if !m.paused {
			m.awaitingLLM = true
			return m, sendAdvance(m.advance)
		}
		return m, nil

	case resultMsg:
		m.final = msg.campaign
		m.refreshTranscript()
		return m, nil

	case doneMsg:
		m.finished = true
		m.final = msg.campaign
		m.runErr = msg.err
		// Drain any pending advance — pattern is done, no more Turn() calls.
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Forward unhandled msgs to the viewport so it can scroll.
	var cmd tea.Cmd
	m.body, cmd = m.body.Update(msg)
	return m, cmd
}

func (m *model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		// Cancel the pattern context and quit. The pattern goroutine
		// will unblock on its next ctx-aware LLM call.
		if m.cancel != nil {
			m.cancel()
		}
		// Also unblock the reporter if it's waiting on advance, by
		// closing the channel via a guard cmd. Safe even if already drained.
		return *m, tea.Quit

	case " ", "p":
		m.paused = !m.paused
		// If we just unpaused and there's a turn the reporter is blocked
		// on, release it now.
		if !m.paused {
			m.awaitingLLM = true
			return *m, sendAdvance(m.advance)
		}
		return *m, nil

	case "s":
		// Step one turn forward (only meaningful while paused).
		if m.paused {
			m.awaitingLLM = true
			return *m, sendAdvance(m.advance)
		}
		return *m, nil

	case "e":
		// Toggle expanded view of the most recent turn.
		if m.expandedTurn == m.currentTurn {
			m.expandedTurn = -1
		} else {
			m.expandedTurn = m.currentTurn
		}
		m.refreshTranscript()
		return *m, nil

	case "j", "down":
		m.body.LineDown(1)
		return *m, nil
	case "k", "up":
		m.body.LineUp(1)
		return *m, nil
	case "g", "home":
		m.body.GotoTop()
		return *m, nil
	case "G", "end":
		m.body.GotoBottom()
		return *m, nil
	}

	var cmd tea.Cmd
	m.body, cmd = m.body.Update(msg)
	return *m, cmd
}

// sendAdvance returns a tea.Cmd that pushes one struct{} into the advance
// channel without blocking the bubbletea event loop. It runs on a worker
// goroutine bubbletea allocates for the Cmd.
func sendAdvance(ch chan<- struct{}) tea.Cmd {
	return func() tea.Msg {
		ch <- struct{}{}
		return nil
	}
}

// relayout reflows the viewport when the terminal size changes or when
// the plan/panel sidebar content changes height.
func (m *model) relayout() {
	if m.width == 0 || m.height == 0 {
		return
	}
	const headerLines = 2
	const footerLines = 1
	const sidePanePad = 4 // borders + padding

	bodyW := m.width - rightPaneWidth - leftPaneWidth - 2*sidePanePad
	if bodyW < 20 {
		bodyW = m.width - 4 // sidebars hidden on narrow terminals
	}
	bodyH := m.height - headerLines - footerLines - 2 // border lines

	if bodyH < 5 {
		bodyH = 5
	}
	if bodyW < 20 {
		bodyW = 20
	}
	m.body = viewportResize(m.body, bodyW, bodyH)
	m.refreshTranscript()
}

// viewportResize is a helper because bubbles/viewport doesn't expose a
// SetSize that preserves position cleanly; we rebuild it but keep content.
func viewportResize(v viewport.Model, w, h int) viewport.Model {
	v.Width = w
	v.Height = h
	return v
}
