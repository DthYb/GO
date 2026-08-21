// Package scan réalise l'inventaire récursif d'un répertoire :
// nombre de fichiers et taille cumulée, ventilés par extension.
package scan

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// ExtStat agrège les statistiques d'une extension donnée.
type ExtStat struct {
	Ext   string // ".go", ".txt", ou "(sans ext)"
	Count int    // nombre de fichiers
	Size  int64  // taille cumulée en octets
}

// Report est le résultat complet d'un scan.
type Report struct {
	Root      string    // répertoire scanné
	Files     int       // nombre total de fichiers
	Dirs      int       // nombre de répertoires (racine exclue)
	TotalSize int64     // taille totale en octets
	ByExt     []ExtStat // statistiques par extension, triées par taille décroissante
}

// Scan parcourt root récursivement avec filepath.WalkDir et agrège les
// statistiques. Les erreurs d'accès à une entrée (permissions...) sont
// remontées : l'appelant décide quoi en faire.
func Scan(root string) (*Report, error) {
	report := &Report{Root: root}
	byExt := make(map[string]*ExtStat)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("accès à %s : %w", path, err)
		}
		if d.IsDir() {
			if path != root {
				report.Dirs++
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil // liens symboliques, sockets... : ignorés
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("stat de %s : %w", path, err)
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == "" {
			ext = "(sans ext)"
		}
		stat, ok := byExt[ext]
		if !ok {
			stat = &ExtStat{Ext: ext}
			byExt[ext] = stat
		}
		stat.Count++
		stat.Size += info.Size()

		report.Files++
		report.TotalSize += info.Size()
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Map → slice triée par taille décroissante, pour un affichage stable.
	for _, stat := range byExt {
		report.ByExt = append(report.ByExt, *stat)
	}
	sort.Slice(report.ByExt, func(i, j int) bool {
		return report.ByExt[i].Size > report.ByExt[j].Size
	})
	return report, nil
}

// Print affiche le rapport sous forme de tableau lisible.
func (r *Report) Print(w io.Writer) {
	fmt.Fprintf(w, "Inventaire de %s\n\n", r.Root)
	fmt.Fprintf(w, "%-14s %8s %12s\n", "EXTENSION", "FICHIERS", "TAILLE")
	for _, stat := range r.ByExt {
		fmt.Fprintf(w, "%-14s %8d %12s\n", stat.Ext, stat.Count, HumanSize(stat.Size))
	}
	fmt.Fprintf(w, "\nTotal : %d fichiers, %d répertoires, %s\n",
		r.Files, r.Dirs, HumanSize(r.TotalSize))
}

// HumanSize formate une taille en octets de façon lisible (Ko, Mo, Go).
func HumanSize(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f Go", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1f Mo", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f Ko", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d o", n)
	}
}
