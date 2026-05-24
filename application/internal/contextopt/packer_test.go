package contextopt

import (
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/investigation"
)

func TestBuildPackOrdering(t *testing.T) {
	cfg := config.NewTestConfig("t")
	cfg.Validation.Commands = []config.ValidationCommand{{Name: "t", Command: "go test ./..."}}
	pack := BuildPack(cfg, PackInput{
		Feature: "f",
		TaskID:  "1",
		ReducedFiles: []FileEntry{
			{RelPath: "low.go", Score: 1, Content: "a"},
			{RelPath: "high.go", Score: 9, Content: "b"},
		},
		Inv: investigation.InvestigationResult{CandidateFiles: []string{"high.go"}},
	})
	if pack.FileExcerpts[0].Path != "high.go" {
		t.Fatalf("order: %v", pack.FileExcerpts[0].Path)
	}
}
