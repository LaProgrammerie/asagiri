package cloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// StatusOptions configures a status check.
type StatusOptions struct {
	Config    *config.Config
	TokenPath string
	Token     string
	CheckAPI  bool
}

// BuildStatus returns cloud connectivity and link state.
func BuildStatus(ctx context.Context, opts StatusOptions) StatusReport {
	report := StatusReport{
		ReportVersion: ReportVersionStatus,
	}
	cfg := opts.Config
	if cfg != nil {
		report.Enabled = cfg.Cloud.Enabled
		report.BaseURL = strings.TrimSpace(cfg.Cloud.BaseURL)
		if report.BaseURL == "" {
			report.BaseURL = config.DefaultCloudBaseURL
		}
		report.TokenPath = strings.TrimSpace(cfg.Cloud.TokenPath)
		if report.TokenPath == "" {
			report.TokenPath = config.DefaultCloudTokenRel
		}
		report.ProjectID = strings.TrimSpace(cfg.Cloud.ProjectID)
		report.Linked = report.ProjectID != ""
	} else {
		report.BaseURL = config.DefaultCloudBaseURL
		report.TokenPath = config.DefaultCloudTokenRel
	}

	tokenPath := strings.TrimSpace(opts.TokenPath)
	if tokenPath == "" && cfg != nil {
		var err error
		tokenPath, err = TokenPath(cfg)
		if err != nil {
			report.Error = RedactError(err)
			return report
		}
	}
	if expanded, err := ExpandPath(report.TokenPath); err == nil && expanded != "" {
		report.TokenPath = expanded
	}

	token := strings.TrimSpace(opts.Token)
	if token == "" && tokenPath != "" {
		expanded, err := ExpandPath(tokenPath)
		if err != nil {
			report.Error = RedactError(err)
			return report
		}
		token, err = LoadToken(expanded)
		if err != nil {
			report.Error = RedactError(err)
			return report
		}
	}
	report.TokenPresent = token != ""

	if !opts.CheckAPI || token == "" {
		return report
	}

	client := NewClient(ClientOptions{BaseURL: report.BaseURL, Token: token})
	me, err := client.Me(ctx)
	if err != nil {
		report.Error = RedactError(err)
		return report
	}
	report.Reachable = true
	report.Me = &me
	return report
}

// RequirePushReady validates config + token + project for push.
func RequirePushReady(cfg *config.Config, token string) error {
	if cfg == nil {
		return fmt.Errorf("cloud: configuration absente")
	}
	if !cfg.Cloud.Enabled {
		return fmt.Errorf("cloud: désactivé — définir cloud.enabled: true dans .asagiri/config.yaml")
	}
	base := strings.TrimSpace(cfg.Cloud.BaseURL)
	if base == "" {
		return fmt.Errorf("cloud: cloud.base_url requis")
	}
	if strings.TrimSpace(cfg.Cloud.ProjectID) == "" {
		return fmt.Errorf("cloud: projet non lié — exécuter `asa cloud link <project-id>` ou définir cloud.project_id")
	}
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("cloud: token absent — exécuter `asa cloud login --token <token>`")
	}
	return nil
}
