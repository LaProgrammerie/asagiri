package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

func newImpactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "impact",
		Short: "Analyser l'impact des changements via le graphe",
	}
	cmd.AddCommand(newImpactAnalyzeCmd())
	return cmd
}

func newImpactAnalyzeCmd() *cobra.Command {
	var (
		file    string
		flow    string
		action  string
		jsonOut bool
	)
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyser l'impact d'un fichier, flow ou action",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := loadContext(mustWd(), false)
			if err != nil {
				return err
			}
			defer c.Close()

			store, err := knowledge.OpenStore(c.RepoRoot)
			if err != nil {
				return err
			}
			defer store.Close()

			analyzer := knowledge.NewImpactAnalyzer(store)
			result, err := analyzer.Analyze(cmd.Context(), knowledge.ImpactRequest{
				File:   file,
				Flow:   flow,
				Action: action,
			})
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			fmt.Fprint(cmd.OutOrStdout(), knowledge.FormatImpactAnalysis(result))
			return nil
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "Fichier source modifié")
	cmd.Flags().StringVar(&flow, "flow", "", "ID de flow")
	cmd.Flags().StringVar(&action, "action", "", "Nom d'action dans le flow")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON")
	return cmd
}
