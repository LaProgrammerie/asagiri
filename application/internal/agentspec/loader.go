package agentspec

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed templates/*.yaml
var defaultTemplates embed.FS

// Loader reads AgentSpec documents from .asagiri/agents or embedded defaults.
type Loader struct {
	repoRoot  string
	agentsDir string
}

// NewLoader creates a loader for the given repository root.
func NewLoader(repoRoot string) *Loader {
	root := strings.TrimSpace(repoRoot)
	return &Loader{
		repoRoot:  root,
		agentsDir: filepath.Join(root, RegistryDir),
	}
}

// AgentsDir returns the absolute registry directory path.
func (l *Loader) AgentsDir() string {
	return l.agentsDir
}

// UsingEmbeddedDefaults reports whether the last LoadAll used embedded templates.
func (l *Loader) UsingEmbeddedDefaults() bool {
	entries, err := listYAMLFiles(l.agentsDir)
	return err != nil || len(entries) == 0
}

// List returns metadata for all available AgentSpec entries.
func (l *Loader) List() ([]Meta, error) {
	specs, err := l.LoadAll()
	if err != nil {
		return nil, err
	}
	out := make([]Meta, 0, len(specs))
	for _, spec := range specs {
		out = append(out, Meta{
			ID:          spec.ID,
			Version:     spec.Version,
			Role:        spec.Role,
			ContentHash: spec.ContentHash,
			Source:      spec.Source,
		})
	}
	return out, nil
}

// LoadDiskOnly reads .asagiri/agents/<id>.yaml without embedded template fallback.
func (l *Loader) LoadDiskOnly(id string) (Spec, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Spec{}, fmt.Errorf("agentspec: id requis")
	}
	for _, ext := range []string{".yaml", ".yml"} {
		path := filepath.Join(l.agentsDir, id+ext)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return Spec{}, fmt.Errorf("agentspec: lecture %s: %w", path, err)
		}
		return Parse(data, path)
	}
	return Spec{}, fmt.Errorf("agentspec: %q introuvable sous %s", id, l.agentsDir)
}

// Load returns one AgentSpec by id.
func (l *Loader) Load(id string) (Spec, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Spec{}, fmt.Errorf("agentspec: id requis")
	}
	specs, err := l.LoadAll()
	if err != nil {
		return Spec{}, err
	}
	for _, spec := range specs {
		if spec.ID == id {
			return spec, nil
		}
	}
	return Spec{}, fmt.Errorf("agentspec: %q introuvable", id)
}

// LoadAll loads every AgentSpec from disk or embedded defaults.
func (l *Loader) LoadAll() ([]Spec, error) {
	files, err := listYAMLFiles(l.agentsDir)
	useEmbedded := err != nil || len(files) == 0

	var specs []Spec
	if useEmbedded {
		specs, err = loadEmbedded()
		if err != nil {
			return nil, err
		}
	} else {
		specs, err = loadFromDir(l.agentsDir, files)
		if err != nil {
			return nil, err
		}
	}

	if err := ValidateDuplicateIDs(specs); err != nil {
		return nil, err
	}
	return specs, nil
}

// Parse parses YAML bytes into a Spec, validates, and computes ContentHash.
func Parse(data []byte, source string) (Spec, error) {
	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return Spec{}, fmt.Errorf("agentspec: parse YAML (%s): %w", source, err)
	}
	filename := ""
	if source != "embedded" && source != "" {
		filename = filepath.Base(source)
	}
	if err := Validate(spec, filename); err != nil {
		return Spec{}, fmt.Errorf("%s: %w", source, err)
	}
	spec.Source = source
	spec.ContentHash = SemanticHash(spec)
	return spec, nil
}

func loadFromDir(dir string, files []string) ([]Spec, error) {
	type rawEntry struct {
		path string
		data []byte
		spec Spec
	}
	entries := make([]rawEntry, 0, len(files))
	for _, name := range files {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("agentspec: lecture %s: %w", path, err)
		}
		var spec Spec
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("agentspec: parse YAML (%s): %w", path, err)
		}
		spec.Source = path
		entries = append(entries, rawEntry{path: path, data: data, spec: spec})
	}
	prelim := make([]Spec, len(entries))
	for i, ent := range entries {
		prelim[i] = ent.spec
	}
	if err := ValidateDuplicateIDs(prelim); err != nil {
		return nil, err
	}

	specs := make([]Spec, 0, len(entries))
	for _, ent := range entries {
		spec, err := Parse(ent.data, ent.path)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}
	return specs, nil
}

func loadEmbedded() ([]Spec, error) {
	entries, err := fs.ReadDir(defaultTemplates, "templates")
	if err != nil {
		return nil, fmt.Errorf("agentspec: templates embarqués: %w", err)
	}
	specs := make([]Spec, 0, len(entries))
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
		specs = append(specs, spec)
	}
	return specs, nil
}

func listYAMLFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		name := ent.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			files = append(files, name)
		}
	}
	return files, nil
}
