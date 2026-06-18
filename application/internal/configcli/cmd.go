package configcli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/cloud"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/configcache"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime/resolve"
	"github.com/spf13/cobra"
)

// RootCommand returns the `asa config` command tree.
func RootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Cache et synchronisation du bundle cloud (registry fallback)",
	}
	cmd.AddCommand(
		newPullCmd(),
		newStatusCmd(),
		newRefreshCmd(),
		newActivateCmd(),
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
		return "", nil, fmt.Errorf("config: %w", err)
	}
	return repoRoot, cfg, nil
}

func cloudClient(cfg *config.Config) (*cloud.Client, string, error) {
	path, err := cloud.TokenPath(cfg)
	if err != nil {
		return nil, "", err
	}
	token, err := cloud.LoadToken(path)
	if err != nil {
		return nil, "", err
	}
	if strings.TrimSpace(token) == "" {
		return nil, "", fmt.Errorf("config: token cloud absent — exécuter `asa cloud login`")
	}
	projectID := strings.TrimSpace(cfg.Cloud.ProjectID)
	if projectID == "" {
		return nil, "", fmt.Errorf("config: cloud.project_id absent — exécuter `asa cloud link`")
	}
	return cloud.NewClient(cloud.ClientOptions{
		BaseURL: cfg.Cloud.BaseURL,
		Token:   token,
	}), projectID, nil
}

func projectSlug(cfg *config.Config) string {
	if s := strings.TrimSpace(cfg.Cloud.Project); s != "" {
		return s
	}
	return "default"
}

func cacheStore() (configcache.Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return configcache.Store{}, err
	}
	return configcache.NewStore(home), nil
}

func newPullCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Télécharge le bundle cloud publié dans le cache local",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, cfg, err := loadRepoConfig()
			if err != nil {
				return err
			}
			client, projectID, err := cloudClient(cfg)
			if err != nil {
				return err
			}
			store, err := cacheStore()
			if err != nil {
				return err
			}
			var pin *int
			if cfg.Cloud.RegistryPinRelease != nil && *cfg.Cloud.RegistryPinRelease > 0 {
				pin = cfg.Cloud.RegistryPinRelease
			}
			fetched, err := client.FetchConfigBundle(cmd.Context(), projectID, pin)
			if err != nil {
				if ptr, ok, _ := store.LoadActive(projectSlug(cfg)); ok {
					if jsonOut {
						return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
							"offline": true,
							"active":  ptr,
							"error":   err.Error(),
						})
					}
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "WARN: fetch impossible (%v), cache existant conservé\n", err)
					return nil
				}
				return err
			}
			ptr, err := store.SaveBundle(projectSlug(cfg), fetched)
			if err != nil {
				return err
			}
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"active":       ptr,
					"release":      fetched.ReleaseVersion,
					"content_hash": fetched.ContentHash,
				})
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Bundle release v%d mis en cache (%s)\n", fetched.ReleaseVersion, fetched.ContentHash)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON")
	return cmd
}

func newStatusCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Affiche l'état du cache config local",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, cfg, err := loadRepoConfig()
			if err != nil {
				return err
			}
			store, err := cacheStore()
			if err != nil {
				return err
			}
			slug := projectSlug(cfg)
			ptr, ok, err := store.LoadActive(slug)
			if err != nil {
				return err
			}
			report := map[string]any{
				"project_slug":      slug,
				"registry_fallback": cfg.Cloud.RegistryFallback,
				"cache_present":     ok,
			}
			if ok {
				report["active"] = ptr
				if bundle, _, meta, loaded, _ := store.CurrentBundle(slug); loaded {
					report["registry_schema"] = meta.RegistrySchemaVersion
					if prov, ok := bundle["provenance"].(map[string]any); ok {
						report["provenance"] = prov
					}
				}
			}
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(report)
			}
			if !ok {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Aucun bundle en cache.")
				return nil
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Cache: release v%d hash=%s fetched=%s\n",
				ptr.ReleaseVersion, ptr.ContentHash, ptr.FetchedAt.Format(time.RFC3339))
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "registry_fallback=%v\n", cfg.Cloud.RegistryFallback)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON")
	return cmd
}

func newRefreshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "refresh",
		Short: "Rafraîchit le cache config si le réseau est disponible",
		RunE: func(cmd *cobra.Command, args []string) error {
			pull := newPullCmd()
			pull.SetContext(cmd.Context())
			pull.SetOut(cmd.OutOrStdout())
			pull.SetErr(cmd.ErrOrStderr())
			return pull.RunE(cmd, nil)
		},
	}
}

func newActivateCmd() *cobra.Command {
	var (
		release int
		jsonOut bool
	)
	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Active une release cloud puis rafraîchit le cache local",
		RunE: func(cmd *cobra.Command, args []string) error {
			if release <= 0 {
				return fmt.Errorf("config activate: --release requis")
			}
			_, cfg, err := loadRepoConfig()
			if err != nil {
				return err
			}
			client, projectID, err := cloudClient(cfg)
			if err != nil {
				return err
			}
			out, err := client.ActivateConfigRelease(context.Background(), projectID, release)
			if err != nil {
				return err
			}
			pin := release
			cfg.Cloud.RegistryPinRelease = &pin
			fetched, err := client.FetchConfigBundle(cmd.Context(), projectID, &release)
			if err != nil {
				return err
			}
			store, err := cacheStore()
			if err != nil {
				return err
			}
			ptr, err := store.SaveBundle(projectSlug(cfg), fetched)
			if err != nil {
				return err
			}
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"activated": out,
					"active":    ptr,
				})
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Release v%d activée et mise en cache\n", release)
			_ = resolve.ResolveAgent(resolve.Options{Config: cfg, Cache: store, ProjectSlug: projectSlug(cfg)})
			return nil
		},
	}
	cmd.Flags().IntVar(&release, "release", 0, "Version de release à activer")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON")
	return cmd
}
