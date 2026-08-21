// gopack — outil CLI de sauvegarde et transfert de fichiers.
// Projet final de la formation Go (TP5).
//
// Sous-commandes :
//
//	gopack scan   <dir>                        inventaire d'un répertoire
//	gopack backup <src> <dst>                  copie récursive avec permissions
//	gopack fetch  [-o dir] [-n N] <url>...     téléchargement concurrent
//
// Chaque sous-commande a son propre flag.NewFlagSet et son aide dédiée
// (gopack <commande> -h). Les erreurs vont sur stderr, le code de sortie
// est 0 en cas de succès, 1 sinon, 2 pour une erreur d'usage.
package main

import (
	"flag"
	"fmt"
	"os"

	"gopack/internal/backup"
	"gopack/internal/fetch"
	"gopack/internal/scan"
)

// version est injectée à la compilation via :
//
//	go build -ldflags "-X main.version=1.0.0"
var version = "dev"

// usage est l'aide globale, affichée sur "gopack", "gopack help" ou "gopack -h".
func usage() {
	fmt.Fprintf(os.Stderr, `gopack %s — sauvegarde et transfert de fichiers

Usage :
  gopack <commande> [options] [arguments]

Commandes :
  scan    <dir>                       Inventaire d'un répertoire (stats par extension)
  backup  <src> <dst>                 Copie récursive avec préservation des permissions
  fetch   [-o dir] [-n N] <url>...    Téléchargement concurrent d'une ou plusieurs URLs
  version                             Affiche la version
  help                                Affiche cette aide

Aide détaillée d'une commande : gopack <commande> -h
`, version)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "scan":
		err = runScan(os.Args[2:])
	case "backup":
		err = runBackup(os.Args[2:])
	case "fetch":
		err = runFetch(os.Args[2:])
	case "version":
		fmt.Println("gopack", version)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "gopack : commande inconnue %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "gopack :", err)
		os.Exit(1)
	}
}

// runScan gère : gopack scan <dir>
func runScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage : gopack scan <dir>")
		fmt.Fprintln(os.Stderr, "\nParcourt <dir> récursivement et affiche le nombre de fichiers")
		fmt.Fprintln(os.Stderr, "et la taille cumulée, ventilés par extension.")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if fs.NArg() != 1 {
		fs.Usage()
		os.Exit(2)
	}

	report, err := scan.Scan(fs.Arg(0))
	if err != nil {
		return err
	}
	report.Print(os.Stdout)
	return nil
}

// runBackup gère : gopack backup <src> <dst>
func runBackup(args []string) error {
	fs := flag.NewFlagSet("backup", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage : gopack backup <src> <dst>")
		fmt.Fprintln(os.Stderr, "\nCopie récursivement <src> vers <dst> en préservant les")
		fmt.Fprintln(os.Stderr, "permissions des fichiers et des répertoires.")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if fs.NArg() != 2 {
		fs.Usage()
		os.Exit(2)
	}

	src, dst := fs.Arg(0), fs.Arg(1)
	copied, err := backup.Copy(src, dst)
	if err != nil {
		return err
	}
	fmt.Printf("Sauvegarde terminée : %d fichier(s) copié(s) de %s vers %s\n", copied, src, dst)
	return nil
}

// runFetch gère : gopack fetch [-o dir] [-n N] <url> [url...]
// Note : avec le package flag, les options se placent AVANT les arguments.
func runFetch(args []string) error {
	fs := flag.NewFlagSet("fetch", flag.ExitOnError)
	outDir := fs.String("o", "downloads", "répertoire de destination")
	workers := fs.Int("n", 4, "nombre maximum de téléchargements simultanés")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage : gopack fetch [-o dir] [-n N] <url> [url...]")
		fmt.Fprintln(os.Stderr, "\nTélécharge une ou plusieurs URLs en parallèle (brique du TP4).")
		fmt.Fprintln(os.Stderr, "Une erreur sur une URL n'interrompt pas les autres téléchargements.")
		fmt.Fprintln(os.Stderr, "\nOptions :")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(2)
	}

	failed, err := fetch.All(fs.Args(), *outDir, *workers)
	if err != nil {
		return err
	}
	if failed > 0 {
		return fmt.Errorf("%d téléchargement(s) en échec", failed)
	}
	fmt.Println("Tous les téléchargements ont réussi.")
	return nil
}
