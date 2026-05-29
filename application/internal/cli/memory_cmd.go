package cli

import (
	"fmt"
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/memory"
	"github.com/LaProgrammerie/asagiri/application/internal/memory/embedder"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/spf13/cobra"
)

func configureMemoryEmbedder(repoRoot string) error {
	cfg, err := config.Load(config.ConfigPath(repoRoot), repoRoot)
	if err != nil {
		return err
	}
	return embedder.ConfigureFromConfig(cfg.Runtime.Memory)
}

func newMemoryDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "doctor",
		Short:   "Vérifier Ollama, dimensions des embeddings et entrées orphelines",
		Example: "  asa memory doctor",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			if err := configureMemoryEmbedder(root); err != nil {
				return err
			}
			store, err := runtime.Open(root)
			if err != nil {
				return err
			}
			defer store.Close()
			checks, err := memory.NewEngine(store).Doctor(cmd.Context())
			if err != nil {
				return err
			}
			return memory.FormatDoctor(cmd.OutOrStdout(), checks)
		},
	}
}

func newMemoryReindexCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "reindex",
		Short:   "Recalculer les embeddings de toutes les entrées mémoire",
		Example: "  asa memory reindex\n  # Après runtime.memory.embedder (hash, ollama, …) dans config.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			if err := configureMemoryEmbedder(root); err != nil {
				return err
			}
			store, err := runtime.Open(root)
			if err != nil {
				return err
			}
			defer store.Close()
			n, err := memory.NewEngine(store).Reindex(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "reindexed: %d\n", n)
			return nil
		},
	}
}
