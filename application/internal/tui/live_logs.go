package tui

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// LogWriter prefixes each line with a timestamp for live logs.
func LogWriter(out io.Writer, prefix string) io.Writer {
	return &logPrefix{out: out, prefix: prefix}
}

type logPrefix struct {
	out    io.Writer
	prefix string
	rest   string
}

func (l *logPrefix) Write(p []byte) (int, error) {
	l.rest += string(p)
	for {
		line, after, cut := strings.Cut(l.rest, "\n")
		if !cut {
			l.rest = line
			break
		}
		l.rest = after
		_, _ = fmt.Fprintf(l.out, "%s %s %s\n", time.Now().UTC().Format(time.RFC3339), l.prefix, line)
	}
	return len(p), nil
}
