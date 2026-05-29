package cli

import (
	"fmt"
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/product"
	uiapp "github.com/LaProgrammerie/asagiri/application/internal/ui/app"
	"github.com/spf13/cobra"
)

func newPrototypeCmd(dryRun *bool) *cobra.Command {
	var productName string
	var stack string
	var style string
	cmd := &cobra.Command{
		Use:   "prototype",
		Short: "Gérer les prototypes produit exécutables",
		RunE:  runUIScreenCommandWithOptions(dryRun, uiapp.ScreenPrototype, nil),
	}
	createCmd := &cobra.Command{
		Use:   `create "<intent>"`,
		Short: "Créer un prototype local déterministe",
		Args:  cobra.ExactArgs(1),
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
			svc := product.NewService(actx.RepoRoot)
			name, err := svc.CreatePrototype(product.CreatePrototypeOptions{
				Intent:  args[0],
				Product: productName,
				Stack:   stack,
				Style:   style,
				DryRun:  actx.DryRun || *dryRun,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "prototype créé: %s\n", name)
			return nil
		},
	}
	createCmd.Flags().StringVar(&productName, "product", "", "Slug produit (sinon dérivé de l'intention)")
	createCmd.Flags().StringVar(&stack, "stack", "react", "Stack prototype")
	createCmd.Flags().StringVar(&style, "style", "minimal", "Style prototype")

	runCmd := &cobra.Command{
		Use:     "run <product>",
		Short:   "Lancer le serveur dev du prototype (npm run dev)",
		Example: "  asa prototype run workspace-saas",
		Args:    cobra.ExactArgs(1),
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
			svc := product.NewService(actx.RepoRoot)
			result, err := svc.RunPrototype(product.PrototypeRunOptions{
				Product: args[0],
				DryRun:  actx.DryRun || *dryRun,
			})
			if err != nil {
				return err
			}
			if actx.DryRun || *dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "prototype %s prêt (%s). cd %s && npm run dev\n", result.Product, result.URL, result.Dir)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "prototype %s démarré (%s) pid=%d dir=%s\n", result.Product, result.URL, result.PID, result.Dir)
			return nil
		},
	}

	patchCmd := &cobra.Command{
		Use:   `patch <product> "<instruction>"`,
		Short: "Patcher un prototype existant",
		Args:  cobra.ExactArgs(2),
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
			svc := product.NewService(actx.RepoRoot)
			if err := svc.PatchPrototype(args[0], args[1], actx.DryRun || *dryRun); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "prototype patché: %s\n", product.Slug(args[0]))
			return nil
		},
	}

	cmd.AddCommand(createCmd, runCmd, patchCmd)
	return cmd
}

func newFlowsCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flows",
		Short: "Extraire et inspecter les flows produit",
	}
	extractCmd := &cobra.Command{
		Use:   "extract <product>",
		Short: "Générer flows et screens depuis le prototype",
		Args:  cobra.ExactArgs(1),
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
			svc := product.NewService(actx.RepoRoot)
			if err := svc.ExtractFlows(args[0], actx.DryRun || *dryRun); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "flows extraits: %s\n", product.Slug(args[0]))
			return nil
		},
	}
	inspectCmd := &cobra.Command{
		Use:   "inspect <product>",
		Short: "Afficher un résumé des flows extraits",
		Args:  cobra.ExactArgs(1),
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
			svc := product.NewService(actx.RepoRoot)
			summary, err := svc.InspectFlows(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), summary)
			return nil
		},
	}
	e2eCmd := &cobra.Command{
		Use:     "e2e <product>",
		Short:   "Générer un squelette de test E2E depuis un flow YAML",
		Example: "  asa flows e2e workspace-saas --flow workspace-onboarding",
		Args:    cobra.ExactArgs(1),
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
			flowID, _ := cmd.Flags().GetString("flow")
			runner, _ := cmd.Flags().GetString("runner")
			svc := product.NewService(actx.RepoRoot)
			result, err := svc.GenerateE2ETest(product.E2EGeneratorOptions{
				Product: args[0],
				FlowID:  flowID,
				Runner:  runner,
				DryRun:  actx.DryRun || *dryRun,
			})
			if err != nil {
				return err
			}
			if actx.DryRun || *dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "e2e skeleton (%s): %s (%d steps)\n", result.Runner, result.Path, result.StepCount)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "e2e généré: %s (%s, %d steps)\n", result.Path, result.Runner, result.StepCount)
			return nil
		},
	}
	e2eCmd.Flags().String("flow", "", "Flow id (fichier .asagiri/products/<product>/flows/<id>.flow.yaml)")
	e2eCmd.Flags().String("runner", "playwright", "Runner cible: playwright ou cypress")
	_ = e2eCmd.MarkFlagRequired("flow")

	reviewCmd := &cobra.Command{
		Use:   "review <product>",
		Short: "Analyser les gaps métier/metrics/observabilité des flows",
		Args:  cobra.ExactArgs(1),
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
			svc := product.NewService(actx.RepoRoot)
			summary, err := svc.ReviewFlows(args[0], actx.DryRun || *dryRun)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), summary)
			return nil
		},
	}
	cmd.AddCommand(extractCmd, inspectCmd, reviewCmd, e2eCmd)
	return cmd
}

func newContractsCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contracts",
		Short: "Extraire les contrats système depuis les flows",
	}
	extractCmd := &cobra.Command{
		Use:   "extract <product>",
		Short: "Générer OpenAPI et contrats dérivés",
		Args:  cobra.ExactArgs(1),
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
			svc := product.NewService(actx.RepoRoot)
			if err := svc.ExtractContracts(args[0], actx.DryRun || *dryRun); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "contracts extraits: %s\n", product.Slug(args[0]))
			return nil
		},
	}
	cmd.AddCommand(extractCmd)
	return cmd
}

func newProductCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "product",
		Short: "Review produit",
	}
	cmd.AddCommand(newProductReviewSubCmd(dryRun))
	return cmd
}

func newProductReviewSubCmd(dryRun *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "review <product>",
		Short: "Analyser flows/screens/contracts d'un produit",
		Args:  cobra.ExactArgs(1),
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
			svc := product.NewService(actx.RepoRoot)
			summary, err := svc.ReviewProduct(args[0], actx.DryRun || *dryRun)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "review produit %s: %s\n", product.Slug(args[0]), summary)
			return nil
		},
	}
}

func newSpecGenerateFromProductCmd(dryRun *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "generate-from-product <product>",
		Short: "Générer specs et tasks depuis un produit extrait",
		Args:  cobra.ExactArgs(1),
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
			svc := product.NewService(actx.RepoRoot)
			if err := svc.GenerateSpecFromProduct(args[0], actx.DryRun || *dryRun); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "spec générée depuis produit: %s\n", product.Slug(args[0]))
			return nil
		},
	}
}

func newArchitectureCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "architecture",
		Short: "Projeter les implications système depuis les flows",
	}
	deriveCmd := &cobra.Command{
		Use:   "derive <product>",
		Short: "Dériver API/async/sécurité/observabilité/infra",
		Args:  cobra.ExactArgs(1),
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
			svc := product.NewService(actx.RepoRoot)
			summary, err := svc.DeriveArchitecture(args[0], actx.DryRun || *dryRun)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), summary)
			return nil
		},
	}
	cmd.AddCommand(deriveCmd)
	return cmd
}
