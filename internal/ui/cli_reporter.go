package ui

import (
	"github.com/akhilsharma/redteam-box/internal/agents"
)

// CLIReporter is the printf/ANSI-color implementation of agents.Reporter.
// It is what runs when stdout is not driving a Bubble Tea program — i.e.
// the classic CLI output you see today.
type CLIReporter struct{}

// NewCLIReporter returns the default CLI reporter.
func NewCLIReporter() *CLIReporter { return &CLIReporter{} }

func (CLIReporter) Banner(title string)        { Banner(title) }
func (CLIReporter) Section(name string)        { Section(name) }
func (CLIReporter) Field(label, value string)  { Field(label, value) }
func (CLIReporter) Plan(plan string)           { Plan(plan) }

func (CLIReporter) Turn(t agents.Turn) {
	Turn(t.Index, t.Attacker, t.Prompt, t.Reply, t.Score, t.Reason)
}

func (CLIReporter) Result(c *agents.Campaign) {
	Result(c.Succeeded, c.BestScore, c.BestPrompt, c.BestReply)
}
