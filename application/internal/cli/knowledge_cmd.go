package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/extractors"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
)

func newKnowledgeCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "knowledge",
		Short: "Construire et interroger le graphe de connaissance",
		RunE:  runUIScreenCommand(dryRun, "knowledge"),
	}
	cmd.AddCommand(
		newKnowledgeBuildCmd(),
		newKnowledgeQueryCmd(),
		newKnowledgeExplainCmd(),
		newKnowledgeSnapshotCmd(),
	)
	return cmd
}

func newKnowledgeBuildCmd() *cobra.Command {
	var (
		incremental      bool
		scope            string
		includeCode      bool
		includeFlows     bool
		includeTests     bool
		includeInfra     bool
		includeADR       bool
		includeRuntime   bool
		includeContracts bool
		jsonOut          bool
	)
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Construire ou mettre à jour le graphe de connaissance",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := loadContext(mustWd(), false)
			if err != nil {
				return err
			}
			defer c.Close()

			staleReport, _ := knowledge.DefaultStalenessDetector().Check(cmd.Context(), c.RepoRoot)
			if !jsonOut && c.Config.Knowledge.WarnOnStale && staleReport.Stale {
				_, _ = fmt.Fprint(cmd.OutOrStdout(), knowledge.FormatStaleness(staleReport))
			}

			req := knowledge.BuildRequestFromConfig(c.RepoRoot, c.Config)
			knowledge.ApplyBuildFlags(&req, knowledge.BuildCLIFlags{
				Incremental:         incremental,
				IncrementalSet:      cmd.Flags().Changed("incremental"),
				Scope:               scope,
				ScopeSet:            cmd.Flags().Changed("scope"),
				IncludeFlows:        includeFlows,
				IncludeFlowsSet:     cmd.Flags().Changed("include-flows"),
				IncludeContracts:    includeContracts,
				IncludeContractsSet: cmd.Flags().Changed("include-contracts"),
				IncludeCode:         includeCode,
				IncludeCodeSet:      cmd.Flags().Changed("include-code"),
				IncludeTests:        includeTests,
				IncludeTestsSet:     cmd.Flags().Changed("include-tests"),
				IncludeInfra:        includeInfra,
				IncludeInfraSet:     cmd.Flags().Changed("include-infra"),
				IncludeADR:          includeADR,
				IncludeADRSet:       cmd.Flags().Changed("include-adr"),
				IncludeRuntime:      includeRuntime,
				IncludeRuntimeSet:   cmd.Flags().Changed("include-runtime"),
			})

			result, err := knowledge.DefaultBuilder().Build(cmd.Context(), req)
			if err != nil {
				return err
			}
			result.StaleFiles = staleReport.FilesChanged

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(struct {
					knowledge.BuildResult
					Staleness knowledge.StalenessReport `json:"staleness,omitempty"`
				}{
					BuildResult: result,
					Staleness:   staleReport,
				})
			}
			knowledge.WriteKnowledgeBuild(cmd.OutOrStdout(), result, staleReport)
			return nil
		},
	}
	cmd.Flags().BoolVar(&incremental, "incremental", false, "Mise à jour incrémentale (rebuild complet si pas de métadonnées)")
	cmd.Flags().StringVar(&scope, "scope", "", "Périmètre produit (ex. product:workspace-saas)")
	cmd.Flags().BoolVar(&includeCode, "include-code", false, "Inclure l'extraction code")
	cmd.Flags().BoolVar(&includeFlows, "include-flows", true, "Inclure les flows produit")
	cmd.Flags().BoolVar(&includeTests, "include-tests", false, "Inclure les tests")
	cmd.Flags().BoolVar(&includeInfra, "include-infra", false, "Inclure l'infrastructure")
	cmd.Flags().BoolVar(&includeADR, "include-adr", false, "Inclure les ADR (docs/decisions)")
	cmd.Flags().BoolVar(&includeRuntime, "include-runtime", false, "Inclure les événements runtime")
	cmd.Flags().BoolVar(&includeContracts, "include-contracts", true, "Inclure contrats OpenAPI, events, permissions et observability")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON")
	return cmd
}

func newKnowledgeSnapshotCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Créer un snapshot du graphe de connaissance",
		RunE: func(cmd *cobra.Command, _ []string) error {
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("knowledge snapshot: --name requis")
			}
			c, err := loadContext(mustWd(), false)
			if err != nil {
				return err
			}
			defer c.Close()

			result, err := knowledge.DefaultSnapshotter().Snapshot(cmd.Context(), knowledge.SnapshotRequest{
				RepoRoot: c.RepoRoot,
				Name:     name,
			})
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Snapshot saved: %s\n", result.Path)
			return nil
		},
	}
	cmd.Flags().String("name", "", "Nom du snapshot")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON")
	return cmd
}

func newKnowledgeQueryCmd() *cobra.Command {
	var (
		nodeID   string
		nodeType string
		fromNode string
		startID  string
		edgeType string
		maxDepth int
		limit    int
		jsonOut  bool
	)
	cmd := &cobra.Command{
		Use:   "query [phrase]",
		Short: "Interroger le graphe de connaissance",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := loadContext(mustWd(), false)
			if err != nil {
				return err
			}
			defer c.Close()

			store, err := knowledge.OpenStore(c.RepoRoot)
			if err != nil {
				return err
			}
			defer func() { _ = store.Close() }()

			q := knowledge.NewQuerier(store)
			var result knowledge.GraphQueryResult

			if len(args) > 0 {
				phrase := strings.Join(args, " ")
				if parsed, ok := knowledge.ParseQueryPhrase(phrase); ok {
					result, err = q.RunPhraseQuery(cmd.Context(), parsed)
				} else {
					return fmt.Errorf("phrase non reconnue: %q (utiliser les flags --node, --type, --from)", phrase)
				}
			} else {
				gq := knowledge.GraphQuery{
					NodeID:     nodeID,
					FromNodeID: fromNode,
					StartID:    startID,
					MaxDepth:   maxDepth,
					Limit:      limit,
				}
				if nodeType != "" {
					gq.NodeType = knowledge.NodeType(nodeType)
				}
				if edgeType != "" {
					gq.EdgeType = knowledge.EdgeType(edgeType)
				}
				result, err = q.Query(cmd.Context(), gq)
			}
			if err != nil {
				return err
			}

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			printQueryResult(cmd.OutOrStdout(), result)
			return nil
		},
	}
	cmd.Flags().StringVar(&nodeID, "node", "", "Filtrer par ID de nœud")
	cmd.Flags().StringVar(&nodeType, "type", "", "Filtrer par type de nœud")
	cmd.Flags().StringVar(&fromNode, "from", "", "Filtrer les arêtes sortantes depuis ce nœud")
	cmd.Flags().StringVar(&startID, "start", "", "Démarrer une traversée BFS depuis ce nœud")
	cmd.Flags().StringVar(&edgeType, "edge-type", "", "Filtrer par type d'arête")
	cmd.Flags().IntVar(&maxDepth, "max-depth", 0, "Profondeur BFS maximale")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limite de résultats")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON")
	return cmd
}

func newKnowledgeExplainCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "explain <flow> <action> <symbol>",
		Short: "Expliquer le lien le plus court entre flow/action et un symbole",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := loadContext(mustWd(), false)
			if err != nil {
				return err
			}
			defer c.Close()

			store, err := knowledge.OpenStore(c.RepoRoot)
			if err != nil {
				return err
			}
			defer func() { _ = store.Close() }()

			q := knowledge.NewQuerier(store)
			result, err := q.ExplainShortestPath(cmd.Context(), knowledge.ExplainRequest{
				Flow:   args[0],
				Action: args[1],
				Symbol: args[2],
			})
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprint(cmd.OutOrStdout(), knowledge.FormatKnowledgeExplain(result))
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON")
	return cmd
}

func printQueryResult(out interface{ Write([]byte) (int, error) }, result knowledge.GraphQueryResult) {
	_, _ = fmt.Fprintf(out, "Nodes (%d)\n", len(result.Nodes))
	for _, n := range result.Nodes {
		_, _ = fmt.Fprintf(out, "  %s  %s\n", n.ID, n.Name)
	}
	_, _ = fmt.Fprintf(out, "Edges (%d)\n", len(result.Edges))
	for _, e := range result.Edges {
		_, _ = fmt.Fprintf(out, "  %s  %s -> %s\n", e.Type, e.From, e.To)
	}
}
