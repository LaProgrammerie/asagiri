package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
	"github.com/spf13/cobra"
)

func newInvestigateCmd(dryRun *bool) *cobra.Command {
	var (
		taskID          string
		flow            string
		runID           string
		fromFailedTests bool
		depth           string
		maxFiles        int
		maxDurationMin  int
		noCloud         bool
		estimateOnly    bool
		output          string
	)
	cmd := &cobra.Command{
		Use:   `investigate "<symptom>"`,
		Short: "Investigation locale structurée",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := loadContext(mustWd(), *dryRun)
			if err != nil {
				return err
			}
			defer c.Close()
			req := investigation.Request{
				Symptom:         args[0],
				Feature:         args[0],
				Flow:            flow,
				TaskID:          taskID,
				RunID:           runID,
				FromFailedTests: fromFailedTests,
				Depth:           investigation.Depth(depth),
				MaxFiles:        maxFiles,
				NoCloud:         noCloud,
				EstimateOnly:    estimateOnly,
				Output:          output,
				RepoRoot:        c.RepoRoot,
			}
			if maxDurationMin > 0 {
				req.MaxDuration = time.Duration(maxDurationMin) * time.Minute
			}
			if req.Output == "" {
				req.Output = "markdown"
			}
			rep, err := investigation.RunInvestigation(cmd.Context(), req, c.Config)
			if err != nil {
				return err
			}
			if !estimateOnly && !c.DryRun && !*dryRun {
				_ = investigation.FeedMemory(c.RepoRoot, rep)
			}
			if output == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(rep)
			}
			reportPath := ""
			if !estimateOnly {
				reportPath = fmt.Sprintf(".asagiri/investigations/%s/report.md", rep.ID)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Investigation: %s\n", rep.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Scope flow: %s | hypotheses: %d | evidence: %d\n",
				rep.Scope.Flow, len(rep.Hypotheses), len(rep.Evidence))
			if reportPath != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Report: %s\n", reportPath)
			}
			if rep.ContextPackPath != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Context pack: %s\n", rep.ContextPackPath)
			}
			if rep.ReplayPackPath != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Replay pack: %s\n", rep.ReplayPackPath)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Estimated tokens: %d | cost: €%.4f\n", rep.EstimateTokens, rep.EstimateCostEUR)
			for _, h := range rep.RootCauseCandidates {
				fmt.Fprintf(cmd.OutOrStdout(), "  candidate (%.2f): %s\n", h.Score, h.Statement)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "ID de tâche")
	cmd.Flags().StringVar(&flow, "flow", "", "Flow produit")
	cmd.Flags().StringVar(&runID, "run", "", "ID de run")
	cmd.Flags().BoolVar(&fromFailedTests, "from-failed-tests", false, "Cibler les tests en échec")
	cmd.Flags().StringVar(&depth, "depth", "standard", "Profondeur: quick|standard|deep|ci")
	cmd.Flags().IntVar(&maxFiles, "max-files", 0, "Nombre max de fichiers dans le context pack")
	cmd.Flags().IntVar(&maxDurationMin, "max-duration", 0, "Durée max (minutes)")
	cmd.Flags().BoolVar(&noCloud, "no-cloud", false, "Investigation 100% locale")
	cmd.Flags().BoolVar(&estimateOnly, "estimate-only", false, "Estimation sans rapport complet")
	cmd.Flags().StringVar(&output, "output", "markdown", "Sortie: markdown|json")

	var impactFlow, impactChange, impactProduct string
	impactCmd := &cobra.Command{
		Use:   "impact",
		Short: "Analyser l'impact d'un changement produit",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := mustWd()
			if impactFlow == "" || impactChange == "" {
				return fmt.Errorf("--flow and --change are required")
			}
			rep, err := investigation.RunImpact(cmd.Context(), investigation.ImpactRequest{
				Flow: impactFlow, Change: impactChange, ProductID: impactProduct, RepoRoot: root,
			})
			if err != nil {
				return err
			}
			path, err := investigation.WriteImpactReport(root, rep)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Impact: %s\nAffected: %d files\nReport: %s\n",
				rep.ID, len(rep.Affected), path)
			return nil
		},
	}
	impactCmd.Flags().StringVar(&impactFlow, "flow", "", "Flow concerné")
	impactCmd.Flags().StringVar(&impactChange, "change", "", "Description du changement")
	impactCmd.Flags().StringVar(&impactProduct, "product", "workspace-saas", "Produit")

	graphCmd := &cobra.Command{
		Use:   "graph <investigation-id>",
		Short: "Afficher le graphe root-cause d'une investigation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := mustWd()
			path := fmt.Sprintf(".asagiri/investigations/%s/graph.json", args[0])
			raw, err := os.ReadFile(filepath.Join(root, path))
			if err != nil {
				return err
			}
			var g investigation.RootCauseGraph
			if err := json.Unmarshal(raw, &g); err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), investigation.FormatGraphPlain(g))
			return nil
		},
	}
	cmd.AddCommand(graphCmd, impactCmd)
	return cmd
}
