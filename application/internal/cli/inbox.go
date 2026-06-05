package cli

import (
	"context"
	"fmt"
	"os"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/source"
	"github.com/spf13/cobra"
)

func newInboxCmd(dryRun *bool) *cobra.Command {
	var sourceFilter string
	_ = dryRun

	cmd := &cobra.Command{
		Use:   "inbox",
		Short: "Lister les specs candidates depuis les sources",
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

			reg := newSourceRegistry(ctx, notionClientFromConfig(ctx))
			var rows []intent.InboxRow

			listSource := func(name string, src source.Source) error {
				items, err := src.List(context.Background())
				if err != nil {
					return err
				}
				for _, it := range items {
					updated := ""
					if !it.UpdatedAt.IsZero() {
						updated = it.UpdatedAt.Format("2006-01-02 15:04")
					}
					rows = append(rows, intent.InboxRow{
						Source:  name,
						Status:  it.Status,
						Updated: updated,
						Feature: it.Feature,
						Path:    it.PathHint,
					})
				}
				return nil
			}

			if sourceFilter == "" || sourceFilter == "local" {
				if err := listSource("local", reg.Local); err != nil {
					return err
				}
			}
			if sourceFilter == "" || sourceFilter == "notion" {
				if reg.Notion != nil {
					if err := listSource("notion", reg.Notion); err != nil {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "notion: %v\n", err)
					}
				} else if sourceFilter == "notion" {
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "notion: source non configurée (sources.notion.enabled + NOTION_TOKEN)")
				}
			}

			if len(rows) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Inbox vide.")
				return nil
			}
			intent.PrintInboxTable(cmd.OutOrStdout(), rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&sourceFilter, "source", "", "Filtrer: notion|local")
	return cmd
}
