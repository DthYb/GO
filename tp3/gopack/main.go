// gopack — TP3 : l'inventaire du TP2 refactoré en projet structuré.
//
// Usage :
//
//	go run . [texte|json] [extension]
//
// Exemples :
//
//	go run .              → tableau texte coloré, inventaire complet
//	go run . json         → JSON indenté
//	go run . json .pdf    → JSON des seuls .pdf
//	go run . texte .xyz   → avertissement "aucun résultat" (code 0)
//	go run . texte xyz    → erreur d'usage (code 1)
//
// Rôle de main : assembler inventory (domaine) et format (affichage),
// et être le SEUL endroit qui décide quoi faire des erreurs.
package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"

	"gopack/format"
	"gopack/inventory"
)

// Codes de sortie : constantes nommées plutôt que 0/1 magiques.
const (
	sortieOK     = 0
	sortieErreur = 1
)

func main() {
	sortie, ext := lireArguments()

	rapport, err := construireRapport(donnees(), sortie, ext)
	if err != nil {
		// errors.Is traverse toute la chaîne de wrapping (%w) : peu
		// importe combien de couches ont enrichi l'erreur, la
		// sentinelle d'origine est retrouvée. Une comparaison directe
		// err == inventory.ErrAucunResultat échouerait, elle.
		switch {
		case errors.Is(err, inventory.ErrAucunResultat):
			// Cas bénin : un avertissement suffit, code de sortie 0.
			color.Yellow("Aucun résultat : %v", err)
			os.Exit(sortieOK)
		case errors.Is(err, inventory.ErrExtensionInvalide):
			// Faute de frappe utilisateur : on explique l'usage.
			fmt.Fprintln(os.Stderr, color.RedString("Erreur : %v", err))
			afficherUsage()
			os.Exit(sortieErreur)
		default:
			// Erreur inattendue : message complet (chaîne de wrapping).
			fmt.Fprintln(os.Stderr, color.RedString("Erreur : %v", err))
			os.Exit(sortieErreur)
		}
	}

	fmt.Println(rapport)
}

// lireArguments extrait la sortie souhaitée ("texte" par défaut)
// et l'éventuel filtre d'extension.
func lireArguments() (sortie, ext string) {
	sortie = "texte"
	if len(os.Args) > 1 {
		sortie = os.Args[1]
	}
	if len(os.Args) > 2 {
		ext = os.Args[2]
	}
	return sortie, ext
}

// construireRapport filtre l'inventaire si demandé, choisit le
// Formatter, et retourne le rapport prêt à afficher.
//
// Couche intermédiaire type : elle ne gère PAS les erreurs, elle les
// wrappe avec son contexte et les remonte au main.
func construireRapport(fichiers []inventory.File, sortie, ext string) (string, error) {
	if ext != "" {
		filtres, err := inventory.FiltrerParExtension(fichiers, ext)
		if err != nil {
			return "", fmt.Errorf("construction du rapport : %w", err)
		}
		fichiers = filtres
	}

	// La variable est typée par l'INTERFACE : à partir d'ici, main
	// ignore totalement quelle implémentation travaille.
	var fmtr format.Formatter
	switch sortie {
	case "texte":
		fmtr = format.TextFormatter{Colored: true}
	case "json":
		// Pas de couleur en JSON : la sortie doit rester parsable.
		fmtr = format.JSONFormatter{}
	default:
		return "", fmt.Errorf("sortie inconnue %q (attendu : texte ou json)", sortie)
	}

	return fmtr.Format(fichiers)
}

// donnees fournit l'inventaire d'exemple (repris du TP2).
// Au jour 4, cette fonction sera remplacée par un vrai scan de
// répertoire — sans toucher aux packages inventory et format.
func donnees() []inventory.File {
	fichiers := []inventory.File{
		{Name: "main.go", Size: 4_096, Extension: ".go", Modified: time.Date(2026, 7, 1, 10, 15, 0, 0, time.UTC)},
		{Name: "utils.go", Size: 2_048, Extension: ".go", Modified: time.Date(2026, 7, 3, 16, 40, 0, 0, time.UTC)},
		{Name: "README.md", Size: 8_192, Extension: ".md", Modified: time.Date(2026, 6, 20, 9, 0, 0, 0, time.UTC)},
		{Name: "lisezmoi.md", Size: 1_024, Extension: ".md", Modified: time.Date(2026, 6, 21, 11, 30, 0, 0, time.UTC)},
		{Name: "rapport.pdf", Size: 1_254_400, Extension: ".pdf", Modified: time.Date(2026, 6, 12, 9, 30, 0, 0, time.UTC)},
		{Name: "presentation.pdf", Size: 3_670_016, Extension: ".pdf", Modified: time.Date(2026, 5, 30, 14, 0, 0, 0, time.UTC)},
		{Name: "logo.png", Size: 46_080, Extension: ".png", Modified: time.Date(2026, 4, 2, 8, 45, 0, 0, time.UTC)},
		{Name: "photo-equipe.png", Size: 2_936_012, Extension: ".png", Modified: time.Date(2026, 4, 2, 8, 50, 0, 0, time.UTC)},
		{Name: "todo.txt", Size: 512, Extension: ".txt", Modified: time.Date(2026, 7, 18, 18, 5, 0, 0, time.UTC)},
		{Name: "backup.zip", Size: 15_728_640, Extension: ".zip", Modified: time.Date(2026, 1, 15, 3, 0, 0, 0, time.UTC)},
		{Name: "archive-2025.zip", Size: 44_040_192, Extension: ".zip", Modified: time.Date(2025, 12, 31, 23, 59, 0, 0, time.UTC)},
		{Name: "config.txt", Size: 730, Extension: ".txt", Modified: time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)},
	}

	// On réutilise les méthodes du domaine : marquage des plus lourds,
	// comme au TP2 (pointer receiver via l'index du slice).
	for i := range fichiers {
		if fichiers[i].Size >= 1_000_000 {
			fichiers[i].Marquer("a-archiver")
		}
	}
	return fichiers
}

// afficherUsage centralise l'aide de la commande.
func afficherUsage() {
	fmt.Fprintln(os.Stderr, "Usage : gopack [texte|json] [extension]")
	fmt.Fprintln(os.Stderr, "Exemples :")
	fmt.Fprintln(os.Stderr, "  gopack")
	fmt.Fprintln(os.Stderr, "  gopack json")
	fmt.Fprintln(os.Stderr, "  gopack texte .pdf")
}
