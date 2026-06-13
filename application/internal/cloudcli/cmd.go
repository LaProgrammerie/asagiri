package cloudcli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/cloud"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/spf13/cobra"
)

// RootCommand returns the `asa cloud` command tree.
func RootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud",
		Short: "Intégration Team Cloud optionnelle (login, sync runs)",
		Long:  "Commandes opt-in vers Asagiri Cloud. Le CLI reste 100 % fonctionnel sans compte.",
	}
	cmd.AddCommand(
		newStatusCmd(),
		newLoginCmd(),
		newLogoutCmd(),
		newLinkCmd(),
		newPushCmd(),
	)
	return cmd
}

func loadRepoConfig() (string, *config.Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}
	repoRoot, err := bootstrap.GitRoot(cwd)
	if err != nil {
		return "", nil, err
	}
	cfg, err := config.Load(config.ConfigPath(repoRoot), repoRoot)
	if err != nil {
		return "", nil, fmt.Errorf("cloud: config: %w", err)
	}
	return repoRoot, cfg, nil
}

func resolveToken(cfg *config.Config) (string, string, error) {
	path, err := cloud.TokenPath(cfg)
	if err != nil {
		return "", "", err
	}
	token, err := cloud.LoadToken(path)
	if err != nil {
		return "", "", err
	}
	return token, path, nil
}

func newStatusCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "État cloud (config, token, API)",
		Example: "  asa cloud status\n  asa cloud status --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg *config.Config
			token := ""
			tokenPath := ""
			if repoRoot, c, err := loadRepoConfig(); err == nil {
				_ = repoRoot
				cfg = c
				token, tokenPath, err = resolveToken(cfg)
				if err != nil {
					return err
				}
			} else {
				expanded, expandErr := cloud.ExpandPath(config.DefaultCloudTokenRel)
				if expandErr == nil {
					tokenPath = expanded
					token, _ = cloud.LoadToken(expanded)
				}
			}

			report := cloud.BuildStatus(cmd.Context(), cloud.StatusOptions{
				Config:    cfg,
				TokenPath: tokenPath,
				Token:     token,
				CheckAPI:  token != "",
			})
			out := cmd.OutOrStdout()
			if jsonOut {
				return cloud.FormatStatusJSON(out, report)
			}
			return cloud.FormatStatusText(out, report)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON sur stdout")
	cmd.SilenceUsage = true
	return cmd
}

func newLoginCmd() *cobra.Command {
	var (
		token    string
		baseURL  string
		jsonOut  bool
	)
	cmd := &cobra.Command{
		Use:     "login",
		Short:   "Enregistrer un token API cloud (hors Git)",
		Example: "  asa cloud login --token <token>\n  asa cloud login --token <token> --base-url https://asagiri-cloud.test",
		RunE: func(cmd *cobra.Command, args []string) error {
			token = strings.TrimSpace(token)
			if token == "" {
				return fmt.Errorf("cloud login: --token requis")
			}

			tokenPath, err := cloud.ExpandPath(config.DefaultCloudTokenRel)
			if err != nil {
				return err
			}
			if repoRoot, cfg, cfgErr := loadRepoConfig(); cfgErr == nil {
				if p, pErr := cloud.TokenPath(cfg); pErr == nil && strings.TrimSpace(p) != "" {
					tokenPath = p
				}
				if strings.TrimSpace(baseURL) != "" {
					if err := cloud.PatchRepoCloud(repoRoot, func(c *config.CloudConfig) {
						c.BaseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
					}); err != nil {
						return err
					}
				}
			}

			if err := cloud.SaveToken(tokenPath, token); err != nil {
				return err
			}

			if jsonOut {
				out := cmd.OutOrStdout()
				enc := json.NewEncoder(out)
				return enc.Encode(map[string]any{"ok": true, "token_path": tokenPath})
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Token cloud enregistré (%s)\n", tokenPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&token, "token", "", "Token API créé côté cloud (requis)")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "URL de base API (met à jour cloud.base_url du dépôt courant)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON sur stdout")
	cmd.SilenceUsage = true
	return cmd
}

func newLogoutCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:     "logout",
		Short:   "Supprimer le token cloud local",
		Example: "  asa cloud logout\n  asa cloud logout --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			tokenPath, err := cloud.ExpandPath(config.DefaultCloudTokenRel)
			if err != nil {
				return err
			}
			if _, cfg, cfgErr := loadRepoConfig(); cfgErr == nil {
				if p, pErr := cloud.TokenPath(cfg); pErr == nil && strings.TrimSpace(p) != "" {
					tokenPath = p
				}
			}
			if err := cloud.RemoveToken(tokenPath); err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				return enc.Encode(map[string]any{"ok": true, "token_path": tokenPath})
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Token cloud supprimé (%s)\n", tokenPath)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON sur stdout")
	cmd.SilenceUsage = true
	return cmd
}

func newLinkCmd() *cobra.Command {
	var (
		slug    string
		enable  bool
		jsonOut bool
	)
	cmd := &cobra.Command{
		Use:     "link [project-id]",
		Short:   "Lier ce dépôt à un projet cloud",
		Example: "  asa cloud link <uuid>\n  asa cloud link --slug my-project --enable",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, cfg, err := loadRepoConfig()
			if err != nil {
				return err
			}
			token, _, err := resolveToken(cfg)
			if err != nil {
				return err
			}
			if strings.TrimSpace(token) == "" {
				return fmt.Errorf("cloud link: token absent — exécuter `asa cloud login --token <token>`")
			}

			projectID := ""
			if len(args) == 1 {
				projectID = strings.TrimSpace(args[0])
			}
			client := cloud.NewClient(cloud.ClientOptions{
				BaseURL: cfg.Cloud.BaseURL,
				Token:   token,
			})
			report, err := cloud.LinkProject(cmd.Context(), client, cloud.LinkOptions{
				RepoRoot:  repoRoot,
				Config:    cfg,
				Token:     token,
				ProjectID: projectID,
				Slug:      slug,
				Enable:    enable,
			})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return cloud.FormatLinkJSON(out, report)
			}
			return cloud.FormatLinkText(out, report)
		},
	}
	cmd.Flags().StringVar(&slug, "slug", "", "Slug du projet cloud (alternative à project-id)")
	cmd.Flags().BoolVar(&enable, "enable", false, "Activer cloud.enabled après liaison")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON sur stdout")
	cmd.SilenceUsage = true
	return cmd
}

func newPushCmd() *cobra.Command {
	var (
		runID   string
		all     bool
		dryRun  bool
		jsonOut bool
	)
	cmd := &cobra.Command{
		Use:     "push",
		Short:   "Pousser des runs ledger vers le cloud",
		Example: "  asa cloud push --dry-run --all\n  asa cloud push --run run-abc\n  asa cloud push --all",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, cfg, err := loadRepoConfig()
			if err != nil {
				return err
			}
			token, _, err := resolveToken(cfg)
			if err != nil {
				return err
			}

			report, err := cloud.Push(context.Background(), cloud.PushOptions{
				RepoRoot: repoRoot,
				Config:   cfg,
				Token:    token,
				RunID:    runID,
				All:      all,
				DryRun:     dryRun,
			})
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()
			if jsonOut {
				if err := cloud.FormatPushJSON(out, report); err != nil {
					return err
				}
				if h := report.Hint; h != "" && dryRun {
					_, _ = fmt.Fprintln(errOut, "→", h)
				}
				return pushExitFromReport(report)
			}
			if err := cloud.FormatPushText(out, report); err != nil {
				return err
			}
			return pushExitFromReport(report)
		},
	}
	cmd.Flags().StringVar(&runID, "run", "", "Run ledger local à pousser")
	cmd.Flags().BoolVar(&all, "all", false, "Pousser tous les runs du ledger")
	cmd.Flags().BoolVar(&dryRun, "dry-run", true, "Simuler sans appeler l'API (défaut)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON sur stdout")
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return cmd
}

func pushExitFromReport(report cloud.PushReport) error {
	if report.Mode == "dry-run" {
		return nil
	}
	for _, item := range report.Items {
		if item.Error != "" {
			return fmt.Errorf("cloud push: échec sur %s", item.LocalRunID)
		}
	}
	return nil
}
