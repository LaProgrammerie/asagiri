package trustcli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/trust"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/replay"
	"github.com/spf13/cobra"
)

// ErrCIFailed is returned when verify trust --ci hits a blocking gate or failed checks.
var ErrCIFailed = errors.New("trust verification failed CI policy")

// Options wires trust CLI commands from the root `asa` package.
type Options struct {
	DryRun          *bool
	LoadWorkContext func(startDir string, dryRun bool) (*WorkContext, error)
	RunRootUI       func(cmd *cobra.Command, args []string) error
}

// RootCommand returns the `asa trust` command tree (sans TUI root par défaut injectée via RunRootUI).
func RootCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trust",
		Short: "Moteur de confiance — gates, replay, synthèse work",
		RunE:  opts.RunRootUI,
	}
	cmd.AddCommand(
		newTrustGatesCmd(),
		newTrustReplayCmd(),
		newTrustDiffCmd(opts),
		newTrustWorkTaskCmd(opts),
		newTrustWorkFeatureCmd(opts),
		newTrustWorkRunCmd(opts),
	)
	return cmd
}

// VerifyCommand returns `asa verify trust` (product flow verification).
func VerifyCommand() *cobra.Command {
	var flowFlag, task, branch, product string
	var strict, jsonOut, ci bool

	cmd := &cobra.Command{
		Use:   "trust <flow>",
		Short: "Exécuter le moteur de confiance sur un flow produit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			flow := args[0]
			if flowFlag != "" {
				flow = flowFlag
			}
			repoRoot, cfg, err := loadTrustRepoConfig()
			if err != nil {
				return err
			}
			eng, closeStore := trustEngine(repoRoot, cfg)
			if closeStore != nil {
				defer closeStore()
			}
			result, err := eng.Verify(cmd.Context(), trust.VerificationRequest{
				Flow:    flow,
				Task:    task,
				Branch:  branch,
				Strict:  strict,
				Product: product,
			})
			if err != nil {
				return err
			}
			if err := writeTrustVerifyOutput(cmd, result, jsonOut); err != nil {
				return err
			}
			if ci && trust.CIShouldFail(result.Report, strict) {
				return ErrCIFailed
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&flowFlag, "flow", "", "ID de flow (remplace l'argument positionnel)")
	cmd.Flags().StringVar(&task, "task", "", "ID de tâche associée")
	cmd.Flags().StringVar(&branch, "branch", "", "Branche git ciblée")
	cmd.Flags().BoolVar(&strict, "strict", false, "Traiter les checks en warn comme des échecs en mode CI")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON du trust report sur stdout")
	cmd.Flags().BoolVar(&ci, "ci", false, "Code de sortie non nul si gate bloqué ou checks en échec")
	cmd.Flags().StringVar(&product, "product", "", "ID produit sous .asagiri/products/")
	return cmd
}

func newTrustGatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gates",
		Short: "Afficher les gates de vérification configurées",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, cfg, err := loadTrustRepoConfig()
			if err != nil {
				return err
			}
			_, _ = fmt.Fprint(cmd.OutOrStdout(), trust.FormatGatesConfig(cfg.Verification))
			return nil
		},
	}
}

func newTrustReplayCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "replay <trust-id>",
		Short: "Rejouer une vérification à partir de replay.yaml",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, cfg, err := loadTrustRepoConfig()
			if err != nil {
				return err
			}
			manifest, err := replay.Load(repoRoot, args[0])
			if err != nil {
				return err
			}
			if manifest.Flow == "" {
				return fmt.Errorf("replay manifest for %q: flow required", args[0])
			}
			eng, closeStore := trustEngine(repoRoot, cfg)
			if closeStore != nil {
				defer closeStore()
			}
			req := trust.VerificationRequest{
				Flow:       manifest.Flow,
				Branch:     manifest.Branch,
				CheckTypes: manifest.Checks,
				Strict:     true,
			}
			result, err := eng.Verify(cmd.Context(), req)
			if err != nil {
				return err
			}
			if err := writeTrustVerifyOutput(cmd, result, jsonOut); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "replayed from: %s\n", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON du trust report sur stdout")
	return cmd
}

func loadTrustRepoConfig() (string, *config.Config, error) {
	startDir, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}
	repoRoot, err := bootstrap.GitRoot(startDir)
	if err != nil {
		return "", nil, err
	}
	cfg, err := config.Load(config.ConfigPath(repoRoot), repoRoot)
	if err != nil {
		return "", nil, err
	}
	return repoRoot, cfg, nil
}

func trustEngine(repoRoot string, cfg *config.Config) (*trust.Engine, func()) {
	eng := trust.NewEngine(repoRoot)
	eng.Gates = trust.NewGateEvaluator(&cfg.Verification)
	eng.Config = cfg
	store, err := runtime.Open(repoRoot)
	if err != nil {
		return eng, nil
	}
	eng.Emitter = trust.NewRuntimeEmitter(store)
	return eng, func() { _ = store.Close() }
}

func writeTrustVerifyOutput(cmd *cobra.Command, result trust.VerificationResult, jsonOut bool) error {
	out := cmd.OutOrStdout()
	if jsonOut {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result.Report); err != nil {
			return fmt.Errorf("encode trust report: %w", err)
		}
		return nil
	}
	_, _ = fmt.Fprint(out, trust.FormatTerminalSummary(result.Report))
	_, _ = fmt.Fprintf(out, "\nTrust ID: %s\n", result.TrustID)
	_, _ = fmt.Fprintf(out, "Reports:\n  %s\n  %s\n", result.MDPath, result.JSONPath)
	return nil
}

// RunVerify is a test hook for integration tests.
func RunVerify(ctx context.Context, repoRoot string, cfg *config.Config, req trust.VerificationRequest) (trust.VerificationResult, error) {
	eng, closeStore := trustEngine(repoRoot, cfg)
	if closeStore != nil {
		defer closeStore()
	}
	return eng.Verify(ctx, req)
}

func osGetwdMust() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}
