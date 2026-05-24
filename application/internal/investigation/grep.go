package investigation

import (
	"context"
	"strings"
	"time"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
)

// Grep runs ripgrep when available, else git grep (specv3 §9).
func Grep(ctx context.Context, repoRoot, pattern string, invCfg config.InvestigationConfig) ([]string, error) {
	sec := invCfg.CommandTimeoutSec
	if sec <= 0 {
		sec = 60
	}
	timeout := time.Duration(sec) * time.Second
	maxB := invCfg.MaxGrepOutputBytes
	if maxB <= 0 {
		maxB = 256 * 1024
	}
	out, err := RunCommand(ctx, timeout, "rg", "-n", "--max-count", "20", pattern, repoRoot)
	if err == nil {
		return limitLines(string(out), maxB), nil
	}
	out2, err2 := RunCommand(ctx, timeout, "git", "-C", repoRoot, "grep", "-n", pattern)
	if err2 != nil {
		return nil, err2
	}
	return limitLines(string(out2), maxB), nil
}

func limitLines(s string, maxBytes int) []string {
	lines := strings.Split(s, "\n")
	var out []string
	var acc int
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		if acc+len(ln)+1 > maxBytes {
			break
		}
		acc += len(ln) + 1
		out = append(out, ln)
	}
	return out
}
