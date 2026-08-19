// Package inventory contient le domaine de gopack : le type File,
// les filtres, les statistiques et le tri de l'inventaire.
//
// Règle du package : aucune notion d'affichage ici — c'est le rôle
// du package format. Les erreurs sont wrappées et remontées à
// l'appelant, qui décide seul de la réaction.
package inventory

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Erreurs sentinelles : exportées pour que l'appelant puisse les
// discriminer avec errors.Is, même après plusieurs niveaux de wrapping.
var (
	ErrExtensionInvalide = errors.New("extension invalide (attendu : \".ext\")")
	ErrTailleNegative    = errors.New("taille minimale négative")
	ErrAucunResultat     = errors.New("aucun fichier ne correspond")
)

// File représente un fichier de l'inventaire.
// Les tags JSON donnent des clés propres en minuscules ; omitempty
// fait disparaître la clé "tag" quand le champ est vide.
type File struct {
	Name      string    `json:"name"`          // nom du fichier, ex : "rapport.pdf"
	Size      int64     `json:"size"`          // taille en octets
	Extension string    `json:"extension"`     // extension avec le point, ex : ".pdf"
	Modified  time.Time `json:"modified"`      // date de dernière modification
	Tag       string    `json:"tag,omitempty"` // étiquette libre, vide par défaut
}

// Stats regroupe les compteurs d'une extension.
type Stats struct {
	Nombre       int   `json:"count"`
	TailleTotale int64 `json:"total_size"`
}

// FiltrerParExtension retourne un nouveau slice contenant les fichiers
// dont l'extension correspond. Le slice d'origine n'est jamais modifié.
//
// Erreurs possibles (toutes wrappées avec le contexte) :
//   - ErrExtensionInvalide si ext est vide ou sans point initial
//   - ErrAucunResultat si aucun fichier ne correspond
func FiltrerParExtension(fichiers []File, ext string) ([]File, error) {
	if ext == "" || !strings.HasPrefix(ext, ".") {
		// %w conserve la sentinelle dans la chaîne d'erreurs :
		// errors.Is(err, ErrExtensionInvalide) restera vrai chez l'appelant.
		return nil, fmt.Errorf("filtre extension %q : %w", ext, ErrExtensionInvalide)
	}

	var resultat []File
	for _, f := range fichiers {
		if f.Extension == ext {
			resultat = append(resultat, f)
		}
	}
	if len(resultat) == 0 {
		return nil, fmt.Errorf("extension %q : %w", ext, ErrAucunResultat)
	}
	return resultat, nil
}

// FiltrerParTailleMin retourne les fichiers d'au moins tailleMin octets.
func FiltrerParTailleMin(fichiers []File, tailleMin int64) ([]File, error) {
	if tailleMin < 0 {
		return nil, fmt.Errorf("filtre taille %d : %w", tailleMin, ErrTailleNegative)
	}

	var resultat []File
	for _, f := range fichiers {
		if f.Size >= tailleMin {
			resultat = append(resultat, f)
		}
	}
	if len(resultat) == 0 {
		return nil, fmt.Errorf("taille minimale %d octets : %w", tailleMin, ErrAucunResultat)
	}
	return resultat, nil
}

// StatistiquesParExtension agrège nombre et taille totale par extension.
func StatistiquesParExtension(fichiers []File) map[string]Stats {
	stats := make(map[string]Stats)
	for _, f := range fichiers {
		// La valeur d'une map n'est pas adressable :
		// on lit, on modifie, on réécrit.
		s := stats[f.Extension]
		s.Nombre++
		s.TailleTotale += f.Size
		stats[f.Extension] = s
	}
	return stats
}

// TrierParTaille retourne une copie triée par taille décroissante ;
// le slice d'origine reste intact.
func TrierParTaille(fichiers []File) []File {
	tries := make([]File, len(fichiers))
	copy(tries, fichiers)
	sort.Slice(tries, func(i, j int) bool {
		return tries[i].Size > tries[j].Size
	})
	return tries
}

// TailleLisible retourne la taille du fichier en unité humaine.
// Value receiver : lecture seule.
func (f File) TailleLisible() string {
	return tailleLisible(f.Size)
}

// Renommer change le nom et recalcule l'extension.
// Pointer receiver : la mutation doit toucher le File d'origine.
func (f *File) Renommer(nouveauNom string) {
	f.Name = nouveauNom
	f.Extension = filepath.Ext(nouveauNom)
}

// Marquer pose une étiquette sur le fichier.
func (f *File) Marquer(tag string) {
	f.Tag = tag
}

// Seuils de conversion : 1 Ko = 1024 octets.
const (
	octetsParKo int64 = 1024
	octetsParMo int64 = 1024 * 1024
)

// tailleLisible est un helper INTERNE : minuscule initiale, donc
// invisible hors du package. Le monde extérieur passe par la
// méthode exportée File.TailleLisible.
func tailleLisible(octets int64) string {
	switch {
	case octets >= octetsParMo:
		return fmt.Sprintf("%.1f Mo", float64(octets)/float64(octetsParMo))
	case octets >= octetsParKo:
		return fmt.Sprintf("%.1f Ko", float64(octets)/float64(octetsParKo))
	default:
		return fmt.Sprintf("%d o", octets)
	}
}
