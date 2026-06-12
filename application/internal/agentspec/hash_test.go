package agentspec_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/stretchr/testify/require"
)

func TestSemanticHashStable(t *testing.T) {
	spec := agentspec.Spec{
		ID:           "dev",
		Version:      "1.0.0",
		Role:         agentspec.RoleDev,
		SystemPrompt: "prompt",
		OutputContract: agentspec.OutputContract{
			Format: agentspec.OutputAsagiriV1,
		},
	}
	h1 := agentspec.SemanticHash(spec)
	h2 := agentspec.SemanticHash(spec)
	require.Equal(t, h1, h2)
	require.Len(t, h1, 64)
}

func TestSemanticHashIgnoresNonDeterministicMetadata(t *testing.T) {
	base := agentspec.Spec{
		ID:           "dev",
		Version:      "1.0.0",
		Role:         agentspec.RoleDev,
		SystemPrompt: "prompt",
		OutputContract: agentspec.OutputContract{
			Format: agentspec.OutputAsagiriV1,
		},
		Metadata: map[string]any{
			"labels": map[string]any{"locale": "fr"},
		},
	}
	withTimestamps := base
	withTimestamps.Metadata = map[string]any{
		"labels":       map[string]any{"locale": "fr"},
		"updated_at":   "2026-06-08T12:00:00Z",
		"content_hash": "should-be-ignored",
		"synced_at":    "2026-06-08T13:00:00Z",
	}

	require.Equal(t, agentspec.SemanticHash(base), agentspec.SemanticHash(withTimestamps))
}

func TestSemanticHashExternalIgnoresLastSyncedHash(t *testing.T) {
	base := agentspec.Spec{
		ID:           "dev",
		Version:      "1.0.0",
		Role:         agentspec.RoleDev,
		SystemPrompt: "prompt",
		OutputContract: agentspec.OutputContract{
			Format: agentspec.OutputAsagiriV1,
		},
		External: &agentspec.ExternalSpec{
			Kind: "cursor-agent",
			Path: "~/.cursor/agents/dev.md",
		},
	}
	withSynced := base
	withSynced.External = &agentspec.ExternalSpec{
		Kind:           "cursor-agent",
		Path:           "~/.cursor/agents/dev.md",
		LastSyncedHash: "abc123",
	}
	require.Equal(t, agentspec.SemanticHash(base), agentspec.SemanticHash(withSynced))
}

func TestSemanticHashProviderTargetsOrderIndependent(t *testing.T) {
	a := agentspec.Spec{
		ID:              "dev",
		Version:         "1.0.0",
		Role:            agentspec.RoleDev,
		ProviderTargets: []string{"cursor-cli", "kiro-cli"},
		SystemPrompt:    "p",
		OutputContract:  agentspec.OutputContract{Format: agentspec.OutputText},
	}
	b := a
	b.ProviderTargets = []string{"kiro-cli", "cursor-cli"}
	require.Equal(t, agentspec.SemanticHash(a), agentspec.SemanticHash(b))
}

func TestParseSetsContentHash(t *testing.T) {
	loader := agentspec.NewLoader(t.TempDir())
	spec, err := loader.Load("dev")
	require.NoError(t, err)
	require.Equal(t, agentspec.SemanticHash(spec), spec.ContentHash)
}
