package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
)

type richUI struct {
	out    io.Writer
	border lipgloss.Style
}

func newRichUI(out io.Writer) *richUI {
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	return &richUI{out: out, border: border}
}

func (r *richUI) Box(title, body string) {
	block := r.border.Render(lipgloss.JoinVertical(lipgloss.Left, lipgloss.NewStyle().Bold(true).Render(title), body))
	_, _ = fmt.Fprintln(r.out, block)
}

func (r *richUI) ProgressLine(label string, pct float64) {
	newPlainUI(r.out).ProgressLine(label, pct)
}

func (r *richUI) Event(kind string, payload any) {
	_, _ = fmt.Fprintf(r.out, "[%s] %v\n", kind, payload)
}

func (r *richUI) Printf(format string, args ...any) {
	_, _ = fmt.Fprintf(r.out, format, args...)
}
