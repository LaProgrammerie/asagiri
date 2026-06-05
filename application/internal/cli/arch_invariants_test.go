// Feature: audit-coherence-consolidation
//
// Test exemple (R8.5 — invariant ADR-027) : contrôle statique des frontières
// d'architecture de la couche UI. La règle est double :
//
//  1. `internal/ui` n'importe AUCUNE logique métier interdite : les décisions de
//     routing (`internal/routing`) et de policy (`internal/policy`) restent hors
//     UI. L'UI ne doit jamais embarquer ces moteurs ; elle consomme leurs
//     résultats via le bus (ADR-027).
//  2. `internal/ui` reste cliente du bus : au moins un fichier de production de
//     la couche UI importe `internal/ui/bus`, preuve que la couche s'appuie sur
//     le bus comme point d'entrée plutôt que sur des appels métier directs.
//
// Le contrôle est purement statique : on parcourt les sources Go de production
// sous `application/internal/ui/` (fichiers `_test.go` exclus : ce sont des
// échafaudages de test, pas la logique livrée) et on analyse leurs imports via
// go/parser. Aucune exécution de binaire, aucun effet de bord.
//
// _Requirements: 8.5_
package cli

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// archModulePath est le chemin de module déclaré dans go.mod. Sert de préfixe
// aux chemins d'import internes du dépôt.
const archModulePath = "github.com/LaProgrammerie/asagiri"

// archUIPkgRel / archForbiddenPkgs / archBusPkg sont les chemins d'import
// (relatifs au module) des packages concernés par l'invariant ADR-027.
const (
	archUIPkgRel = archModulePath + "/application/internal/ui"
	archBusPkg   = archModulePath + "/application/internal/ui/bus"
)

// archForbiddenPkgs liste les packages de logique métier qui doivent rester hors
// de la couche UI. Le routing (décision d'agent) et la policy (rôles autorisés /
// interdits) sont des moteurs métier : l'UI en consomme les sorties via le bus,
// elle ne les importe pas directement.
var archForbiddenPkgs = []string{
	archModulePath + "/application/internal/routing",
	archModulePath + "/application/internal/policy",
}

// TestUINoForbiddenBusinessLogicImports vérifie qu'aucun fichier de production de
// la couche UI n'importe les packages métier interdits (routing/policy). Toute
// violation est listée fichier → import, ce qui pointe directement la fuite de
// logique métier dans l'UI (ADR-027, exigence 8.5).
func TestUINoForbiddenBusinessLogicImports(t *testing.T) {
	uiDir := filepath.Join(moduleRootForArchTest(t), filepath.FromSlash("application/internal/ui"))

	var violations []string
	forEachUIProductionFile(t, uiDir, func(path string, imports []string) {
		for _, imp := range imports {
			for _, forbidden := range archForbiddenPkgs {
				if imp == forbidden || strings.HasPrefix(imp, forbidden+"/") {
					rel := relPathForArchTest(t, path)
					violations = append(violations, rel+" → "+imp)
				}
			}
		}
	})

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("invariant ADR-027 violé : internal/ui importe de la logique métier interdite\n"+
			"(routing/policy doivent rester hors UI ; l'UI consomme leurs résultats via le bus)\n  %s",
			strings.Join(violations, "\n  "))
	}
}

// TestUIRemainsBusClient garantit que la couche UI reste cliente du bus : au
// moins un fichier de production sous internal/ui importe internal/ui/bus. Si
// cette assertion tombe à zéro, l'UI a cessé de passer par le bus comme point
// d'entrée — régression de l'invariant ADR-027 (exigence 8.5).
func TestUIRemainsBusClient(t *testing.T) {
	uiDir := filepath.Join(moduleRootForArchTest(t), filepath.FromSlash("application/internal/ui"))

	busClients := 0
	forEachUIProductionFile(t, uiDir, func(_ string, imports []string) {
		for _, imp := range imports {
			if imp == archBusPkg {
				busClients++
				return
			}
		}
	})

	if busClients == 0 {
		t.Fatalf("aucun fichier de production sous internal/ui n'importe %q : "+
			"l'UI n'apparaît plus cliente du bus (régression ADR-027)", archBusPkg)
	}
}

// forEachUIProductionFile parcourt récursivement uiDir, parse chaque fichier Go
// de production (hors *_test.go) en mode ImportsOnly, et invoque fn avec le
// chemin du fichier et la liste de ses chemins d'import. Échoue le test si un
// fichier source attendu ne peut être parsé.
func forEachUIProductionFile(t *testing.T, uiDir string, fn func(path string, imports []string)) {
	t.Helper()

	fset := token.NewFileSet()
	walked := 0
	err := filepath.WalkDir(uiDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}

		file, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			t.Fatalf("parse %s: %v", path, perr)
		}

		imports := make([]string, 0, len(file.Imports))
		for _, spec := range file.Imports {
			imports = append(imports, strings.Trim(spec.Path.Value, `"`))
		}
		walked++
		fn(path, imports)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", uiDir, err)
	}
	if walked == 0 {
		t.Fatalf("aucun fichier de production Go trouvé sous %s : "+
			"le contrôle statique ADR-027 ne couvrirait rien", uiDir)
	}
}

// moduleRootForArchTest remonte depuis le répertoire courant jusqu'à la racine du
// module (présence de go.mod ET du dossier application/internal/ui), de sorte que
// le test localise les sources UI quel que soit le répertoire d'exécution.
func moduleRootForArchTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		_, goModErr := os.Stat(filepath.Join(dir, "go.mod"))
		uiPath := filepath.Join(dir, filepath.FromSlash("application/internal/ui"))
		_, uiErr := os.Stat(uiPath)
		if goModErr == nil && uiErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("impossible de localiser la racine du module (go.mod + application/internal/ui)")
		}
		dir = parent
	}
}

// relPathForArchTest réduit un chemin absolu à un chemin relatif à la racine du
// module pour des messages d'erreur lisibles ; renvoie le chemin brut en cas
// d'échec de la relativisation.
func relPathForArchTest(t *testing.T, path string) string {
	t.Helper()
	root := moduleRootForArchTest(t)
	if rel, err := filepath.Rel(root, path); err == nil {
		return filepath.ToSlash(rel)
	}
	return path
}
