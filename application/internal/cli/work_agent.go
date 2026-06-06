package cli

import (
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/spf13/cobra"
)

func resolveWorkAgent(cmd *cobra.Command, flagValue string, fromCfg func(*config.Config) string, cfg *config.Config) string {
	if cmd != nil && cmd.Flags().Changed("agent") {
		return flagValue
	}
	if cfg != nil {
		return fromCfg(cfg)
	}
	return flagValue
}
