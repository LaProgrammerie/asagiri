// Feature: audit-coherence-consolidation, Property 16
//
// Test de propriété : parité Plain_Output / JSON_Output (R6 — AUD-007).
//
// Property 16 (design.md) : pour toute sortie d'une Unitary_Command ou du
// Guided_Path (y compris onboarding / ready), l'ensemble des champs d'information
// du Plain_Output est égal à celui du JSON_Output, indépendamment du mode de
// rendu. Autrement dit, aucun champ d'information n'est conditionné au mode de
// rendu (exigences 6.8 et 7.5).
//
// Stratégie. Le seul DTO de sortie purement atteignable dans le package `cli`
// disposant à la fois d'un rendu plain réel (`formatRoutingExplanation`) et d'un
// encodage JSON réel (tags de struct) est `RoutingExplanation` — la sortie non
// interactive de `asa explain routing`, une Unitary_Command mise en avant par le
// Guided_Path. Le test de propriété (P16) génère des valeurs variées de ce DTO,
// rend chaque valeur dans les deux modes, extrait l'ensemble des champs
// d'information de chaque rendu de façon indépendante, et asserte l'égalité
// ensembliste (clés ET valeurs). Les sorties onboarding / ready sont couvertes
// par les tests exemple en fin de fichier, qui exercent les rendus réels de bout
// en bout via `onboarding.Ready` (plain vs json).
//
// Conventions : une seule propriété par test (`quick.Check`), ≥ 100 itérations,
// générateur déterministe (`testing/quick`, source fixée). Préfixe `p16` unique
// pour éviter toute collision avec les autres tests du package cli.
//
// Validates: Requirements 6.8, 7.5
package cli

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"testing/quick"

	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
)

// p16Reasons est l'ensemble fermé des raisons de routage exposées par le Router
// (exigence 4.6), complété par la chaîne vide pour couvrir le cas dégénéré.
var p16Reasons = []string{
	"prefer_local", "no_cloud", "cloud_heavy", "cloud_fast", "default", "",
}

// p16Explanation enveloppe un RoutingExplanation pour que testing/quick le génère
// via un générateur contraint à l'espace d'entrée réaliste.
type p16Explanation struct {
	value RoutingExplanation
}

// Generate satisfait testing/quick.Generator : il produit un RoutingExplanation
// varié mais réaliste. Les champs texte sont contraints à un alphabet
// d'identifiants (lettres, chiffres, `_`, `-`, espace) sans saut de ligne ni
// `": "`, afin que le format plain `clé: valeur` reste inversible : la propriété
// porte sur la parité des champs d'information, pas sur l'échappement du format.
func (p16Explanation) Generate(rnd *rand.Rand, _ int) reflect.Value {
	v := RoutingExplanation{
		StepClass: p16Token(rnd),
		Agent:     p16Token(rnd),
		Model:     p16Token(rnd),
		Local:     rnd.Intn(2) == 0,
		Reason:    p16Reasons[rnd.Intn(len(p16Reasons))],
	}
	return reflect.ValueOf(p16Explanation{value: v})
}

// p16Token génère un jeton de l'espace d'entrée réaliste : éventuellement vide
// (pour couvrir la parité sur champ vide), sinon une suite de caractères
// d'identifiant sans `\n` ni `": "`.
func p16Token(rnd *rand.Rand) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_- "
	n := rnd.Intn(12) // 0..11 → inclut la chaîne vide
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte(alphabet[rnd.Intn(len(alphabet))])
	}
	return b.String()
}

// TestProperty16RoutingExplanationPlainJSONParity vérifie Property 16 sur le DTO
// RoutingExplanation : pour toute valeur générée, l'ensemble des champs
// d'information rendu en plain est exactement égal à celui rendu en JSON, clés et
// valeurs comprises, indépendamment du mode de rendu.
//
// Validates: Requirements 6.8, 7.5
func TestProperty16RoutingExplanationPlainJSONParity(t *testing.T) {
	property := func(in p16Explanation) bool {
		plainFields := p16PlainFields(formatRoutingExplanation(in.value))

		raw, err := json.Marshal(in.value)
		if err != nil {
			return false
		}
		jsonFields := p16JSONFields(raw)

		return reflect.DeepEqual(plainFields, jsonFields)
	}

	// testing/quick avec une source fixée : déterministe, ≥ 100 itérations.
	cfg := &quick.Config{MaxCount: 200, Rand: rand.New(rand.NewSource(16))}
	if err := quick.Check(property, cfg); err != nil {
		t.Fatalf("Property 16 (parité plain/json de RoutingExplanation) violée : %v", err)
	}
}

// p16PlainFields extrait l'ensemble des champs d'information d'un Plain_Output au
// format `clé: valeur` (une paire par ligne), tel que produit par
// formatRoutingExplanation. Retourne une map clé→valeur canonique.
func p16PlainFields(plain string) map[string]string {
	fields := map[string]string{}
	for _, line := range strings.Split(strings.TrimRight(plain, "\n"), "\n") {
		if line == "" {
			continue
		}
		idx := strings.Index(line, ": ")
		if idx < 0 {
			// Ligne sans séparateur : clé seule, valeur vide.
			fields[line] = ""
			continue
		}
		fields[line[:idx]] = line[idx+2:]
	}
	return fields
}

// p16JSONFields extrait l'ensemble des champs d'information d'un JSON_Output en
// le désérialisant dans une map générique, puis en canonicalisant chaque valeur
// en chaîne. La désérialisation générique révèle les clés réellement émises (et
// non celles supposées par la struct), ce qui permet de détecter un champ présent
// dans un mode et absent de l'autre.
func p16JSONFields(raw []byte) map[string]string {
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return map[string]string{"__unmarshal_error__": err.Error()}
	}
	fields := make(map[string]string, len(decoded))
	for k, v := range decoded {
		fields[k] = p16CanonValue(v)
	}
	return fields
}

// p16CanonValue rend une valeur JSON décodée sous une forme textuelle stable,
// alignée sur la façon dont le rendu plain imprime les mêmes types (`%s` pour les
// chaînes, `%t` pour les booléens). RoutingExplanation n'expose que des chaînes
// et des booléens ; le cas par défaut sert de filet pour tout champ futur.
func p16CanonValue(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

// --- Couverture onboarding / ready (rendus réels, de bout en bout) -----------
//
// Les tests suivants ne sont pas des tests de propriété : ils exercent les rendus
// plain et json réels de `onboarding.Ready` sur un même dépôt, puis comparent
// l'ensemble des champs d'information (statut ready, score, identifiants de
// checks). Ils complètent Property 16 pour la clause « sorties onboarding /
// ready » sans recourir à des mocks (rendus de production uniquement).

// TestReadyOutputPlainJSONInformationParity vérifie que `asa ready` expose la
// même information (ready, score, ensemble des checks) en mode plain et en mode
// JSON pour un dépôt donné. _Requirements: 6.8, 7.5_
func TestReadyOutputPlainJSONInformationParity(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)
	writeExampleConfig(t, repo)

	var plainBuf, jsonBuf strings.Builder

	if _, err := onboarding.Ready(repo, onboarding.Options{CheckOnly: true, Plain: true}, &plainBuf); err != nil {
		t.Fatalf("ready plain: %v", err)
	}
	jsonRes, err := onboarding.Ready(repo, onboarding.Options{CheckOnly: true, JSON: true}, &jsonBuf)
	if err != nil {
		t.Fatalf("ready json: %v", err)
	}

	// Information extraite du Plain_Output.
	plainReady, plainScore, plainChecks := p16ParseReadyPlain(t, plainBuf.String())

	// Information extraite du JSON_Output (rendu de production).
	var report onboarding.Report
	if err := json.Unmarshal([]byte(jsonBuf.String()), &report); err != nil {
		t.Fatalf("json.Unmarshal report: %v\n%s", err, jsonBuf.String())
	}
	jsonChecks := p16CheckIDSet(report.Checks)

	// Cohérence interne : le rendu JSON doit refléter le Result retourné.
	if report.Score != jsonRes.Report.Score || report.Ready != jsonRes.Report.Ready {
		t.Fatalf("JSON_Output incohérent avec le Result retourné : json={ready:%v score:%d} result={ready:%v score:%d}",
			report.Ready, report.Score, jsonRes.Report.Ready, jsonRes.Report.Score)
	}

	if plainReady != report.Ready {
		t.Errorf("parité ready violée : plain=%v json=%v", plainReady, report.Ready)
	}
	if plainScore != report.Score {
		t.Errorf("parité score violée : plain=%d json=%d", plainScore, report.Score)
	}
	if !reflect.DeepEqual(plainChecks, jsonChecks) {
		t.Errorf("parité de l'ensemble des checks violée :\n plain=%v\n  json=%v",
			p16SortedKeys(plainChecks), p16SortedKeys(jsonChecks))
	}
}

// p16ParseReadyPlain extrait (ready, score, ensemble des IDs de checks) du
// Plain_Output de formatReadyPlain : entête `Readiness: <STATUT> (score N/100)`
// puis lignes `[statut] id[: message]`.
func p16ParseReadyPlain(t *testing.T, plain string) (bool, int, map[string]struct{}) {
	t.Helper()
	ready := false
	score := -1
	checks := map[string]struct{}{}

	for _, line := range strings.Split(plain, "\n") {
		switch {
		case strings.HasPrefix(line, "Readiness:"):
			// "NOT READY" contient "READY" : tester d'abord la négation.
			ready = !strings.Contains(line, "NOT READY")
			if i := strings.Index(line, "(score "); i >= 0 {
				rest := line[i+len("(score "):]
				if j := strings.Index(rest, "/100)"); j >= 0 {
					if n, err := strconv.Atoi(strings.TrimSpace(rest[:j])); err == nil {
						score = n
					}
				}
			}
		case strings.HasPrefix(line, "["):
			// Ligne de check : "[status] id" éventuellement suivi de ": message".
			close := strings.Index(line, "] ")
			if close < 0 {
				continue
			}
			rest := line[close+2:]
			id := rest
			if c := strings.Index(rest, ": "); c >= 0 {
				id = rest[:c]
			}
			id = strings.TrimSpace(id)
			if id != "" {
				checks[id] = struct{}{}
			}
		}
	}

	if score < 0 {
		t.Fatalf("score introuvable dans le Plain_Output :\n%s", plain)
	}
	return ready, score, checks
}

// p16CheckIDSet projette une liste de checks sur l'ensemble de leurs IDs.
func p16CheckIDSet(checks []onboarding.Check) map[string]struct{} {
	set := make(map[string]struct{}, len(checks))
	for _, c := range checks {
		set[c.ID] = struct{}{}
	}
	return set
}

// p16SortedKeys retourne les clés triées d'un ensemble, pour des messages
// d'erreur déterministes.
func p16SortedKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
