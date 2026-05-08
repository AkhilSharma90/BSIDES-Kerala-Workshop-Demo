package agents

// Reporter is the surface every pattern uses to publish progress to a
// frontend. The CLI uses a printf-based implementation; the TUI uses
// one that pumps messages into a Bubble Tea program. Patterns never
// import a frontend package directly — they only know about Reporter.
//
// Methods are called from the pattern's goroutine. Implementations must
// be safe to call concurrently from that goroutine but do not need to
// be safe across multiple goroutines (patterns are single-threaded).
type Reporter interface {
	// Banner is the top-of-run announcement (e.g. "redteam-box · mixture · MockTarget").
	Banner(title string)
	// Section announces a new logical block within a run.
	Section(name string)
	// Field renders a labeled value such as "goal: ...".
	Field(label, value string)
	// Plan renders the strategist's plan text (multi-line OK).
	Plan(plan string)
	// Turn renders one completed turn.
	Turn(t Turn)
	// Result renders the final campaign outcome.
	Result(c *Campaign)
}

// NoopReporter is the safe default when no Reporter is configured.
type NoopReporter struct{}

func (NoopReporter) Banner(string)         {}
func (NoopReporter) Section(string)        {}
func (NoopReporter) Field(string, string)  {}
func (NoopReporter) Plan(string)           {}
func (NoopReporter) Turn(Turn)             {}
func (NoopReporter) Result(*Campaign)      {}
