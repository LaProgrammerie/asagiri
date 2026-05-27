package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const relDir = ".asagiri/skills"

// LoadAll reads skill YAML files under `.asagiri/skills/**` (spec-my-A §24.14).
func LoadAll(repoRoot string) ([]Skill, error) {
	root := filepath.Join(repoRoot, relDir)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, nil
	}
	var out []Skill
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var s Skill
		if err := yaml.Unmarshal(raw, &s); err != nil {
			return fmt.Errorf("skills: parse %s: %w", path, err)
		}
		if s.ID == "" {
			s.ID = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		}
		if s.Name == "" {
			s.Name = s.ID
		}
		rel, _ := filepath.Rel(repoRoot, path)
		s.Path = filepath.ToSlash(rel)
		out = append(out, s)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// Match returns skills whose scope overlaps the given tags.
func Match(all []Skill, tags []string) []Skill {
	if len(tags) == 0 {
		return all
	}
	var out []Skill
tagLoop:
	for _, s := range all {
		for _, sc := range s.Scope {
			for _, t := range tags {
				if strings.EqualFold(sc, t) {
					out = append(out, s)
					continue tagLoop
				}
			}
		}
	}
	return out
}
