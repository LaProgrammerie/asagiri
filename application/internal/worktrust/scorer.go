package worktrust

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

var dimensionLabels = map[DimensionID]string{
	DimSpecificationAlignment: "Alignement spec",
	DimImplementationQuality:  "Qualité impl.",
	DimValidationStrength:     "Force validation",
	DimGateConfidence:         "Confiance gates",
	DimHumanConfidence:        "Confiance humaine",
	DimResidualRisk:           "Risque résiduel",
}

var dimensionWeights = map[DimensionID]float64{
	DimGateConfidence:         0.25,
	DimValidationStrength:     0.25,
	DimHumanConfidence:        0.20,
	DimSpecificationAlignment: 0.15,
	DimImplementationQuality:  0.10,
	DimResidualRisk:           0.05,
}

const (
	thresholdTrusted    = 80
	thresholdAcceptable = 60
	thresholdRisky      = 40
)

func scoreDimensions(s taskSignals) []WorkTrustDimension {
	spec := scoreSpecificationAlignment(s)
	impl := scoreImplementationQuality(s)
	validation := scoreValidationStrength(s)
	gateConf := scoreGateConfidence(s)
	human := scoreHumanConfidence(s)
	residual := scoreResidualRisk(s, spec, impl, validation, gateConf, human)

	return []WorkTrustDimension{
		spec, impl, validation, gateConf, human, residual,
	}
}

func scoreSpecificationAlignment(s taskSignals) WorkTrustDimension {
	var parts []float64
	var sources []string

	if s.gateActive("plan") {
		if sc, ok := gateEntryScore(s, "plan"); ok {
			parts = append(parts, sc)
			sources = append(sources, "plan")
		}
	}
	if s.gateActive(gates.EnrichGateName) {
		if sc, ok := gateEntryScore(s, gates.EnrichGateName); ok {
			parts = append(parts, sc)
			sources = append(sources, gates.EnrichGateName)
		} else if taskBeforeEnrich(s.Task.Status) {
			return unevaluatedDim(DimSpecificationAlignment, "enrich gate not reached")
		} else {
			parts = append(parts, 45)
			sources = append(sources, gates.EnrichGateName)
		}
	}

	if len(parts) == 0 {
		return unevaluatedDim(DimSpecificationAlignment, "no active spec gates")
	}
	score := average(parts)
	return WorkTrustDimension{
		ID:          DimSpecificationAlignment,
		Label:       dimensionLabels[DimSpecificationAlignment],
		Score:       score,
		Status:      dimensionStatusFromScore(score),
		Summary:     dimSummary(sources, score),
		SourceGates: sources,
	}
}

func scoreImplementationQuality(s taskSignals) WorkTrustDimension {
	var parts []float64
	var sources []string

	if s.gateActive("governance") {
		if sc, ok := gateEntryScore(s, "governance"); ok {
			parts = append(parts, sc)
			sources = append(sources, "governance")
		} else if taskBeforeGovernance(s.Task.Status) {
			return unevaluatedDim(DimImplementationQuality, "governance gate not reached")
		} else {
			parts = append(parts, 45)
			sources = append(sources, "governance")
		}
	}

	statusScore := statusImplementationScore(s.Task.Status)
	if statusScore >= 0 {
		parts = append(parts, statusScore)
	}

	if len(parts) == 0 {
		return unevaluatedDim(DimImplementationQuality, "no implementation signals")
	}
	score := average(parts)
	return WorkTrustDimension{
		ID:          DimImplementationQuality,
		Label:       dimensionLabels[DimImplementationQuality],
		Score:       score,
		Status:      dimensionStatusFromScore(score),
		Summary:     dimSummary(sources, score),
		SourceGates: sources,
	}
}

func scoreValidationStrength(s taskSignals) WorkTrustDimension {
	var parts []float64
	var sources []string

	if s.Validation != nil {
		sc := validationDocScore(s.Validation)
		parts = append(parts, sc)
		sources = append(sources, "validation")
	} else if taskNeedsValidationArtifact(s.Task.Status) {
		parts = append(parts, 40)
		sources = append(sources, "validation")
	}

	if s.gateActive(gates.VerifyEvidenceGateName) {
		if sc, ok := gateEntryScore(s, gates.VerifyEvidenceGateName); ok {
			parts = append(parts, sc)
			sources = append(sources, gates.VerifyEvidenceGateName)
		} else if !taskNeedsValidationArtifact(s.Task.Status) {
			return unevaluatedDim(DimValidationStrength, "verify_evidence not reached")
		} else {
			parts = append(parts, 45)
			sources = append(sources, gates.VerifyEvidenceGateName)
		}
	}

	if len(parts) == 0 {
		return unevaluatedDim(DimValidationStrength, "no validation signals")
	}
	score := average(parts)
	if s.VerifyEvOK && s.Validation != nil && validationDocScore(s.Validation) >= 85 {
		if score < 90 {
			score = 90
		}
	}
	return WorkTrustDimension{
		ID:          DimValidationStrength,
		Label:       dimensionLabels[DimValidationStrength],
		Score:       score,
		Status:      dimensionStatusFromScore(score),
		Summary:     dimSummary(sources, score),
		SourceGates: sources,
	}
}

func scoreGateConfidence(s taskSignals) WorkTrustDimension {
	var parts []float64
	var sources []string
	for _, name := range []string{"plan", gates.EnrichGateName, "governance", gates.HumanReviewGateName, gates.VerifyEvidenceGateName} {
		if !s.gateActive(name) {
			continue
		}
		if sc, ok := gateEntryScore(s, name); ok {
			parts = append(parts, sc)
			sources = append(sources, name)
		} else if gateExpectedForStatus(name, s.Task.Status) {
			parts = append(parts, 40)
			sources = append(sources, name)
		}
	}
	if len(parts) == 0 {
		return unevaluatedDim(DimGateConfidence, "no active work gates")
	}
	score := average(parts)
	return WorkTrustDimension{
		ID:          DimGateConfidence,
		Label:       dimensionLabels[DimGateConfidence],
		Score:       score,
		Status:      dimensionStatusFromScore(score),
		Summary:     dimSummary(sources, score),
		SourceGates: sources,
	}
}

func scoreHumanConfidence(s taskSignals) WorkTrustDimension {
	if !s.gateActive(gates.HumanReviewGateName) {
		return unevaluatedDim(DimHumanConfidence, "human_review gate inactive")
	}
	if s.HasPending {
		return WorkTrustDimension{
			ID:          DimHumanConfidence,
			Label:       dimensionLabels[DimHumanConfidence],
			Score:       15,
			Status:      DimStatusFailed,
			Summary:     "human review pending",
			SourceGates: []string{gates.HumanReviewGateName},
		}
	}
	if sc, ok := gateEntryScore(s, gates.HumanReviewGateName); ok {
		return WorkTrustDimension{
			ID:          DimHumanConfidence,
			Label:       dimensionLabels[DimHumanConfidence],
			Score:       sc,
			Status:      dimensionStatusFromScore(sc),
			Summary:     "human_review " + s.Entries[gates.HumanReviewGateName].Status,
			SourceGates: []string{gates.HumanReviewGateName},
		}
	}
	if taskBeforeHumanReview(s.Task.Status) {
		return unevaluatedDim(DimHumanConfidence, "human_review not reached")
	}
	return WorkTrustDimension{
		ID:          DimHumanConfidence,
		Label:       dimensionLabels[DimHumanConfidence],
		Score:       40,
		Status:      DimStatusWeak,
		Summary:     "human_review missing",
		SourceGates: []string{gates.HumanReviewGateName},
	}
}

func scoreResidualRisk(s taskSignals, others ...WorkTrustDimension) WorkTrustDimension {
	risk := 100.0
	var notes []string

	for gateName, entry := range s.Entries {
		if !s.gateActive(gateName) {
			continue
		}
		st := strings.ToLower(strings.TrimSpace(entry.Status))
		switch st {
		case string(gates.VerdictFail):
			risk -= 35
			notes = append(notes, gateName+" fail")
		case string(gates.VerdictWarn):
			if !gates.GateEntrySatisfied(s.warnAdvisory(gateName), entry) {
				risk -= 15
				notes = append(notes, gateName+" warn")
			} else {
				risk -= 5
			}
		default:
			if st != string(gates.VerdictPass) && gateExpectedForStatus(gateName, s.Task.Status) {
				risk -= 10
			}
		}
	}

	for _, name := range []string{gates.EnrichGateName, "governance", gates.VerifyEvidenceGateName} {
		if s.gateActive(name) && gateExpectedForStatus(name, s.Task.Status) {
			if _, ok := s.Entries[name]; !ok {
				risk -= 12
				notes = append(notes, name+" missing")
			}
		}
	}

	if s.HasPending {
		risk = minFloat(risk, 20)
		notes = append(notes, "HR pending")
	}

	st := strings.ToLower(strings.TrimSpace(s.Task.Status))
	switch st {
	case asagiri.StatusVerifyFailed, asagiri.StatusFailed:
		risk = minFloat(risk, 25)
		notes = append(notes, st)
	case asagiri.StatusReviewFailed:
		risk = minFloat(risk, 35)
		notes = append(notes, st)
	}

	if risk < 0 {
		risk = 0
	}
	summary := "low residual risk"
	if len(notes) > 0 {
		summary = strings.Join(notes, "; ")
	}
	return WorkTrustDimension{
		ID:      DimResidualRisk,
		Label:   dimensionLabels[DimResidualRisk],
		Score:   risk,
		Status:  dimensionStatusFromScore(risk),
		Summary: summary,
	}
}

func gateEntryScore(s taskSignals, gateName string) (float64, bool) {
	entry, ok := s.Entries[gateName]
	if !ok {
		return 0, false
	}
	st := strings.ToLower(strings.TrimSpace(entry.Status))
	base := 0.0
	switch st {
	case string(gates.VerdictPass):
		base = 95
	case string(gates.VerdictWarn):
		if gates.GateEntrySatisfied(s.warnAdvisory(gateName), entry) {
			base = 72
		} else {
			base = 50
		}
	case string(gates.VerdictFail):
		base = 18
	default:
		return 0, false
	}
	if entry.Confidence > 0 && entry.Confidence <= 1 {
		conf := entry.Confidence * 100
		if st == string(gates.VerdictPass) {
			base = (base + conf) / 2
		}
	}
	return base, true
}

func validationDocScore(doc *validationEvidenceDocument) float64 {
	if doc == nil || len(doc.Commands) == 0 {
		return 50
	}
	for _, c := range doc.Commands {
		if c.ExitCode != 0 {
			return 35
		}
	}
	return 88
}

func statusImplementationScore(status string) float64 {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case asagiri.StatusMerged, asagiri.StatusReadyForPR, asagiri.StatusReviewed:
		return 95
	case asagiri.StatusVerified:
		return 85
	case asagiri.StatusImplemented:
		return 75
	case asagiri.StatusRunning:
		return 55
	case asagiri.StatusEnriched, asagiri.StatusPlanned, asagiri.StatusPending:
		return 45
	case asagiri.StatusVerifyFailed, asagiri.StatusReviewFailed, asagiri.StatusFailed:
		return 20
	default:
		return -1
	}
}

func computeOverall(dims []WorkTrustDimension) float64 {
	var sum, wSum float64
	for _, d := range dims {
		if d.Score < 0 {
			continue
		}
		w := dimensionWeights[d.ID]
		if w == 0 {
			w = 0.1
		}
		sum += d.Score * w
		wSum += w
	}
	if wSum == 0 {
		return 50
	}
	return sum / wSum
}

func applyStatusCaps(status string, overall float64) float64 {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case asagiri.StatusVerifyFailed, asagiri.StatusFailed:
		if overall > 35 {
			return 35
		}
	case asagiri.StatusReviewFailed:
		if overall > 45 {
			return 45
		}
	}
	return overall
}

func computeVerdict(s taskSignals, overall float64, dims []WorkTrustDimension) Verdict {
	if s.HasPending {
		return VerdictBlocked
	}
	st := strings.ToLower(strings.TrimSpace(s.Task.Status))
	if st == asagiri.StatusVerifyFailed || st == asagiri.StatusFailed {
		return VerdictBlocked
	}
	if hasBlockingGateFail(s) {
		return VerdictBlocked
	}
	if overall < thresholdRisky && hasAnyGateFail(s) {
		return VerdictBlocked
	}
	if st == asagiri.StatusReviewFailed {
		return VerdictRisky
	}
	if overall >= thresholdTrusted && !hasAnyGateFail(s) && !hasNonAdvisoryWarn(s) {
		return VerdictTrusted
	}
	if overall >= thresholdAcceptable {
		return VerdictAcceptable
	}
	if overall >= thresholdRisky {
		return VerdictRisky
	}
	return VerdictBlocked
}

func hasBlockingGateFail(s taskSignals) bool {
	blocking := []string{gates.EnrichGateName, "governance", gates.HumanReviewGateName}
	for _, name := range blocking {
		entry, ok := s.Entries[name]
		if !ok || !s.gateActive(name) {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(entry.Status), string(gates.VerdictFail)) {
			return true
		}
	}
	return false
}

func hasAnyGateFail(s taskSignals) bool {
	for name, entry := range s.Entries {
		if !s.gateActive(name) {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(entry.Status), string(gates.VerdictFail)) {
			return true
		}
	}
	return false
}

func hasNonAdvisoryWarn(s taskSignals) bool {
	for name, entry := range s.Entries {
		if !s.gateActive(name) {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(entry.Status), string(gates.VerdictWarn)) &&
			!gates.GateEntrySatisfied(s.warnAdvisory(name), entry) {
			return true
		}
	}
	return false
}

func unevaluatedDim(id DimensionID, summary string) WorkTrustDimension {
	return WorkTrustDimension{
		ID:      id,
		Label:   dimensionLabels[id],
		Score:   UnevaluatedScore,
		Status:  DimStatusUnevaluated,
		Summary: summary,
	}
}

func dimensionStatusFromScore(score float64) DimensionStatus {
	if score < 0 {
		return DimStatusUnevaluated
	}
	switch {
	case score >= 80:
		return DimStatusStrong
	case score >= 60:
		return DimStatusModerate
	case score >= 40:
		return DimStatusWeak
	default:
		return DimStatusFailed
	}
}

func dimSummary(sources []string, score float64) string {
	if len(sources) == 0 {
		return ""
	}
	return strings.Join(sources, ", ") + " — " + formatScore(score)
}

func formatScore(score float64) string {
	if score < 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.0f/100", score)
}

func average(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func verdictSummary(v Verdict, overall float64) string {
	return string(v) + " — " + scoreLabel(overall)
}

func scoreLabel(overall float64) string {
	if overall < 0 {
		return "score n/a"
	}
	n := int(overall + 0.5)
	if n < 0 {
		n = 0
	}
	if n > 100 {
		n = 100
	}
	return fmt.Sprintf("score %d/100", n)
}

func mapFindingSeverity(severity, gateStatus string) string {
	s := strings.ToLower(strings.TrimSpace(severity))
	switch s {
	case "critical", "high", "medium", "low", "info":
		return s
	case "fail", "4":
		return "high"
	case "warn", "3":
		return "medium"
	case "pass", "2", "1":
		return "info"
	}
	if strings.EqualFold(gateStatus, string(gates.VerdictFail)) {
		return "high"
	}
	return "medium"
}

func taskBeforeEnrich(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case asagiri.StatusPending, asagiri.StatusPlanned, "":
		return true
	default:
		return false
	}
}

func taskBeforeGovernance(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case asagiri.StatusPending, asagiri.StatusPlanned, asagiri.StatusEnriched, asagiri.StatusRunning, "":
		return true
	default:
		return false
	}
}

func taskBeforeHumanReview(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case asagiri.StatusPending, asagiri.StatusPlanned, asagiri.StatusEnriched, asagiri.StatusRunning, "":
		return true
	default:
		return false
	}
}

func taskNeedsValidationArtifact(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case asagiri.StatusVerified, asagiri.StatusReviewed, asagiri.StatusReadyForPR, asagiri.StatusMerged,
		asagiri.StatusReviewFailed:
		return true
	default:
		return false
	}
}

func gateExpectedForStatus(gateName, status string) bool {
	st := strings.ToLower(strings.TrimSpace(status))
	switch gateName {
	case gates.EnrichGateName:
		return !taskBeforeEnrich(st)
	case "governance":
		return st == asagiri.StatusImplemented || st == asagiri.StatusVerified ||
			st == asagiri.StatusReviewFailed || st == asagiri.StatusReviewed
	case gates.HumanReviewGateName:
		return st == asagiri.StatusImplemented || st == asagiri.StatusVerified ||
			st == asagiri.StatusReviewFailed || st == asagiri.StatusReviewed
	case gates.VerifyEvidenceGateName:
		return taskNeedsValidationArtifact(st) || st == asagiri.StatusImplemented
	default:
		return false
	}
}
