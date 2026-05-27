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
		Short: "Engineering analysis layer (structural graphs)",
	}
	var productID string
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build and write analysis graphs for a product",
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
			fmt.Fprintf(cmd.OutOrStdout(), "analysis graphs: %s (%d kinds)\n", path, len(b.Graphs))
			for k, g := range b.Graphs {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s: %d nodes, %d edges\n", k, len(g.Nodes), len(g.Edges))
			}
			return nil
		},
	}
	buildCmd.Flags().StringVar(&productID, "product", "", "Product id under .asagiri/products/")
	cmd.AddCommand(buildCmd)
	return cmd
}
