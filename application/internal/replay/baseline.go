package replay

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const baselineHashesRel = "reports/baseline-hashes.json"

type baselineIndex struct {
	ReplayID string            `json:"replay_id"`
	Hashes   map[string]string `json:"hashes"`
}

// WriteBaselineHashes records content hashes for replay artefacts (strict mode §6.2).
func WriteBaselineHashes(replayDir, replayID string) error {
	hashes, err := hashReplayTree(replayDir)
	if err != nil {
		return err
	}
	idx := baselineIndex{ReplayID: replayID, Hashes: hashes}
	body, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Join(replayDir, "reports")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "baseline-hashes.json"), append(body, '\n'), 0o644)
}

// CompareBaseline returns divergences when current artefact hashes differ from capture baseline.
func CompareBaseline(replayDir string) ([]Divergence, error) {
	path := filepath.Join(replayDir, baselineHashesRel)
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var idx baselineIndex
	if err := json.Unmarshal(body, &idx); err != nil {
		return nil, fmt.Errorf("parse baseline hashes: %w", err)
	}
	current, err := hashReplayTree(replayDir)
	if err != nil {
		return nil, err
	}
	var out []Divergence
	for rel, expected := range idx.Hashes {
		got, ok := current[rel]
		if !ok {
			out = append(out, Divergence{
				Kind:     DivergenceOutput,
				Severity: "error",
				Message:  fmt.Sprintf("baseline artefact removed: %s", rel),
				Path:     rel,
			})
			continue
		}
		if got != expected {
			out = append(out, Divergence{
				Kind:     DivergenceOutput,
				Severity: "error",
				Message:  fmt.Sprintf("artefact content changed since capture: %s", rel),
				Path:     rel,
			})
		}
	}
	return out, nil
}

func hashReplayTree(replayDir string) (map[string]string, error) {
	out := map[string]string{}
	err := filepath.WalkDir(replayDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(replayDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == ManifestName || strings.HasSuffix(rel, ".gz") ||
			rel == baselineHashesRel || rel == "reports/last-session.json" {
			return nil
		}
		if !isTextArtifact(path) {
			return nil
		}
		data, err := ReadMaybeCompressed(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		out[rel] = hex.EncodeToString(sum[:])
		return nil
	})
	return out, err
}
