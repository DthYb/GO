// Package backup copie récursivement un répertoire vers une destination
// en préservant les permissions des fichiers et des répertoires.
package backup

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Copy copie récursivement src vers dst. Les répertoires sont créés avec
// les permissions du répertoire source, les fichiers avec celles du
// fichier source. Un compteur d'avancement est affiché sur stdout.
// Retourne le nombre de fichiers copiés.
func Copy(src, dst string) (int, error) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return 0, fmt.Errorf("source inaccessible : %w", err)
	}
	if !srcInfo.IsDir() {
		return 0, fmt.Errorf("%s n'est pas un répertoire", src)
	}

	// Garde-fou : interdire dst à l'intérieur de src (copie infinie).
	if inside, err := isInside(src, dst); err != nil {
		return 0, err
	} else if inside {
		return 0, fmt.Errorf("%s est à l'intérieur de %s : copie impossible", dst, src)
	}

	// Racine de destination, avec les permissions de la racine source.
	if err := os.MkdirAll(dst, srcInfo.Mode().Perm()); err != nil {
		return 0, fmt.Errorf("création de %s : %w", dst, err)
	}

	copied := 0
	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("accès à %s : %w", path, err)
		}

		// Chemin relatif à la source → même chemin sous la destination.
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("stat de %s : %w", path, err)
		}

		switch {
		case d.IsDir():
			if path == src {
				return nil // la racine est déjà créée
			}
			// Répertoire recréé avec ses permissions d'origine.
			return os.MkdirAll(target, info.Mode().Perm())
		case d.Type().IsRegular():
			if err := copyFile(path, target, info.Mode().Perm()); err != nil {
				return err
			}
			copied++
			// \r réécrit la même ligne : compteur d'avancement minimaliste.
			fmt.Printf("\rCopie... %d fichier(s)", copied)
			return nil
		default:
			return nil // liens symboliques, etc. : ignorés volontairement
		}
	})
	if copied > 0 {
		fmt.Println() // termine proprement la ligne du compteur
	}
	if err != nil {
		return copied, err
	}
	return copied, nil
}

// copyFile copie un fichier en préservant ses permissions, via io.Copy
// (streaming : le fichier n'est jamais chargé entièrement en mémoire).
func copyFile(src, dst string, perm fs.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("ouverture de %s : %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("création de %s : %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copie %s → %s : %w", src, dst, err)
	}
	// Chmod explicite : O_CREATE applique perm modulo le umask du process.
	return os.Chmod(dst, perm)
}

// isInside indique si path est situé à l'intérieur de root.
func isInside(root, path string) (bool, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return false, err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return false, err
	}
	return rel == "." || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".."), nil
}
