package plan

import (
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestNormalizeMarkdownTasks(t *testing.T) {
	doc := &spec.Document{
		Feature: "agentflow-test",
		Tasks: `
- [ ] Ajouter la commande plan
- [ ] Ajouter la commande dev
`,
	}
	tasks, err := Normalize("agentflow-test", doc)
	require.NoError(t, err)
	require.Len(t, tasks, 2)
	require.Equal(t, "agentflow-test-001", tasks[0].ID)
	require.Equal(t, "Ajouter la commande plan", tasks[0].Title)
}

func TestNormalizeFallbackTask(t *testing.T) {
	doc := &spec.Document{Feature: "x", Active: "texte libre"}
	tasks, err := Normalize("x", doc)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	require.Contains(t, tasks[0].Title, "Implémenter")
}
