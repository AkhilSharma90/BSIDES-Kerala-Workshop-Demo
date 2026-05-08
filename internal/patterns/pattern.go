// Package patterns implements the three attack strategies the talk
// centers on. Each pattern is a Pattern: it takes a target and returns
// a populated Campaign describing what happened.
package patterns

import (
	"context"
	"fmt"

	"github.com/akhilsharma/redteam-box/internal/agents"
	"github.com/akhilsharma/redteam-box/internal/target"
)

// Pattern is the contract every attack pattern satisfies.
type Pattern interface {
	Name() string
	Description() string
	Run(ctx context.Context, t target.Target) (*agents.Campaign, error)
}

// Available is the registry of patterns the CLI exposes.
func Available(opts Options) []Pattern {
	return []Pattern{
		NewCrescendo(opts),
		NewEvolutionary(opts),
		NewMixture(opts),
	}
}

// ByName looks up a pattern by its CLI name. Returns nil if not found.
func ByName(name string, opts Options) Pattern {
	for _, p := range Available(opts) {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// Options bundles the knobs every pattern accepts. Patterns ignore
// fields they don't use — keeps the surface area uniform.
type Options struct {
	Goal        string
	MaxTurns    int
	Generations int
	Population  int
	Rounds      int
	Verbose     bool
	// Reporter receives all user-facing progress events. If nil, a
	// no-op reporter is used so tests and library callers don't need
	// to wire one up.
	Reporter agents.Reporter
}

// reporter returns the configured Reporter or a no-op fallback.
func (o Options) reporter() agents.Reporter {
	if o.Reporter == nil {
		return agents.NoopReporter{}
	}
	return o.Reporter
}

// must is a tiny helper used by patterns that want to surface a
// configuration error early rather than silently default.
func must(o Options) error {
	if o.Goal == "" {
		return fmt.Errorf("pattern: goal is required")
	}
	return nil
}
