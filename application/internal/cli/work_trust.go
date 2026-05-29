package cli

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/trust"
)

func resolveWorkStrictTrust(repoRoot, trustFlowFlag, featureHint string) (flowID, productID string, err error) {
	if flow := strings.TrimSpace(trustFlowFlag); flow != "" {
		flowID, productID, err = trust.ResolveProductFlow(repoRoot, flow)
		if err != nil {
			return "", "", fmt.Errorf("strict-trust: %w (pass --trust-flow with a flow id from .asagiri/products/*/flows/)", err)
		}
		return flowID, productID, nil
	}
	feature := strings.TrimSpace(featureHint)
	if feature == "" {
		return "", "", fmt.Errorf("strict-trust: pass --trust-flow with a product flow id")
	}
	flowID, productID, err = trust.ResolveProductFlow(repoRoot, feature)
	if err != nil {
		return "", "", fmt.Errorf(
			"strict-trust: could not resolve product flow from feature %q: %w; pass --trust-flow",
			feature, err,
		)
	}
	return flowID, productID, nil
}
