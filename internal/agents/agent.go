// Package agents implements the four specialist roles the orchestrator
// drives: Strategist, PayloadGenerator, Executor, Evaluator. Each agent
// is intentionally small — one responsibility — and they communicate
// through a shared Campaign value rather than method chaining, which is
// what makes this a true multi-agent system rather than a pipeline.
package agents

import (
	"context"
	"time"

	"github.com/akhilsharma/redteam-box/internal/llm"
	"github.com/akhilsharma/redteam-box/internal/target"
)

// Turn is one full exchange in the campaign: the prompt the attacker
// sent, the target's reply, and the evaluator's score for that reply.
type Turn struct {
	Index    int
	Attacker string // which expert/persona produced this prompt
	Prompt   string
	Reply    string
	Score    int    // 0-100
	Reason   string // evaluator's rationale
	At       time.Time
}

// Campaign is the shared state every agent reads from and writes to.
type Campaign struct {
	Goal        string
	Plan        string
	Turns       []Turn
	BestScore   int
	BestPrompt  string
	BestReply   string
	Succeeded   bool
	Notes       []string
}

// AppendTurn records a Turn and updates the running best.
func (c *Campaign) AppendTurn(t Turn) {
	c.Turns = append(c.Turns, t)
	if t.Score > c.BestScore {
		c.BestScore = t.Score
		c.BestPrompt = t.Prompt
		c.BestReply = t.Reply
	}
	if t.Score >= 80 {
		c.Succeeded = true
	}
}

// Agent is the common interface — every specialist satisfies it.
type Agent interface {
	Name() string
}

// Driver is the function signature the patterns call to actually run a
// full campaign against a target. It returns the populated Campaign.
type Driver func(ctx context.Context, t target.Target) (*Campaign, error)

// Common dependency bundle. Each agent gets one of these and can use
// whichever pieces it needs. Keeping the bundle simple and explicit
// avoids the framework drift that creeps into agent codebases.
type Deps struct {
	LLM     llm.Client
	Verbose bool
}
