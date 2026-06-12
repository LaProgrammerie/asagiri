package agentcontext_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentcontext"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/stretchr/testify/require"
)

func sampleSpec() agentspec.Spec {
	return agentspec.Spec{
		ID:           "dev",
		Version:      "1.0.0",
		Role:         agentspec.RoleDev,
		SystemPrompt: "Tu es développeur.",
		Instructions: []string{"Lire le handoff"},
		Constraints:  []string{"Scope strict"},
		OutputContract: agentspec.OutputContract{
			Format:         agentspec.OutputAsagiriV1,
			RequiredFields: []string{"status", "summary"},
		},
		ContentHash: "abc123",
	}
}

func sampleInput() agentcontext.Input {
	return agentcontext.Input{
		Spec:           sampleSpec(),
		Feature:        "feat-x",
		TaskID:         "task-42",
		RunID:          "run-7",
		Phase:          "dev",
		UserTaskPrompt: "Implémente la tâche.",
		ContextFiles:   []string{"docs/ai/active/handoff.md"},
		AllowedPaths:   []string{"application/**"},
		ForbiddenPaths: []string{".env"},
	}
}

func TestBuildDeterministicContext(t *testing.T) {
	a := agentcontext.Build(sampleInput())
	b := agentcontext.Build(sampleInput())
	require.Equal(t, a, b)
	require.Equal(t, agentcontext.ModeOrchestrated, a.Mode)
	require.Equal(t, "dev", a.AgentID)
	require.Equal(t, "abc123", a.AgentHash)
}

func TestContextHashStable(t *testing.T) {
	ctx := agentcontext.Build(sampleInput())
	h1 := agentcontext.ContextHash(ctx)
	h2 := agentcontext.ContextHash(ctx)
	require.Equal(t, h1, h2)
	require.Len(t, h1, 64)
}

func TestRenderContainsOrchestratedModeAndOutputContract(t *testing.T) {
	ctx := agentcontext.Build(sampleInput())
	prompt := agentcontext.RenderPrompt(ctx)
	require.Contains(t, prompt, "Tu es exécuté par Asagiri en mode orchestré")
	require.Contains(t, prompt, "mode: orchestrated")
	require.Contains(t, prompt, "## Output contract")
	require.Contains(t, prompt, "format: asagiri-v1")
	require.Contains(t, prompt, "required_fields:")
	require.Contains(t, prompt, "status")
	require.Contains(t, prompt, "## Task prompt")
	require.Contains(t, prompt, "Implémente la tâche.")
}

func TestRenderDeterministicRegardlessOfSliceOrder(t *testing.T) {
	in := sampleInput()
	in.ContextFiles = []string{"b.md", "a.md"}
	in.Spec.Instructions = []string{"z", "a"}
	p1 := agentcontext.RenderPrompt(agentcontext.Build(in))

	in.ContextFiles = []string{"a.md", "b.md"}
	in.Spec.Instructions = []string{"a", "z"}
	p2 := agentcontext.RenderPrompt(agentcontext.Build(in))
	require.Equal(t, p1, p2)
}

func TestWriteLogsPaths(t *testing.T) {
	repo := t.TempDir()
	in := sampleInput()
	ctx, prompt, err := agentcontext.WriteFromInput(repo, in)
	require.NoError(t, err)

	dir := agentcontext.AgentLogDir(repo, ctx.TaskID, ctx.AgentID)
	require.Equal(t, filepath.Join(repo, ".asagiri", "logs", "task-42", "agents", "dev"), dir)

	contextPath := filepath.Join(dir, "context.json")
	promptPath := filepath.Join(dir, "prompt.md")
	require.FileExists(t, contextPath)
	require.FileExists(t, promptPath)

	body, err := os.ReadFile(promptPath)
	require.NoError(t, err)
	require.Equal(t, prompt, string(body))
}

func TestWriteLogsRequiresTaskAndAgentID(t *testing.T) {
	ctx := agentcontext.ExecutionContext{AgentID: "dev"}
	err := agentcontext.WriteLogs(t.TempDir(), ctx, "x")
	require.Error(t, err)
	require.Contains(t, err.Error(), "task_id")
}
