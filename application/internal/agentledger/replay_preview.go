package agentledger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const ReplayPreviewReportVersion = "agent-run-replay-preview-v1"

// ReplayPreviewOptions controls optional fields in the preview report.
type ReplayPreviewOptions struct {
	IncludePrompt bool
}

// PreviewArtifact is one log artifact with optional inlined content.
type PreviewArtifact struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Exists     bool   `json:"exists"`
	SizeBytes  *int64 `json:"size_bytes,omitempty"`
	ModifiedAt string `json:"modified_at,omitempty"`
	Content    string `json:"content,omitempty"`
}

// ReplayPreviewReport reconstructs exploitable inputs for a past run (read-only).
type ReplayPreviewReport struct {
	ReportVersion string            `json:"report_version"`
	RunID         string            `json:"run_id"`
	TaskID        string            `json:"task_id"`
	Feature       string            `json:"feature"`
	AgentID       string            `json:"agent_id"`
	Role          string            `json:"role,omitempty"`
	Provider      string            `json:"provider,omitempty"`
	Phase         string            `json:"phase,omitempty"`
	StartedAt     string            `json:"started_at"`
	EndedAt       string            `json:"ended_at"`
	DurationMS    int64             `json:"duration_ms"`
	ExitCode      int               `json:"exit_code"`
	ContractValid *bool             `json:"contract_valid,omitempty"`
	PromptHash    string            `json:"prompt_hash"`
	ContextHash   string            `json:"context_hash,omitempty"`
	OutputHash    string            `json:"output_hash"`
	LogDir        string            `json:"log_dir"`
	DryRun        bool              `json:"dry_run,omitempty"`
	Artifacts     []PreviewArtifact `json:"artifacts"`
}

// ReplayPreview loads a run and reads on-disk artifacts for offline replay preview.
func ReplayPreview(repoRoot, runID string, opts ReplayPreviewOptions) (ReplayPreviewReport, error) {
	repoRoot = strings.TrimSpace(repoRoot)
	runID = strings.TrimSpace(runID)
	if repoRoot == "" {
		return ReplayPreviewReport{}, fmt.Errorf("agentledger: repo_root requis")
	}
	if runID == "" {
		return ReplayPreviewReport{}, fmt.Errorf("agentledger: run_id requis")
	}
	entry, ok, err := findByRunID(repoRoot, runID)
	if err != nil {
		return ReplayPreviewReport{}, err
	}
	if !ok {
		return ReplayPreviewReport{}, fmt.Errorf("agentledger: run %q introuvable", runID)
	}
	return buildReplayPreviewReport(repoRoot, entry, opts), nil
}

func buildReplayPreviewReport(repoRoot string, e Entry, opts ReplayPreviewOptions) ReplayPreviewReport {
	return ReplayPreviewReport{
		ReportVersion: ReplayPreviewReportVersion,
		RunID:         e.RunID,
		TaskID:        e.TaskID,
		Feature:       e.Feature,
		AgentID:       e.AgentID,
		Role:          e.Role,
		Provider:      e.Provider,
		Phase:         e.Phase,
		StartedAt:     e.StartedAt,
		EndedAt:       e.EndedAt,
		DurationMS:    e.DurationMS,
		ExitCode:      e.ExitCode,
		ContractValid: e.ContractValid,
		PromptHash:    e.PromptHash,
		ContextHash:   e.ContextHash,
		OutputHash:    e.OutputHash,
		LogDir:        e.LogDir,
		DryRun:        e.DryRun,
		Artifacts:     resolvePreviewArtifacts(repoRoot, e.LogDir, opts.IncludePrompt),
	}
}

func resolvePreviewArtifacts(repoRoot, logDir string, includePrompt bool) []PreviewArtifact {
	artifacts := make([]PreviewArtifact, 0, len(inspectArtifactNames))
	for _, name := range inspectArtifactNames {
		rel := artifactRelPath(logDir, name)
		base := statArtifact(repoRoot, rel, name)
		preview := PreviewArtifact{
			Name:       base.Name,
			Path:       base.Path,
			Exists:     base.Exists,
			SizeBytes:  base.SizeBytes,
			ModifiedAt: base.ModifiedAt,
		}
		if base.Exists && shouldInlinePreviewContent(name, includePrompt) {
			preview.Content = readArtifactContent(repoRoot, rel)
		}
		artifacts = append(artifacts, preview)
	}
	return artifacts
}

func shouldInlinePreviewContent(name string, includePrompt bool) bool {
	switch name {
	case "prompt.md":
		return includePrompt
	case "invocation.json", "context.json", "contract.json":
		return true
	default:
		return false
	}
}

func readArtifactContent(repoRoot, relPath string) string {
	abs := filepath.Join(repoRoot, filepath.FromSlash(relPath))
	body, err := os.ReadFile(abs)
	if err != nil {
		return ""
	}
	return string(body)
}
