package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/akhilsharma/redteam-box/internal/patterns"
)

var patternsCmd = &cobra.Command{
	Use:   "patterns",
	Short: "List the attack patterns this tool implements.",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("Available attack patterns:")
		for _, p := range patterns.Available(patterns.Options{Goal: "list"}) {
			fmt.Printf("  • %-13s %s\n", p.Name(), p.Description())
		}
		return nil
	},
}
