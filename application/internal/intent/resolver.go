package intent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// IntentResolver resolves free-text instructions (specv2 §5).
type IntentResolver interface {
	Resolve(ctx context.Context, input IntentInput) (ResolvedIntent, error)
}

// HybridResolver implements deterministic + fuzzy + Ollama fallback.
type HybridResolver struct {
	Ollama OllamaClient
}

// OllamaClient resolves ambiguous intents via local LLM.
type OllamaClient interface {
	ResolveIntent(ctx context.Context, instruction string, candidates []string) (ResolvedIntent, error)
}

// NewHybridResolver returns the default resolver.
func NewHybridResolver() *HybridResolver {
	return &HybridResolver{Ollama: &HTTPOllamaClient{}}
}

var (
	notionURLRe  = regexp.MustCompile(`(?i)https?://(?:www\.)?notion\.(?:so|site)/[^\s]+`)
	featureSlugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
)

// Resolve implements IntentResolver.
func (r *HybridResolver) Resolve(ctx context.Context, input IntentInput) (ResolvedIntent, error) {
	if input.Config != nil && !input.Config.Intent.Enabled {
		return ResolvedIntent{}, ErrIntentDisabled
	}

	raw := strings.TrimSpace(input.RawInstruction)
	if raw == "" {
		return ResolvedIntent{}, ErrLowConfidence
	}

	out := r.resolveDeterministic(raw, input.StateSnapshot)
	if out.Action != IntentUnknown && out.Confidence >= minConfidence(input.Config) {
		return out, nil
	}

	if feat, score, ok := matchFeatureFromText(raw, input.StateSnapshot); ok {
		if out.Action == IntentUnknown {
			out.Action = IntentDevelop
		}
		out.Feature = feat.Name
		out.Confidence = score
		out.Reason = "feature match"
		if out.Confidence >= minConfidence(input.Config) {
			return out, nil
		}
	}

	candidates := rankCandidates(raw, input.StateSnapshot)
	minConf := minConfidence(input.Config)
	if out.Confidence < minConf && len(candidates) > 0 {
		if input.Config != nil && input.Config.Intent.Resolver.UseOllamaFallback && r.Ollama != nil {
			llm, err := r.Ollama.ResolveIntent(ctx, raw, candidates)
			if err == nil && llm.Action != IntentUnknown && llm.Confidence >= minConf {
				return llm, nil
			}
		}
		if !input.Interactive && (input.Config == nil || input.Config.Intent.Resolver.AskWhenBelowConfidence) {
			return ResolvedIntent{}, &AmbiguityError{
				Instruction: raw,
				Candidates:  candidates,
				Confidence:  out.Confidence,
			}
		}
	}

	if out.Action != IntentUnknown && out.Confidence >= minConf {
		return out, nil
	}
	if out.Action != IntentUnknown && out.Confidence < minConf && !input.Interactive {
		return ResolvedIntent{}, &AmbiguityError{Instruction: raw, Candidates: candidates, Confidence: out.Confidence}
	}
	if out.Action != IntentUnknown {
		return out, nil
	}
	return ResolvedIntent{Action: IntentUnknown, Confidence: 0, Reason: "no match"}, ErrLowConfidence
}

func minConfidence(cfg *config.Config) float64 {
	if cfg == nil || cfg.Intent.Resolver.MinConfidence == 0 {
		return 0.75
	}
	return cfg.Intent.Resolver.MinConfidence
}

func (r *HybridResolver) resolveDeterministic(raw string, snap StateSnapshot) ResolvedIntent {
	lower := strings.ToLower(raw)

	if notionURLRe.MatchString(raw) {
		return ResolvedIntent{
			Action:       IntentImport,
			Source:       "notion",
			SourceRef:    notionURLRe.FindString(raw),
			RequiresSync: true,
			Confidence:   0.95,
			Reason:       "notion url",
		}
	}

	action := IntentDevelop
	switch {
	case containsAny(lower, "sync", "synchronise", "synchronize", "synchroniser"):
		action = IntentSync
	case containsAny(lower, "continue", "continuer", "reprends", "reprendre", "resume", "reprend"):
		action = IntentResume
	case containsAny(lower, "verify", "vérifier", "verifier", "vérifie", "valide", "validation"):
		action = IntentVerify
	case wordMatch(lower, "import", "importer", "importe"):
		action = IntentImport
	case containsAny(lower, "review", "revue", "reviewer"):
		action = IntentReview
	case containsAny(lower, "fix", "corrige", "corriger", "répare", "repare", "tests"):
		action = IntentFix
	case containsAny(lower, "status", "état", "etat", "state"):
		action = IntentStatus
	case containsAny(lower, "développe", "developpe", "develop", "dev ", " impl"):
		action = IntentDevelop
	}

	feature := extractFeatureSlug(raw, snap)
	conf := 0.5
	if feature != "" {
		if _, score, ok := FindFeature(snap, feature); ok {
			conf = score
		} else if featureSlugRe.MatchString(feature) {
			conf = 0.82
		}
	}

	return ResolvedIntent{
		Action:     action,
		Feature:    feature,
		Confidence: conf,
		Reason:     "deterministic",
	}
}

func extractFeatureSlug(raw string, snap StateSnapshot) string {
	lower := strings.ToLower(raw)
	for _, f := range snap.Features {
		if strings.Contains(lower, f.Name) {
			return f.Name
		}
	}
	tokens := tokenize(raw)
	for _, t := range tokens {
		if verbSlugs[t] || len(t) < 3 {
			continue
		}
		if featureSlugRe.MatchString(t) {
			return t
		}
	}
	// quoted slug
	if i := strings.Index(raw, `"`); i >= 0 {
		rest := raw[i+1:]
		if j := strings.Index(rest, `"`); j > 0 {
			return slugFeature(rest[:j])
		}
	}
	return ""
}

func matchFeatureFromText(raw string, snap StateSnapshot) (FeatureState, float64, bool) {
	tokens := tokenize(raw)
	var best FeatureState
	var bestScore float64
	for _, f := range snap.Features {
		for _, tok := range tokens {
			if len(tok) < 3 {
				continue
			}
			sc := fuzzyScore(tok, f.Name)
			if sc > bestScore {
				bestScore = sc
				best = f
			}
		}
		lower := strings.ToLower(raw)
		if strings.Contains(lower, f.Name) {
			return f, 0.9, true
		}
		// match slug tokens: "import csv" → import-csv
		parts := strings.Split(f.Name, "-")
		if len(parts) > 1 {
			all := true
			for _, p := range parts {
				if p == "" || !strings.Contains(lower, p) {
					all = false
					break
				}
			}
			if all {
				return f, 0.88, true
			}
		}
	}
	if bestScore >= 0.6 {
		return best, bestScore, true
	}
	return FeatureState{}, 0, false
}

func rankCandidates(raw string, snap StateSnapshot) []string {
	type scored struct {
		name  string
		score float64
	}
	lower := strings.ToLower(raw)
	var list []scored
	for _, f := range snap.Features {
		sc := fuzzyScore(f.Name, lower)
		for _, tok := range tokenize(raw) {
			if sc2 := fuzzyScore(tok, f.Name); sc2 > sc {
				sc = sc2
			}
		}
		list = append(list, scored{f.Name, sc})
	}
	sortScored := func(i, j int) bool { return list[i].score > list[j].score }
	_ = sortScored
	// simple sort
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].score > list[i].score {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	out := make([]string, 0, 3)
	for _, s := range list {
		if s.score < 0.3 {
			continue
		}
		out = append(out, s.name)
		if len(out) >= 3 {
			break
		}
	}
	return out
}

func containsAny(s string, parts ...string) bool {
	for _, p := range parts {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

func wordMatch(s string, words ...string) bool {
	tokens := tokenize(s)
	for _, w := range words {
		for _, t := range tokens {
			if t == w {
				return true
			}
		}
	}
	return false
}

var verbSlugs = map[string]bool{
	"develop": true, "developpe": true, "dev": true, "continue": true,
	"resume": true, "reprends": true, "reprend": true, "verify": true, "review": true,
	"fix": true, "sync": true, "import": true, "status": true, "do": true, "something": true,
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	repl := strings.NewReplacer(",", " ", ".", " ", ":", " ", ";", " ", "'", " ", "\"", " ")
	s = repl.Replace(s)
	return strings.Fields(s)
}

// HTTPOllamaClient calls Ollama for JSON intent resolution.
type HTTPOllamaClient struct {
	HTTP    *http.Client
	BaseURL string
	Model   string
}

func (c *HTTPOllamaClient) ResolveIntent(ctx context.Context, instruction string, candidates []string) (ResolvedIntent, error) {
	base := c.BaseURL
	if base == "" {
		base = "http://localhost:11434"
	}
	model := c.Model
	if model == "" {
		model = "qwen2.5-coder:7b"
	}
	prompt := buildOllamaPrompt(instruction, candidates)
	body, _ := json.Marshal(map[string]any{
		"model":  model,
		"stream": false,
		"format": "json",
		"prompt": prompt,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/generate", strings.NewReader(string(body)))
	if err != nil {
		return ResolvedIntent{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return ResolvedIntent{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return ResolvedIntent{}, ErrLowConfidence
	}
	var gen struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gen); err != nil {
		return ResolvedIntent{}, err
	}
	var parsed struct {
		Action     string  `json:"action"`
		Feature    string  `json:"feature"`
		TaskID     string  `json:"task_id"`
		Confidence float64 `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(gen.Response), &parsed); err != nil {
		return ResolvedIntent{}, err
	}
	return ResolvedIntent{
		Action:     IntentAction(parsed.Action),
		Feature:    slugFeature(parsed.Feature),
		TaskID:     parsed.TaskID,
		Confidence: parsed.Confidence,
		Reason:     "ollama",
	}, nil
}

func buildOllamaPrompt(instruction string, candidates []string) string {
	var b strings.Builder
	b.WriteString("Return ONLY JSON: {\"action\":\"develop|resume|verify|review|fix|import|sync|status\",\"feature\":\"slug\",\"task_id\":\"\",\"confidence\":0.0}\n")
	b.WriteString("Instruction: ")
	b.WriteString(instruction)
	b.WriteString("\nCandidates: ")
	b.WriteString(strings.Join(candidates, ", "))
	return b.String()
}

// ParseNotionPageID extracts page id from notion URL.
func ParseNotionPageID(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	path := strings.Trim(u.Path, "/")
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "-")
	if len(parts) == 0 {
		return path
	}
	last := parts[len(parts)-1]
	last = strings.ReplaceAll(last, "/", "")
	if len(last) >= 32 {
		return last
	}
	return path
}
