package cost

import (
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func testTokenEstConfig() config.TokenEstimationConfig {
	return config.TokenEstimationConfig{
		DefaultCharsPerToken:  4,
		CodeCharsPerToken:     3.2,
		MarkdownCharsPerToken: 4.2,
		JSONCharsPerToken:     3.6,
		AnthropicCharsPerToken: 3.5,
		GoogleCharsPerToken:   4.0,
		LocalEncoding:         "cl100k_base",
	}
}

func TestClassifyModel(t *testing.T) {
	require.Equal(t, ProviderOpenAI, ClassifyModel("gpt-4o"))
	require.Equal(t, ProviderAnthropic, ClassifyModel("claude-3-5-sonnet"))
	require.Equal(t, ProviderGoogle, ClassifyModel("gemini-2.0-flash"))
	require.Equal(t, ProviderLocal, ClassifyModel("qwen2.5-coder"))
	require.Equal(t, ProviderUnknown, ClassifyModel(""))
}

func TestCountTokensOpenAIExact(t *testing.T) {
	cfg := testTokenEstConfig()
	text := "Hello, world!"
	n, method := CountTokensDetailed(text, "gpt-4", ContentDefault, cfg)
	require.Equal(t, MethodTiktoken, method)
	require.Equal(t, 4, n)

	// Repeated chars: tiktoken batches differently from chars/4 heuristic.
	repeated := strings.Repeat("x", 40)
	tik, _ := CountTokensDetailed(repeated, "gpt-4", ContentDefault, cfg)
	approx := EstimateTokensApprox(len(repeated), ContentDefault, cfg)
	require.Equal(t, 5, tik)
	require.Equal(t, 10, approx)
}

func TestCountTokensAnthropicHeuristic(t *testing.T) {
	cfg := testTokenEstConfig()
	text := strings.Repeat("a", 35)
	n, method := CountTokensDetailed(text, "claude-3-opus", ContentDefault, cfg)
	require.Equal(t, MethodAnthropicHeuristic, method)
	require.Equal(t, 10, n) // ceil(35/3.5)
}

func TestCountTokensLocalTiktoken(t *testing.T) {
	cfg := testTokenEstConfig()
	text := "Hello, world!"
	n, method := CountTokensDetailed(text, "qwen2.5", ContentDefault, cfg)
	require.Equal(t, MethodLocalTiktoken, method)
	require.Equal(t, 4, n)
}

func TestCountTokensFallbackApprox(t *testing.T) {
	cfg := testTokenEstConfig()
	text := strings.Repeat("z", 40)
	n, method := CountTokensDetailed(text, "unknown-model-xyz", ContentDefault, cfg)
	require.Equal(t, MethodApprox, method)
	require.Equal(t, 10, n)
}

func TestCountTokensDisableProviderTokenizer(t *testing.T) {
	cfg := testTokenEstConfig()
	cfg.DisableProviderTokenizer = true
	text := "Hello, world!"
	n, method := CountTokensDetailed(text, "gpt-4", ContentDefault, cfg)
	require.Equal(t, MethodApprox, method)
	require.Equal(t, 4, n) // 13 chars / 4 -> ceil(3.25) = 4
}

func TestCountTokensEmpty(t *testing.T) {
	cfg := testTokenEstConfig()
	n, method := CountTokensDetailed("", "gpt-4", ContentDefault, cfg)
	require.Equal(t, 0, n)
	require.Equal(t, MethodApprox, method)
}

func TestEstimateFromTextForModelUsesTokenizer(t *testing.T) {
	cfg := testTokenEstConfig()
	repeated := strings.Repeat("x", 40)
	modelAware := EstimateFromTextForModel(repeated, "gpt-4", ContentDefault, cfg)
	approxOnly := EstimateFromText(repeated, ContentDefault, cfg)
	require.Equal(t, 5, modelAware)
	require.Equal(t, 10, approxOnly)
}
