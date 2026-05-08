// Package cmd wires the CLI together. The top-level binary is
// `redteam-box`; subcommands live in their own files in this package.
package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "redteam-box",
	Short: "Autonomous multi-agent AI attack tool — BSIDES Kerala 2026 workshop.",
	Long: `redteam-box is a Go CLI that runs autonomous AI attack campaigns
against a target AI system. It implements three patterns:

  crescendo     turn-by-turn escalation (Microsoft Research, Russinovich et al.)
  evolutionary  GA-style population search
  mixture       Mixture of Attackers — multi-persona / multi-LLM panel

Run "redteam-box demo" for an offline end-to-end demo. No API keys
required for the demo. To attack a real target, set ANTHROPIC_API_KEY
/ OPENAI_API_KEY / GOOGLE_API_KEY and pass --target real.

This tool is for authorized security research and education only.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(attackCmd)
	rootCmd.AddCommand(demoCmd)
	rootCmd.AddCommand(patternsCmd)
}
