package cost

import (
	"math"
	"strings"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
)

// ContentKind selects chars-per-token ratio (specv3 §5.3).
type ContentKind string

const (
	ContentDefault ContentKind = "default"
	ContentCode    ContentKind = "code"
	ContentMarkdown ContentKind = "markdown"
	ContentJSON    ContentKind = "json"
)

// EstimateTokensApprox returns token count from character length using cfg ratios.
func EstimateTokensApprox(chars int, kind ContentKind, cfg config.TokenEstimationConfig) int {
	if chars <= 0 {
		return 0
	}
	var cpt float64
	switch kind {
	case ContentCode:
		cpt = cfg.CodeCharsPerToken
	case ContentMarkdown:
		cpt = cfg.MarkdownCharsPerToken
	case ContentJSON:
		cpt = cfg.JSONCharsPerToken
	default:
		cpt = cfg.DefaultCharsPerToken
	}
	if cpt <= 0 {
		return 0
	}
	t := float64(chars) / cpt
	if t <= 0 {
		return 0
	}
	return int(math.Ceil(t))
}

// ClassifyPath guesses content kind from file extension (best-effort).
func ClassifyPath(path string) ContentKind {
	// minimal extension map — callers may override with explicit kind
	switch {
	case hasSuffixFold(path, ".go", ".rs", ".php", ".java", ".ts", ".tsx", ".js", ".jsx", ".py", ".sql"):
		return ContentCode
	case hasSuffixFold(path, ".md", ".mdx"):
		return ContentMarkdown
	case hasSuffixFold(path, ".json", ".yaml", ".yml"):
		return ContentJSON
	default:
		return ContentDefault
	}
}

func hasSuffixFold(p string, suf ...string) bool {
	lp := strings.ToLower(p)
	for _, s := range suf {
		if strings.HasSuffix(lp, strings.ToLower(s)) {
			return true
		}
	}
	return false
}

// EstimateFromText is a convenience for raw text with explicit kind.
func EstimateFromText(text string, kind ContentKind, cfg config.TokenEstimationConfig) int {
	return EstimateTokensApprox(len(text), kind, cfg)
}
