// Package ui prints transcripts in a way that reads well during a live
// demo. No external dependency — just ANSI escapes.
package ui

import (
	"fmt"
	"strings"
)

const (
	reset  = "\x1b[0m"
	bold   = "\x1b[1m"
	dim    = "\x1b[2m"
	red    = "\x1b[31m"
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
	blue   = "\x1b[34m"
	purple = "\x1b[35m"
	cyan   = "\x1b[36m"
)

func Banner(title string) {
	bar := strings.Repeat("─", len(title)+4)
	fmt.Printf("\n%s%s%s\n", bold, bar, reset)
	fmt.Printf("%s  %s  %s\n", bold, title, reset)
	fmt.Printf("%s%s%s\n\n", bold, bar, reset)
}

func Section(label string) {
	fmt.Printf("\n%s%s▌ %s%s\n", bold, cyan, label, reset)
}

func Field(label, value string) {
	fmt.Printf("  %s%s:%s %s\n", dim, label, reset, value)
}

func Plan(plan string) {
	fmt.Printf("%s%splan%s\n", bold, purple, reset)
	for _, line := range strings.Split(strings.TrimSpace(plan), "\n") {
		fmt.Printf("    %s\n", line)
	}
}

func Turn(idx int, attacker, prompt, reply string, score int, reason string) {
	color := red
	switch {
	case score >= 80:
		color = green
	case score >= 50:
		color = yellow
	}
	fmt.Printf("\n%sturn %d%s  %s%s%s\n", bold, idx, reset, dim, attacker, reset)
	fmt.Printf("  %s→ attacker:%s %s\n", blue, reset, oneline(prompt, 220))
	fmt.Printf("  %s← target:%s   %s\n", purple, reset, oneline(reply, 220))
	fmt.Printf("  %sscore:%s     %s%d%s  %s%s%s\n",
		dim, reset, color, score, reset, dim, reason, reset)
}

func Result(succeeded bool, best int, prompt, reply string) {
	if succeeded {
		fmt.Printf("\n%s%s★ campaign succeeded%s  best score: %s%d%s\n",
			bold, green, reset, bold, best, reset)
	} else {
		fmt.Printf("\n%s%s× campaign did not reach success threshold%s  best score: %s%d%s\n",
			bold, yellow, reset, bold, best, reset)
	}
	fmt.Printf("  %sbest prompt:%s %s\n", dim, reset, oneline(prompt, 240))
	fmt.Printf("  %sbest reply:%s  %s\n", dim, reset, oneline(reply, 240))
	fmt.Println()
}

func oneline(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ⏎ ")
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
