package tui

import (
	"fmt"
	"io"
)

// RenderDashboard prints a compact multi-line status block.
func RenderDashboard(out io.Writer, feature, task, budget string, inputTok, outTok int, cost string) {
	fmt.Fprintf(out, "Asagiri Work\n")
	fmt.Fprintf(out, "══════════════\n")
	fmt.Fprintf(out, "Feature  %s\n", feature)
	fmt.Fprintf(out, "Task     %s\n", task)
	fmt.Fprintf(out, "Budget   %s\n", budget)
	fmt.Fprintf(out, "Tokens   in=%d out=%d\n", inputTok, outTok)
	fmt.Fprintf(out, "Cost     %s\n", cost)
}
