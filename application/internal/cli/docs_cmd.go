package cli

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/cli/docgen"
	"github.com/spf13/cobra"
)

func newDocsCmd() *cobra.Command {
	const defaultOutputDir = "docs-site/content/docs/en/cli/generated"

	var output string
	gen := &cobra.Command{
		Use:   "generate-cli",
		Short: "Render MDX reference pages derived from this CLI's cobra tree",
		Long: strings.TrimSpace(`
The generator walks every command except the root binary and writes one deterministic MDX file per reachable node.
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := docgen.Generate(cmd.Root(), output); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "generated CLI docs in %s\n", output)
			return nil
		},
	}
	gen.Flags().StringVar(&output, "output", defaultOutputDir, "Directory receiving generated MDX files")

	docs := &cobra.Command{
		Use:   "docs",
		Short: "Documentation tooling for Asagiri artefacts",
	}
	docs.AddCommand(gen)

	return docs
}
