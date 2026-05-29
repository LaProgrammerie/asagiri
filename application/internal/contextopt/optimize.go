package contextopt

import (
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// ComputeOptimize compares raw collected entries vs reduced pack (specv3 §8, §16.3).
func ComputeOptimize(entries []FileEntry, reduced []FileEntry, pack ContextPack, cfg config.TokenEstimationConfig) OptimizeResult {
	rawTok := SumEntryTokens(entries, cfg)
	optTok := PackApproxTokens(pack, cfg)
	if optTok <= 0 && len(reduced) > 0 {
		optTok = SumEntryTokens(reduced, cfg)
	}
	res := OptimizeResult{OriginalTokens: rawTok, OptimizedTokens: optTok}
	if rawTok > 0 && optTok < rawTok {
		res.SavingsRatio = float64(rawTok-optTok) / float64(rawTok)
	}
	return res
}

// SumEntryTokens estimates tokens for raw file entries before packing.
func SumEntryTokens(entries []FileEntry, cfg config.TokenEstimationConfig) int {
	var n int
	for _, e := range entries {
		ratio := cfg.DefaultCharsPerToken
		if ratio <= 0 {
			ratio = 4
		}
		switch e.Language {
		case KindCode:
			if cfg.CodeCharsPerToken > 0 {
				ratio = cfg.CodeCharsPerToken
			}
		case KindMarkdown:
			if cfg.MarkdownCharsPerToken > 0 {
				ratio = cfg.MarkdownCharsPerToken
			}
		case KindJSON:
			if cfg.JSONCharsPerToken > 0 {
				ratio = cfg.JSONCharsPerToken
			}
		}
		n += int(float64(len(e.Content))/ratio + 0.999)
	}
	return n
}
