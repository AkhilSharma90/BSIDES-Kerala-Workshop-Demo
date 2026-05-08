package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/akhilsharma/redteam-box/internal/agents"
)

// model is the Bubble Tea model. It holds every piece of UI state.
type model struct {
	// static run info
	patternName string
	targetName  string
	goal        string

	// dynamic campaign state
	plan         string
	panel        string
	turns        []agents.Turn
	bestScore    int
	currentTurn  int  // 1-based index of the latest turn shown
	expandedTurn int  // -1 = none expanded, else index of expanded turn
	awaitingLLM  bool // true while a Turn() event hasn't arrived yet

	// run lifecycle
	paused   bool
	finished bool
	final    *agents.Campaign
	runErr   error

	// layout / sub-components
	width   int
	height  int
	body    viewport.Model
	spin    spinner.Model

	// channels for communication with the pattern goroutine
	advance chan<- struct{}
	cancel  context.CancelFunc
}

// initial returns a fresh model populated with the static run info.
func initial(patternName, targetName, goal string,
	advance chan<- struct{}, cancel context.CancelFunc) model {

	body := viewport.New(80, 20)
	body.SetContent("")

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot

	return model{
		patternName:  patternName,
		targetName:   targetName,
		goal:         goal,
		expandedTurn: -1,
		body:         body,
		spin:         sp,
		advance:      advance,
		cancel:       cancel,
		awaitingLLM:  true,
	}
}

func (m model) Init() tea.Cmd {
	return m.spin.Tick
}
