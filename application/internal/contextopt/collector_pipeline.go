package contextopt

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// CollectForPipeline gathers context files. When investigation already produced
// candidate paths, it avoids a full-repo tree walk and reads only those files
// plus feature spec paths (consolidation: no double scan).
func CollectForPipeline(repoRoot, feature string, cfg *config.Config, opts CollectOpts, candidateFiles []string) ([]FileEntry, error) {
	if len(candidateFiles) == 0 {
		return Collect(repoRoot, feature, cfg, opts)
	}
	if cfg == nil {
		return nil, fmt.Errorf("contextopt: config nil")
	}
	if opts.MaxFiles <= 0 {
		opts.MaxFiles = 500
	}
	if opts.MaxBytes <= 0 {
		opts.MaxBytes = 2 << 20
	}
	seen := map[string]struct{}{}
	var paths []string
	if feature != "" {
		paths = append(paths,
			filepath.Join(".kiro", "specs", feature),
			filepath.Join(".asagiri", "specs", feature),
		)
	}
	paths = append(paths, cfg.Sources.Local.Paths...)
	for _, c := range candidateFiles {
		c = strings.TrimSpace(c)
		if c != "" {
			paths = append(paths, c)
		}
	}

	var out []FileEntry
	var total int64
	for _, rel := range paths {
		abs := filepath.Join(repoRoot, filepath.Clean(rel))
		entries, err := readTreeFiles(abs, rel, seen, repoRoot, opts.MaxFiles-len(out), opts.MaxBytes-total)
		if err != nil {
			continue
		}
		for _, e := range entries {
			out = append(out, e)
			total += e.Size
			if len(out) >= opts.MaxFiles || total >= opts.MaxBytes {
				return out, nil
			}
		}
	}
	return out, nil
}
