package agentspec

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// EmbeddedTemplate is one shipped default AgentSpec template.
type EmbeddedTemplate struct {
	ID   string `json:"id"`
	Data []byte `json:"-"`
}

// ListEmbeddedTemplates returns embedded template bytes sorted by id.
func ListEmbeddedTemplates() ([]EmbeddedTemplate, error) {
	entries, err := fs.ReadDir(defaultTemplates, "templates")
	if err != nil {
		return nil, fmt.Errorf("agentspec: templates embarqués: %w", err)
	}
	out := make([]EmbeddedTemplate, 0, len(entries))
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".yaml") {
			continue
		}
		path := filepath.Join("templates", ent.Name())
		data, err := defaultTemplates.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("agentspec: lecture template %s: %w", ent.Name(), err)
		}
		spec, err := Parse(data, "embedded")
		if err != nil {
			return nil, fmt.Errorf("agentspec: template %s: %w", ent.Name(), err)
		}
		out = append(out, EmbeddedTemplate{ID: spec.ID, Data: append([]byte(nil), data...)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
