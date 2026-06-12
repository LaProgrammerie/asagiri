package agentledger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const InspectReportVersion = "agent-run-v1"

var inspectArtifactNames = []string{
	"prompt.md",
	"invocation.json",
	"context.json",
	"contract.json",
}

// Artifact describes one on-disk log artifact for a run.
type Artifact struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Exists     bool   `json:"exists"`
	SizeBytes  *int64 `json:"size_bytes,omitempty"`
	ModifiedAt string `json:"modified_at,omitempty"`
}

// InspectReport is the read model for `asa agents run <run_id>`.
type InspectReport struct {
	ReportVersion string     `json:"report_version"`
	RunID         string     `json:"run_id"`
	TaskID        string     `json:"task_id"`
	Feature       string     `json:"feature"`
	AgentID       string     `json:"agent_id"`
	Role          string     `json:"role,omitempty"`
	Provider      string     `json:"provider,omitempty"`
	Phase         string     `json:"phase,omitempty"`
	StartedAt     string     `json:"started_at"`
	EndedAt       string     `json:"ended_at"`
	DurationMS    int64      `json:"duration_ms"`
	ExitCode      int        `json:"exit_code"`
	ContractValid *bool      `json:"contract_valid,omitempty"`
	PromptHash    string     `json:"prompt_hash"`
	ContextHash   string     `json:"context_hash,omitempty"`
	OutputHash    string     `json:"output_hash"`
	LogDir        string     `json:"log_dir"`
	DryRun        bool       `json:"dry_run,omitempty"`
	Artifacts     []Artifact `json:"artifacts"`
}

// Inspect loads one ledger entry by run_id and resolves log artifacts (read-only).
func Inspect(repoRoot, runID string) (InspectReport, error) {
	repoRoot = strings.TrimSpace(repoRoot)
	runID = strings.TrimSpace(runID)
	if repoRoot == "" {
		return InspectReport{}, fmt.Errorf("agentledger: repo_root requis")
	}
	if runID == "" {
		return InspectReport{}, fmt.Errorf("agentledger: run_id requis")
	}
	entry, ok, err := findByRunID(repoRoot, runID)
	if err != nil {
		return InspectReport{}, err
	}
	if !ok {
		return InspectReport{}, fmt.Errorf("agentledger: run %q introuvable", runID)
	}
	return buildInspectReport(repoRoot, entry), nil
}

func findByRunID(repoRoot, runID string) (Entry, bool, error) {
	report, err := List(repoRoot, ListOptions{})
	if err != nil {
		return Entry{}, false, err
	}
	for _, e := range report.Entries {
		if e.RunID == runID {
			return e, true, nil
		}
	}
	return Entry{}, false, nil
}

func buildInspectReport(repoRoot string, e Entry) InspectReport {
	return InspectReport{
		ReportVersion: InspectReportVersion,
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
		Artifacts:     resolveArtifacts(repoRoot, e.LogDir),
	}
}

func resolveArtifacts(repoRoot, logDir string) []Artifact {
	logDir = strings.TrimSpace(logDir)
	artifacts := make([]Artifact, 0, len(inspectArtifactNames))
	for _, name := range inspectArtifactNames {
		rel := artifactRelPath(logDir, name)
		artifacts = append(artifacts, statArtifact(repoRoot, rel, name))
	}
	return artifacts
}

func artifactRelPath(logDir, name string) string {
	if logDir == "" {
		return filepath.ToSlash(name)
	}
	return filepath.ToSlash(filepath.Join(logDir, name))
}

func statArtifact(repoRoot, relPath, name string) Artifact {
	art := Artifact{
		Name:   name,
		Path:   relPath,
		Exists: false,
	}
	abs := filepath.Join(repoRoot, filepath.FromSlash(relPath))
	info, err := os.Stat(abs)
	if err != nil {
		if !os.IsNotExist(err) {
			return art
		}
		return art
	}
	if info.IsDir() {
		return art
	}
	size := info.Size()
	art.Exists = true
	art.SizeBytes = &size
	art.ModifiedAt = info.ModTime().UTC().Format(time.RFC3339Nano)
	return art
}
