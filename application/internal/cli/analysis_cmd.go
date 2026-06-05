package cli

import (
	"fmt"
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis"
	"github.com/spf13/cobra"
)

func newAnalysisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analysis",
		Short: "Analyse structurelle du code (graphes de dépendances)",
	}
	var productID string
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Construire les graphes d'analyse pour un produit",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			if productID == "" {
				productID, err = analysis.DefaultProduct(root)
				if err != nil {
					return err
				}
			}
			b, err := analysis.BuildAll(root, productID)
			if err != nil {
				return err
			}
			path, err := analysis.WriteBundle(root, b)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "analysis graphs: %s (%d kinds)\n", path, len(b.Graphs))
			for k, g := range b.Graphs {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s: %d nodes, %d edges\n", k, len(g.Nodes), len(g.Edges))
			}
			return nil
		},
	}
	buildCmd.Flags().StringVar(&productID, "product", "", "Product id under .asagiri/products/")
	cmd.AddCommand(buildCmd)
	return cmd
}
