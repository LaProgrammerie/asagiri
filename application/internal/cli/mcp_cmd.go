package cli

import (
	"fmt"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/mcp"
	"github.com/spf13/cobra"
)

func newMcpCmd(dryRun *bool) *cobra.Command {
	root := &cobra.Command{Use: "mcp", Short: "Serveur MCP local"}
	root.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Démarrer MCP stdio",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := loadContext(mustWd(), *dryRun)
			if err != nil {
				return err
			}
			defer c.Close()
			if !c.Config.MCP.Enabled {
				return fmt.Errorf("MCP désactivé : mcp.enabled: true dans .agentflow/config.yaml")
			}
			return mcp.ServeStdio(c.RepoRoot, c.Config)
		},
	})
	return root
}
