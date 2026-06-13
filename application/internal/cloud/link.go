package cloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// LinkOptions configures project linking.
type LinkOptions struct {
	RepoRoot  string
	Config    *config.Config
	Token     string
	ProjectID string
	Slug      string
	Enable    bool
}

// LinkProject resolves a cloud project and persists cloud.project_id in repo config.
func LinkProject(ctx context.Context, client *Client, opts LinkOptions) (LinkReport, error) {
	repoRoot := strings.TrimSpace(opts.RepoRoot)
	if repoRoot == "" {
		return LinkReport{}, fmt.Errorf("cloud link: repo_root requis")
	}
	projectID := strings.TrimSpace(opts.ProjectID)
	slug := strings.TrimSpace(opts.Slug)
	if projectID == "" && slug == "" {
		return LinkReport{}, fmt.Errorf("cloud link: project-id ou --slug requis")
	}

	var project Project
	if projectID != "" {
		projects, err := client.ListProjects(ctx)
		if err != nil {
			return LinkReport{}, err
		}
		found := false
		for _, p := range projects {
			if p.ID == projectID {
				project = p
				found = true
				break
			}
		}
		if !found {
			return LinkReport{}, fmt.Errorf("cloud link: projet %q introuvable ou inaccessible", projectID)
		}
	} else {
		projects, err := client.ListProjects(ctx)
		if err != nil {
			return LinkReport{}, err
		}
		for _, p := range projects {
			if p.Slug == slug {
				project = p
				break
			}
		}
		if project.ID == "" {
			return LinkReport{}, fmt.Errorf("cloud link: aucun projet avec slug %q", slug)
		}
	}

	enable := opts.Enable
	if err := PatchRepoCloud(repoRoot, func(c *config.CloudConfig) {
		c.ProjectID = project.ID
		if enable {
			c.Enabled = true
		}
	}); err != nil {
		return LinkReport{}, err
	}

	return LinkReport{
		ReportVersion: ReportVersionLink,
		ProjectID:     project.ID,
		ProjectName:   project.Name,
		ProjectSlug:   project.Slug,
		ConfigPath:    config.ConfigPath(repoRoot),
	}, nil
}
