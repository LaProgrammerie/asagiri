package cli

import (
	"github.com/LaProgrammerie/asagiri/application/internal/knowledgecli"
	"github.com/spf13/cobra"
)

func newKnowledgeCmd(dryRun *bool) *cobra.Command {
	return knowledgecli.RootCommand(knowledgecli.Options{
		DryRun:      dryRun,
		LoadContext: adaptKnowledgeContext,
		RunRootUI:   runUIScreenCommand(dryRun, "knowledge"),
	})
}

func adaptKnowledgeContext(startDir string, dryRun bool) (*knowledgecli.Context, error) {
	ctx, err := loadContext(startDir, dryRun)
	if err != nil {
		return nil, err
	}
	return knowledgecli.NewContext(ctx.RepoRoot, ctx.Config, ctx.Close), nil
}
