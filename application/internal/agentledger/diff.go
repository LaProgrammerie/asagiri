package agentledger

import (
	"fmt"
	"strconv"
	"strings"
)

const DiffReportVersion = "agent-run-diff-v1"

// FieldDiff compares one metadata field between two runs.
type FieldDiff struct {
	Field string `json:"field"`
	Left  string `json:"left"`
	Right string `json:"right"`
	Equal bool   `json:"equal"`
}

// ArtifactSide summarizes one artifact on one side of the diff.
type ArtifactSide struct {
	Path       string `json:"path"`
	Exists     bool   `json:"exists"`
	SizeBytes  *int64 `json:"size_bytes,omitempty"`
	ModifiedAt string `json:"modified_at,omitempty"`
}

// ArtifactDiff compares one named artifact between two runs.
type ArtifactDiff struct {
	Name            string       `json:"name"`
	Left            ArtifactSide `json:"left"`
	Right           ArtifactSide `json:"right"`
	ExistsEqual     bool         `json:"exists_equal"`
	SizeEqual       bool         `json:"size_equal"`
	ModifiedAtEqual bool         `json:"modified_at_equal"`
	Equal           bool         `json:"equal"`
}

// DiffReport is the read model for `asa agents diff`.
type DiffReport struct {
	ReportVersion string         `json:"report_version"`
	LeftRunID     string         `json:"left_run_id"`
	RightRunID    string         `json:"right_run_id"`
	Identical     bool           `json:"identical"`
	Fields        []FieldDiff    `json:"fields"`
	Artifacts     []ArtifactDiff `json:"artifacts"`
}

// Diff compares two ledger runs and their on-disk artifacts (read-only).
func Diff(repoRoot, leftRunID, rightRunID string) (DiffReport, error) {
	repoRoot = strings.TrimSpace(repoRoot)
	leftRunID = strings.TrimSpace(leftRunID)
	rightRunID = strings.TrimSpace(rightRunID)
	if repoRoot == "" {
		return DiffReport{}, fmt.Errorf("agentledger: repo_root requis")
	}
	if leftRunID == "" || rightRunID == "" {
		return DiffReport{}, fmt.Errorf("agentledger: left_run_id et right_run_id requis")
	}
	left, err := Inspect(repoRoot, leftRunID)
	if err != nil {
		return DiffReport{}, err
	}
	right, err := Inspect(repoRoot, rightRunID)
	if err != nil {
		return DiffReport{}, err
	}
	return buildDiffReport(left, right), nil
}

func buildDiffReport(left, right InspectReport) DiffReport {
	fields := []FieldDiff{
		diffField("task_id", left.TaskID, right.TaskID),
		diffField("feature", left.Feature, right.Feature),
		diffField("agent_id", left.AgentID, right.AgentID),
		diffField("role", left.Role, right.Role),
		diffField("provider", left.Provider, right.Provider),
		diffField("phase", left.Phase, right.Phase),
		diffField("started_at", left.StartedAt, right.StartedAt),
		diffField("ended_at", left.EndedAt, right.EndedAt),
		diffField("duration_ms", strconv.FormatInt(left.DurationMS, 10), strconv.FormatInt(right.DurationMS, 10)),
		diffField("exit_code", strconv.Itoa(left.ExitCode), strconv.Itoa(right.ExitCode)),
		diffField("prompt_hash", left.PromptHash, right.PromptHash),
		diffField("context_hash", left.ContextHash, right.ContextHash),
		diffField("output_hash", left.OutputHash, right.OutputHash),
		diffContractField(left.ContractValid, right.ContractValid),
		diffField("log_dir", left.LogDir, right.LogDir),
		diffBoolField("dry_run", left.DryRun, right.DryRun),
	}
	artifacts := diffArtifacts(left.Artifacts, right.Artifacts)
	identical := true
	for _, f := range fields {
		if !f.Equal {
			identical = false
			break
		}
	}
	if identical {
		for _, a := range artifacts {
			if !a.Equal {
				identical = false
				break
			}
		}
	}
	return DiffReport{
		ReportVersion: DiffReportVersion,
		LeftRunID:     left.RunID,
		RightRunID:    right.RunID,
		Identical:     identical,
		Fields:        fields,
		Artifacts:     artifacts,
	}
}

func diffField(name, left, right string) FieldDiff {
	return FieldDiff{
		Field: name,
		Left:  left,
		Right: right,
		Equal: left == right,
	}
}

func diffBoolField(name string, left, right bool) FieldDiff {
	return FieldDiff{
		Field: name,
		Left:  strconv.FormatBool(left),
		Right: strconv.FormatBool(right),
		Equal: left == right,
	}
}

func diffContractField(left, right *bool) FieldDiff {
	return FieldDiff{
		Field: "contract_valid",
		Left:  formatContractValid(left),
		Right: formatContractValid(right),
		Equal: contractValidEqual(left, right),
	}
}

func formatContractValid(v *bool) string {
	if v == nil {
		return ""
	}
	return strconv.FormatBool(*v)
}

func contractValidEqual(left, right *bool) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	return *left == *right
}

func diffArtifacts(left, right []Artifact) []ArtifactDiff {
	leftByName := mapArtifacts(left)
	rightByName := mapArtifacts(right)
	out := make([]ArtifactDiff, 0, len(inspectArtifactNames))
	for _, name := range inspectArtifactNames {
		out = append(out, diffArtifact(name, leftByName[name], rightByName[name]))
	}
	return out
}

func mapArtifacts(artifacts []Artifact) map[string]Artifact {
	m := make(map[string]Artifact, len(artifacts))
	for _, a := range artifacts {
		m[a.Name] = a
	}
	return m
}

func diffArtifact(name string, left, right Artifact) ArtifactDiff {
	existsEqual := left.Exists == right.Exists
	sizeEqual := int64PtrEqual(left.SizeBytes, right.SizeBytes)
	modEqual := left.ModifiedAt == right.ModifiedAt
	equal := existsEqual && sizeEqual && modEqual
	return ArtifactDiff{
		Name: name,
		Left: ArtifactSide{
			Path:       left.Path,
			Exists:     left.Exists,
			SizeBytes:  left.SizeBytes,
			ModifiedAt: left.ModifiedAt,
		},
		Right: ArtifactSide{
			Path:       right.Path,
			Exists:     right.Exists,
			SizeBytes:  right.SizeBytes,
			ModifiedAt: right.ModifiedAt,
		},
		ExistsEqual:     existsEqual,
		SizeEqual:       sizeEqual,
		ModifiedAtEqual: modEqual,
		Equal:           equal,
	}
}

func int64PtrEqual(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
