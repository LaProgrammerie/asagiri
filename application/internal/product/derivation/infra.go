package derivation

import "strings"

func deriveAPIRequirements(flows []FlowInput) []string {
	var out []string
	for _, flow := range flows {
		for _, action := range flow.StepActions {
			if strings.HasPrefix(strings.ToUpper(action.ContractRef), "POST ") || strings.HasPrefix(strings.ToUpper(action.ContractRef), "GET ") || strings.HasPrefix(strings.ToUpper(action.ContractRef), "PUT ") {
				out = append(out, "endpoint:"+action.ContractRef)
			}
			if action.ContractRef == "" {
				out = append(out, "endpoint:TODO:"+flow.ID+"."+action.Action)
			}
		}
	}
	return out
}

func deriveInfraRequirements(flows []FlowInput) []string {
	var out []string
	for _, flow := range flows {
		for _, action := range flow.StepActions {
			name := strings.ToLower(action.Action)
			if strings.Contains(name, "invite") || strings.Contains(name, "email") {
				out = append(out, "queue:invitation_delivery")
				out = append(out, "worker:invitation_retry")
				out = append(out, "provider:email")
			}
		}
	}
	return out
}

func deriveInfrastructureRequirements(flows []FlowInput) []string {
	var out []string
	for _, flow := range flows {
		if len(flow.Metrics) > 0 {
			out = append(out, "service:metrics_backend")
		}
		for _, action := range flow.StepActions {
			if strings.Contains(strings.ToLower(action.Action), "invite") {
				out = append(out, "service:queue")
			}
		}
	}
	return out
}
