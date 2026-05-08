package tui

import (
	"context"

	"github.com/akhilsharma/redteam-box/internal/agents"
)

// reporter is the agents.Reporter implementation the TUI hands to the
// patterns. It pushes one tea.Msg into the events channel for each
// reported event, and on Turn() it additionally blocks on `advance`
// so the TUI can pause between turns when the user wants step mode.
//
// All sends and the advance-receive select on ctx.Done() so that when
// the user quits the TUI, the pattern goroutine unwinds cleanly even
// when the underlying LLM client (e.g. the mock) is not ctx-aware.
type reporter struct {
	ctx     context.Context
	events  chan<- any
	advance <-chan struct{}
}

func newReporter(ctx context.Context, events chan<- any, advance <-chan struct{}) *reporter {
	return &reporter{ctx: ctx, events: events, advance: advance}
}

func (r *reporter) send(ev any) {
	select {
	case r.events <- ev:
	case <-r.ctx.Done():
	}
}

func (r *reporter) Banner(title string)        { r.send(bannerMsg{title: title}) }
func (r *reporter) Section(name string)        { r.send(sectionMsg{name: name}) }
func (r *reporter) Field(label, value string)  { r.send(fieldMsg{label: label, value: value}) }
func (r *reporter) Plan(plan string)           { r.send(planMsg{plan: plan}) }

// Turn sends a turnMsg and then blocks until the TUI signals on advance.
// Both operations honour ctx so that quitting the TUI unblocks the
// pattern even if its LLM client is not itself ctx-aware.
func (r *reporter) Turn(t agents.Turn) {
	r.send(turnMsg{turn: t})
	select {
	case <-r.advance:
	case <-r.ctx.Done():
	}
}

func (r *reporter) Result(c *agents.Campaign) { r.send(resultMsg{campaign: c}) }
