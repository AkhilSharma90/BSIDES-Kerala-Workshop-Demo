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

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Run all three patterns against the mock target end-to-end.",
	Long: `The demo runs the three patterns sequentially against the offline
mock target. It is the recommended way to see what the tool does
on a fresh clone — no API keys, no network calls.`,
	RunE: runDemo,
}

func runDemo(_ *cobra.Command, _ []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ui.Banner("redteam-box · offline demo")
	ui.Field("note", "running all three patterns against the mock target — no API keys needed")

	patternNames := []string{"crescendo", "evolutionary", "mixture"}
	for _, name := range patternNames {
		t := target.NewMock()
		opts := patterns.Options{
			Goal:        t.Goal(),
			MaxTurns:    5,
			Generations: 3,
			Population:  4,
			Rounds:      6,
			Verbose:     true,
		}
		p := patterns.ByName(name, opts)
		if p == nil {
			return fmt.Errorf("missing pattern %q", name)
		}

		ui.Banner(fmt.Sprintf("pattern: %s", p.Name()))
		ui.Field("description", p.Description())
		ui.Field("target", t.Name())
		ui.Field("goal", t.Goal())

		camp, err := p.Run(ctx, t)
		if err != nil {
			return fmt.Errorf("%s failed: %w", name, err)
		}
		ui.Result(camp.Succeeded, camp.BestScore, camp.BestPrompt, camp.BestReply)
	}

	ui.Banner("demo complete")
	ui.Field("next", "edit internal/target/target.go to point at a real AI product, set API keys, and run with --target real")
	return nil
}
