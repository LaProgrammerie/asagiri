package cost

import (
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
)

func TestEstimateTokensApprox(t *testing.T) {
	cfg := config.TokenEstimationConfig{
		DefaultCharsPerToken:  4,
		CodeCharsPerToken:     3.2,
		MarkdownCharsPerToken: 4.2,
		JSONCharsPerToken:     3.6,
	}
	if n := EstimateTokensApprox(40, ContentDefault, cfg); n != 10 {
		t.Fatalf("default: got %d", n)
	}
	if n := EstimateTokensApprox(32, ContentCode, cfg); n != 10 {
		t.Fatalf("code: got %d", n)
	}
}
