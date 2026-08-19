// Package format transforme un inventaire en représentation prête
// à afficher (texte en colonnes, JSON, ...).
//
// Le cœur du package est l'interface Formatter : main manipule des
// Formatter sans jamais connaître l'implémentation concrète.
package format

import (
	"errors"

	"gopack/inventory"
)

// ErrRienAFormater est retournée quand l'inventaire à formater est vide.
var ErrRienAFormater = errors.New("rien à formater : inventaire vide")

// Formatter transforme un inventaire en chaîne prête à afficher.
//
// Interface volontairement minuscule (1 méthode) et définie côté
// consommateur. L'implémentation est IMPLICITE : toute struct dotée
// de cette méthode est un Formatter, sans le déclarer nulle part.
type Formatter interface {
	Format(fichiers []inventory.File) (string, error)
}
