package knowledge

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SourceCategory labels extractor source groups for incremental builds.
type SourceCategory string

const (
	SourceCategoryFlows     SourceCategory = "flows"
	SourceCategoryContracts SourceCategory = "contracts"
	SourceCategoryCode      SourceCategory = "code"
	SourceCategoryTests     SourceCategory = "tests"
	SourceCategoryInfra     SourceCategory = "infra"
	SourceCategoryADR       SourceCategory = "adr"
	SourceCategoryRuntime   SourceCategory = "runtime"
)

// SourceIndex captures mtimes of knowledge inputs (spec-my-E §21).
type SourceIndex struct {
	CategoryMTimes map[SourceCategory]int64
}

// ScanSourceIndex walks repository inputs and returns per-category max modification times.
func ScanSourceIndex(repoRoot string, products []string, req BuildRequest) (SourceIndex, error) {
	if repoRoot == "" {
		return SourceIndex{}, fmt.Errorf("scan knowledge sources: repo root required")
	}
	idx := SourceIndex{CategoryMTimes: make(map[SourceCategory]int64)}

	includeFlows := req.IncludeFlows
	includeContracts := req.IncludeContracts
	if !includeFlows && !includeContracts && !req.IncludeCode && !req.IncludeTests &&
		!req.IncludeInfra && !req.IncludeADR && !req.IncludeRuntime {
		includeFlows = true
		includeContracts = true
	}

	if includeFlows {
		for _, product := range products {
			flowDir := filepath.Join(repoRoot, ".asagiri", "products", product, "flows")
			max, err := maxMtimeUnder(flowDir, func(path string) bool {
				return strings.HasSuffix(path, ".flow.yaml")
			})
			if err != nil {
				return SourceIndex{}, err
			}
			idx.mergeMax(SourceCategoryFlows, max)
		}
	}
	if includeContracts {
		for _, product := range products {
			contractDir := filepath.Join(repoRoot, ".asagiri", "products", product, "contracts")
			max, err := maxMtimeUnder(contractDir, func(path string) bool {
				ext := strings.ToLower(filepath.Ext(path))
				return ext == ".yaml" || ext == ".yml" || ext == ".json"
			})
			if err != nil {
				return SourceIndex{}, err
			}
			idx.mergeMax(SourceCategoryContracts, max)
		}
	}
	if req.IncludeCode {
		max, err := maxGoMtime(repoRoot, false)
		if err != nil {
			return SourceIndex{}, err
		}
		idx.mergeMax(SourceCategoryCode, max)
	}
	if req.IncludeTests {
		max, err := maxGoMtime(repoRoot, true)
		if err != nil {
			return SourceIndex{}, err
		}
		idx.mergeMax(SourceCategoryTests, max)
	}
	if req.IncludeInfra {
		for _, root := range []string{
			filepath.Join(repoRoot, "infrastructure"),
			filepath.Join(repoRoot, "terraform"),
		} {
			max, err := maxMtimeUnder(root, func(path string) bool {
				ext := strings.ToLower(filepath.Ext(path))
				return ext == ".tf" || ext == ".yaml" || ext == ".yml"
			})
			if err != nil {
				return SourceIndex{}, err
			}
			idx.mergeMax(SourceCategoryInfra, max)
		}
	}
	if req.IncludeADR {
		max, err := maxMtimeUnder(filepath.Join(repoRoot, "docs", "decisions"), func(path string) bool {
			return strings.HasSuffix(strings.ToLower(path), ".md")
		})
		if err != nil {
			return SourceIndex{}, err
		}
		idx.mergeMax(SourceCategoryADR, max)
	}
	if req.IncludeRuntime {
		max, err := maxMtimeUnder(filepath.Join(repoRoot, ".asagiri", "runtime"), func(path string) bool {
			ext := strings.ToLower(filepath.Ext(path))
			return ext == ".yaml" || ext == ".yml" || ext == ".json"
		})
		if err != nil {
			return SourceIndex{}, err
		}
		idx.mergeMax(SourceCategoryRuntime, max)
	}
	return idx, nil
}

func (idx *SourceIndex) mergeMax(cat SourceCategory, unix int64) {
	if unix <= 0 {
		return
	}
	if prev := idx.CategoryMTimes[cat]; unix > prev {
		idx.CategoryMTimes[cat] = unix
	}
}

func maxMtimeUnder(root string, accept func(path string) bool) (int64, error) {
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	if !info.IsDir() {
		return 0, nil
	}
	var maxUnix int64
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !accept(path) {
			return nil
		}
		stat, err := d.Info()
		if err != nil {
			return err
		}
		unix := stat.ModTime().UTC().Unix()
		if unix > maxUnix {
			maxUnix = unix
		}
		return nil
	})
	return maxUnix, err
}

func maxGoMtime(repoRoot string, testFiles bool) (int64, error) {
	var maxUnix int64
	err := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case "vendor", "node_modules", ".git", ".asagiri":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		isTest := strings.HasSuffix(d.Name(), "_test.go")
		if testFiles != isTest {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if strings.Contains(rel, "/vendor/") || strings.Contains(rel, "/testdata/") || strings.HasPrefix(rel, "testdata/") {
			return nil
		}
		stat, err := d.Info()
		if err != nil {
			return err
		}
		unix := stat.ModTime().UTC().Unix()
		if unix > maxUnix {
			maxUnix = unix
		}
		return nil
	})
	return maxUnix, err
}

func buildSourceMTimesMap(idx SourceIndex) map[string]int64 {
	out := make(map[string]int64, len(idx.CategoryMTimes))
	for cat, unix := range idx.CategoryMTimes {
		out[string(cat)] = unix
	}
	return out
}

func storedSourceMTimes(meta map[string]any) map[string]int64 {
	out := make(map[string]int64)
	raw, ok := meta["source_mtimes"]
	if !ok {
		return out
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return out
	}
	for k, v := range m {
		switch t := v.(type) {
		case float64:
			out[k] = int64(t)
		case int64:
			out[k] = t
		case int:
			out[k] = int64(t)
		}
	}
	return out
}

func parseBuiltAt(meta map[string]any) (time.Time, bool) {
	raw, ok := meta["built_at"]
	if !ok {
		return time.Time{}, false
	}
	s, ok := raw.(string)
	if !ok || s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return time.Time{}, false
		}
	}
	return t.UTC(), true
}

func categoryUnchanged(stored map[string]int64, cat SourceCategory, current int64) bool {
	prev, ok := stored[string(cat)]
	if !ok || prev == 0 {
		return false
	}
	return prev == current && current > 0
}

// CountFilesChangedSince counts source files newer than builtAt across tracked categories.
func CountFilesChangedSince(repoRoot string, products []string, req BuildRequest, builtAt time.Time) (int, error) {
	if builtAt.IsZero() {
		return 0, nil
	}
	cutoff := builtAt.UTC()
	var changed int

	includeFlows := req.IncludeFlows
	includeContracts := req.IncludeContracts
	if !includeFlows && !includeContracts && !req.IncludeCode && !req.IncludeTests &&
		!req.IncludeInfra && !req.IncludeADR && !req.IncludeRuntime {
		includeFlows = true
		includeContracts = true
	}

	countUnder := func(root string, accept func(path string) bool) error {
		return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !accept(path) {
				return nil
			}
			stat, err := d.Info()
			if err != nil {
				return err
			}
			if stat.ModTime().UTC().After(cutoff) {
				changed++
			}
			return nil
		})
	}

	if includeFlows {
		for _, product := range products {
			root := filepath.Join(repoRoot, ".asagiri", "products", product, "flows")
			if err := countUnder(root, func(path string) bool {
				return strings.HasSuffix(path, ".flow.yaml")
			}); err != nil && !os.IsNotExist(err) {
				return 0, err
			}
		}
	}
	if includeContracts {
		for _, product := range products {
			root := filepath.Join(repoRoot, ".asagiri", "products", product, "contracts")
			if err := countUnder(root, func(path string) bool {
				ext := strings.ToLower(filepath.Ext(path))
				return ext == ".yaml" || ext == ".yml" || ext == ".json"
			}); err != nil && !os.IsNotExist(err) {
				return 0, err
			}
		}
	}
	if req.IncludeCode {
		if n, err := countGoChanged(repoRoot, false, cutoff); err != nil {
			return 0, err
		} else {
			changed += n
		}
	}
	if req.IncludeTests {
		if n, err := countGoChanged(repoRoot, true, cutoff); err != nil {
			return 0, err
		} else {
			changed += n
		}
	}
	if req.IncludeADR {
		root := filepath.Join(repoRoot, "docs", "decisions")
		if err := countUnder(root, func(path string) bool {
			return strings.HasSuffix(strings.ToLower(path), ".md")
		}); err != nil && !os.IsNotExist(err) {
			return 0, err
		}
	}
	return changed, nil
}

func countGoChanged(repoRoot string, testFiles bool, cutoff time.Time) (int, error) {
	var changed int
	err := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case "vendor", "node_modules", ".git", ".asagiri":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		isTest := strings.HasSuffix(d.Name(), "_test.go")
		if testFiles != isTest {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if strings.Contains(rel, "/vendor/") || strings.Contains(rel, "/testdata/") || strings.HasPrefix(rel, "testdata/") {
			return nil
		}
		stat, err := d.Info()
		if err != nil {
			return err
		}
		if stat.ModTime().UTC().After(cutoff) {
			changed++
		}
		return nil
	})
	return changed, err
}
