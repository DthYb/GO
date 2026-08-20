// TP4 — Téléchargeur concurrent : squelette de départ.
//
// Complétez les TODO dans l'ordre des étapes du README.
// Le squelette compile tel quel : go run ./starter
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileInfo reflète le JSON renvoyé par GET /files.
type FileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// Result est le compte-rendu d'un téléchargement (envoyé dans le
// channel de résultats à l'étape 3).
type Result struct {
	Name     string
	Bytes    int64
	Duration time.Duration
	Err      error
}

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "URL du serveur de fichiers")
	outDir := flag.String("dir", "downloads", "répertoire de destination")
	workers := flag.Int("n", 4, "nombre maximum de téléchargements simultanés") // Bonus étape 5
	flag.Parse()

	// Récupération de la liste des fichiers (fourni).
	files, err := fetchList(*serverURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erreur :", err)
		os.Exit(1)
	}
	fmt.Printf("%d fichiers disponibles sur le serveur (concurrence : %d)\n", len(files), *workers)

	// Étape 2 : le répertoire de destination doit exister avant d'écrire.
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "erreur :", err)
		os.Exit(1)
	}

	start := time.Now()

	// Étape 3 : une goroutine par fichier, un channel de résultats, un
	// sémaphore (bonus étape 5) pour limiter la concurrence à *workers.
	results := make(chan Result)
	sem := make(chan struct{}, *workers)
	var wg sync.WaitGroup

	for _, f := range files {
		wg.Add(1)
		go func(f FileInfo) {
			defer wg.Done()

			sem <- struct{}{}        // prend un jeton (bloque si *workers en cours)
			defer func() { <-sem }() // rend le jeton à la fin

			t0 := time.Now()
			n, err := downloadFile(*serverURL, f.Name, *outDir)
			results <- Result{Name: f.Name, Bytes: n, Duration: time.Since(t0), Err: err}
		}(f)
	}

	// Goroutine dédiée : ferme le channel une fois tous les téléchargements
	// terminés. C'est ce close() qui termine le range ci-dessous.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Étape 3/4 : collecte au fil de l'eau, avec compteur d'avancement.
	// Une erreur sur un fichier n'arrête jamais le lot.
	var done, failed int
	for res := range results {
		done++
		if res.Err != nil {
			failed++
			fmt.Fprintf(os.Stderr, "[%d/%d] ❌ %-20s %v\n", done, len(files), res.Name, res.Err)
			continue
		}
		fmt.Printf("[%d/%d] ✅ %-20s %8s en %v\n",
			done, len(files), res.Name, humanSize(res.Bytes), res.Duration.Round(time.Millisecond))
	}

	// Étape 4 : bilan final.
	fmt.Printf("\nTerminé en %v — %d succès, %d échec(s)\n",
		time.Since(start).Round(time.Millisecond), done-failed, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

// fetchList interroge GET /files et décode la réponse JSON (fourni).
func fetchList(serverURL string) ([]FileInfo, error) {
	resp, err := http.Get(serverURL + "/files")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("statut inattendu : %s", resp.Status)
	}

	var files []FileInfo
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("décodage JSON : %w", err)
	}
	return files, nil
}

// downloadFile télécharge serverURL/files/name et l'écrit dans outDir
// avec les permissions 0644. Retourne le nombre d'octets écrits.
func downloadFile(serverURL, name, outDir string) (int64, error) {
	resp, err := http.Get(serverURL + "/files/" + name)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("statut inattendu : %s", resp.Status)
	}

	// os.Create utiliserait 0666 (modulé par le umask) : os.OpenFile impose
	// explicitement 0644 (lecture/écriture propriétaire, lecture seule ailleurs).
	dest := filepath.Join(outDir, name)
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	// io.Copy streame la réponse vers le fichier sans tout charger en mémoire.
	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return n, fmt.Errorf("copie vers %s : %w", dest, err)
	}
	return n, nil
}

// humanSize formate une taille en octets de façon lisible (Ko, Mo).
func humanSize(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f Mo", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f Ko", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d o", n)
	}
}
