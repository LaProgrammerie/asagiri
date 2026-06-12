package contextopt

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

func TestComputeOptimizeSavings(t *testing.T) {
	entries := []FileEntry{
		{Content: string(make([]byte, 4000)), Language: KindCode},
	}
	reduced, _ := Reduce(entries, &config.Config{}, ReduceOpts{MaxCharsPerFile: 500})
	pack := BuildPack(&config.Config{}, PackInput{ReducedFiles: reduced, OutputFormat: "text"})
	te := config.TokenEstimationConfig{DefaultCharsPerToken: 4}
	opt := ComputeOptimize(entries, reduced, pack, te)
	if opt.OriginalTokens <= opt.OptimizedTokens {
		t.Fatalf("expected savings, orig=%d opt=%d", opt.OriginalTokens, opt.OptimizedTokens)
	}
	if opt.SavingsRatio <= 0 {
		t.Fatalf("expected positive ratio, got %f", opt.SavingsRatio)
	}
}
