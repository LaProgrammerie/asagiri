package agentcontract

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agent"
	"github.com/LaProgrammerie/asagiri/application/internal/agentcontext"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"gopkg.in/yaml.v3"
)

const (
	ErrCodeMissingRequiredField = "missing_required_field"
	ErrCodeInvalidJSON          = "invalid_json"
	ErrCodeInvalidYAML          = "invalid_yaml"
	ErrCodeUnknownFormat        = "unknown_format"
	ErrCodeEmptyOutput          = "empty_output"
)

// ContractError is one validation issue.
type ContractError struct {
	Code    string `json:"code"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// ContractValidationResult is the stable JSON report for output contract validation.
type ContractValidationResult struct {
	Valid            bool            `json:"valid"`
	Format           string          `json:"format"`
	Errors           []ContractError `json:"errors"`
	Warnings         []string        `json:"warnings"`
	ExtractedSummary string          `json:"extracted_summary,omitempty"`
}

// ValidateOutput checks raw agent stdout against spec.output_contract.
func ValidateOutput(spec agentspec.Spec, rawOutput string) ContractValidationResult {
	format := strings.TrimSpace(spec.OutputContract.Format)
	result := ContractValidationResult{
		Format:   format,
		Errors:   []ContractError{},
		Warnings: []string{},
	}
	if format == "" {
		result.Errors = append(result.Errors, ContractError{
			Code:    ErrCodeUnknownFormat,
			Message: "output_contract.format vide",
		})
		return finalize(result)
	}

	raw := strings.TrimSpace(rawOutput)
	switch format {
	case agentspec.OutputText:
		return finalize(validateText(raw, result))
	case agentspec.OutputAsagiriV1:
		return finalize(validateAsagiriV1(raw, spec.OutputContract.RequiredFields, result))
	case agentspec.OutputGateYAML:
		return finalize(validateGateOutput(raw, spec.OutputContract.RequiredFields, result, true))
	case agentspec.OutputGateJSON:
		return finalize(validateGateOutput(raw, spec.OutputContract.RequiredFields, result, false))
	default:
		result.Errors = append(result.Errors, ContractError{
			Code:    ErrCodeUnknownFormat,
			Message: fmt.Sprintf("format %q non supporté", format),
		})
		return finalize(result)
	}
}

func validateText(raw string, result ContractValidationResult) ContractValidationResult {
	if raw == "" {
		result.Errors = append(result.Errors, ContractError{
			Code:    ErrCodeEmptyOutput,
			Message: "sortie texte vide",
		})
		return result
	}
	result.ExtractedSummary = truncateSummary(raw)
	return result
}

func validateAsagiriV1(raw string, required []string, result ContractValidationResult) ContractValidationResult {
	if raw == "" {
		result.Errors = append(result.Errors, ContractError{
			Code:    ErrCodeEmptyOutput,
			Message: "sortie asagiri-v1 vide",
		})
		return result
	}

	payload := extractJSONPayload(raw)
	if payload == "" {
		result.Errors = append(result.Errors, ContractError{
			Code:    ErrCodeInvalidJSON,
			Message: "aucun objet JSON détecté dans la sortie",
		})
		return result
	}

	var doc map[string]json.RawMessage
	if err := json.Unmarshal([]byte(payload), &doc); err != nil {
		result.Errors = append(result.Errors, ContractError{
			Code:    ErrCodeInvalidJSON,
			Message: err.Error(),
		})
		return result
	}

	fields := requiredFields(required, []string{"status"})
	for _, field := range fields {
		if !jsonFieldPresent(doc, field) {
			result.Errors = append(result.Errors, ContractError{
				Code:    ErrCodeMissingRequiredField,
				Field:   field,
				Message: fmt.Sprintf("champ requis %q absent", field),
			})
		}
	}

	if len(result.Errors) > 0 {
		return result
	}

	parsed, ok := agent.ParseResult(raw)
	if !ok {
		result.Warnings = append(result.Warnings, "JSON parsé mais agent.ParseResult a refusé la forme AgentResult")
	} else {
		result.ExtractedSummary = truncateSummary(parsed.Summary)
		if result.ExtractedSummary == "" {
			result.ExtractedSummary = truncateSummary(parsed.Status)
		}
	}
	return result
}

func validateGateOutput(raw string, required []string, result ContractValidationResult, expectYAML bool) ContractValidationResult {
	if raw == "" {
		result.Errors = append(result.Errors, ContractError{
			Code:    ErrCodeEmptyOutput,
			Message: "sortie gate vide",
		})
		return result
	}

	payload := gatesExtractPayload(raw)
	if payload == "" {
		result.Errors = append(result.Errors, ContractError{
			Code:    ErrCodeEmptyOutput,
			Message: "payload gate introuvable",
		})
		return result
	}

	if expectYAML {
		if err := probeYAML(payload); err != nil {
			result.Errors = append(result.Errors, ContractError{
				Code:    ErrCodeInvalidYAML,
				Message: err.Error(),
			})
			return result
		}
	} else if err := probeJSON(payload); err != nil {
		result.Errors = append(result.Errors, ContractError{
			Code:    ErrCodeInvalidJSON,
			Message: err.Error(),
		})
		return result
	}

	blockKey := inferGateBlockKey(required, payload)
	cfg := gates.ParseConfig{BlockKey: blockKey}
	parsed := gates.ParseResult(raw, cfg)
	if strings.TrimSpace(string(parsed.Status)) == "" {
		result.Errors = append(result.Errors, ContractError{
			Code:    ErrCodeMissingRequiredField,
			Field:   "status",
			Message: "status gate absent ou invalide",
		})
	}
	if parsed.ParseError != "" {
		code := ErrCodeInvalidYAML
		if !expectYAML {
			code = ErrCodeInvalidJSON
		}
		result.Errors = append(result.Errors, ContractError{
			Code:    code,
			Message: parsed.ParseError,
		})
	}

	for _, field := range requiredFields(required, []string{"status"}) {
		if field == "status" {
			continue
		}
		if !gateFieldPresent(payload, blockKey, field) {
			result.Errors = append(result.Errors, ContractError{
				Code:    ErrCodeMissingRequiredField,
				Field:   field,
				Message: fmt.Sprintf("champ requis %q absent du bloc gate", field),
			})
		}
	}

	if len(result.Errors) == 0 {
		result.ExtractedSummary = fmt.Sprintf("status=%s confidence=%.2f", parsed.Status, parsed.Confidence)
	}
	return result
}

func requiredFields(fromSpec, defaults []string) []string {
	if len(fromSpec) == 0 {
		return append([]string(nil), defaults...)
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(fromSpec))
	for _, f := range fromSpec {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if _, ok := seen[f]; ok {
			continue
		}
		seen[f] = struct{}{}
		out = append(out, f)
	}
	return out
}

func jsonFieldPresent(doc map[string]json.RawMessage, field string) bool {
	_, ok := doc[field]
	return ok
}

func extractJSONPayload(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if idx := strings.Index(raw, "{"); idx >= 0 {
		return strings.TrimSpace(raw[idx:])
	}
	return ""
}

func probeYAML(payload string) error {
	var doc any
	return yaml.Unmarshal([]byte(payload), &doc)
}

func probeJSON(payload string) error {
	var doc any
	return json.Unmarshal([]byte(payload), &doc)
}

func inferGateBlockKey(required []string, payload string) string {
	for _, f := range required {
		f = strings.TrimSpace(f)
		if strings.HasSuffix(f, "_gate") || f == "gate" {
			return f
		}
	}
	var doc map[string]any
	if err := yaml.Unmarshal([]byte(payload), &doc); err == nil {
		for k := range doc {
			if strings.HasSuffix(k, "_gate") || k == "gate" {
				return k
			}
		}
		if len(doc) == 1 {
			for k := range doc {
				return k
			}
		}
	}
	if err := json.Unmarshal([]byte(payload), &doc); err == nil {
		for k := range doc {
			if strings.HasSuffix(k, "_gate") || k == "gate" {
				return k
			}
		}
		if len(doc) == 1 {
			for k := range doc {
				return k
			}
		}
	}
	lower := strings.ToLower(payload)
	for _, token := range []string{"enrich_gate:", "governance:", "verify_evidence_gate:", "human_review:", "plan_gate:"} {
		if strings.Contains(lower, token) {
			return strings.TrimSuffix(strings.TrimSpace(token), ":")
		}
	}
	return "gate"
}

func gateFieldPresent(payload, blockKey, field string) bool {
	var wrapper map[string]map[string]any
	if err := yaml.Unmarshal([]byte(payload), &wrapper); err == nil {
		if block, ok := wrapper[blockKey]; ok {
			_, ok := block[field]
			return ok
		}
	}
	if err := json.Unmarshal([]byte(payload), &wrapper); err == nil {
		if block, ok := wrapper[blockKey]; ok {
			_, ok := block[field]
			return ok
		}
	}
	return false
}

func gatesExtractPayload(stdout string) string {
	s := strings.TrimSpace(stdout)
	if s == "" {
		return ""
	}
	if fenced := extractYAMLFence(s); fenced != "" {
		return fenced
	}
	return s
}

func extractYAMLFence(s string) string {
	lower := strings.ToLower(s)
	start := strings.Index(lower, "```yaml")
	if start < 0 {
		start = strings.Index(lower, "```yml")
	}
	if start < 0 {
		start = strings.Index(lower, "```json")
	}
	if start < 0 {
		return ""
	}
	rest := s[start:]
	if i := strings.Index(rest, "\n"); i >= 0 {
		rest = rest[i+1:]
	}
	end := strings.Index(rest, "```")
	if end < 0 {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(rest[:end])
}

func truncateSummary(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 160 {
		return s
	}
	return s[:157] + "..."
}

func finalize(r ContractValidationResult) ContractValidationResult {
	r.Valid = len(r.Errors) == 0
	if r.Errors == nil {
		r.Errors = []ContractError{}
	}
	if r.Warnings == nil {
		r.Warnings = []string{}
	}
	return r
}

// WriteLog persists contract.json under the orchestrated agent log directory.
func WriteLog(repoRoot, taskID, agentID string, result ContractValidationResult) error {
	if strings.TrimSpace(taskID) == "" || strings.TrimSpace(agentID) == "" {
		return fmt.Errorf("agentcontract: task_id et agent_id requis")
	}
	dir := agentcontext.AgentLogDir(repoRoot, taskID, agentID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("agentcontract: création répertoire logs: %w", err)
	}
	body, err := json.MarshalIndent(finalize(result), "", "  ")
	if err != nil {
		return fmt.Errorf("agentcontract: marshal contract.json: %w", err)
	}
	body = append(body, '\n')
	path := filepath.Join(dir, "contract.json")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("agentcontract: écriture %s: %w", path, err)
	}
	return nil
}
