// Sortie texte en colonnes alignées, avec couleurs optionnelles
// via le package externe github.com/fatih/color.
package format

import (
	"fmt"
	"strings"

	"github.com/fatih/color"

	"gopack/inventory"
)

// TextFormatter produit un tableau texte lisible par un humain.
type TextFormatter struct {
	Colored bool // true : en-tête et tags colorisés (séquences ANSI)
}

// Format implémente l'interface Formatter — implicitement : aucune
// déclaration "implements" n'existe en Go, la méthode suffit.
func (t TextFormatter) Format(fichiers []inventory.File) (string, error) {
	if len(fichiers) == 0 {
		return "", ErrRienAFormater
	}

	// Les Sprint* de fatih/color retournent la string entourée de
	// séquences ANSI. Quand Colored est faux, on garde des fonctions
	// neutres : le reste du code ne change pas.
	enTete := fmt.Sprint
	tag := fmt.Sprint
	if t.Colored {
		enTete = color.New(color.FgCyan, color.Bold).Sprint
		tag = color.New(color.FgYellow).Sprint
	}

	// strings.Builder : on CONSTRUIT la sortie au lieu de l'imprimer,
	// c'est le contrat de l'interface (retourner une string).
	var b strings.Builder
	ligne := "%-25s %-10s %-13s %s\n"
	fmt.Fprintf(&b, ligne, enTete("NOM"), enTete("TAILLE"), enTete("MODIFIÉ"), enTete("TAG"))
	for _, f := range fichiers {
		fmt.Fprintf(&b, ligne,
			f.Name,
			f.TailleLisible(),
			f.Modified.Format("2006-01-02"),
			tag(f.Tag),
		)
	}
	return b.String(), nil
}
