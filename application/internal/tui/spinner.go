package tui

import (
	"fmt"
	"io"
	"time"
)

// Spinner renders a simple ASCII spinner on io.Writer (plain path).
type Spinner struct {
	out    io.Writer
	frames []rune
	i      int
	label  string
}

// NewSpinner yields a lightweight spinner.
func NewSpinner(out io.Writer, label string) *Spinner {
	return &Spinner{out: out, frames: []rune(`-\|/`), label: label}
}

// Tick prints one frame; call from a goroutine with delay.
func (s *Spinner) Tick() {
	if s == nil || s.out == nil {
		return
	}
	f := s.frames[s.i%len(s.frames)]
	s.i++
	_, _ = fmt.Fprintf(s.out, "\r%c %s", f, s.label)
}

// Stop clears the line.
func (s *Spinner) Stop(final string) {
	if s.out == nil {
		return
	}
	_, _ = fmt.Fprintf(s.out, "\r%s\n", final)
}

// Sleep is a test-friendly delay wrapper.
func Sleep(d time.Duration) {
	time.Sleep(d)
}
