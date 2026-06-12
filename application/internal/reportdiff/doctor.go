package reportdiff

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/doctor"
)

func DiffDoctor(before, after doctor.Report, paths ReportPaths) DoctorDiff {
	diff := DoctorDiff{
		ReportVersion: ReportVersion,
		Scope:         "doctor",
		Paths:         paths,
		Ready: BoolDelta{
			Before:  before.Ready,
			After:   after.Ready,
			Changed: before.Ready != after.Ready,
		},
		Warnings: countDelta(len(before.Warnings), len(after.Warnings)),
		Failures: countDelta(len(before.Failures), len(after.Failures)),
		NextAction: nextActionDelta(
			firstDoctorCLI(before.NextActions),
			firstDoctorCLI(after.NextActions),
		),
	}
	bv, av := doctorTrustVerdict(before), doctorTrustVerdict(after)
	if bv != "" || av != "" {
		diff.TrustVerdict = verdictDelta(bv, av)
	}
	return diff
}

func doctorTrustVerdict(r doctor.Report) string {
	if r.Trust == nil {
		return ""
	}
	return strings.TrimSpace(r.Trust.Verdict)
}

func firstDoctorCLI(actions []doctor.Action) string {
	if len(actions) == 0 {
		return ""
	}
	return strings.TrimSpace(actions[0].CLI)
}

func countDelta(before, after int) CountDelta {
	return CountDelta{
		Before:  before,
		After:   after,
		Delta:   after - before,
		Changed: before != after,
	}
}
