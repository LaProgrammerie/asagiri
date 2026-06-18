package onboarding

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/cloud"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// CloudProbeResult is the outcome of a cloud API connectivity check.
type CloudProbeResult struct {
	Email    string
	Projects []cloud.Project
}

// CloudSkipped reports whether the wizard should leave cloud disabled.
func CloudSkipped(mode string) bool {
	switch strings.TrimSpace(strings.ToLower(mode)) {
	case "", "skip", "local", "non", "no", "false", "n":
		return true
	default:
		return false
	}
}

// CloudConnect reports whether the wizard should configure cloud.
func CloudConnect(mode string) bool {
	switch strings.TrimSpace(strings.ToLower(mode)) {
	case "connect", "yes", "oui", "true", "o", "y":
		return true
	default:
		return false
	}
}

// DefaultCloudOnboardBaseURL is the suggested API URL during onboarding.
func DefaultCloudOnboardBaseURL() string {
	return "https://asagiri-cloud.test"
}

// DefaultCloudBaseURL returns the default API base URL for onboarding prompts.
func DefaultCloudBaseURL() string {
	return DefaultCloudOnboardBaseURL()
}

// ProbeCloud validates the token and lists accessible projects.
func ProbeCloud(ctx context.Context, baseURL, token string) (CloudProbeResult, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	token = strings.TrimSpace(token)
	if baseURL == "" {
		return CloudProbeResult{}, fmt.Errorf("cloud: base_url requis")
	}
	if token == "" {
		return CloudProbeResult{}, fmt.Errorf("cloud: token requis")
	}
	client := cloud.NewClient(cloud.ClientOptions{BaseURL: baseURL, Token: token})
	me, err := client.Me(ctx)
	if err != nil {
		return CloudProbeResult{}, fmt.Errorf("cloud: token invalide ou API injoignable: %w", err)
	}
	projects, err := client.ListProjects(ctx)
	if err != nil {
		return CloudProbeResult{}, fmt.Errorf("cloud: liste projets: %w", err)
	}
	return CloudProbeResult{Email: me.Email, Projects: projects}, nil
}

// TestCloudConnection probes the API and fills project choices on the form.
func TestCloudConnection(ctx context.Context, repoRoot string, f *Form) error {
	if f == nil {
		return fmt.Errorf("cloud: formulaire requis")
	}
	if CloudSkipped(f.Answers.CloudMode) {
		return fmt.Errorf("cloud: activez « Connecter Asagiri Cloud » pour tester")
	}
	errs := validateCloudCredentials(repoRoot, f)
	if len(errs) > 0 {
		f.Errors = errs
		return fmt.Errorf("cloud: %s", firstMapValue(errs))
	}
	baseURL := strings.TrimRight(strings.TrimSpace(f.Answers.CloudBaseURL), "/")
	token, _ := ResolveCloudToken(repoRoot, nil, f.Answers.CloudToken)
	result, err := ProbeCloud(ctx, baseURL, token)
	if err != nil {
		f.CloudProbed = false
		f.Answers.CloudConnectedEmail = ""
		f.Errors = map[string]string{"cloud_token": cloud.RedactError(err)}
		return err
	}
	f.CloudProbed = true
	f.Answers.CloudConnectedEmail = result.Email
	f.Answers.CloudProjectOptions = EncodeCloudProjectOptions(result.Projects)
	f.CloudProjectChoices = projectChoiceLabels(result.Projects)
	if strings.TrimSpace(f.Answers.CloudProjectID) == "" && len(result.Projects) == 1 {
		f.Answers.CloudProjectID = result.Projects[0].ID
		f.Answers.CloudProjectName = ProjectLabel(result.Projects[0])
	}
	f.Errors = map[string]string{}
	return nil
}

func validateCloudCredentials(repoRoot string, f *Form) map[string]string {
	errors := map[string]string{}
	baseURL := strings.TrimRight(strings.TrimSpace(f.Answers.CloudBaseURL), "/")
	if baseURL == "" {
		errors["cloud_base_url"] = "URL requise"
	}
	token, tokenErr := ResolveCloudToken(repoRoot, nil, f.Answers.CloudToken)
	if tokenErr != nil {
		errors["cloud_token"] = "lecture token: " + tokenErr.Error()
		return errors
	}
	if token == "" {
		errors["cloud_token"] = "token requis"
	}
	return errors
}

// ProjectDisplayName returns the short label shown in project pickers.
func ProjectDisplayName(p cloud.Project) string {
	if name := strings.TrimSpace(p.Name); name != "" {
		return name
	}
	if slug := strings.TrimSpace(p.Slug); slug != "" {
		return slug
	}
	return p.ID
}

// EncodeCloudProjectOptions serializes id|label pairs for the TUI.
func EncodeCloudProjectOptions(projects []cloud.Project) string {
	parts := make([]string, 0, len(projects))
	for _, p := range projects {
		parts = append(parts, p.ID+"|"+ProjectDisplayName(p))
	}
	return strings.Join(parts, ",")
}

// DecodeCloudProjectOptions parses encoded project options (label → id).
func DecodeCloudProjectOptions(encoded string) map[string]string {
	out := map[string]string{}
	for _, part := range strings.Split(encoded, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, label, ok := strings.Cut(part, "|")
		if !ok || strings.TrimSpace(id) == "" {
			continue
		}
		label = strings.TrimSpace(label)
		if label == "" {
			label = id
		}
		out[label] = id
	}
	return out
}

// ProjectIDForLabel resolves a display label to a project id.
func ProjectIDForLabel(encoded, label string) string {
	label = strings.TrimSpace(label)
	if label == "" {
		return ""
	}
	if id, ok := DecodeCloudProjectOptions(encoded)[label]; ok {
		return id
	}
	for l, id := range DecodeCloudProjectOptions(encoded) {
		if strings.EqualFold(l, label) {
			return id
		}
	}
	return label
}

// projectChoiceLabels returns display names for project select controls.
func projectChoiceLabels(projects []cloud.Project) []string {
	out := make([]string, 0, len(projects))
	for _, p := range projects {
		out = append(out, ProjectDisplayName(p))
	}
	return out
}

// ValidateCloudContinue checks the user can leave the cloud step.
func ValidateCloudContinue(repoRoot string, f *Form) map[string]string {
	if CloudSkipped(f.Answers.CloudMode) {
		return nil
	}
	errs := validateCloudCredentials(repoRoot, f)
	if len(errs) > 0 {
		return errs
	}
	if !f.CloudProbed {
		return map[string]string{"cloud_test": "cliquez sur « Tester la connexion »"}
	}
	if strings.TrimSpace(f.Answers.CloudProjectID) == "" {
		return map[string]string{"cloud_project_id": "choisissez un projet"}
	}
	return nil
}

// ProjectLabel returns a human-readable project label.
func ProjectLabel(p cloud.Project) string {
	name := strings.TrimSpace(p.Name)
	slug := strings.TrimSpace(p.Slug)
	switch {
	case name != "" && slug != "":
		return fmt.Sprintf("%s (%s)", name, slug)
	case name != "":
		return name
	case slug != "":
		return slug
	default:
		return p.ID
	}
}

// FindProject returns the project with the given ID.
func FindProject(projects []cloud.Project, id string) (cloud.Project, bool) {
	id = strings.TrimSpace(id)
	for _, p := range projects {
		if p.ID == id {
			return p, true
		}
	}
	return cloud.Project{}, false
}

// ResolveCloudToken returns the wizard token or an existing token file contents.
func ResolveCloudToken(repoRoot string, cfg *config.Config, wizardToken string) (string, error) {
	if t := strings.TrimSpace(wizardToken); t != "" {
		return t, nil
	}
	if cfg == nil {
		var err error
		cfg, err = config.Load(config.ConfigPath(repoRoot), repoRoot)
		if err != nil {
			return "", nil
		}
	}
	path, err := cloud.TokenPath(cfg)
	if err != nil {
		return "", err
	}
	expanded, err := cloud.ExpandPath(path)
	if err != nil {
		return "", err
	}
	return cloud.LoadToken(expanded)
}

// ApplyCloudAnswers persists token (if provided) and merges cloud config fields.
func ApplyCloudAnswers(repoRoot string, cfg *config.Config, answers Answers) error {
	if cfg == nil {
		return fmt.Errorf("cloud onboard: config absente")
	}
	if CloudSkipped(answers.CloudMode) {
		return nil
	}
	baseURL := strings.TrimRight(strings.TrimSpace(answers.CloudBaseURL), "/")
	if baseURL == "" {
		baseURL = config.DefaultCloudBaseURL
	}
	projectID := strings.TrimSpace(answers.CloudProjectID)
	if projectID == "" {
		return fmt.Errorf("cloud onboard: project_id requis")
	}
	token := strings.TrimSpace(answers.CloudToken)
	if token != "" {
		path, err := cloud.TokenPath(cfg)
		if err != nil {
			return err
		}
		expanded, err := cloud.ExpandPath(path)
		if err != nil {
			return err
		}
		if err := cloud.SaveToken(expanded, token); err != nil {
			return err
		}
	}
	cfg.Cloud.Enabled = true
	cfg.Cloud.BaseURL = baseURL
	cfg.Cloud.ProjectID = projectID
	if cfg.Cloud.TokenPath == "" {
		cfg.Cloud.TokenPath = config.DefaultCloudTokenRel
	}
	return nil
}

// MergeCloudPatch applies optional cloud settings from wizard answers.
func MergeCloudPatch(cfg *config.Config, answers Answers) {
	if cfg == nil || CloudSkipped(answers.CloudMode) {
		return
	}
	baseURL := strings.TrimRight(strings.TrimSpace(answers.CloudBaseURL), "/")
	if baseURL == "" {
		baseURL = config.DefaultCloudBaseURL
	}
	cfg.Cloud.Enabled = true
	cfg.Cloud.BaseURL = baseURL
	cfg.Cloud.ProjectID = strings.TrimSpace(answers.CloudProjectID)
	if cfg.Cloud.TokenPath == "" {
		cfg.Cloud.TokenPath = config.DefaultCloudTokenRel
	}
}

// ValidateCloudStep validates cloud wizard inputs and optionally probes the API.
func ValidateCloudStep(ctx context.Context, repoRoot string, f *Form, probe bool) map[string]string {
	if CloudSkipped(f.Answers.CloudMode) {
		return nil
	}
	if probe {
		if err := TestCloudConnection(ctx, repoRoot, f); err != nil {
			if len(f.Errors) > 0 {
				return f.Errors
			}
			return map[string]string{"cloud_token": err.Error()}
		}
		return nil
	}
	return ValidateCloudContinue(repoRoot, f)
}

// PickProjectInteractive lists projects and prompts for a selection on stdin.
func PickProjectInteractive(projects []cloud.Project, in io.Reader, out io.Writer, def string) (cloud.Project, error) {
	if len(projects) == 0 {
		return cloud.Project{}, fmt.Errorf("cloud: aucun projet disponible")
	}
	for i, p := range projects {
		_, _ = fmt.Fprintf(out, "  %d) %s [%s]\n", i+1, ProjectLabel(p), p.ID)
	}
	if len(projects) == 1 {
		return projects[0], nil
	}
	line := promptString(in, out, "Projet (numéro ou UUID)", def)
	if idx, err := strconv.Atoi(line); err == nil && idx >= 1 && idx <= len(projects) {
		return projects[idx-1], nil
	}
	if p, ok := FindProject(projects, line); ok {
		return p, nil
	}
	return cloud.Project{}, fmt.Errorf("cloud: sélection projet invalide %q", line)
}

func firstMapValue(m map[string]string) string {
	for _, v := range m {
		return v
	}
	return "validation échouée"
}
