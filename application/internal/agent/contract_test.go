package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/pkg/agentflow"
)

func TestAgentContextResultRoundTrip(t *testing.T) {
	task := &agentflow.Task{
		ID:    "t-1",
		Title: "Do thing",
		Scope: agentflow.TaskScope{AllowedPaths: []string{"application/**"}},
	}
	ctx := BuildContext("run-1", task, []string{"spec.md"})
	res := DryRunResult("ok")
	repo := t.TempDir()
	if err := WriteLogs(repo, "t-1", ctx, res); err != nil {
		t.Fatal(err)
	}
	var readCtx agentflow.AgentContext
	body, _ := os.ReadFile(filepath.Join(repo, ".agentflow", "logs", "t-1", "context.json"))
	if err := json.Unmarshal(body, &readCtx); err != nil {
		t.Fatal(err)
	}
	if readCtx.TaskID != "t-1" {
		t.Fatalf("task id %q", readCtx.TaskID)
	}
}
