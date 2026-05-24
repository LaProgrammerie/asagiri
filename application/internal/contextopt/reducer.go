package contextopt

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// ReduceOpts configures trimming of large files.
type ReduceOpts struct {
	MaxCharsPerFile int
	HeadLines       int
	TailLines       int
}

// Reduce trims oversized FileEntry.Content, recording warnings (specv3 §8.3).
func Reduce(entries []FileEntry, _ *config.Config, opts ReduceOpts) ([]FileEntry, []string) {
	if opts.MaxCharsPerFile <= 0 {
		opts.MaxCharsPerFile = 24000
	}
	if opts.HeadLines <= 0 {
		opts.HeadLines = 40
	}
	if opts.TailLines <= 0 {
		opts.TailLines = 15
	}
	var warns []string
	out := make([]FileEntry, len(entries))
	copy(out, entries)
	for i := range out {
		if len(out[i].Content) <= opts.MaxCharsPerFile {
			continue
		}
		excerpt := excerptWithBoundaries(out[i].Content, opts.MaxCharsPerFile, opts.HeadLines, opts.TailLines)
		warns = append(warns, fmt.Sprintf("truncated %s (%d -> %d chars)", out[i].RelPath, len(out[i].Content), len(excerpt)))
		out[i].Content = excerpt
	}
	return out, warns
}

func excerptWithBoundaries(content string, maxChars, headLines, tailLines int) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= headLines+tailLines {
		return trimRunes(content, maxChars)
	}
	head := strings.Join(lines[:headLines], "\n")
	tail := strings.Join(lines[len(lines)-tailLines:], "\n")
	merged := head + "\n/* … truncated … */\n" + tail
	return trimRunes(merged, maxChars)
}

func trimRunes(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n…"
}
