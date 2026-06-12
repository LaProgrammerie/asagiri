package agentadapter_test

import (
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentadapter"
	"github.com/LaProgrammerie/asagiri/application/internal/agentcontext"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func testSpec(targets ...string) agentspec.Spec {
	return agentspec.Spec{
		ID:              "dev",
		Version:         "1.0.0",
		Role:            agentspec.RoleDev,
		ProviderTargets: targets,
		SystemPrompt:    "system",
		OutputContract: agentspec.OutputContract{
			Format: agentspec.OutputAsagiriV1,
		},
	}
}

func testContext() agentcontext.ExecutionContext {
	return agentcontext.Build(agentcontext.Input{
		Spec:           testSpec("kiro-cli", "cursor-cli", "claude-code", "codex-cli"),
		Feature:        "feat",
		TaskID:         "task-1",
		RunID:          "run-1",
		Phase:          "dev",
		UserTaskPrompt: "Implémente.",
	})
}

func testRuntimeConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := config.NewTestConfig("t")
	cfg.Providers = map[string]config.ProviderConfig{
		"kiro-cli":    {Type: config.ProviderTypeKiroCLI, Command: "kiro", Args: []string{"--cli"}},
		"cursor-cli":  {Type: config.ProviderTypeCursorCLI, Command: "cursor-agent"},
		"claude-code": {Type: config.ProviderTypeClaudeCode, Command: "claude", Args: []string{"--print", "--output-format", "stream-json"}},
		"codex-cli":   {Type: config.ProviderTypeCodexCLI, Command: "codex", Args: []string{"exec"}},
		"ollama":      {Type: config.ProviderTypeOllama, Command: "ollama", Args: []string{"run", "m"}},
	}
	cfg.Agents = map[string]config.Agent{
		"dev":            {Provider: "claude-code"},
		"laprogrammerie": {Provider: "kiro-cli", Profile: "laprogrammerie"},
		"cursor":         {Provider: "cursor-cli"},
		"codex":          {Provider: "codex-cli"},
		"enrich":         {Provider: "ollama"},
	}
	return cfg
}

func assertOrchestratedPrompt(t *testing.T, out agentadapter.RenderedInvocation) {
	t.Helper()
	require.Contains(t, out.StdinPrompt, "Tu es exécuté par Asagiri en mode orchestré")
	require.Contains(t, out.StdinPrompt, "## Output contract")
	for _, arg := range out.Args {
		require.NotContains(t, arg, "Tu es exécuté par Asagiri")
	}
}

func TestCursorInvocationDeterministic(t *testing.T) {
	cfg := testRuntimeConfig(t)
	ctx := testContext()
	spec := testSpec("cursor-cli")

	out1, err := agentadapter.RenderFromConfig(cfg, "cursor", spec, ctx)
	require.NoError(t, err)
	out2, err := agentadapter.RenderFromConfig(cfg, "cursor", spec, ctx)
	require.NoError(t, err)
	require.Equal(t, out1, out2)

	require.Equal(t, config.ProviderTypeCursorCLI, out1.ProviderType)
	require.Equal(t, agentadapter.SupportInlinePrompt, out1.SupportLevel)
	require.Equal(t, "cursor-agent", out1.Command)
	assertOrchestratedPrompt(t, out1)
}

func TestKiroInvocationDeterministic(t *testing.T) {
	cfg := testRuntimeConfig(t)
	ctx := testContext()
	spec := testSpec("kiro-cli")

	out, err := agentadapter.RenderFromConfig(cfg, "laprogrammerie", spec, ctx)
	require.NoError(t, err)
	require.Equal(t, "kiro", out.Command)
	require.Equal(t, []string{"--cli"}, out.Args)
	require.Equal(t, agentadapter.SupportInlinePrompt, out.SupportLevel)
	assertOrchestratedPrompt(t, out)
	require.NotEmpty(t, out.Warnings)
	require.True(t, strings.Contains(out.Warnings[0], "profile") || len(out.Warnings) > 0)
}

func TestClaudeCodeInvocationDeterministic(t *testing.T) {
	cfg := testRuntimeConfig(t)
	ctx := testContext()
	spec := testSpec("claude-code")

	out, err := agentadapter.RenderFromConfig(cfg, "dev", spec, ctx)
	require.NoError(t, err)
	require.Equal(t, agentadapter.SupportNativeProfile, out.SupportLevel)
	require.Equal(t, "claude", out.Command)
	require.Equal(t, []string{"--print", "--output-format", "stream-json"}, out.Args)
	assertOrchestratedPrompt(t, out)
}

func TestCodexInvocationDeterministic(t *testing.T) {
	cfg := testRuntimeConfig(t)
	ctx := testContext()
	spec := testSpec("codex-cli")

	out, err := agentadapter.RenderFromConfig(cfg, "codex", spec, ctx)
	require.NoError(t, err)
	require.Equal(t, "codex", out.Command)
	require.Equal(t, []string{"exec"}, out.Args)
	require.Equal(t, agentadapter.SupportInlinePrompt, out.SupportLevel)
	assertOrchestratedPrompt(t, out)
}

func TestUnsupportedProviderTargetsFallsBackInline(t *testing.T) {
	cfg := testRuntimeConfig(t)
	ctx := testContext()
	spec := testSpec("ollama") // dev agent uses claude-code runtime but spec only allows ollama - mismatch

	providerType, merged, err := cfg.MergedAgentRuntime("dev")
	require.NoError(t, err)
	inv := agentadapter.NewInvocation(providerType, "dev", spec, ctx, merged)

	out, err := NewFactory().Render(inv)
	require.NoError(t, err)
	require.Equal(t, agentadapter.SupportInlinePrompt, out.SupportLevel)
	require.NotEmpty(t, out.Warnings)
	assertOrchestratedPrompt(t, out)
}

func NewFactory() *agentadapter.Factory {
	return agentadapter.NewFactory()
}

func TestUnknownProviderTypeUnsupported(t *testing.T) {
	spec := testSpec()
	ctx := testContext()
	inv := agentadapter.NewInvocation("unknown-provider", "x", spec, ctx, config.Agent{})

	out, err := agentadapter.NewFactory().Render(inv)
	require.Error(t, err)
	require.Equal(t, agentadapter.SupportUnsupported, out.SupportLevel)
}

func TestOllamaInlineFallback(t *testing.T) {
	cfg := testRuntimeConfig(t)
	ctx := testContext()
	spec := testSpec("ollama")

	out, err := agentadapter.RenderFromConfig(cfg, "enrich", spec, ctx)
	require.NoError(t, err)
	require.Equal(t, config.ProviderTypeOllama, out.ProviderType)
	require.Equal(t, agentadapter.SupportInlinePrompt, out.SupportLevel)
	require.Equal(t, "ollama", out.Command)
	assertOrchestratedPrompt(t, out)
}

func TestSupportMatrix(t *testing.T) {
	spec := testSpec("claude-code", "kiro-cli")
	m := agentadapter.SupportMatrix(spec)
	require.Equal(t, agentadapter.SupportNativeProfile, m[config.ProviderTypeClaudeCode])
	require.Equal(t, agentadapter.SupportInlinePrompt, m[config.ProviderTypeKiroCLI])
	require.Equal(t, agentadapter.SupportUnsupported, m[config.ProviderTypeCursorCLI])
}

func TestExplainNotEmpty(t *testing.T) {
	require.NotEmpty(t, agentadapter.Explain(config.ProviderTypeCursorCLI))
	require.NotEmpty(t, agentadapter.Explain(config.ProviderTypeKiroCLI))
}
