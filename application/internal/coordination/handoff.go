package coordination

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultHandoffsRel = ".asagiri/handoffs"
)

// Handoff is a structured agent-to-agent transfer (spec-my-D §8).
type Handoff struct {
	ID          string    `yaml:"id,omitempty" json:"id,omitempty"`
	From        AgentRole `yaml:"from" json:"from"`
	To          AgentRole `yaml:"to" json:"to"`
	Summary     string    `yaml:"summary" json:"summary"`
	Files       []string  `yaml:"files,omitempty" json:"files,omitempty"`
	Constraints []string  `yaml:"constraints,omitempty" json:"constraints,omitempty"`
	Confidence  float64   `yaml:"confidence,omitempty" json:"confidence,omitempty"`
	CreatedAt   string    `yaml:"created_at,omitempty" json:"created_at,omitempty"`
}

// AgentResult feeds handoff construction after a node completes.
type AgentResult struct {
	NodeID     string
	Role       AgentRole
	AgentRef   string
	Summary    string
	Files      []string
	Constraints []string
	Confidence float64
	TargetRole AgentRole
}

// HandoffBuilder persists structured handoffs.
type HandoffBuilder interface {
	Build(ctx context.Context, result AgentResult) (Handoff, error)
}

// DefaultHandoffBuilder writes YAML under repoRoot/.asagiri/handoffs/<id>/handoff.yaml.
type DefaultHandoffBuilder struct {
	RepoRoot     string
	HandoffsPath string
}

// Build validates the result and persists a handoff artefact.
func (b *DefaultHandoffBuilder) Build(_ context.Context, result AgentResult) (Handoff, error) {
	if b == nil || strings.TrimSpace(b.RepoRoot) == "" {
		return Handoff{}, fmt.Errorf("%w: repo root required", ErrHandoffPersist)
	}
	if err := ValidateRole(result.Role); err != nil {
		return Handoff{}, err
	}
	if result.TargetRole == "" {
		return Handoff{}, fmt.Errorf("%w: target role required", ErrInvalidHandoff)
	}
	if err := ValidateRole(result.TargetRole); err != nil {
		return Handoff{}, err
	}
	if strings.TrimSpace(result.Summary) == "" {
		return Handoff{}, fmt.Errorf("%w: summary required", ErrInvalidHandoff)
	}

	id := handoffID(result.NodeID)
	h := Handoff{
		ID:          id,
		From:        result.Role,
		To:          result.TargetRole,
		Summary:     strings.TrimSpace(result.Summary),
		Files:       append([]string(nil), result.Files...),
		Constraints: append([]string(nil), result.Constraints...),
		Confidence:  result.Confidence,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	base, err := b.handoffsBase()
	if err != nil {
		return Handoff{}, err
	}
	dir := filepath.Join(base, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Handoff{}, fmt.Errorf("%w: %v", ErrHandoffPersist, err)
	}
	data, err := yaml.Marshal(h)
	if err != nil {
		return Handoff{}, fmt.Errorf("%w: %v", ErrHandoffPersist, err)
	}
	path := filepath.Join(dir, "handoff.yaml")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return Handoff{}, fmt.Errorf("%w: %v", ErrHandoffPersist, err)
	}
	return h, nil
}

func (b *DefaultHandoffBuilder) handoffsBase() (string, error) {
	rel := strings.TrimSpace(b.HandoffsPath)
	if rel == "" {
		rel = DefaultHandoffsRel
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("%w: handoffs path must be relative to repo", ErrHandoffPersist)
	}
	clean := filepath.Clean(rel)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: handoffs path must not escape repo", ErrHandoffPersist)
	}
	base := filepath.Join(b.RepoRoot, clean)
	relToRepo, err := filepath.Rel(b.RepoRoot, base)
	if err != nil || strings.HasPrefix(relToRepo, "..") {
		return "", fmt.Errorf("%w: handoffs path must not escape repo", ErrHandoffPersist)
	}
	return base, nil
}

func handoffID(nodeID string) string {
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, nodeID)
	if safe == "" {
		safe = "handoff"
	}
	return fmt.Sprintf("%s-%d", safe, time.Now().Unix())
}
