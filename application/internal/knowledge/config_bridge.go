package knowledge

import "github.com/LaProgrammerie/asagiri/application/internal/config"

// BuildRequestFromConfig maps config defaults into a build request (CLI flags override separately).
func BuildRequestFromConfig(repoRoot string, cfg *config.Config) BuildRequest {
	req := BuildRequest{RepoRoot: repoRoot}
	if cfg == nil {
		return req
	}
	k := cfg.Knowledge
	req.IncludeFlows = k.DefaultIncludeFlows
	req.IncludeContracts = k.DefaultIncludeContracts
	req.IncludeCode = k.DefaultIncludeCode
	req.IncludeTests = k.DefaultIncludeTests
	req.Incremental = k.IncrementalByDefault
	return req
}

// ApplyBuildFlags overlays explicit CLI flag values onto the request.
func ApplyBuildFlags(req *BuildRequest, flags BuildCLIFlags) {
	if flags.IncrementalSet {
		req.Incremental = flags.Incremental
	}
	if flags.ScopeSet {
		req.Scope = flags.Scope
	}
	if flags.IncludeFlowsSet {
		req.IncludeFlows = flags.IncludeFlows
	}
	if flags.IncludeContractsSet {
		req.IncludeContracts = flags.IncludeContracts
	}
	if flags.IncludeCodeSet {
		req.IncludeCode = flags.IncludeCode
	}
	if flags.IncludeTestsSet {
		req.IncludeTests = flags.IncludeTests
	}
	if flags.IncludeInfraSet {
		req.IncludeInfra = flags.IncludeInfra
	}
	if flags.IncludeADRSet {
		req.IncludeADR = flags.IncludeADR
	}
	if flags.IncludeRuntimeSet {
		req.IncludeRuntime = flags.IncludeRuntime
	}
}

// BuildCLIFlags tracks whether a knowledge build flag was set on the CLI.
type BuildCLIFlags struct {
	Incremental      bool
	IncrementalSet   bool
	Scope            string
	ScopeSet         bool
	IncludeFlows     bool
	IncludeFlowsSet  bool
	IncludeContracts bool
	IncludeContractsSet bool
	IncludeCode      bool
	IncludeCodeSet   bool
	IncludeTests     bool
	IncludeTestsSet  bool
	IncludeInfra     bool
	IncludeInfraSet  bool
	IncludeADR       bool
	IncludeADRSet    bool
	IncludeRuntime   bool
	IncludeRuntimeSet bool
}
