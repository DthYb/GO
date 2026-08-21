// Package fetch télécharge une ou plusieurs URLs en parallèle.
// C'est la brique héritée du TP4 : goroutines + channel de résultats +
// sync.WaitGroup + channel sémaphore pour limiter la concurrence.
package fetch

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

// Result est le compte-rendu du téléchargement d'une URL.
type Result struct {
	URL      string
	Dest     string
	Bytes    int64
	Duration time.Duration
	Err      error
}

// All télécharge toutes les URLs vers outDir, avec au plus workers
// téléchargements simultanés. Une erreur sur une URL n'arrête pas le lot.
// Retourne le nombre d'échecs.
func All(urls []string, outDir string, workers int) (int, error) {
	if len(urls) == 0 {
		return 0, fmt.Errorf("aucune URL fournie")
	}
	if workers < 1 {
		return 0, fmt.Errorf("le nombre de workers doit être >= 1 (reçu %d)", workers)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return 0, fmt.Errorf("création de %s : %w", outDir, err)
	}

	results := make(chan Result)        // compte-rendus des goroutines
	sem := make(chan struct{}, workers) // sémaphore : limite la concurrence
	var wg sync.WaitGroup

	for i, rawURL := range urls {
		wg.Add(1)
		go func(i int, rawURL string) {
			defer wg.Done()

			sem <- struct{}{}        // prend un jeton
			defer func() { <-sem }() // rend le jeton

			dest := filepath.Join(outDir, fileName(rawURL, i))
			t0 := time.Now()
			n, err := download(rawURL, dest)
			results <- Result{URL: rawURL, Dest: dest, Bytes: n,
				Duration: time.Since(t0), Err: err}
		}(i, rawURL)
	}

	// Fermer le channel une fois toutes les goroutines terminées,
	// pour que le range ci-dessous se termine.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collecte au fil de l'eau, avec compteur d'avancement [i/total].
	var done, failed int
	for res := range results {
		done++
		if res.Err != nil {
			failed++
			fmt.Fprintf(os.Stderr, "[%d/%d] ❌ %s : %v\n", done, len(urls), res.URL, res.Err)
			continue
		}
		fmt.Printf("[%d/%d] ✅ %s (%d octets en %v)\n",
			done, len(urls), res.Dest, res.Bytes, res.Duration.Round(time.Millisecond))
	}
	return failed, nil
}

// download télécharge une URL vers dest (permissions 0644, streaming io.Copy).
func download(rawURL, dest string) (int64, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("statut inattendu : %s", resp.Status)
	}

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return n, fmt.Errorf("écriture de %s : %w", dest, err)
	}
	return n, nil
}

// fileName déduit un nom de fichier local depuis l'URL, avec un nom de
// repli si l'URL ne se termine pas par un nom exploitable.
func fileName(rawURL string, index int) string {
	fallback := fmt.Sprintf("telechargement-%02d", index+1)
	u, err := url.Parse(rawURL)
	if err != nil {
		return fallback
	}
	name := path.Base(u.Path)
	if name == "" || name == "." || name == "/" {
		return fallback
	}
	return name
}
