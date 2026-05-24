package tui

import (
	"fmt"
	"io"
	"strings"
)

type plainUI struct {
	out io.Writer
}

func newPlainUI(out io.Writer) *plainUI {
	return &plainUI{out: out}
}

func (p *plainUI) Box(title, body string) {
	fmt.Fprintf(p.out, "%s\n%s\n%s\n\n", title, strings.Repeat("-", len(title)), strings.TrimSpace(body))
}

func (p *plainUI) ProgressLine(label string, pct float64) {
	bar := renderBar(pct, 24)
	fmt.Fprintf(p.out, "\r%s %6.1f%% %s", bar, pct, label)
}

func (p *plainUI) Event(kind string, payload any) {
	fmt.Fprintf(p.out, "[%s] %v\n", kind, payload)
}

func (p *plainUI) Printf(format string, args ...any) {
	fmt.Fprintf(p.out, format, args...)
}

func renderBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int(float64(width) * pct / 100)
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
