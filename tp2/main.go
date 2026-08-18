// TP2 — Inventaire de fichiers en mémoire (corrigé complet)
//
// Première brique du fil rouge gopack : l'inventaire vit en mémoire
// avec des données en dur ; au jour 4 il sera rempli par un vrai scan
// de répertoire.
//
// Usage :
//
//	go run main.go
//
// Notions : structs, slices, maps, retours multiples (valeur, erreur),
// sort.Slice, méthodes value receiver vs pointer receiver.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ────────────────────────────────────────────────────────────────────
// Étape 1 — La struct File
// ────────────────────────────────────────────────────────────────────

// File représente un fichier de l'inventaire.
type File struct {
	Name      string    // nom du fichier, ex : "rapport.pdf"
	Size      int64     // taille en octets (int64 : le type de os.FileInfo.Size())
	Extension string    // extension avec le point, ex : ".pdf"
	Modified  time.Time // date de dernière modification
	Tag       string    // étiquette libre — zero value "" tant qu'on ne marque pas
}

// ────────────────────────────────────────────────────────────────────
// Étape 2 — Les données en dur : 12 fichiers simulant un dossier projet
// ────────────────────────────────────────────────────────────────────

var inventaire = []File{
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

// ────────────────────────────────────────────────────────────────────
// Étape 3 — Filtrage avec retours multiples (résultat, erreur)
// ────────────────────────────────────────────────────────────────────

// filtrerParExtension retourne un NOUVEAU slice contenant les fichiers
// dont l'extension correspond. Le slice d'origine n'est jamais modifié.
func filtrerParExtension(fichiers []File, ext string) ([]File, error) {
	// Validation en entrée, retour anticipé : le pattern (nil, erreur).
	if ext == "" {
		return nil, errors.New("extension vide")
	}
	if !strings.HasPrefix(ext, ".") {
		return nil, errors.New("l'extension doit commencer par un point, ex : \".pdf\"")
	}

	var resultat []File
	for _, f := range fichiers {
		if f.Extension == ext {
			resultat = append(resultat, f)
		}
	}
	return resultat, nil
}

// filtrerParTailleMin retourne les fichiers d'au moins tailleMin octets.
func filtrerParTailleMin(fichiers []File, tailleMin int64) ([]File, error) {
	if tailleMin < 0 {
		return nil, errors.New("la taille minimale ne peut pas être négative")
	}

	var resultat []File
	for _, f := range fichiers {
		if f.Size >= tailleMin {
			resultat = append(resultat, f)
		}
	}
	return resultat, nil
}

// ────────────────────────────────────────────────────────────────────
// Étape 4 — Statistiques par extension
// ────────────────────────────────────────────────────────────────────

// Stats regroupe les compteurs d'une extension.
type Stats struct {
	Nombre       int   // nombre de fichiers
	TailleTotale int64 // somme des tailles en octets
}

// statistiquesParExtension agrège nombre et taille totale par extension.
func statistiquesParExtension(fichiers []File) map[string]Stats {
	stats := make(map[string]Stats)
	for _, f := range fichiers {
		// La valeur d'une map n'est pas adressable : impossible d'écrire
		// stats[ext].Nombre++. Le pattern : lire, modifier, réécrire.
		// Si la clé est absente, s vaut la zero value {0, 0} : parfait.
		s := stats[f.Extension]
		s.Nombre++
		s.TailleTotale += f.Size
		stats[f.Extension] = s
	}
	return stats
}

// ────────────────────────────────────────────────────────────────────
// Étape 5 — Tri par taille décroissante (sans modifier l'original)
// ────────────────────────────────────────────────────────────────────

// trierParTaille retourne une COPIE triée par taille décroissante.
func trierParTaille(fichiers []File) []File {
	// Copie de travail : sort.Slice trie en place, on protège l'original.
	tries := make([]File, len(fichiers))
	copy(tries, fichiers)

	// La fonction anonyme répond à : "i doit-il passer avant j ?"
	// Avec > on obtient l'ordre décroissant.
	sort.Slice(tries, func(i, j int) bool {
		return tries[i].Size > tries[j].Size
	})
	return tries
}

// ────────────────────────────────────────────────────────────────────
// Étape 6 — Affichage formaté
// ────────────────────────────────────────────────────────────────────

// Seuils de conversion : 1 Ko = 1024 octets.
const (
	octetsParKo int64 = 1024
	octetsParMo int64 = 1024 * 1024
)

// tailleLisible convertit des octets en "512 o", "4.0 Ko" ou "42.0 Mo".
func tailleLisible(octets int64) string {
	// Switch sans expression : chaque case est une condition,
	// évaluée dans l'ordre.
	switch {
	case octets >= octetsParMo:
		return fmt.Sprintf("%.1f Mo", float64(octets)/float64(octetsParMo))
	case octets >= octetsParKo:
		return fmt.Sprintf("%.1f Ko", float64(octets)/float64(octetsParKo))
	default:
		return fmt.Sprintf("%d o", octets)
	}
}

// afficher imprime l'inventaire en colonnes alignées.
func afficher(fichiers []File) {
	fmt.Printf("%-25s %-10s %-13s %s\n", "NOM", "TAILLE", "MODIFIÉ", "TAG")
	for _, f := range fichiers {
		fmt.Printf("%-25s %-10s %-13s %s\n",
			f.Name,
			f.TailleLisible(),               // méthode de l'étape 7
			f.Modified.Format("2006-01-02"), // la date de référence de Go
			f.Tag,
		)
	}
}

// ────────────────────────────────────────────────────────────────────
// Étape 7 — Méthodes & pointer receiver
// ────────────────────────────────────────────────────────────────────

// TailleLisible — VALUE receiver : lecture seule, une copie suffit.
func (f File) TailleLisible() string {
	return tailleLisible(f.Size)
}

// Renommer — POINTER receiver : on veut modifier le File d'origine.
// Le nom change ET l'extension est recalculée à partir du nouveau nom.
func (f *File) Renommer(nouveauNom string) {
	f.Name = nouveauNom
	f.Extension = filepath.Ext(nouveauNom) // "todo-2026.md" → ".md"
}

// Marquer pose une étiquette (préparation de la sauvegarde gopack).
func (f *File) Marquer(tag string) {
	f.Tag = tag
}

// marquerLesLourds marque tous les fichiers d'au moins tailleMin octets
// et retourne le nombre de fichiers marqués.
func marquerLesLourds(fichiers []File, tailleMin int64) int {
	marques := 0
	// PIÈGE : `for _, f := range` donnerait une COPIE de chaque File ;
	// f.Marquer(...) marquerait la copie et l'inventaire resterait vierge.
	// On passe par l'index : fichiers[i] désigne l'élément DU slice, et
	// Go prend automatiquement son adresse pour le pointer receiver.
	for i := range fichiers {
		if fichiers[i].Size >= tailleMin {
			fichiers[i].Marquer("a-archiver")
			marques++
		}
	}
	return marques
}

// trouverIndex retourne l'index du fichier nommé nom, ou -1 si absent.
func trouverIndex(fichiers []File, nom string) int {
	for i := range fichiers {
		if fichiers[i].Name == nom {
			return i
		}
	}
	return -1
}

// ────────────────────────────────────────────────────────────────────
// Scénario complet
// ────────────────────────────────────────────────────────────────────

func main() {
	// ── Étape 1 : un fichier de test ────────────────────────────────
	test := File{Name: "test.txt", Size: 100, Extension: ".txt", Modified: time.Now()}
	fmt.Printf("Étape 1 — %%v  : %v\n", test)
	fmt.Printf("Étape 1 — %%+v : %+v\n", test) // %+v affiche les noms des champs

	// ── Étape 2 : le slice inventaire ───────────────────────────────
	fmt.Println("\nÉtape 2 — inventaire :", len(inventaire), "fichiers")
	for i, f := range inventaire {
		fmt.Printf("  %2d. %s (%d octets)\n", i+1, f.Name, f.Size)
	}
	// append peut réallouer le backing array : on RÉAFFECTE toujours.
	inventaire = append(inventaire, File{
		Name: "notes-perso.txt", Size: 256, Extension: ".txt", Modified: time.Now(),
	})
	fmt.Println("Après append :", len(inventaire), "fichiers")

	// ── Étape 3 : filtrage avec gestion d'erreur ────────────────────
	pdfs, err := filtrerParExtension(inventaire, ".pdf")
	if err != nil {
		fmt.Println("Erreur de filtrage :", err)
		os.Exit(1)
	}
	fmt.Println("\nÉtape 3 — fichiers .pdf :", len(pdfs))

	// Cas d'erreur volontaire : extension sans point.
	if _, err := filtrerParExtension(inventaire, "go"); err != nil {
		fmt.Println("Erreur attendue :", err)
	}

	// Filtres chaînés : les .png de plus de 1 Mo.
	pngs, err := filtrerParExtension(inventaire, ".png")
	if err != nil {
		fmt.Println("Erreur de filtrage :", err)
		os.Exit(1)
	}
	grosPngs, err := filtrerParTailleMin(pngs, 1_000_000)
	if err != nil {
		fmt.Println("Erreur de filtrage :", err)
		os.Exit(1)
	}
	fmt.Println(".png de plus de 1 Mo :", len(grosPngs)) // → 1 (photo-equipe.png)

	// ── Étape 4 : statistiques par extension ────────────────────────
	fmt.Println("\nÉtape 4 — statistiques par extension :")
	stats := statistiquesParExtension(inventaire)

	// Bonus étape 5 : clés triées pour un affichage déterministe
	// (l'itération d'une map est volontairement non ordonnée en Go).
	extensions := make([]string, 0, len(stats))
	for ext := range stats {
		extensions = append(extensions, ext)
	}
	sort.Strings(extensions)
	for _, ext := range extensions {
		s := stats[ext]
		fmt.Printf("  %-6s %2d fichier(s)  %s\n", ext, s.Nombre, tailleLisible(s.TailleTotale))
	}

	// L'idiome ok : distinguer "absent" de "présent avec zero value".
	if s, ok := stats[".zip"]; ok {
		fmt.Println("Les .zip pèsent", tailleLisible(s.TailleTotale))
	}

	// ── Étape 5 : top 3 des plus lourds ─────────────────────────────
	fmt.Println("\nÉtape 5 — top 3 des fichiers les plus lourds :")
	tries := trierParTaille(inventaire)
	for i := 0; i < 3; i++ {
		fmt.Printf("  %d. %-20s %s\n", i+1, tries[i].Name, tries[i].TailleLisible())
	}
	fmt.Println("Premier fichier de l'inventaire (intact) :", inventaire[0].Name)

	// ── Étape 7 : renommage + marquage en masse ─────────────────────
	fmt.Println("\nÉtape 7 — renommage et marquage :")
	if i := trouverIndex(inventaire, "todo.txt"); i >= 0 {
		// inventaire[i] est l'élément du slice : la mutation est visible.
		inventaire[i].Renommer("todo-2026.md")
		fmt.Printf("Renommé : %s (extension : %s)\n", inventaire[i].Name, inventaire[i].Extension)
	}
	marques := marquerLesLourds(inventaire, 1_000_000)
	fmt.Println("Fichiers marqués \"a-archiver\" :", marques) // → 5

	// ── Étape 6 : affichage final ───────────────────────────────────
	fmt.Println("\nÉtape 6 — inventaire final :")
	afficher(inventaire)
}
