package cli

import (
	"fmt"
	"os"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/rag"
	"github.com/spf13/cobra"
)

func newIndexCmd(dryRun *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "index",
		Short: "Indexer le dépôt pour le RAG local (spec §10.3)",
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()
			res, err := rag.Index(rag.IndexOptions{
				RepoRoot: ctx.RepoRoot,
				Paths:    rag.DefaultIndexPaths(),
				DryRun:   ctx.DryRun,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "index: %d fichiers, %d chunks → %s\n", res.Files, res.Chunks, res.DBPath)
			return nil
		},
	}
}
