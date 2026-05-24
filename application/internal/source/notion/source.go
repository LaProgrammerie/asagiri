package notion

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/source"
)

// Source implements source.Source for Notion (specv2 §8).
type Source struct {
	RepoRoot string
	Config   config.NotionSourceConfig
	Client   *Client
}

func (n *Source) Name() string { return "notion" }

func (n *Source) List(ctx context.Context) ([]source.SourceItem, error) {
	if n.Client == nil || n.Client.Token == "" {
		env := n.Config.TokenEnv
		if env == "" {
			env = "NOTION_TOKEN"
		}
		return nil, fmt.Errorf("notion: token manquant (variable %s)", env)
	}
	db := n.Config.SpecsDatabaseID
	if db == "" {
		db = n.Config.DefaultDatabaseID
	}
	if db == "" {
		return nil, nil
	}
	pages, err := n.Client.QueryDatabase(ctx, db)
	if err != nil {
		return nil, err
	}
	items := make([]source.SourceItem, 0, len(pages))
	for _, p := range pages {
		title := TitleFromPage(p, n.Config.TitleProperty)
		feature := slugFeature(title)
		updated, _ := time.Parse(time.RFC3339, p.LastEditedTime)
		items = append(items, source.SourceItem{
			Ref:       source.SourceRef{ID: p.ID, URL: p.URL, Name: feature},
			Feature:   feature,
			Status:    StatusFromPage(p, n.Config.StatusProperty),
			UpdatedAt: updated,
		})
	}
	return items, nil
}

func (n *Source) Fetch(ctx context.Context, ref source.SourceRef) (source.SourceDocument, error) {
	pageID := ref.ID
	if pageID == "" && ref.URL != "" {
		pageID = intent.ParseNotionPageID(ref.URL)
	}
	if pageID == "" {
		return source.SourceDocument{}, fmt.Errorf("notion: page id requis")
	}
	page, err := n.Client.GetPage(ctx, pageID)
	if err != nil {
		return source.SourceDocument{}, err
	}
	blocks, err := n.Client.ListBlockChildren(ctx, page.ID)
	if err != nil {
		return source.SourceDocument{}, err
	}
	md, tasksYAML := BlocksToMarkdown(blocks)
	title := TitleFromPage(page, n.Config.TitleProperty)
	feature := slugFeature(title)
	if strings.TrimSpace(md) == "" {
		return source.SourceDocument{}, fmt.Errorf("notion: spec vide")
	}
	status := StatusFromPage(page, n.Config.StatusProperty)
	updated, _ := time.Parse(time.RFC3339, page.LastEditedTime)
	return source.SourceDocument{
		Feature:         feature,
		Title:           title,
		Markdown:        md,
		TasksYAML:       tasksYAML,
		Status:          status,
		RemoteUpdatedAt: updated,
		Ref:             source.SourceRef{ID: page.ID, URL: page.URL, Name: feature},
	}, nil
}

func (n *Source) Sync(ctx context.Context, ref source.SourceRef, dest source.LocalSpecPath, opts source.SyncOptions) (source.SyncResult, error) {
	doc, err := n.Fetch(ctx, ref)
	if err != nil {
		return source.SyncResult{}, err
	}
	importPath := n.Config.ImportPath
	if dest.Root != "" {
		importPath = dest.Root
	}
	feature := dest.Feature
	if feature == "" {
		feature = doc.Feature
	}
	return source.WriteLocalSpec(n.RepoRoot, importPath, feature, doc, opts)
}

func slugFeature(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return strings.Trim(s, "-")
}
