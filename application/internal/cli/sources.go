package cli

import (
	"fmt"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/source"
	srcnotion "github.com/LaProgrammerie/hyper-fast-builder/application/internal/source/notion"
)

type sourceRegistry struct {
	Local  *source.LocalSource
	Notion *srcnotion.Source
}

func newSourceRegistry(c *appContext, client *srcnotion.Client) *sourceRegistry {
	reg := &sourceRegistry{
		Local: &source.LocalSource{
			RepoRoot: c.RepoRoot,
			Config:   c.Config.Sources.Local,
		},
	}
	if c.Config.Sources.Notion.Enabled {
		token := c.Config.NotionToken()
		if client == nil && token != "" {
			client = &srcnotion.Client{Token: token}
		}
		if client != nil && client.Token != "" {
			reg.Notion = &srcnotion.Source{
				RepoRoot: c.RepoRoot,
				Config:   c.Config.Sources.Notion,
				Client:   client,
			}
		}
	}
	return reg
}

func (r *sourceRegistry) byName(name string) (source.Source, error) {
	switch name {
	case "local":
		if r.Local == nil {
			return nil, fmt.Errorf("source local indisponible")
		}
		return r.Local, nil
	case "notion":
		if r.Notion == nil {
			return nil, fmt.Errorf("source notion désactivée ou token manquant")
		}
		return r.Notion, nil
	default:
		return nil, fmt.Errorf("source inconnue: %s", name)
	}
}
