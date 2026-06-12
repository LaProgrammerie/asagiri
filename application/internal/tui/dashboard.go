package tui

import (
	"fmt"
	"io"
)

// RenderDashboard prints a compact multi-line status block.
func RenderDashboard(out io.Writer, feature, task, budget string, inputTok, outTok int, cost string) {
	_, _ = fmt.Fprintf(out, "Asagiri Work\n")
	_, _ = fmt.Fprintf(out, "══════════════\n")
	_, _ = fmt.Fprintf(out, "Feature  %s\n", feature)
	_, _ = fmt.Fprintf(out, "Task     %s\n", task)
	_, _ = fmt.Fprintf(out, "Budget   %s\n", budget)
	_, _ = fmt.Fprintf(out, "Tokens   in=%d out=%d\n", inputTok, outTok)
	_, _ = fmt.Fprintf(out, "Cost     %s\n", cost)
}
