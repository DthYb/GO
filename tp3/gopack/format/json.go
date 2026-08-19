// Sortie JSON, destinée aux machines (scripts, autres programmes).
package format

import (
	"encoding/json"
	"fmt"

	"gopack/inventory"
)

// JSONFormatter sérialise l'inventaire en JSON indenté.
// Les clés sont pilotées par les tags `json:"..."` de inventory.File.
type JSONFormatter struct{}

// Format implémente l'interface Formatter.
func (j JSONFormatter) Format(fichiers []inventory.File) (string, error) {
	if len(fichiers) == 0 {
		return "", ErrRienAFormater
	}

	// MarshalIndent : JSON lisible ("" de préfixe, 2 espaces d'indentation).
	// Seuls les champs EXPORTÉS de File sont sérialisés : encoding/json
	// est un package externe à inventory, la casse s'applique à lui aussi.
	donnees, err := json.MarshalIndent(fichiers, "", "  ")
	if err != nil {
		// On wrappe avec %w : l'appelant garde accès à l'erreur d'origine.
		return "", fmt.Errorf("encodage JSON : %w", err)
	}
	return string(donnees), nil
}
