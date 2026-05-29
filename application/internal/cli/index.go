package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/rag"
	"github.com/spf13/cobra"
)

func newIndexCmd(dryRun *bool) *cobra.Command {
	var skipEmbeddings bool
	cmd := &cobra.Command{
		Use:     "index",
		Short:   "Indexer le dépôt pour le RAG local (sémantique + mots-clés, spec §10.3)",
		Example: "  asa index\n  asa index --dry-run\n  asa index search \"authentication middleware\"",
		Long: `Indexe le dépôt dans .asagiri/index/chunks.sqlite.

À la construction, les chunks reçoivent des embeddings via runtime.memory.embedder
(même embedder que asa memory reindex, PF-A-01). La recherche utilise la similarité cosinus
par défaut ; --keyword force le mode LIKE SQL.

Reconstruire l’index après changement d’embedder (ollama, cloud, hash).`,
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
				RepoRoot:       ctx.RepoRoot,
				Paths:          rag.DefaultIndexPaths(),
				DryRun:         ctx.DryRun,
				Memory:         ctx.Config.Runtime.Memory,
				SkipEmbeddings: skipEmbeddings,
			})
			if err != nil {
				return err
			}
			if res.EmbedderConfigured {
				fmt.Fprintf(cmd.OutOrStdout(), "index: %d fichiers, %d chunks (%d embeddings) → %s\n",
					res.Files, res.Chunks, res.EmbeddedChunks, res.DBPath)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "index: %d fichiers, %d chunks (keyword-only) → %s\n",
					res.Files, res.Chunks, res.DBPath)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&skipEmbeddings, "skip-embeddings", false, "Indexer sans vecteurs (recherche LIKE uniquement)")
	cmd.AddCommand(newIndexSearchCmd(dryRun))
	return cmd
}

func newIndexSearchCmd(dryRun *bool) *cobra.Command {
	var keyword bool
	var limit int
	cmd := &cobra.Command{
		Use:     "search <query>",
		Short:   "Rechercher des chemins dans l’index RAG",
		Args:    cobra.MinimumNArgs(1),
		Example: "  asa index search \"graph checkpoint resume\"\n  asa index search \"login\" --keyword",
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			actx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer actx.Close()
			if actx.DryRun {
				fmt.Fprintln(cmd.OutOrStdout(), "dry-run: recherche index ignorée")
				return nil
			}
			db, err := rag.OpenDB(actx.RepoRoot)
			if err != nil {
				return err
			}
			defer db.Close()
			query := strings.Join(args, " ")
			paths, err := rag.NewRetriever(db).SearchWithOptions(cmd.Context(), query, rag.SearchOptions{
				Limit:       limit,
				KeywordOnly: keyword,
				Memory:      actx.Config.Runtime.Memory,
			})
			if err != nil {
				return err
			}
			for _, p := range paths {
				fmt.Fprintln(cmd.OutOrStdout(), p)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&keyword, "keyword", false, "Recherche LIKE par mots-clés (sans similarité cosinus)")
	cmd.Flags().IntVar(&limit, "limit", 8, "Nombre maximal de chemins distincts")
	return cmd
}
