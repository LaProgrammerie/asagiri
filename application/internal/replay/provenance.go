package replay

import "time"

// ProvenanceRecord traces an artefact back to its origin (spec §16).
type ProvenanceRecord struct {
	ID           string    `json:"id" yaml:"id"`
	ProducedBy   string    `json:"produced_by" yaml:"produced_by"`
	GraphNode    string    `json:"graph_node,omitempty" yaml:"graph_node,omitempty"`
	SourceCommit string    `json:"source_commit,omitempty" yaml:"source_commit,omitempty"`
	ReplayID     string    `json:"replay,omitempty" yaml:"replay,omitempty"`
	SourcePath   string    `json:"source_path,omitempty" yaml:"source_path,omitempty"`
	CapturedAt   time.Time `json:"captured_at" yaml:"captured_at"`
}

// ProvenanceIndex lists provenance for all captured artefacts in a replay package.
type ProvenanceIndex struct {
	ReplayID string             `json:"replay_id" yaml:"replay_id"`
	Records  []ProvenanceRecord `json:"records" yaml:"records"`
}

// NewProvenanceRecord builds a provenance entry for a captured file.
func NewProvenanceRecord(replayID, artifactID, producedBy, sourcePath, commit, graphNode string) ProvenanceRecord {
	return ProvenanceRecord{
		ID:           artifactID,
		ProducedBy:   producedBy,
		GraphNode:    graphNode,
		SourceCommit: commit,
		ReplayID:     replayID,
		SourcePath:   sourcePath,
		CapturedAt:   time.Now().UTC(),
	}
}
