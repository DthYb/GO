// TP2 — Inventaire de fichiers en mémoire (fichier de départ)
//
// Ce fichier compile tel quel :
//
//	go run main.go
//
// Suivez les TODO étape par étape (voir README.md).
package main

import (
	"fmt"
)

// ────────────────────────────────────────────────────────────────────
// Étape 1 — Définissez ici la struct File
// ────────────────────────────────────────────────────────────────────
// TODO : champs attendus
//   - Name      string    : nom du fichier, ex "rapport.pdf"
//   - Size      int64     : taille en octets
//   - Extension string    : extension avec le point, ex ".pdf"
//   - Modified  time.Time : date de dernière modification
//   - Tag       string    : étiquette libre, vide par défaut (étape 7)

// ────────────────────────────────────────────────────────────────────
// Étape 2 — Les données en dur : 12 fichiers simulant un dossier projet
// ────────────────────────────────────────────────────────────────────
// Décommentez ce bloc une fois la struct File définie,
// et ajoutez "time" à la liste des imports ci-dessus.
/*
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
*/

// ────────────────────────────────────────────────────────────────────
// Étape 3 — Filtrage avec retours multiples (résultat, erreur)
// ────────────────────────────────────────────────────────────────────
// TODO : func filtrerParExtension(fichiers []File, ext string) ([]File, error)
//   - erreur (errors.New) si ext est vide ou ne commence pas par "."
//   - sinon : nouveau slice des fichiers correspondants, et nil

// TODO : func filtrerParTailleMin(fichiers []File, tailleMin int64) ([]File, error)
//   - erreur si tailleMin < 0
//   - sinon : nouveau slice des fichiers de taille >= tailleMin, et nil

// ────────────────────────────────────────────────────────────────────
// Étape 4 — Statistiques par extension
// ────────────────────────────────────────────────────────────────────
// TODO : type Stats struct { Nombre int; TailleTotale int64 }
// TODO : func statistiquesParExtension(fichiers []File) map[string]Stats

// ────────────────────────────────────────────────────────────────────
// Étape 5 — Tri par taille décroissante (sans modifier l'original)
// ────────────────────────────────────────────────────────────────────
// TODO : func trierParTaille(fichiers []File) []File
//   - copie de travail (make + copy) puis sort.Slice

// ────────────────────────────────────────────────────────────────────
// Étape 6 — Affichage formaté
// ────────────────────────────────────────────────────────────────────
// TODO : func afficher(fichiers []File)
// TODO : func tailleLisible(octets int64) string   → "512 o", "4.0 Ko", "42.0 Mo"

// ────────────────────────────────────────────────────────────────────
// Étape 7 — Mini-challenge : méthodes & pointer receiver
// ────────────────────────────────────────────────────────────────────
// TODO : func (f File) TailleLisible() string          (value receiver)
// TODO : func (f *File) Renommer(nouveauNom string)     (pointer receiver)
// TODO : func (f *File) Marquer(tag string)             (pointer receiver)
// TODO : func marquerLesLourds(fichiers []File, tailleMin int64) int

func main() {
	fmt.Println("TP2 — inventaire de fichiers : à vous de jouer !")

	// TODO Étape 1 : instancier un File de test et l'afficher (%v puis %+v)

	// TODO Étape 2 : afficher len(inventaire), parcourir avec range,
	//                puis append d'un 13e fichier

	// TODO Étape 3 : tester les filtres (cas OK et cas d'erreur),
	//                puis les chaîner (.png de plus de 1 Mo)

	// TODO Étape 4 : afficher les stats par extension + idiome ok sur ".zip"

	// TODO Étape 5 : afficher le top 3 des fichiers les plus lourds

	// TODO Étape 6 : appeler afficher(inventaire)

	// TODO Étape 7 : renommer "todo.txt" en "todo-2026.md",
	//                marquer les fichiers >= 1 Mo, réafficher l'inventaire
}
