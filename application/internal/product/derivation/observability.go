package derivation

import "strings"

func deriveObservabilityRequirements(flows []FlowInput) []string {
	var out []string
	for _, flow := range flows {
		out = append(out, "trace:"+flow.ID+".start")
		out = append(out, "trace:"+flow.ID+".complete")
		for _, metric := range flow.Metrics {
			metric = strings.TrimSpace(metric)
			if metric == "" {
				continue
			}
			out = append(out, "metric:"+metric)
		}
		for _, action := range flow.StepActions {
			if len(action.Errors) > 0 {
				out = append(out, "log:"+flow.ID+"."+action.Action+".failed")
			}
			if strings.Contains(strings.ToLower(action.Action), "invite") {
				out = append(out, "metric:invitation_delivery_success_rate")
			}
		}
	}
	return out
}
