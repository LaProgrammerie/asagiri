package agentspec

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
)

// metadataKeysExcludedFromHash are non-deterministic or derived metadata fields.
var metadataKeysExcludedFromHash = map[string]struct{}{
	"content_hash":     {},
	"updated_at":       {},
	"synced_at":        {},
	"last_synced_at":   {},
	"last_synced_hash": {},
	"created_at":       {},
}

type semanticPayload struct {
	ID              string            `json:"id"`
	Version         string            `json:"version"`
	Role            string            `json:"role"`
	ProviderTargets []string          `json:"provider_targets,omitempty"`
	SystemPrompt    string            `json:"system_prompt"`
	Instructions    []string          `json:"instructions,omitempty"`
	Constraints     []string          `json:"constraints,omitempty"`
	OutputContract  OutputContract    `json:"output_contract"`
	External        *externalSemantic `json:"external,omitempty"`
	Metadata        map[string]any    `json:"metadata,omitempty"`
}

type externalSemantic struct {
	Kind string `json:"kind,omitempty"`
	Path string `json:"path,omitempty"`
}

// SemanticHash returns a stable SHA-256 hex digest of the semantic AgentSpec content.
// Timestamps and derived hash fields in metadata are ignored.
func SemanticHash(spec Spec) string {
	payload := semanticPayload{
		ID:              spec.ID,
		Version:         spec.Version,
		Role:            spec.Role,
		ProviderTargets: append([]string(nil), spec.ProviderTargets...),
		SystemPrompt:    spec.SystemPrompt,
		Instructions:    append([]string(nil), spec.Instructions...),
		Constraints:     append([]string(nil), spec.Constraints...),
		OutputContract:  spec.OutputContract,
		External:        semanticExternal(spec.External),
		Metadata:        semanticMetadata(spec.Metadata),
	}
	sort.Strings(payload.ProviderTargets)
	sort.Strings(payload.Instructions)
	sort.Strings(payload.Constraints)
	if len(payload.OutputContract.RequiredFields) > 0 {
		fields := append([]string(nil), payload.OutputContract.RequiredFields...)
		sort.Strings(fields)
		payload.OutputContract.RequiredFields = fields
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func semanticExternal(ext *ExternalSpec) *externalSemantic {
	if ext == nil {
		return nil
	}
	kind := strings.TrimSpace(ext.Kind)
	path := strings.TrimSpace(ext.Path)
	if kind == "" && path == "" {
		return nil
	}
	return &externalSemantic{Kind: kind, Path: path}
}

func semanticMetadata(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	keys := make([]string, 0, len(in))
	for k := range in {
		if _, skip := metadataKeysExcludedFromHash[k]; skip {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		out[k] = in[k]
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
