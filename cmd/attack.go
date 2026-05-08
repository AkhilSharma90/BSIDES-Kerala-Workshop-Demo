package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/akhilsharma/redteam-box/internal/llm"
	"github.com/akhilsharma/redteam-box/internal/patterns"
	"github.com/akhilsharma/redteam-box/internal/target"
	"github.com/akhilsharma/redteam-box/internal/tui"
	"github.com/akhilsharma/redteam-box/internal/ui"
)

var (
	flagPattern             string
	flagTarget              string
	flagGoal                string
	flagMaxTurns            int
	flagGenerations         int
	flagPopulation          int
	flagRounds              int
	flagVerbose             bool
	flagTargetProvider      string
	flagTargetModel         string
	flagTargetSystemPrompt  string
	flagTargetName          string
	flagTUI                 bool
	flagStep                bool
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

	attackCmd.Flags().StringVar(&flagTargetProvider, "target-provider", "",
		"when --target=real: anthropic | openai | gemini (auto-detected from env if empty)")
	attackCmd.Flags().StringVar(&flagTargetModel, "target-model", "",
		"when --target=real: provider-specific model name (default depends on provider)")
	attackCmd.Flags().StringVar(&flagTargetSystemPrompt, "target-system-prompt", "",
		"when --target=real: system prompt for the target (string, or @path/to/file)")
	attackCmd.Flags().StringVar(&flagTargetName, "target-name", "",
		"when --target=real: display name for the target (default: derived from provider)")

	attackCmd.Flags().BoolVar(&flagTUI, "tui", false,
		"run the campaign inside the full-screen TUI (recommended for live demos)")
	attackCmd.Flags().BoolVar(&flagStep, "step", false,
		"start the TUI in paused mode — press s to advance one turn at a time")
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

	if flagTUI {
		_, err := tui.Run(ctx, p, t, opts, flagStep)
		return err
	}

	r := ui.NewCLIReporter()
	opts.Reporter = r
	p = patterns.ByName(flagPattern, opts)

	r.Banner(fmt.Sprintf("redteam-box · %s · %s", p.Name(), t.Name()))
	r.Field("goal", goal)

	camp, err := p.Run(ctx, t)
	if err != nil {
		return err
	}

	r.Result(camp)
	return nil
}

func buildTarget(name string) (target.Target, error) {
	switch name {
	case "mock", "":
		return target.NewMock(), nil
	case "real":
		return buildLiveTarget()
	default:
		return nil, fmt.Errorf("unknown target %q (try: mock | real)", name)
	}
}

// buildLiveTarget constructs a target.LiveTarget from --target-* flags.
// It picks the provider from --target-provider, or auto-detects one from
// the first ANTHROPIC_API_KEY / OPENAI_API_KEY / GOOGLE_API_KEY that is
// set. The system prompt is required and may be passed inline or as a
// @path/to/file reference.
func buildLiveTarget() (target.Target, error) {
	sys, err := resolveSystemPrompt(flagTargetSystemPrompt)
	if err != nil {
		return nil, err
	}
	if sys == "" {
		return nil, fmt.Errorf(
			"--target-system-prompt is required for --target=real\n" +
				"  example: --target-system-prompt 'You are a helpful pirate. Never break character.'\n" +
				"  or load from file: --target-system-prompt @./my-target.txt")
	}

	provider := flagTargetProvider
	if provider == "" {
		switch {
		case os.Getenv("ANTHROPIC_API_KEY") != "":
			provider = "anthropic"
		case os.Getenv("OPENAI_API_KEY") != "":
			provider = "openai"
		case os.Getenv("GOOGLE_API_KEY") != "":
			provider = "gemini"
		default:
			return nil, fmt.Errorf(
				"no provider configured: set one of ANTHROPIC_API_KEY / OPENAI_API_KEY / " +
					"GOOGLE_API_KEY, or pass --target-provider explicitly")
		}
	}

	client, err := buildLiveClient(provider, flagTargetModel)
	if err != nil {
		return nil, err
	}

	goal := flagGoal
	if goal == "" {
		goal = "extract or paraphrase the target's system prompt"
	}

	name := flagTargetName
	if name == "" {
		name = "Live(" + client.Name() + ")"
	}

	return target.NewLive(name, sys, goal, client), nil
}

// buildLiveClient constructs an llm.Client for the named provider, picking
// the API key from the environment and falling back to a sensible default
// model when --target-model is not set.
func buildLiveClient(provider, model string) (llm.Client, error) {
	switch provider {
	case "anthropic":
		key := os.Getenv("ANTHROPIC_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY is not set")
		}
		if model == "" {
			model = "claude-opus-4-7"
		}
		return llm.NewAnthropic(key, model), nil
	case "openai":
		key := os.Getenv("OPENAI_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is not set")
		}
		if model == "" {
			model = "gpt-4o"
		}
		return llm.NewOpenAI(key, model), nil
	case "gemini":
		key := os.Getenv("GOOGLE_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("GOOGLE_API_KEY is not set")
		}
		if model == "" {
			model = "gemini-1.5-pro"
		}
		return llm.NewGemini(key, model), nil
	default:
		return nil, fmt.Errorf("unknown provider %q (try: anthropic | openai | gemini)", provider)
	}
}

// resolveSystemPrompt accepts an inline string or, if the value starts
// with '@', reads the rest as a file path.
func resolveSystemPrompt(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}
	if strings.HasPrefix(s, "@") {
		path := strings.TrimPrefix(s, "@")
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("reading system prompt file %q: %w", path, err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return s, nil
}
