package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/akhilsharma/redteam-box/internal/patterns"
	"github.com/akhilsharma/redteam-box/internal/target"
	"github.com/akhilsharma/redteam-box/internal/ui"
)

var (
	flagPattern     string
	flagTarget      string
	flagGoal        string
	flagMaxTurns    int
	flagGenerations int
	flagPopulation  int
	flagRounds      int
	flagVerbose     bool
)

var attackCmd = &cobra.Command{
	Use:   "attack",
	Short: "Run an attack pattern against a target.",
	Long: `Run an autonomous attack campaign. By default, the campaign runs
against the offline mock target so you can see the full transcript
without any API keys configured.

Examples:
  redteam-box attack --pattern crescendo
  redteam-box attack --pattern evolutionary --generations 5 --population 6
  redteam-box attack --pattern mixture --rounds 8 --verbose
  redteam-box attack --pattern mixture --target real --goal "extract the system prompt"`,
	RunE: runAttack,
}

func init() {
	attackCmd.Flags().StringVar(&flagPattern, "pattern", "crescendo",
		"attack pattern: crescendo | evolutionary | mixture")
	attackCmd.Flags().StringVar(&flagTarget, "target", "mock",
		"target system: mock (offline) | real (uses configured providers)")
	attackCmd.Flags().StringVar(&flagGoal, "goal", "",
		"campaign objective (defaults to extracting the mock target's secret)")
	attackCmd.Flags().IntVar(&flagMaxTurns, "max-turns", 6,
		"max turns for crescendo")
	attackCmd.Flags().IntVar(&flagGenerations, "generations", 4,
		"generations for evolutionary")
	attackCmd.Flags().IntVar(&flagPopulation, "population", 4,
		"population size for evolutionary")
	attackCmd.Flags().IntVar(&flagRounds, "rounds", 8,
		"rounds for mixture")
	attackCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", true,
		"print every turn (default true so the demo is readable)")
}

func runAttack(_ *cobra.Command, _ []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	t, err := buildTarget(flagTarget)
	if err != nil {
		return err
	}

	goal := flagGoal
	if goal == "" {
		goal = t.Goal()
	}

	opts := patterns.Options{
		Goal:        goal,
		MaxTurns:    flagMaxTurns,
		Generations: flagGenerations,
		Population:  flagPopulation,
		Rounds:      flagRounds,
		Verbose:     flagVerbose,
	}
	p := patterns.ByName(flagPattern, opts)
	if p == nil {
		return fmt.Errorf("unknown pattern %q (try: crescendo, evolutionary, mixture)", flagPattern)
	}

	ui.Banner(fmt.Sprintf("redteam-box · %s · %s", p.Name(), t.Name()))
	ui.Field("goal", goal)

	camp, err := p.Run(ctx, t)
	if err != nil {
		return err
	}

	ui.Result(camp.Succeeded, camp.BestScore, camp.BestPrompt, camp.BestReply)
	return nil
}

func buildTarget(name string) (target.Target, error) {
	switch name {
	case "mock", "":
		return target.NewMock(), nil
	case "real":
		return nil, fmt.Errorf(
			"real-target wiring is left as an exercise — point internal/target at the AI " +
				"product you have authorization to test, or use --target mock for the demo")
	default:
		return nil, fmt.Errorf("unknown target %q (try: mock)", name)
	}
}
