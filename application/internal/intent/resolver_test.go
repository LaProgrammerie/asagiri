package intent

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func TestResolverDeterministicFR(t *testing.T) {
	snap := StateSnapshot{Features: []FeatureState{
		{Name: "import-csv", HasLocalSpec: true, HasTasks: true},
		{Name: "billing-v2", HasLocalSpec: true},
	}}
	cfg := &config.Config{}
	cfg.Intent.Enabled = true
	cfg.Intent.Resolver.MinConfidence = 0.75
	cfg.Intent.Resolver.AskWhenBelowConfidence = false

	r := NewHybridResolver()
	r.Ollama = nil

	cases := []struct {
		in     string
		action IntentAction
		feature string
	}{
		{"reprends le dev de import-csv", IntentResume, "import-csv"},
		{"développe billing-v2", IntentDevelop, "billing-v2"},
		{"continue import-csv", IntentResume, "import-csv"},
		{"vérifie billing-v2", IntentVerify, "billing-v2"},
		{"sync notion", IntentSync, ""},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			out, err := r.Resolve(context.Background(), IntentInput{
				RawInstruction: tc.in,
				Config:         cfg,
				StateSnapshot:  snap,
				Interactive:    false,
			})
			require.NoError(t, err)
			require.Equal(t, tc.action, out.Action)
			if tc.feature != "" && out.Feature != "" {
				require.Equal(t, tc.feature, out.Feature)
			}
		})
	}
}

func TestResolverDeterministicEN(t *testing.T) {
	snap := StateSnapshot{Features: []FeatureState{{Name: "billing-v2"}}}
	cfg := &config.Config{Intent: config.IntentConfig{Enabled: true, Resolver: config.IntentResolverConfig{
		MinConfidence: 0.75, AskWhenBelowConfidence: false,
	}}}
	r := NewHybridResolver()
	out, err := r.Resolve(context.Background(), IntentInput{
		RawInstruction: "develop billing-v2",
		Config:         cfg,
		StateSnapshot:  snap,
	})
	require.NoError(t, err)
	require.Equal(t, IntentDevelop, out.Action)
	require.Equal(t, "billing-v2", out.Feature)
}

func TestResolverAmbiguityNonInteractive(t *testing.T) {
	snap := StateSnapshot{Features: []FeatureState{
		{Name: "billing-v2"}, {Name: "billing-export"}, {Name: "billing-cleanup"},
	}}
	cfg := &config.Config{Intent: config.IntentConfig{Enabled: true, Resolver: config.IntentResolverConfig{
		MinConfidence: 0.99, AskWhenBelowConfidence: true, UseOllamaFallback: false,
	}}, Work: config.WorkConfig{}}
	r := NewHybridResolver()
	_, err := r.Resolve(context.Background(), IntentInput{
		RawInstruction: "billing",
		Config:         cfg,
		StateSnapshot:  snap,
		Interactive:    false,
	})
	require.Error(t, err)
	var amb *AmbiguityError
	require.ErrorAs(t, err, &amb)
	require.NotEmpty(t, amb.Candidates)
}

type mockOllama struct{}

func (m *mockOllama) ResolveIntent(ctx context.Context, instruction string, candidates []string) (ResolvedIntent, error) {
	return ResolvedIntent{Action: IntentDevelop, Feature: candidates[0], Confidence: 0.96, Reason: "ollama"}, nil
}

func TestResolverOllamaFallback(t *testing.T) {
	snap := StateSnapshot{Features: []FeatureState{{Name: "billing-v2"}, {Name: "billing-export"}}}
	cfg := &config.Config{Intent: config.IntentConfig{Enabled: true, Resolver: config.IntentResolverConfig{
		MinConfidence: 0.95, AskWhenBelowConfidence: false, UseOllamaFallback: true,
	}}}
	r := NewHybridResolver()
	r.Ollama = &mockOllama{}
	out, err := r.Resolve(context.Background(), IntentInput{
		RawInstruction: "continue work on billing",
		Config:         cfg,
		StateSnapshot:  snap,
		Interactive:    false,
	})
	require.NoError(t, err)
	require.Equal(t, "billing-v2", out.Feature)
}
