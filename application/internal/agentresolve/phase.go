package agentresolve

// Phase identifies a workflow agent invocation kind.
type Phase string

const (
	PhaseDev            Phase = "dev"
	PhaseEnrich         Phase = "enrich"
	PhaseEnrichGate     Phase = "enrich_gate"
	PhaseGovernance     Phase = "governance"
	PhaseReview         Phase = "review"
	PhaseVerifyEvidence Phase = "verify_evidence"
	PhasePlanGate       Phase = "plan_gate"
	PhaseHumanReview    Phase = "human_review"
)
