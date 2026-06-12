package agentledger

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const ExportReportVersion = "agent-run-export-v1"

// ExportOptions controls where and how a run bundle is written.
type ExportOptions struct {
	OutputDir     string
	IncludePrompt bool
}

// ExportFile describes one file written into the bundle.
type ExportFile struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

// ExportManifest is persisted as manifest.json inside the bundle.
type ExportManifest struct {
	ReportVersion string       `json:"report_version"`
	RunID         string       `json:"run_id"`
	ExportedAt    string       `json:"exported_at"`
	IncludePrompt bool         `json:"include_prompt"`
	Files         []ExportFile `json:"files"`
}

// ExportReport is returned to the CLI (stdout with --json).
type ExportReport struct {
	ReportVersion string       `json:"report_version"`
	RunID         string       `json:"run_id"`
	OutputDir     string       `json:"output_dir"`
	IncludePrompt bool         `json:"include_prompt"`
	ManifestPath  string       `json:"manifest_path"`
	Files         []ExportFile `json:"files"`
}

// Export writes a read-only bundle for one agent run under outputDir.
func Export(repoRoot, runID string, opts ExportOptions) (ExportReport, error) {
	repoRoot = strings.TrimSpace(repoRoot)
	runID = strings.TrimSpace(runID)
	if repoRoot == "" {
		return ExportReport{}, fmt.Errorf("agentledger: repo_root requis")
	}
	if runID == "" {
		return ExportReport{}, fmt.Errorf("agentledger: run_id requis")
	}
	entry, ok, err := findByRunID(repoRoot, runID)
	if err != nil {
		return ExportReport{}, err
	}
	if !ok {
		return ExportReport{}, fmt.Errorf("agentledger: run %q introuvable", runID)
	}

	outputDir := strings.TrimSpace(opts.OutputDir)
	if outputDir == "" {
		outputDir = filepath.Join(repoRoot, ".asagiri", "exports", "agents", runID)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return ExportReport{}, fmt.Errorf("agentledger: mkdir export: %w", err)
	}

	inspectReport := buildInspectReport(repoRoot, entry)
	previewReport := buildReplayPreviewReport(repoRoot, entry, ReplayPreviewOptions{
		IncludePrompt: opts.IncludePrompt,
	})

	files := make([]ExportFile, 0, 8)
	written, err := writeExportJSON(outputDir, "ledger-entry.json", entry)
	if err != nil {
		return ExportReport{}, err
	}
	files = append(files, written)

	written, err = writeExportJSON(outputDir, "inspect.json", inspectReport)
	if err != nil {
		return ExportReport{}, err
	}
	files = append(files, written)

	written, err = writeExportJSON(outputDir, "replay-preview.json", previewReport)
	if err != nil {
		return ExportReport{}, err
	}
	files = append(files, written)

	artifactDir := filepath.Join(outputDir, "artifacts")
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		return ExportReport{}, fmt.Errorf("agentledger: mkdir artifacts: %w", err)
	}
	for _, name := range inspectArtifactNames {
		rel := artifactRelPath(entry.LogDir, name)
		src := filepath.Join(repoRoot, filepath.FromSlash(rel))
		if _, err := os.Stat(src); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return ExportReport{}, fmt.Errorf("agentledger: stat %s: %w", rel, err)
		}
		dst := filepath.Join(artifactDir, name)
		if err := copyFile(src, dst); err != nil {
			return ExportReport{}, err
		}
		written, err = fileExportEntry(filepath.Join("artifacts", name), dst)
		if err != nil {
			return ExportReport{}, err
		}
		files = append(files, written)
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	manifest := ExportManifest{
		ReportVersion: ExportReportVersion,
		RunID:         runID,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339Nano),
		IncludePrompt: opts.IncludePrompt,
		Files:         files,
	}
	manifestPath := filepath.Join(outputDir, "manifest.json")
	if err := writeJSONFile(manifestPath, manifest); err != nil {
		return ExportReport{}, err
	}
	manifestEntry, err := fileExportEntry("manifest.json", manifestPath)
	if err != nil {
		return ExportReport{}, err
	}
	files = append(files, manifestEntry)
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })

	return ExportReport{
		ReportVersion: ExportReportVersion,
		RunID:         runID,
		OutputDir:     filepath.ToSlash(outputDir),
		IncludePrompt: opts.IncludePrompt,
		ManifestPath:  "manifest.json",
		Files:         files,
	}, nil
}

func writeExportJSON(dir, name string, v any) (ExportFile, error) {
	path := filepath.Join(dir, name)
	if err := writeJSONFile(path, v); err != nil {
		return ExportFile{}, err
	}
	return fileExportEntry(name, path)
}

func writeJSONFile(path string, v any) error {
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("agentledger: marshal %s: %w", filepath.Base(path), err)
	}
	body = append(body, '\n')
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("agentledger: write %s: %w", path, err)
	}
	return nil
}

func fileExportEntry(relPath, absPath string) (ExportFile, error) {
	info, err := os.Stat(absPath)
	if err != nil {
		return ExportFile{}, fmt.Errorf("agentledger: stat %s: %w", absPath, err)
	}
	hash, err := hashFile(absPath)
	if err != nil {
		return ExportFile{}, err
	}
	return ExportFile{
		Path:      filepath.ToSlash(relPath),
		SizeBytes: info.Size(),
		SHA256:    hash,
	}, nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("agentledger: open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("agentledger: hash %s: %w", path, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("agentledger: open %s: %w", src, err)
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("agentledger: create %s: %w", dst, err)
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("agentledger: copy %s: %w", dst, err)
	}
	return nil
}
