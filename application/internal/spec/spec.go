package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// Document is the normalized source used by planning.
type Document struct {
	Feature      string `json:"feature"`
	Requirements string `json:"requirements"`
	Tasks        string `json:"tasks"`
	Design       string `json:"design"`
	Active       string `json:"active"`
	Source       string `json:"source"`
}

type Reader struct {
	RepoRoot string
	Config   *config.Config
}

func NewReader(repoRoot string, cfg *config.Config) *Reader {
	return &Reader{RepoRoot: repoRoot, Config: cfg}
}

// ReadFeature reads .kiro/specs/<feature> files, with fallback to active spec.
func (r *Reader) ReadFeature(feature string) (*Document, error) {
	if r.Config == nil {
		return nil, fmt.Errorf("spec reader: config nil")
	}
	feature = strings.TrimSpace(feature)
	if feature == "" {
		return nil, fmt.Errorf("feature requise")
	}

	base := filepath.Join(r.RepoRoot, r.Config.Specs.KiroPath, feature)
	doc := &Document{Feature: feature, Source: "kiro"}

	doc.Requirements = readIfExists(filepath.Join(base, "requirements.md"))
	doc.Tasks = readIfExists(filepath.Join(base, "tasks.md"))
	doc.Design = readIfExists(filepath.Join(base, "design.md"))

	if doc.Requirements == "" && doc.Tasks == "" && doc.Design == "" {
		activePath := filepath.Join(r.RepoRoot, r.Config.Specs.ActiveSpecPath)
		active := readIfExists(activePath)
		if active == "" {
			return nil, fmt.Errorf("spec introuvable pour %q", feature)
		}
		doc.Active = active
		doc.Source = "active"
	}

	return doc, nil
}

func readIfExists(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func (d *Document) CombinedText() string {
	chunks := make([]string, 0, 4)
	for _, chunk := range []string{d.Requirements, d.Tasks, d.Design, d.Active} {
		if strings.TrimSpace(chunk) != "" {
			chunks = append(chunks, strings.TrimSpace(chunk))
		}
	}
	return strings.Join(chunks, "\n\n")
}
