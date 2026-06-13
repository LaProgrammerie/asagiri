package cloud

import (
	"fmt"
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"gopkg.in/yaml.v3"
)

// PatchRepoCloud updates cloud fields in .asagiri/config.yaml.
func PatchRepoCloud(repoRoot string, patch func(*config.CloudConfig)) error {
	repoRoot = strings.TrimSpace(repoRoot)
	if repoRoot == "" {
		return fmt.Errorf("cloud: repo_root requis")
	}
	cfgPath := config.ConfigPath(repoRoot)
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("cloud: lecture config: %w", err)
	}
	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("cloud: parse config: %w", err)
	}

	cfg, err := config.Load(cfgPath, repoRoot)
	if err != nil {
		return fmt.Errorf("cloud: load config: %w", err)
	}
	patch(&cfg.Cloud)
	if cfg.Cloud.BaseURL == "" {
		cfg.Cloud.BaseURL = config.DefaultCloudBaseURL
	}
	if cfg.Cloud.TokenPath == "" {
		cfg.Cloud.TokenPath = config.DefaultCloudTokenRel
	}

	cloudNode := map[string]any{
		"enabled":    cfg.Cloud.Enabled,
		"base_url":   cfg.Cloud.BaseURL,
		"token_path": cfg.Cloud.TokenPath,
	}
	if id := strings.TrimSpace(cfg.Cloud.ProjectID); id != "" {
		cloudNode["project_id"] = id
	}
	root["cloud"] = cloudNode

	out, err := yaml.Marshal(root)
	if err != nil {
		return fmt.Errorf("cloud: marshal config: %w", err)
	}
	if err := os.WriteFile(cfgPath, out, 0o644); err != nil {
		return fmt.Errorf("cloud: écriture config: %w", err)
	}
	return nil
}
