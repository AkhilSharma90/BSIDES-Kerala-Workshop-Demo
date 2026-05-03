package agents

import (
	"context"

	"github.com/akhilsharma/redteam-box/internal/target"
)

// Executor is the agent that actually sends payloads to the target and
// hands the reply back. It exists as a separate role because in real
// engagements the executor often handles transport-level concerns
// (retries, rate limits, session cookies, MFA flows) that you don't
// want bleeding into payload generation.
type Executor struct {
	deps   Deps
	target target.Target
}

func NewExecutor(d Deps, t target.Target) *Executor {
	return &Executor{deps: d, target: t}
}

func (e *Executor) Name() string { return "executor" }

// Send delivers a single payload and returns the target's reply.
func (e *Executor) Send(ctx context.Context, payload string) (string, error) {
	return e.target.Ask(ctx, payload)
}

// Reset clears the target's conversation state — useful between
// independent attempts in the evolutionary pattern.
func (e *Executor) Reset() { e.target.Reset() }
