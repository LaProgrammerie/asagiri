package trust

import (
	"context"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/confidence"
)

// LowConfidenceThreshold triggers trust.low_confidence events (spec §18).
const LowConfidenceThreshold = 0.7

// RuntimeEmitter implements EventEmitter via runtime.VerificationEmitter.
type RuntimeEmitter struct {
	inner *runtime.VerificationEmitter
	flow  string
}

// NewRuntimeEmitter wires trust verification events to a runtime store.
func NewRuntimeEmitter(store *runtime.Store) *RuntimeEmitter {
	if store == nil {
		return nil
	}
	return &RuntimeEmitter{inner: &runtime.VerificationEmitter{Store: store}}
}

// Emit publishes a verification runtime event for the configured flow.
func (e *RuntimeEmitter) Emit(_ context.Context, name string, payload map[string]any) error {
	if e == nil || e.inner == nil {
		return nil
	}
	if payload == nil {
		payload = map[string]any{}
	}
	return e.inner.Emit(name, e.flow, payload)
}

func (e *Engine) emitLifecycleEvents(ctx context.Context, scope VerificationScope, checks []VerificationCheck, conf confidence.Report, gate GateEvaluation) {
	if e.Emitter == nil {
		return
	}
	if re, ok := e.Emitter.(*RuntimeEmitter); ok {
		re.flow = scope.Flow
	}
	base := map[string]any{
		"trust_id": scope.TrustID,
		"flow":     scope.Flow,
		"branch":   scope.Branch,
	}

	completed := clonePayload(base)
	completed["overall_confidence"] = conf.Overall
	completed["gate_status"] = string(gate.Status)
	_ = e.Emitter.Emit(ctx, runtime.EventVerificationCompleted, completed)

	e.emitConditionalEvents(ctx, checks, conf, base)
}

func (e *Engine) emitConditionalEvents(ctx context.Context, checks []VerificationCheck, conf confidence.Report, base map[string]any) {
	if conf.Overall > 0 && conf.Overall < LowConfidenceThreshold {
		p := clonePayload(base)
		p["overall_confidence"] = conf.Overall
		_ = e.Emitter.Emit(ctx, runtime.EventTrustLowConfidence, p)
	}
	if securityIssueDetected(checks) {
		_ = e.Emitter.Emit(ctx, runtime.EventSecurityIssueDetected, clonePayload(base))
	}
	if flowIntegrityFailed(checks) {
		_ = e.Emitter.Emit(ctx, runtime.EventFlowIntegrityFailed, clonePayload(base))
	}
	if contractBreakingChangeDetected(checks) {
		_ = e.Emitter.Emit(ctx, runtime.EventContractBreakingChangeDetected, clonePayload(base))
	}
}

func clonePayload(base map[string]any) map[string]any {
	out := make(map[string]any, len(base))
	for k, v := range base {
		out[k] = v
	}
	return out
}

func securityIssueDetected(checks []VerificationCheck) bool {
	for _, c := range checks {
		if c.Type != CheckSecurity && c.Type != CheckFlows {
			continue
		}
		if c.Status == CheckStatusFailed {
			return true
		}
		for _, f := range c.Findings {
			if f.Severity == SeverityError || f.Severity == SeverityCritical {
				if containsCategory(f.Category, "security") {
					return true
				}
			}
		}
	}
	return false
}

func flowIntegrityFailed(checks []VerificationCheck) bool {
	for _, c := range checks {
		if c.Type != CheckFlows {
			continue
		}
		if c.Status == CheckStatusFailed {
			return true
		}
		for _, f := range c.Findings {
			if f.Severity == SeverityError && containsCategory(f.Category, "flow.integrity") {
				return true
			}
		}
	}
	return false
}

func contractBreakingChangeDetected(checks []VerificationCheck) bool {
	for _, c := range checks {
		if c.Type != CheckContracts && c.Type != CheckBackwardCompatibility {
			continue
		}
		if c.Status == CheckStatusFailed {
			return true
		}
		for _, f := range c.Findings {
			if containsCategory(f.Category, "compatibility.contract", "breaking") {
				if f.Severity == SeverityError || f.Severity == SeverityWarning {
					return true
				}
			}
		}
	}
	return false
}

func containsCategory(category string, prefixes ...string) bool {
	for _, p := range prefixes {
		if category == p || strings.HasPrefix(category, p) {
			return true
		}
	}
	return false
}
