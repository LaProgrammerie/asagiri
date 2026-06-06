package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/product"
)

type productLayerOutcome struct {
	Handled       bool
	ProductID     string
	StopWorkFlow  bool
	ResolvedPatch intent.ResolvedIntent
}

func handleProductLayer(
	w io.Writer,
	instruction string,
	scope intent.IntentScope,
	resolved intent.ResolvedIntent,
	opts intent.WorkOptions,
	repoRoot string,
) (productLayerOutcome, error) {
	out := productLayerOutcome{ResolvedPatch: resolved}
	if !intent.ShouldRunProductLayer(scope) {
		return out, nil
	}
	if resolved.Action != intent.IntentDevelop && resolved.Action != intent.IntentUnknown {
		return out, nil
	}

	productID := product.ResolveProductID(repoRoot, instruction, resolved.Feature)
	state := product.InspectArtifacts(repoRoot, productID)
	out.ProductID = productID
	out.Handled = true

	if opts.DryRun || opts.PlanOnly {
		_, _ = fmt.Fprint(w, product.FormatLayerDryRun(state, opts.PlanOnly))
		out.StopWorkFlow = true
		return out, nil
	}

	if !state.NeedsPreparation() {
		out.ResolvedPatch.Feature = productID
		return out, nil
	}

	if !opts.Yes && !opts.Interactive {
		return out, &intent.ConfirmationRequiredError{
			Message: intent.ProductLayerNonInteractiveMessage(instruction),
		}
	}
	if opts.Interactive && !opts.Yes {
		ok, err := confirmPrompt(os.Stderr, os.Stdin, productLayerConfirmMsg())
		if err != nil {
			return out, err
		}
		if !ok {
			return out, fmt.Errorf("annulé par l'utilisateur")
		}
	}

	svc := product.NewService(repoRoot)
	res, err := svc.PrepareLayer(product.PrepareLayerOptions{
		Intent:  instruction,
		Product: productID,
		DryRun:  false,
	})
	if err != nil {
		return out, err
	}
	_, _ = fmt.Fprint(w, product.FormatLayerResult(res))
	out.ResolvedPatch.Feature = productID
	if len(res.Generated) > 0 {
		out.StopWorkFlow = true
	}
	return out, nil
}

func productLayerConfirmMsg() string {
	return strings.TrimSpace(`Product-level intent detected.

Asagiri can prepare the Product Layer before coding.

Steps:
1. Product model
2. Prototype
3. Flows
4. Contracts
5. Specs
6. Tasks

Continue?`)
}
