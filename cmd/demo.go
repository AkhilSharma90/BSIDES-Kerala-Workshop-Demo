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

	r := ui.NewCLIReporter()
	r.Banner("redteam-box · offline demo")
	r.Field("note", "running all three patterns against the mock target — no API keys needed")

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
			Reporter:    r,
		}
		p := patterns.ByName(name, opts)
		if p == nil {
			return fmt.Errorf("missing pattern %q", name)
		}

		r.Banner(fmt.Sprintf("pattern: %s", p.Name()))
		r.Field("description", p.Description())
		r.Field("target", t.Name())
		r.Field("goal", t.Goal())

		camp, err := p.Run(ctx, t)
		if err != nil {
			return fmt.Errorf("%s failed: %w", name, err)
		}
		r.Result(camp)
	}

	r.Banner("demo complete")
	r.Field("next", "run with --target real --target-system-prompt @path.txt to attack any AI product you have authorization to test")
	return nil
}
