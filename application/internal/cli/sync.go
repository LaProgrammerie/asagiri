package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/source"
	"github.com/spf13/cobra"
)

func newSyncCmd(dryRun *bool) *cobra.Command {
	var pageURL, feature string
	var force bool
	_ = dryRun

	cmd := &cobra.Command{
		Use:   "sync [notion|all]",
		Short: "Synchroniser une source externe vers le repo local",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(startDir, false)
			if err != nil {
				return err
			}
			defer ctx.Close()

			target := "notion"
			if len(args) > 0 {
				target = args[0]
			}

			reg := newSourceRegistry(ctx, notionClientFromConfig(ctx))
			opts := source.SyncOptions{Force: force, Interactive: isInteractive()}

			syncOne := func(src source.Source, ref source.SourceRef, feat string) error {
				dest := source.LocalSpecPath{
					Root:    ctx.Config.Sources.Notion.ImportPath,
					Feature: feat,
				}
				if !force && opts.Interactive {
					specPath := fmt.Sprintf("%s/%s/spec.md", dest.Root, feat)
					if _, err := os.Stat(ctx.Config.Resolve(ctx.RepoRoot, specPath)); err == nil {
						if err := requireConfirm(intent.WorkOptions{Interactive: true}, "Écraser la spec locale modifiée?"); err != nil {
							return err
						}
					}
				}
				if ctx.DryRun || *dryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "dry-run: sync %s → %s/%s\n", ref.URL, dest.Root, feat)
					return nil
				}
				res, err := src.Sync(context.Background(), ref, dest, opts)
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "synced %s → %s\n", res.Feature, res.Path)
				return nil
			}

			switch target {
			case "all":
				if reg.Notion != nil {
					items, err := reg.Notion.List(context.Background())
					if err != nil {
						return err
					}
					for _, it := range items {
						if err := syncOne(reg.Notion, it.Ref, it.Feature); err != nil {
							return err
						}
					}
				}
				return nil
			case "notion":
				src, err := reg.byName("notion")
				if err != nil {
					return err
				}
				ref := source.SourceRef{}
				feat := feature
				if pageURL != "" {
					ref.URL = pageURL
					ref.ID = intent.ParseNotionPageID(pageURL)
				}
				if feat != "" {
					ref.Name = feat
					items, _ := src.List(context.Background())
					for _, it := range items {
						if it.Feature == feat {
							ref = it.Ref
							break
						}
					}
				}
				return syncOne(src, ref, feat)
			default:
				return fmt.Errorf("source sync inconnue: %s", target)
			}
		},
	}
	cmd.Flags().StringVar(&pageURL, "page", "", "URL page Notion")
	cmd.Flags().StringVar(&feature, "feature", "", "Feature slug")
	cmd.Flags().BoolVar(&force, "force", false, "Écraser spec locale sans confirmation")
	return cmd
}
