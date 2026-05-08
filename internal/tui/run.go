package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/akhilsharma/redteam-box/internal/agents"
	"github.com/akhilsharma/redteam-box/internal/patterns"
	"github.com/akhilsharma/redteam-box/internal/target"
)

// Run executes a single attack pattern under the TUI. It blocks until
// the user quits or the pattern finishes, then returns the populated
// Campaign (which may be nil if the user quit before any turn completed).
//
// The function spins up two goroutines:
//   - the pattern goroutine, which calls reporter methods that pump events
//     back here via a channel and (for Turn) block on `advance` so the TUI
//     can pause between turns.
//   - the events-pump goroutine, which forwards reporter events into the
//     bubbletea program's message queue.
//
// The bubbletea program owns the main goroutine.
func Run(ctx context.Context, p patterns.Pattern, t target.Target,
	opts patterns.Options, paused bool) (*agents.Campaign, error) {

	events := make(chan any, 64)
	advance := make(chan struct{}, 1)

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	rep := newReporter(runCtx, events, advance)
	opts.Reporter = rep
	// re-bind the pattern with the reporter installed; patterns capture
	// opts at construction time.
	p = patterns.ByName(p.Name(), opts)

	m := initial(p.Name(), t.Name(), opts.Goal, advance, cancel)
	m.paused = paused

	prog := tea.NewProgram(m, tea.WithAltScreen(), tea.WithContext(runCtx))

	// Pump reporter events into the bubbletea program until ctx is
	// cancelled (i.e. the user quit). We don't close `events` because
	// the pattern goroutine may still be unwinding and would panic on
	// a send to a closed channel.
	go func() {
		for {
			select {
			case ev := <-events:
				prog.Send(ev)
			case <-runCtx.Done():
				return
			}
		}
	}()

	// Run the pattern on its own goroutine. When done, push a doneMsg
	// into the program — if the program has already quit, Send is a
	// no-op, which is exactly what we want.
	var camp *agents.Campaign
	var runErr error
	go func() {
		camp, runErr = p.Run(runCtx, t)
		prog.Send(doneMsg{campaign: camp, err: runErr})
	}()

	if _, err := prog.Run(); err != nil {
		return camp, err
	}
	return camp, runErr
}
