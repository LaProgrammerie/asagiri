package cost

import (
	"strings"
	"sync"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/pkoukk/tiktoken-go"
)

// TokenCountMethod records which counting strategy was used.
// Estimates are not billing-grade unless MethodTiktoken matched a known OpenAI encoding.
type TokenCountMethod string

const (
	MethodTiktoken           TokenCountMethod = "tiktoken"
	MethodAnthropicHeuristic TokenCountMethod = "anthropic_heuristic"
	MethodLocalTiktoken      TokenCountMethod = "local_tiktoken"
	MethodGoogleHeuristic    TokenCountMethod = "google_heuristic"
	MethodApprox             TokenCountMethod = "approx"
)

// ProviderKind groups models for tokenizer selection.
type ProviderKind string

const (
	ProviderOpenAI    ProviderKind = "openai"
	ProviderAnthropic ProviderKind = "anthropic"
	ProviderGoogle    ProviderKind = "google"
	ProviderLocal     ProviderKind = "local"
	ProviderUnknown   ProviderKind = "unknown"
)

var tiktokenCache sync.Map // encoding name -> *tiktoken.Tiktoken

// ClassifyModel maps a model or agent id to a provider bucket (best-effort, case-insensitive).
func ClassifyModel(model string) ProviderKind {
	m := strings.ToLower(strings.TrimSpace(model))
	if m == "" {
		return ProviderUnknown
	}
	switch {
	case isOpenAIModel(m):
		return ProviderOpenAI
	case isAnthropicModel(m):
		return ProviderAnthropic
	case isGoogleModel(m):
		return ProviderGoogle
	case isLocalModel(m):
		return ProviderLocal
	default:
		return ProviderUnknown
	}
}

func isOpenAIModel(m string) bool {
	return strings.HasPrefix(m, "gpt-") ||
		strings.HasPrefix(m, "o1") ||
		strings.HasPrefix(m, "o3") ||
		strings.HasPrefix(m, "o4") ||
		strings.HasPrefix(m, "text-embedding") ||
		strings.HasPrefix(m, "text-davinci") ||
		strings.HasPrefix(m, "davinci") ||
		strings.Contains(m, "openai")
}

func isAnthropicModel(m string) bool {
	return strings.Contains(m, "claude") ||
		strings.Contains(m, "opus") ||
		strings.Contains(m, "sonnet") ||
		strings.Contains(m, "haiku") ||
		strings.Contains(m, "anthropic")
}

func isGoogleModel(m string) bool {
	return strings.Contains(m, "gemini") ||
		strings.Contains(m, "gemma") ||
		strings.Contains(m, "google")
}

func isLocalModel(m string) bool {
	return strings.Contains(m, "llama") ||
		strings.Contains(m, "qwen") ||
		strings.Contains(m, "mistral") ||
		strings.Contains(m, "phi") ||
		strings.Contains(m, "codellama") ||
		strings.Contains(m, "deepseek") ||
		strings.Contains(m, "ollama") ||
		strings.Contains(m, "local")
}

// CountTokens counts tokens for text using a provider tokenizer when cfg allows and the
// model is recognized; otherwise falls back to chars-per-token heuristics (EstimateTokensApprox).
func CountTokens(text string, model string, kind ContentKind, cfg config.TokenEstimationConfig) int {
	n, _ := CountTokensDetailed(text, model, kind, cfg)
	return n
}

// CountTokensDetailed returns the token count and the method that produced it.
func CountTokensDetailed(text string, model string, kind ContentKind, cfg config.TokenEstimationConfig) (int, TokenCountMethod) {
	if len(text) == 0 {
		return 0, MethodApprox
	}
	if cfg.DisableProviderTokenizer {
		return EstimateTokensApprox(len(text), kind, cfg), MethodApprox
	}
	provider := ClassifyModel(model)
	switch provider {
	case ProviderOpenAI:
		if n, ok := countOpenAITokens(text, model); ok {
			return n, MethodTiktoken
		}
	case ProviderAnthropic:
		return countAnthropicHeuristic(text, kind, cfg), MethodAnthropicHeuristic
	case ProviderGoogle:
		return countGoogleHeuristic(text, kind, cfg), MethodGoogleHeuristic
	case ProviderLocal:
		if n, ok := countLocalTokens(text, cfg); ok {
			return n, MethodLocalTiktoken
		}
	}
	return EstimateTokensApprox(len(text), kind, cfg), MethodApprox
}

// EstimateFromTextForModel is the model-aware counterpart to EstimateFromText.
func EstimateFromTextForModel(text string, model string, kind ContentKind, cfg config.TokenEstimationConfig) int {
	return CountTokens(text, model, kind, cfg)
}

func countOpenAITokens(text, model string) (int, bool) {
	enc, err := tiktoken.EncodingForModel(model)
	if err != nil {
		enc, err = encodingByName("cl100k_base")
		if err != nil {
			return 0, false
		}
	}
	return len(enc.Encode(text, nil, nil)), true
}

func countLocalTokens(text string, cfg config.TokenEstimationConfig) (int, bool) {
	name := strings.TrimSpace(cfg.LocalEncoding)
	if name == "" {
		name = "cl100k_base"
	}
	enc, err := encodingByName(name)
	if err != nil {
		return 0, false
	}
	// Llama/Qwen/Ollama models use BPE vocabularies that differ from OpenAI; cl100k_base
	// is a tiktoken-style approximation for order-of-magnitude local cost estimates.
	return len(enc.Encode(text, nil, nil)), true
}

func countAnthropicHeuristic(text string, kind ContentKind, cfg config.TokenEstimationConfig) int {
	cpt := cfg.AnthropicCharsPerToken
	if cpt <= 0 {
		cpt = 3.5
	}
	// Anthropic does not publish a stable offline tokenizer; this ratio tracks typical
	// English prose (~3.5 chars/token). Not billing-exact — use Anthropic count_tokens API for that.
	if kind == ContentCode && cfg.CodeCharsPerToken > 0 {
		cpt = cfg.CodeCharsPerToken
	}
	return EstimateTokensApprox(len(text), kind, withCharsPerToken(cfg, cpt))
}

func countGoogleHeuristic(text string, kind ContentKind, cfg config.TokenEstimationConfig) int {
	cpt := cfg.GoogleCharsPerToken
	if cpt <= 0 {
		cpt = 4.0
	}
	return EstimateTokensApprox(len(text), kind, withCharsPerToken(cfg, cpt))
}

func withCharsPerToken(cfg config.TokenEstimationConfig, cpt float64) config.TokenEstimationConfig {
	out := cfg
	out.DefaultCharsPerToken = cpt
	return out
}

func encodingByName(name string) (*tiktoken.Tiktoken, error) {
	if v, ok := tiktokenCache.Load(name); ok {
		return v.(*tiktoken.Tiktoken), nil
	}
	enc, err := tiktoken.GetEncoding(name)
	if err != nil {
		return nil, err
	}
	actual, _ := tiktokenCache.LoadOrStore(name, enc)
	return actual.(*tiktoken.Tiktoken), nil
}
